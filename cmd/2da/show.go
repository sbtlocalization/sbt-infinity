// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package twoda

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

func NewShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <2da-file>",
		Aliases: []string{"cat"},
		Short: "Display contents of a 2DA file",
		Long:  `Display contents of a 2DA file from game archives.`,
		Example: `  Display EFFTEXT.2DA contents:

      sbt-inf 2da show EFFTEXT -k chitin.key

  Output in JSONL format:

      sbt-inf 2da show EFFTEXT -k chitin.key -j`,
		Args: cobra.ExactArgs(1),
		RunE: runShow,
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSONL format")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	resourceName := args[0]
	isJson, _ := cmd.Flags().GetBool("json")

	keyPath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		return fmt.Errorf("error resolving key path: %w", err)
	}

	if !strings.HasSuffix(strings.ToUpper(resourceName), ".2DA") {
		resourceName = resourceName + ".2DA"
	}

	infFs := fs.NewInfinityFs(keyPath, fs.WithTypeFilter(fs.FileType_2DA))

	file, err := infFs.Open(resourceName)
	if err != nil {
		return fmt.Errorf("error opening 2DA file %s: %w", resourceName, err)
	}
	defer file.Close()

	if isJson {
		return outputJsonl(file)
	}
	return outputRaw(file)
}

func outputRaw(r io.Reader) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading 2DA file: %w", err)
	}
	fmt.Print(string(content))
	return nil
}

type orderedMap struct {
	keys   []string
	values map[string]string
}

func (o *orderedMap) MarshalJSON() ([]byte, error) {
	var buf strings.Builder
	buf.WriteByte('{')
	for i, key := range o.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(key)
		valJSON, _ := json.Marshal(o.values[key])
		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.Write(valJSON)
	}
	buf.WriteByte('}')
	return []byte(buf.String()), nil
}

func outputJsonl(r io.Reader) error {
	const rowKeyCol = "row_key"

	twoda, err := parser.ParseTwoDA(r)
	if err != nil {
		return fmt.Errorf("error parsing 2DA file: %w", err)
	}

	for _, rowKey := range twoda.RowKeys {
		row, _ := twoda.Row(rowKey)

		output := &orderedMap{
			keys:   make([]string, 0, len(twoda.Columns)+1),
			values: make(map[string]string, len(twoda.Columns)+1),
		}

		output.keys = append(output.keys, rowKeyCol)
		output.values[rowKeyCol] = rowKey

		for i, col := range twoda.Columns {
			output.keys = append(output.keys, col)
			if i < len(row) {
				output.values[col] = row[i]
			} else {
				output.values[col] = twoda.DefaultValue
			}
		}

		jsonData, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	}

	return nil
}
