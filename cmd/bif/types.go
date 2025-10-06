// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package bif

import (
	"fmt"

	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

func NewTypesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "types",
		Short:   "Lists all known resourse types",
		Long: `Shows all types which can be passed to type filter.
https://gibberlings3.github.io/iesdp/file_formats/general.htm can be visited
for types description.`,

		Run:  runTypesBif,
		Args: cobra.MaximumNArgs(0),
	}

	return cmd
}

// `bif types` handler
func runTypesBif(cmd *cobra.Command, args []string) {
	for number, name := range *fs.GetAllTypes() {
		fmt.Printf("%5s %6d\n", name, number)
	}
}
