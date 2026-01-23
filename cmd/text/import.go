// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/tealeg/xlsx/v4"
	"github.com/samber/lo"
	"github.com/sbtlocalization/sbt-infinity/text"
	"github.com/sbtlocalization/sbt-infinity/utils"
	"github.com/spf13/cobra"
)

type xlsxTlkRow struct {
	Key            uint32
	Text           string
	HasText        bool
	HasToken       bool
	HasSound       bool
	SoundFile      string
	VolumeVariance uint32
	PitchVariance  uint32
}

func NewImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "import",
		Aliases: []string{"im"},
		Short:   "Import XLSX file to TLK files",
		Long: `Import an XLSX file (produced by 'text export') to TLK files.

The command creates dialog.tlk and optionally dialogf.tlk (for feminine text)
in the specified output directory.`,
		Example: `  Import dialog.xlsx to TLK files:
    sbt-inf text import --input dialog.xlsx --output ./lang/en_US/`,
		Args: cobra.NoArgs,
		RunE: runImport,
	}

	cmd.Flags().StringP("input", "i", "", "input XLSX `file` path")
	cmd.Flags().StringP("output", "o", "", "output `directory` path (writes dialog.tlk and dialogf.tlk)")
	cmd.Flags().StringP("separator", "s", " // ", "separator for male/female text variants")
	cmd.Flags().Uint16("lang-code", 0, "language code for TLK header")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose output")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagFilename("input", "xlsx")
	cmd.MarkFlagDirname("output")

	return cmd
}

func runImport(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	separator, _ := cmd.Flags().GetString("separator")
	langCode, _ := cmd.Flags().GetUint16("lang-code")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if !strings.HasSuffix(strings.ToLower(inputPath), ".xlsx") {
		return fmt.Errorf("input file must be an xlsx file: %s", inputPath)
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	if verbose {
		fmt.Printf("Reading XLSX file: %s\n", inputPath)
	}

	rows, err := parseXlsxForTlk(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse XLSX file: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("input file contains no data rows")
	}

	if verbose {
		fmt.Printf("Found %d entries\n", len(rows))
	}

	maleEntries, femaleEntries, hasFemale := buildTlkEntries(rows, separator)

	opts := text.TlkWriteOptions{Lang: langCode}

	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	dialogPath := filepath.Join(outputPath, "dialog.tlk")
	if err := text.WriteTlkFile(dialogPath, maleEntries, opts); err != nil {
		return fmt.Errorf("failed to write dialog.tlk: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully wrote: %s (%d entries)\n", dialogPath, len(maleEntries))
	}

	if hasFemale {
		dialogfPath := filepath.Join(outputPath, "dialogf.tlk")
		if err := text.WriteTlkFile(dialogfPath, femaleEntries, opts); err != nil {
			return fmt.Errorf("failed to write dialogf.tlk: %w", err)
		}

		if verbose {
			fmt.Printf("Successfully wrote: %s (%d entries)\n", dialogfPath, len(femaleEntries))
		}
	} else if verbose {
		fmt.Println("No female text variants found, skipping dialogf.tlk")
	}

	return nil
}

func parseXlsxForTlk(path string) ([]xlsxTlkRow, error) {
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

	// Find column indices
	keyIdx, textIdx, soundIdx := -1, -1, -1
	hasTextIdx, hasSoundIdx, hasTokenIdx := -1, -1, -1
	volumeIdx, pitchIdx := -1, -1

	colIdx := 0
	headerRow.ForEachCell(func(cell *xlsx.Cell) error {
		switch strings.ToLower(cell.Value) {
		case "key":
			keyIdx = colIdx
		case "source or translation":
			textIdx = colIdx
		case "sound file":
			soundIdx = colIdx
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
	if textIdx == -1 {
		return nil, fmt.Errorf("xlsx file missing required 'source or translation' column")
	}

	var rows []xlsxTlkRow
	maxRows := sheet.MaxRow
	for rowIdx := 1; rowIdx < maxRows; rowIdx++ {
		row, err := sheet.Row(rowIdx)
		if err != nil {
			return nil, fmt.Errorf("unable to read row %d: %w", rowIdx+1, err)
		}

		keyCell := row.GetCell(keyIdx)
		keyVal, err := keyCell.Int64()
		if err != nil {
			return nil, fmt.Errorf("invalid key value at row %d: %q is not a valid number", rowIdx+1, keyCell.Value)
		}

		textVal := row.GetCell(textIdx).Value

		// Parse optional columns
		soundFile := ""
		if soundIdx != -1 {
			soundFile = row.GetCell(soundIdx).Value
		}

		hasText := textVal != ""
		if hasTextIdx != -1 {
			hasText = row.GetCell(hasTextIdx).Bool()
		}

		hasSound := soundFile != ""
		if hasSoundIdx != -1 {
			hasSound = row.GetCell(hasSoundIdx).Bool()
		}

		hasToken := false
		if hasTokenIdx != -1 {
			hasToken = row.GetCell(hasTokenIdx).Bool()
		}

		var volumeVariance, pitchVariance uint32
		if volumeIdx != -1 {
			if v, err := row.GetCell(volumeIdx).Int64(); err == nil {
				volumeVariance = uint32(v)
			}
		}
		if pitchIdx != -1 {
			if v, err := row.GetCell(pitchIdx).Int64(); err == nil {
				pitchVariance = uint32(v)
			}
		}

		rows = append(rows, xlsxTlkRow{
			Key:            uint32(keyVal),
			Text:           textVal,
			HasText:        hasText,
			HasToken:       hasToken,
			HasSound:       hasSound,
			SoundFile:      soundFile,
			VolumeVariance: volumeVariance,
			PitchVariance:  pitchVariance,
		})
	}

	return rows, nil
}

// Converts XLSX rows to TLK entries, filling gaps with empty entries.
// Returns (maleEntries, femaleEntries, hasFemaleVariants)
func buildTlkEntries(rows []xlsxTlkRow, separator string) ([]text.TlkWriteEntry, []text.TlkWriteEntry, bool) {
	maxKey := lo.MaxBy(rows, func(a, b xlsxTlkRow) bool {
		return a.Key > b.Key
	}).Key

	// Initialize arrays with empty entries
	maleEntries := make([]text.TlkWriteEntry, maxKey+1)
	femaleEntries := make([]text.TlkWriteEntry, maxKey+1)
	for i := range maleEntries {
		maleEntries[i] = text.NewEmptyTlkEntry()
		femaleEntries[i] = text.NewEmptyTlkEntry()
	}

	hasFemale := false

	// Fill in actual data
	for _, row := range rows {
		maleText, femaleText, hasSplit := utils.SplitMaleFemaleText(row.Text, separator)
		if hasSplit {
			hasFemale = true
		} else {
			femaleText = maleText // Same text for both if no split
		}

		maleEntries[row.Key] = text.TlkWriteEntry{
			Text:           maleText,
			HasText:        row.HasText,
			HasSound:       row.HasSound,
			HasToken:       row.HasToken,
			AudioName:      row.SoundFile,
			VolumeVariance: row.VolumeVariance,
			PitchVariance:  row.PitchVariance,
		}

		femaleEntries[row.Key] = text.TlkWriteEntry{
			Text:           femaleText,
			HasText:        row.HasText,
			HasSound:       row.HasSound,
			HasToken:       row.HasToken,
			AudioName:      row.SoundFile,
			VolumeVariance: row.VolumeVariance,
			PitchVariance:  row.PitchVariance,
		}
	}

	return maleEntries, femaleEntries, hasFemale
}
