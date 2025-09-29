// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

// runListBif handles the `bif ls` command execution
func runListBif(cmd *cobra.Command, args []string) {
	keyFilePath := args[0]
	isJson, _ := cmd.Flags().GetBool(Bif_Flag_JSON)
	filterBifContent(cmd, keyFilePath, func(index int, name string, bifPath string, resType parser.Key_ResType) {
		outputFound(isJson, index, name, bifPath, resType)
	})
}

// TODO: allow user to pass some format ?
func outputFound(isJson bool, index int, name string, bifPath string, resType parser.Key_ResType) {
	if isJson {
		output := struct {
			Index   int                `json:"index"`
			Name    string             `json:"name"`
			Bif     string             `json:"bif"`
			ResType parser.Key_ResType `json:"type"`
		}{
			Index:   index,
			Name:    name,
			Bif:     bifPath,
			ResType: resType,
		}
		jsonData, err := json.Marshal(output)
		if err != nil {
			log.Fatalf("error marshaling JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("%d %s %s 0x%x\n", index, name, bifPath, resType)
	}
}
