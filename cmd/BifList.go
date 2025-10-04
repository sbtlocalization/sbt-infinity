// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

// runListBif handles the `bif ls` command execution
func runListBif(cmd *cobra.Command, args []string) {
	typeRawInput, _ := cmd.Flags().GetString("type")
	filterRawInput, _ := cmd.Flags().GetString("filter")
	isJson, _ := cmd.Flags().GetBool("json")

	keyFilePath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		log.Fatalf("Error with .key path: %v\n", err)
	}

	contentFilter := getContentFilter(filterRawInput)

	resFs := fs.NewInfinityFs(keyFilePath, getFileTypeFilter(typeRawInput)...)

	for _, v := range resFs.ListResourses(contentFilter) {
		var index int
		if v.FileIndex != 0 {
			index = int(v.FileIndex)
		} else {
			index = int(v.TilesetIndex)
		}

		if isJson {
			output := struct {
				Index   int                `json:"index"`
				Name    string             `json:"name"`
				Bif     string             `json:"bif"`
				ResType parser.Key_ResType `json:"type"`
			}{
				Index:   index,
				Name:    v.FullName,
				Bif:     v.BifFile,
				ResType: v.Type.ToParserType(),
			}
			jsonData, err := json.Marshal(output)
			if err != nil {
				log.Fatalf("error marshaling JSON: %v", err)
			}
			fmt.Println(string(jsonData))
		} else {
			fmt.Printf("%d %s %s 0x%x\n", index, v.FullName, v.BifFile, v.Type.ToParserType())
		}
	}

}
