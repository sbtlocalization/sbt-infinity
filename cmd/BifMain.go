// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

var bif_log_level_verbose = false

// mainBifCmd represents the bif command which has subcommands
var mainBifCmd = &cobra.Command{
	Use:   "bif ls|ex path-to-chitin.key [-o output-dir][-t resource-type][-f regex-filter]",
	Short: "unpack or list BIF files into resources",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Error: must also specify list(ls) or extract(ex) command")
	},
}

var listBifCmd = &cobra.Command{
	Use:   "ls path-to-chitin.key [-j=json][-t resource-type][-f regex-filter]",
	Short: "list all BIFF files and resources attached to KEY file",
	Long: `list all BIFF files and resources attached to KEY file.

	Additional filter may be passed to unpack only specific resources
	`,
	Run: runListBif,
}

var extractBifCmd = &cobra.Command{
	Use:   "ex path-to-chitin.key [-o output-dir][-t resource-type][-f regex-filter]",
	Short: "unpack(extracts) BIF files into resources",
	Long: `unpack(extracts) BIF files into set of resources.
	Structure of resources is read from chitin.key,
	so all related .bif files picked automatically.

	Additional filter may be passed to unpack only specific resources
	`,
	Run: runExtractBif,
}

func init() {
	rootCmd.AddCommand(mainBifCmd)
	mainBifCmd.AddCommand(listBifCmd)
	mainBifCmd.AddCommand(extractBifCmd)

	mainBifCmd.PersistentFlags().StringP("type", "t", "", "Resourse type filter. Comma separated integers (dec or hex) or extension names (like DLG). Take type number from https://gibberlings3.github.io/iesdp/file_formats/general.htm")
	mainBifCmd.PersistentFlags().StringP("filter", "f", "", "Regex for resourse name filtering")

	listBifCmd.Flags().BoolP("json", "j", false, "Decorate output as JSON")

	extractBifCmd.Flags().StringP("output", "o", "", "Output directory for resource files (default: current directory)")
}

// Parses argument like `-t 1011,0x409,1022,DLG,bmp` into list of Key_ResType
// TODO: remove duplicate types
func getFileTypeFilter(rawInput string) (filter []fs.FileType) {
	if len(rawInput) == 0 {
		return filter
	}

	tokens := strings.Split(rawInput, ",")

	if len(tokens) == 0 {
		return filter
	}

	for _, value := range tokens {
		if fType := fs.FileTypeFromExtension(value); fType.IsValid() {
			filter = append(filter, fType)
		} else if parsed, err := strconv.ParseInt(value, 0, 32); err == nil {
			resType := fs.FileType(parsed)
			if resType.IsValid() {
				filter = append(filter, resType)
			} else {
				log.Fatalf("Value 0x%x (%d) does not match known type\n", parsed, parsed)
			}
		} else {
			log.Fatalf("Value %s does not match known type\n", value)
		}
	}

	return filter
}

func getContentFilter(rawInput string) *regexp.Regexp {
	if len(rawInput) == 0 {
		return nil
	}

	compiled, err := regexp.Compile(rawInput)
	if err != nil {
		log.Fatalf("Value %s is not Regexp: %v\n", rawInput, err)
		return nil
	}

	return compiled
}
