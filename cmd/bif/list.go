// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package bif

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

func NewLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [-t type][flags]... [-j]",
		Aliases: []string{"ls"},
		Short:   "List game engine resources contained in BIF files",
		Long: `List game engine resources contained in BIF files.

Additional filter may be passed to list only specific resources.`,
		Example: `List all resourses which have dialog type and have 'arca' part in name:

sbt-inf bif ls -k chitin.key -f *arca* -t dlg`,
		Run:  runListBif,
		Args: cobra.MaximumNArgs(0),
	}

	cmd.Flags().BoolP("json", "j", false, "Decorate output as JSON")

	return cmd
}

// runListBif handles the `bif ls` command execution
func runListBif(cmd *cobra.Command, args []string) {
	typeRawInput, _ := cmd.Flags().GetStringSlice("type")
	bifFilterRawInput, _ := cmd.Flags().GetString("bif-filter")
	filterRawInput, _ := cmd.Flags().GetString("filter")
	isJson, _ := cmd.Flags().GetBool("json")

	keyFilePath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		log.Fatalf("Error with .key path: %v\n", err)
	}

	resFs := fs.NewInfinityFs(keyFilePath,
		fs.WithTypeFilter(getFileTypeFilter(typeRawInput)...),
		fs.WithBifFilter(bifFilterRawInput),
		fs.WithContentFilter(filterRawInput),
	)

	for _, v := range resFs.ListResources() {
		if isJson {
			output := struct {
				Name    string      `json:"name"`
				Bif     string      `json:"bif"`
				ResType fs.FileType `json:"type"`
			}{
				Name:    v.FullName,
				Bif:     v.BifFile,
				ResType: v.Type,
			}
			jsonData, err := json.Marshal(output)
			if err != nil {
				log.Fatalf("error marshaling JSON: %v", err)
			}
			fmt.Println(string(jsonData))
		} else {
			fmt.Printf("%s %s 0x%x\n", v.FullName, v.BifFile, v.Type.ToParserType())
		}
	}

}
