// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package bif

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bif ls|ex path-to-chitin.key [-o output-dir][-t resource-type][-f regex-filter]",
		Short: "unpack or list BIF files into resources",

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Error: must also specify list(ls) or extract(ex) command")
		},
	}

	cmd.AddCommand(NewLsCommand())
	cmd.AddCommand(NewExportCommand())
	return cmd
}

// Parses argument like `-t 1011,0x409,1022,DLG,bmp` into list of Key_ResType
// TODO: remove duplicate types
func getFileTypeFilter(tokens []string) (filter []fs.FileType) {
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
