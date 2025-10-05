// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package bif

import (
	"log"
	"regexp"
	"strconv"

	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bif",
		Short: "Works with raw BIF files which are bound to `chitin.key`",
	}

	cmd.PersistentFlags().StringSliceP("type", "t", nil, "Resourse type filter. Comma separated integers (dec or hex) or extension names (like DLG). Use `bif types` command to see all types.")
	cmd.PersistentFlags().StringP("filter", "f", "", "Regex for resourse name filtering")

	cmd.AddCommand(NewLsCommand())
	cmd.AddCommand(NewExportCommand())
	cmd.AddCommand(NewTypesCommand())
	return cmd
}

// Parses argument like `-t 1011,0x409,1022,DLG,bmp` into list of Key_ResType
func getFileTypeFilter(tokens []string) (filter []fs.FileType) {
	if len(tokens) == 0 {
		return
	}

	typeSet := make(map[fs.FileType]bool)

	for _, value := range tokens {
		if fType := fs.FileTypeFromExtension(value); fType.IsValid() {
			typeSet[fType] = true
		} else if parsed, err := strconv.ParseInt(value, 0, 32); err == nil {
			resType := fs.FileType(parsed)
			if resType.IsValid() {
				typeSet[resType] = true
			} else {
				log.Fatalf("Value 0x%x (%d) does not match known type\n", parsed, parsed)
			}
		} else {
			log.Fatalf("Value %s does not match known type\n", value)
		}
	}

	for key := range typeSet {
		filter = append(filter, key)
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
