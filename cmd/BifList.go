// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"

	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

// runListBif handles the `bif ls` command execution
func runListBif(cmd *cobra.Command, args []string) {
	initLogF(cmd)

	keyFilePath := args[0]
	printLogF("bif ls called with key file: %s\n", keyFilePath)

	keyFile, realFile := parseKeyFile(keyFilePath)
	// Close file on this level to avoid keep interface opened
	// TODO: use NewInfinityFs ?
	defer realFile.Close()

	// Display KEY file information
	printLogF("KEY file parsed successfully!\n")
	printLogF("BIF files count: %d\n", keyFile.NumBiffEntries)
	printLogF("Packed resource count: %d\n", keyFile.NumResEntries)

	typeFilter := getTypeFilter(cmd)
	printLogF("Active type filters: %v\n", typeFilter)

	contentFilter := getContentFilter(cmd)

	resEntries, _ := keyFile.ResEntries()

	isJson, _ := cmd.Flags().GetBool(Bif_Flag_JSON)

	for key, value := range resEntries {
		if len(typeFilter) > 0 && !slices.Contains(typeFilter, value.Type) {
			continue
		}

		index := key
		resourseName := value.Name
		bifFile, _ := value.Locator.BiffFile()
		bifFilePath, _ := bifFile.FilePath()

		if contentFilter != nil {
			if !(contentFilter.MatchString(strconv.Itoa(index)) ||
				contentFilter.MatchString(resourseName) ||
				contentFilter.MatchString(bifFilePath)) {
				continue
			}
		}

		outputFound(isJson, index, resourseName, bifFilePath, value.Type)
	}
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
