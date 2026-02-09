// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package tra

import (
	"fmt"
	"os"
	"strings"

	"codeberg.org/tealeg/xlsx/v4"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

type flagMetadata struct {
	HasText        bool
	HasSound       bool
	HasToken       bool
	VolumeVariance uint32
	PitchVariance  uint32
}

func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export",
		Aliases: []string{"ex"},
		Short:   "Export TRA to XLSX format",
		Long: `Export TRA translations to XLSX format, optionally using original XLSX for flag metadata.

The output XLSX can be used with 'text import' to generate TLK files.`,
		Example: `  Export TRA to XLSX with default flags:
    sbt-inf tra export -i dialog.tra -o dialog.xlsx

  Export TRA to XLSX using flags from original XLSX:
    sbt-inf tra export -i dialog.tra --original original.xlsx -o dialog.xlsx`,
		Args: cobra.NoArgs,
		RunE: runExport,
	}

	cmd.Flags().StringP("input", "i", "", "input TRA `file` path")
	cmd.Flags().String("original", "", "original XLSX `file` (source of flag metadata)")
	cmd.Flags().StringP("output", "o", "", "output XLSX `file` path")
	cmd.Flags().StringP("separator", "s", " // ", "separator for male/female text variants")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose output")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagFilename("input", "tra")
	cmd.MarkFlagFilename("original", "xlsx")
	cmd.MarkFlagFilename("output", "xlsx")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	originalPath, _ := cmd.Flags().GetString("original")
	outputPath, _ := cmd.Flags().GetString("output")
	separator, _ := cmd.Flags().GetString("separator")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if !strings.HasSuffix(strings.ToLower(inputPath), ".tra") {
		return fmt.Errorf("input file must be a .tra file: %s", inputPath)
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	if !strings.HasSuffix(strings.ToLower(outputPath), ".xlsx") {
		outputPath = outputPath + ".xlsx"
	}

	if verbose {
		fmt.Printf("Reading TRA file: %s\n", inputPath)
	}

	traFile, err := parser.ParseTraFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse TRA file: %w", err)
	}

	if verbose {
		fmt.Printf("Found %d entries\n", len(traFile.Entries))
	}

	// Load flag metadata from original XLSX if provided
	var flagMap map[uint32]flagMetadata
	if originalPath != "" {
		if _, err := os.Stat(originalPath); os.IsNotExist(err) {
			return fmt.Errorf("original file does not exist: %s", originalPath)
		}

		if verbose {
			fmt.Printf("Loading flag metadata from: %s\n", originalPath)
		}

		flagMap, err = loadFlagMetadata(originalPath)
		if err != nil {
			return fmt.Errorf("failed to load flag metadata: %w", err)
		}

		if verbose {
			fmt.Printf("Loaded flags for %d entries\n", len(flagMap))
		}
	}

	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("Strings")
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	headerRow := sheet.AddRow()
	headers := []string{"Key", "Source or Translation", "Sound File", "Has Text", "Has Sound", "Has Token", "Volume Variance", "Pitch Variance"}
	for _, h := range headers {
		cell := headerRow.AddCell()
		cell.SetValue(h)
	}

	for _, entry := range traFile.Entries {
		text := combineMaleFemaleText(entry.MaleText, entry.FemaleText, separator)

		flags := defaultFlagMetadata(text, entry.SoundFile)
		if flagMap != nil {
			if f, ok := flagMap[entry.ID]; ok {
				flags = f
			}
		}

		row := sheet.AddRow()
		row.AddCell().SetInt64(int64(entry.ID))
		row.AddCell().SetValue(text)
		row.AddCell().SetValue(entry.SoundFile)
		row.AddCell().SetBool(flags.HasText)
		row.AddCell().SetBool(flags.HasSound)
		row.AddCell().SetBool(flags.HasToken)
		row.AddCell().SetInt64(int64(flags.VolumeVariance))
		row.AddCell().SetInt64(int64(flags.PitchVariance))
	}

	if err := xlsxFile.Save(outputPath); err != nil {
		return fmt.Errorf("failed to save XLSX file: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully wrote XLSX file: %s (%d entries)\n", outputPath, len(traFile.Entries))
	}

	return nil
}

func loadFlagMetadata(xlsxPath string) (map[uint32]flagMetadata, error) {
	xlsxFile, err := xlsx.OpenFile(xlsxPath)
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

	// Find column indices
	keyIdx := -1
	hasTextIdx, hasSoundIdx, hasTokenIdx := -1, -1, -1
	volumeIdx, pitchIdx := -1, -1

	colIdx := 0
	headerRow.ForEachCell(func(cell *xlsx.Cell) error {
		switch strings.ToLower(cell.Value) {
		case "key":
			keyIdx = colIdx
		case "has text":
			hasTextIdx = colIdx
		case "has sound":
			hasSoundIdx = colIdx
		case "has token":
			hasTokenIdx = colIdx
		case "volume variance":
			volumeIdx = colIdx
		case "pitch variance":
			pitchIdx = colIdx
		}
		colIdx++
		return nil
	})

	if keyIdx == -1 {
		return nil, fmt.Errorf("xlsx file missing required 'key' column")
	}

	flagMap := make(map[uint32]flagMetadata)
	maxRows := sheet.MaxRow
	for rowIdx := 1; rowIdx < maxRows; rowIdx++ {
		row, err := sheet.Row(rowIdx)
		if err != nil {
			continue
		}

		keyCell := row.GetCell(keyIdx)
		keyVal, err := keyCell.Int64()
		if err != nil {
			continue
		}

		flags := flagMetadata{
			HasText:        true,
			HasSound:       false,
			HasToken:       true,
			VolumeVariance: 0,
			PitchVariance:  0,
		}

		if hasTextIdx != -1 {
			flags.HasText = row.GetCell(hasTextIdx).Bool()
		}
		if hasSoundIdx != -1 {
			flags.HasSound = row.GetCell(hasSoundIdx).Bool()
		}
		if hasTokenIdx != -1 {
			flags.HasToken = row.GetCell(hasTokenIdx).Bool()
		}
		if volumeIdx != -1 {
			if v, err := row.GetCell(volumeIdx).Int64(); err == nil {
				flags.VolumeVariance = uint32(v)
			}
		}
		if pitchIdx != -1 {
			if v, err := row.GetCell(pitchIdx).Int64(); err == nil {
				flags.PitchVariance = uint32(v)
			}
		}

		flagMap[uint32(keyVal)] = flags
	}

	return flagMap, nil
}

func defaultFlagMetadata(text, soundFile string) flagMetadata {
	return flagMetadata{
		HasText:        text != "",
		HasSound:       soundFile != "",
		HasToken:       true,
		VolumeVariance: 0,
		PitchVariance:  0,
	}
}

func combineMaleFemaleText(male, female, separator string) string {
	if female == "" || male == female {
		return male
	}
	return male + separator + female
}
