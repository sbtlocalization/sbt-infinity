// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package ids

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

func NewShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <ids-file>",
		Aliases: []string{"cat"},
		Short:   "Display contents of an IDS file",
		Long:    `Display contents of an IDS file from game archives.`,
		Example: `  Display ACTION.IDS contents:

      sbt-inf ids show ACTION -k chitin.key

  Output in JSONL format:

      sbt-inf ids show ACTION -k chitin.key -j`,
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

	if !strings.HasSuffix(strings.ToUpper(resourceName), ".IDS") {
		resourceName = resourceName + ".IDS"
	}

	infFs := fs.NewInfinityFs(keyPath, fs.WithTypeFilter(fs.FileType_IDS))

	file, err := infFs.Open(resourceName)
	if err != nil {
		return fmt.Errorf("error opening IDS file %s: %w", resourceName, err)
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
		return fmt.Errorf("error reading IDS file: %w", err)
	}
	fmt.Print(string(content))
	return nil
}

func outputJsonl(r io.Reader) error {
	ids, err := parser.ParseIds(r)
	if err != nil {
		return fmt.Errorf("error parsing IDS file: %w", err)
	}

	values := make([]int32, 0, len(ids.Entries))
	for value := range ids.Entries {
		values = append(values, value)
	}
	slices.Sort(values)

	for _, value := range values {
		jsonData, err := json.Marshal(struct {
			Value      int32  `json:"value"`
			Identifier string `json:"identifier"`
		}{
			Value:      value,
			Identifier: ids.Entries[value],
		})
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	}

	return nil
}
