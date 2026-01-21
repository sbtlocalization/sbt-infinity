// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package tra

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/sbtlocalization/sbt-infinity/tra"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update TRA file entries from CSV diff files",
		Long: `Update TRA file entries using CSV diff files (output of 'csv diff').

The CSV files must have 'id' and 'text' columns. Updates are matched by ID.
At least one of --male-csv or --female-csv must be provided.`,
		Example: `  Update male texts:
    sbt-inf tra update -i dialog.tra -m male_diff.csv -o dialog_updated.tra

  Update female texts:
    sbt-inf tra update -i dialog.tra -f female_diff.csv -o dialog_updated.tra

  Update both:
    sbt-inf tra update -i dialog.tra -m male.csv -f female.csv -o updated.tra`,
		Args: cobra.NoArgs,
		RunE: runUpdate,
	}

	cmd.Flags().StringP("input", "i", "", "input TRA `file` path")
	cmd.Flags().StringP("output", "o", "", "output TRA `file` path")
	cmd.Flags().StringP("male-csv", "m", "", "CSV `file` with male text updates")
	cmd.Flags().StringP("female-csv", "f", "", "CSV `file` with female text updates")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose output")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagFilename("input", "tra")
	cmd.MarkFlagFilename("output", "tra")
	cmd.MarkFlagFilename("male-csv", "csv")
	cmd.MarkFlagFilename("female-csv", "csv")
	cmd.MarkFlagsOneRequired("male-csv", "female-csv")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	maleCsvPath, _ := cmd.Flags().GetString("male-csv")
	femaleCsvPath, _ := cmd.Flags().GetString("female-csv")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if !strings.HasSuffix(strings.ToLower(inputPath), ".tra") {
		return fmt.Errorf("input file must be a .tra file: %s", inputPath)
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	if !strings.HasSuffix(strings.ToLower(outputPath), ".tra") {
		outputPath = outputPath + ".tra"
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

	// Load male CSV updates
	var maleUpdates map[uint32]string
	if maleCsvPath != "" {
		if verbose {
			fmt.Printf("Reading male CSV: %s\n", maleCsvPath)
		}
		maleUpdates, err = loadCSVUpdates(maleCsvPath)
		if err != nil {
			return fmt.Errorf("failed to load male CSV: %w", err)
		}
		if verbose {
			fmt.Printf("Found %d male updates\n", len(maleUpdates))
		}
	}

	// Load female CSV updates
	var femaleUpdates map[uint32]string
	if femaleCsvPath != "" {
		if verbose {
			fmt.Printf("Reading female CSV: %s\n", femaleCsvPath)
		}
		femaleUpdates, err = loadCSVUpdates(femaleCsvPath)
		if err != nil {
			return fmt.Errorf("failed to load female CSV: %w", err)
		}
		if verbose {
			fmt.Printf("Found %d female updates\n", len(femaleUpdates))
		}
	}

	maleCount, femaleCount := 0, 0
	for i := range traFile.Entries {
		entry := &traFile.Entries[i]

		if maleUpdates != nil {
			if newText, ok := maleUpdates[entry.ID]; ok {
				entry.MaleText = newText
				maleCount++
			}
		}

		if femaleUpdates != nil {
			if newText, ok := femaleUpdates[entry.ID]; ok {
				entry.FemaleText = newText
				femaleCount++
			}
		}
	}

	if verbose {
		fmt.Printf("Applied %d male updates, %d female updates\n", maleCount, femaleCount)
	}

	err = tra.WriteFile(outputPath, traFile.Entries)
	if err != nil {
		return fmt.Errorf("failed to write TRA file: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully wrote TRA file: %s\n", outputPath)
	}

	return nil
}

// Loads a CSV file and returns a map of id -> text.
func loadCSVUpdates(path string) (map[uint32]string, error) {
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

	// Find column indices from header
	header := records[0]
	idIdx, textIdx := -1, -1
	for i, col := range header {
		switch col {
		case "id":
			idIdx = i
		case "text":
			textIdx = i
		}
	}

	if idIdx == -1 || textIdx == -1 {
		return nil, fmt.Errorf("CSV file must have 'id' and 'text' columns")
	}

	updates := make(map[uint32]string)
	for _, record := range records[1:] {
		if len(record) <= idIdx || len(record) <= textIdx {
			continue
		}

		id, err := strconv.ParseUint(record[idIdx], 10, 32)
		if err != nil {
			continue // Skip invalid IDs
		}

		updates[uint32(id)] = record[textIdx]
	}

	return updates, nil
}
