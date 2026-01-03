// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/tealeg/xlsx/v4"
	"github.com/spf13/cobra"
)

type XlsxRow struct {
	Key       uint32
	Text      string
	SoundFile string
}

func NewConvertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "to-tra",
		Aliases: []string{"convert"},
		Short:   "Convert XLSX file to TRA file",
		Long: `Convert an XLSX file (produced by 'text export') to TRA file.

The TRA format is used by WeiDU and other Infinity Engine modding tools.`,
		Example: `  Convert dialog.xlsx to dialog.tra:
    sbt-inf text convert --input dialog.xlsx --output dialog.tra`,
		Args: cobra.NoArgs,
		RunE: runConvert,
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

func runConvert(cmd *cobra.Command, args []string) error {
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

	err = writeTraFile(outputPath, rows, separator)
	if err != nil {
		return fmt.Errorf("failed to write TRA file: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully wrote TRA file: %s\n", outputPath)
	}

	return nil
}

func parseXlsxFile(path string) ([]XlsxRow, error) {
	xlsxFile, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open xlsx file: %w", err)
	}

	// Get the first sheet
	if len(xlsxFile.Sheets) == 0 {
		return nil, fmt.Errorf("xlsx file has no sheets")
	}
	sheet := xlsxFile.Sheets[0]

	// Find column indices from header row
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

	var rows []XlsxRow
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

		rows = append(rows, XlsxRow{
			Key:       uint32(keyVal),
			Text:      text,
			SoundFile: soundFile,
		})
	}

	return rows, nil
}

func wrapWithDelimiters(text string) string {
	switch {
	case !strings.Contains(text, "~"):
		return "~" + text + "~"
	case !strings.Contains(text, "%"):
		return "%" + text + "%"
	case !strings.Contains(text, "\""):
		return "\"" + text + "\""
	default:
		return "~~~~~" + text + "~~~~~"
	}
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

func formatTraText(text string, separator string) string {
	if male, female, hasSplit := splitMaleFemaleText(text, separator); hasSplit {
		return wrapWithDelimiters(male) + " " + wrapWithDelimiters(female)
	} else {
		return wrapWithDelimiters(text)
	}
}

func formatTraEntry(id uint32, text string, soundFile string, maxIdWidth int, separator string) string {
	entry := fmt.Sprintf("@%-*d = %s", maxIdWidth, id, formatTraText(text, separator))
	if soundFile != "" {
		entry += fmt.Sprintf(" [%s]", soundFile)
	}

	return entry
}

func writeTraFile(path string, rows []XlsxRow, separator string) error {
	outputDir := filepath.Dir(path)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("unable to create output directory: %w", err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}
	defer file.Close()

	// Calculate max ID width for alignment
	maxId := uint32(0)
	for _, row := range rows {
		if row.Key > maxId {
			maxId = row.Key
		}
	}
	maxIdWidth := len(fmt.Sprintf("%d", maxId))

	for _, row := range rows {
		line := formatTraEntry(row.Key, row.Text, row.SoundFile, maxIdWidth, separator)
		_, err := fmt.Fprintln(file, line)
		if err != nil {
			return fmt.Errorf("error writing entry %d: %w", row.Key, err)
		}
	}

	return nil
}
