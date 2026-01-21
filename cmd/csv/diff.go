// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type diffMode struct {
	changed bool
	added   bool
	removed bool
}

type csvData struct {
	header  []string
	records [][]string
	idIdx   int
	textIdx int
}

func NewDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <file1.csv> <file2.csv>",
		Short: "Compare two CSV files by id and text columns",
		Long: `Compare two CSV files by their 'id' and 'text' columns.
Both files must have 'id' and 'text' columns (and may have additional columns).

Modes:
  c - changed: rows where ID exists in both files but text differs (default)
  a - added: rows in file2 that don't exist in file1
  r - removed: rows in file1 that don't exist in file2
  all - shorthand for c+a+r`,
		Example: `  Diff two files and save to output file:
    sbt-inf csv diff old.csv new.csv -o diff.csv

  Show changed and added rows:
    sbt-inf csv diff old.csv new.csv --mode c+a`,
		Args: cobra.ExactArgs(2),
		RunE: runDiff,
	}

	cmd.Flags().StringP("output", "o", "", "output file path (default: stdout)")
	cmd.Flags().StringArrayP("mode", "m", []string{"c"},
		"diff `modes`: c=changed, a=added, r=removed, all=a+c+r")

	return cmd
}

func parseMode(modes []string) (diffMode, error) {
	dm := diffMode{}
	for _, mode := range modes {
		for _, m := range strings.Split(mode, "+") {
			switch m {
			case "c":
				dm.changed = true
			case "a":
				dm.added = true
			case "r":
				dm.removed = true
			case "all":
				dm.changed, dm.added, dm.removed = true, true, true
			default:
				return dm, fmt.Errorf("unknown mode: %s (valid: c, a, r, all)", m)
			}
		}
	}
	return dm, nil
}

func runDiff(cmd *cobra.Command, args []string) error {
	outputPath, _ := cmd.Flags().GetString("output")
	modeFlags, _ := cmd.Flags().GetStringArray("mode")

	mode, err := parseMode(modeFlags)
	if err != nil {
		return err
	}

	file1, err := loadCSV(args[0])
	if err != nil {
		return fmt.Errorf("error reading first file: %w", err)
	}

	file2, err := loadCSV(args[1])
	if err != nil {
		return fmt.Errorf("error reading second file: %w", err)
	}

	diffRows := collectDiffs(file1, file2, mode)
	header := selectHeader(file1, file2, mode)

	return writeCSV(outputPath, header, diffRows)
}

func collectDiffs(file1, file2 *csvData, mode diffMode) [][]string {
	file1Map := buildTextMap(file1)
	file2IDs := buildIDSet(file2)

	var diffRows [][]string

	if mode.changed || mode.added {
		diffRows = append(diffRows, collectChangedAndAdded(file2, file1Map, mode)...)
	}

	if mode.removed {
		diffRows = append(diffRows, collectRemoved(file1, file2IDs)...)
	}

	return diffRows
}

func buildTextMap(data *csvData) map[string]string {
	m := make(map[string]string)
	for _, record := range data.records {
		m[record[data.idIdx]] = record[data.textIdx]
	}
	return m
}

func buildIDSet(data *csvData) map[string]struct{} {
	m := make(map[string]struct{})
	for _, record := range data.records {
		m[record[data.idIdx]] = struct{}{}
	}
	return m
}

func collectChangedAndAdded(file2 *csvData, file1Map map[string]string, mode diffMode) [][]string {
	var rows [][]string
	for _, record := range file2.records {
		id := record[file2.idIdx]
		text := record[file2.textIdx]

		file1Text, exists := file1Map[id]
		if exists && mode.changed && file1Text != text {
			rows = append(rows, record)
		} else if !exists && mode.added {
			rows = append(rows, record)
		}
	}
	return rows
}

func collectRemoved(file1 *csvData, file2IDs map[string]struct{}) [][]string {
	var rows [][]string
	for _, record := range file1.records {
		if _, exists := file2IDs[record[file1.idIdx]]; !exists {
			rows = append(rows, record)
		}
	}
	return rows
}

func selectHeader(file1, file2 *csvData, mode diffMode) []string {
	if !mode.changed && !mode.added && mode.removed {
		return file1.header
	}
	return file2.header
}

func writeCSV(outputPath string, header []string, rows [][]string) error {
	var writer *csv.Writer
	if outputPath != "" {
		outFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer outFile.Close()
		writer = csv.NewWriter(outFile)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}

	return nil
}

func loadCSV(path string) (*csvData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	header := records[0]
	idIdx, textIdx := findColumnIndices(header)

	if idIdx == -1 || textIdx == -1 {
		return nil, fmt.Errorf("CSV file must have 'id' and 'text' columns")
	}

	return &csvData{
		header:  header,
		records: records[1:],
		idIdx:   idIdx,
		textIdx: textIdx,
	}, nil
}

func findColumnIndices(header []string) (int, int) {
	idIdx, textIdx := -1, -1
	for i, col := range header {
		switch col {
		case "id":
			idIdx = i
		case "text":
			textIdx = i
		}
	}
	return idIdx, textIdx
}
