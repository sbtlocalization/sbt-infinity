// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("List case is not implemented yet")
	},
}

var extractBifCmd = &cobra.Command{
	Use:   "ex path-to-chitin.key [-o output-dir][-t resource-type][-f regex-filter]",
	Short: "unpack(extracts) BIF files into resources",
	Long: `unpack(extracts) BIF files into set of resources.
	Structure of resources is read from chitin.key,
	so all related .bif files picked automatically.

	Additional filter may be passed to unpack only specific resources
	`,
	Args: cobra.ExactArgs(1),
	Run:  runExtractBif,
}

func init() {
	rootCmd.AddCommand(mainBifCmd)
	mainBifCmd.AddCommand(listBifCmd)
	mainBifCmd.AddCommand(extractBifCmd)

	mainBifCmd.PersistentFlags().StringP("type", "t", "", "Resourse type filter. Take type number from https://gibberlings3.github.io/iesdp/file_formats/general.htm")
	mainBifCmd.PersistentFlags().StringP("filter", "f", "", "Regex for resourse name filtering")

	listBifCmd.Flags().BoolP("json", "j", false, "Decorate output as JSON")

	extractBifCmd.Flags().StringP("output", "o", "", "Output directory for resource files (default: current directory)")
}
