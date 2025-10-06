// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package bif

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

func NewTypesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "Lists all known resourse types",
		Long: `Shows all types which can be passed to type filter.
https://gibberlings3.github.io/iesdp/file_formats/general.htm can be visited
for types description.`,

		Run:  runTypesBif,
		Args: cobra.MaximumNArgs(0),
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")

	return cmd
}

// `bif types` handler
func runTypesBif(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	types := *fs.GetAllTypes()
	var exts []string
	for _, name := range types {
		exts = append(exts, name)
	}
	slices.Sort(exts)

	if jsonOutput {
		for _, e := range exts {
			entry := typeEntry{
				Type:  e,
				Value: int(fs.FileTypeFromExtension(e)),
			}
			data, _ := json.Marshal(entry)
			fmt.Println(string(data))
		}
	} else {
		for _, e := range exts {
			fmt.Printf("%5s %6d\n", e, fs.FileTypeFromExtension(e))
		}
	}
}

type typeEntry struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}
