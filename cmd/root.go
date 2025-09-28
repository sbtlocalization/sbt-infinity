// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"os"

	"github.com/sbtlocalization/sbt-infinity/cmd/dialog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sbt-inf",
	Short: "A set of tools for Infinity Engine games",
	Long: `SBT Infinity Tools is a collection of utilities designed to assist with the localization and 
modification of games based on the Infinity Engine, such as Baldur's Gate, Baldur's Gate II, 
and Planescape: Torment.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(dialog.NewCommand())
}
