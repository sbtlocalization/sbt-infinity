// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/utils"
	"github.com/spf13/cobra"
)

func NewConvertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "convert",
		Aliases: []string{"cv"},
		Short:   "Convert XLSX file to JSON",
		Long: `Convert an XLSX file (produced by 'text export') to a JSON file.

The output JSON maps string IDs to arrays of text variants:
  {"1": ["male text", "female text"], "2": ["male text"], ...}

If female text is absent, the array contains only the male text.`,
		Example: `  Convert dialog.xlsx to JSON (stdout):
    sbt-inf text convert --input dialog.xlsx

  Convert dialog.xlsx to a file:
    sbt-inf text convert --input dialog.xlsx --output dialog.json`,
		Args: cobra.NoArgs,
		RunE: runConvert,
	}

	cmd.Flags().StringP("input", "i", "", "input XLSX `file` path")
	cmd.Flags().StringP("output", "o", "", "output JSON `file` path (default: stdout)")
	cmd.Flags().StringP("separator", "s", " // ", "separator for male/female text variants")

	cmd.MarkFlagRequired("input")
	cmd.MarkFlagFilename("input", "xlsx")
	cmd.MarkFlagFilename("output", "json")

	return cmd
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	separator, _ := cmd.Flags().GetString("separator")

	if !strings.HasSuffix(strings.ToLower(inputPath), ".xlsx") {
		return fmt.Errorf("input file must be an xlsx file: %s", inputPath)
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	rows, err := parseXlsxForTlk(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse XLSX file: %w", err)
	}

	jsonData, err := buildConvertJSON(rows, separator)
	if err != nil {
		return fmt.Errorf("failed to build JSON: %w", err)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Print(string(jsonData))
	}

	return nil
}

func buildConvertJSON(rows []xlsxTlkRow, separator string) ([]byte, error) {
	type entry struct {
		key   uint32
		texts []string
	}

	var entries []entry
	for _, row := range rows {
		male, female, hasSplit := utils.SplitMaleFemaleText(row.Text, separator)
		if male == "" && !hasSplit {
			continue
		}
		texts := []string{male}
		if hasSplit && female != "" {
			texts = append(texts, female)
		}
		entries = append(entries, entry{key: row.Key, texts: texts})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	result := make(map[string][]string, len(entries))
	for _, e := range entries {
		result[strconv.FormatUint(uint64(e.key), 10)] = e.texts
	}

	return json.MarshalIndent(result, "", "    ")
}
