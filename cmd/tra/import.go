// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package tra

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"codeberg.org/tealeg/xlsx/v4"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/sbtlocalization/sbt-infinity/tra"
	"github.com/spf13/cobra"
)

type xlsxRow struct {
	Key       uint32
	Text      string
	SoundFile string
}

func NewImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "import",
		Aliases: []string{"im"},
		Short:   "Import XLSX file to TRA file",
		Long: `Import an XLSX file (produced by 'text export') to TRA file.

The TRA format is used by WeiDU and other Infinity Engine modding tools.`,
		Example: `  Import dialog.xlsx to dialog.tra:
    sbt-inf tra import --input dialog.xlsx --output dialog.tra`,
		Args: cobra.NoArgs,
		RunE: runImport,
	}

	cmd.Flags().StringP("input", "i", "", "input XLSX `file` path")
	cmd.Flags().StringP("output", "o", "", "output TRA `file` path")
	cmd.Flags().StringP("separator", "s", " // ", "separator for male/female text variants")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose output")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagFilename("input", "xlsx")
	cmd.MarkFlagFilename("output", "tra")

	return cmd
}

func runImport(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	separator, _ := cmd.Flags().GetString("separator")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if !strings.HasSuffix(strings.ToLower(inputPath), ".xlsx") {
		return fmt.Errorf("input file must be an xlsx file: %s", inputPath)
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	if !strings.HasSuffix(strings.ToLower(outputPath), ".tra") {
		outputPath = outputPath + ".tra"
	}

	if verbose {
		fmt.Printf("Reading XLSX file: %s\n", inputPath)
	}

	rows, err := parseXlsxFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse XLSX file: %w", err)
	}

	if verbose {
		fmt.Printf("Found %d entries\n", len(rows))
	}

	entries := xlsxRowsToTraEntries(rows, separator)
	err = tra.WriteFile(outputPath, entries)
	if err != nil {
		return fmt.Errorf("failed to write TRA file: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully wrote TRA file: %s\n", outputPath)
	}

	return nil
}

func parseXlsxFile(path string) ([]xlsxRow, error) {
	xlsxFile, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open xlsx file: %w", err)
	}

	if len(xlsxFile.Sheets) == 0 {
		return nil, fmt.Errorf("xlsx file has no sheets")
	}
	sheet := xlsxFile.Sheets[0]

	headerRow, err := sheet.Row(0)
	if err != nil {
		return nil, fmt.Errorf("unable to read header row: %w", err)
	}

	keyIdx, textIdx, soundIdx := -1, -1, -1
	colIdx := 0
	headerRow.ForEachCell(func(cell *xlsx.Cell) error {
		switch strings.ToLower(cell.Value) {
		case "key":
			keyIdx = colIdx
		case "source or translation":
			textIdx = colIdx
		case "sound file":
			soundIdx = colIdx
		}
		colIdx++
		return nil
	})

	if keyIdx == -1 {
		return nil, fmt.Errorf("xlsx file missing required 'key' column")
	}
	if textIdx == -1 {
		return nil, fmt.Errorf("xlsx file missing required 'source or translation' column")
	}

	var rows []xlsxRow
	maxRows := sheet.MaxRow
	for rowIdx := 1; rowIdx < maxRows; rowIdx++ {
		row, err := sheet.Row(rowIdx)
		if err != nil {
			return nil, fmt.Errorf("unable to read row %d: %w", rowIdx+1, err)
		}

		keyCell := row.GetCell(keyIdx)
		keyVal, err := strconv.ParseUint(keyCell.Value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid key value at row %d: %q is not a valid number", rowIdx+1, keyCell.Value)
		}

		text := row.GetCell(textIdx).Value

		soundFile := ""
		if soundIdx != -1 {
			soundFile = row.GetCell(soundIdx).Value
		}

		rows = append(rows, xlsxRow{
			Key:       uint32(keyVal),
			Text:      text,
			SoundFile: soundFile,
		})
	}

	return rows, nil
}

// Splits "male <sep> female" text into male/female variants
// Returns (male, female, hasSplit)
func splitMaleFemaleText(text string, separator string) (string, string, bool) {
	if separator == "" {
		return text, "", false
	}

	idx := strings.Index(text, separator)
	if idx == -1 {
		return text, "", false
	}

	male := strings.TrimSpace(text[:idx])
	female := strings.TrimSpace(text[idx+len(separator):])
	return male, female, true
}

func xlsxRowsToTraEntries(rows []xlsxRow, separator string) []parser.TraEntry {
	entries := make([]parser.TraEntry, len(rows))
	for i, row := range rows {
		male, female, _ := splitMaleFemaleText(row.Text, separator)
		entries[i] = parser.TraEntry{
			ID:         row.Key,
			MaleText:   male,
			FemaleText: female,
			SoundFile:  row.SoundFile,
		}
	}
	return entries
}
