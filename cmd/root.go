// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"os"
	"runtime/pprof"

	twoda "github.com/sbtlocalization/sbt-infinity/cmd/2da"
	"github.com/sbtlocalization/sbt-infinity/cmd/bif"
	"github.com/sbtlocalization/sbt-infinity/cmd/csv"
	"github.com/sbtlocalization/sbt-infinity/cmd/dialog"
	"github.com/sbtlocalization/sbt-infinity/cmd/text"
	"github.com/sbtlocalization/sbt-infinity/cmd/tra"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sbt-inf",
	Short: "A set of tools for Infinity Engine games",
	Long: `SBT Infinity Tools is a collection of utilities designed to assist with the localization and 
modification of games based on the Infinity Engine, such as Baldur's Gate, Baldur's Gate II, 
and Planescape: Torment.`,
	Version: "8",

	PersistentPreRunE:  startProfiling,
	PersistentPostRunE: stopProfiling,
}

var profilingFile *os.File

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(twoda.NewCommand())
	rootCmd.AddCommand(bif.NewCommand())
	rootCmd.AddCommand(csv.NewCommand())
	rootCmd.AddCommand(dialog.NewCommand())
	rootCmd.AddCommand(text.NewCommand())
	rootCmd.AddCommand(tra.NewCommand())

	rootCmd.PersistentFlags().BoolP("profile", "p", false, "Enable profiling")
	rootCmd.PersistentFlags().MarkHidden("profile")

	rootCmd.PersistentFlags().StringP("config", "c", "sbt-inf.toml", "`path` to config file")
	rootCmd.PersistentFlags().StringP("game", "g", "", "game `name` from config to use (default - first one in A-Z order)")
	rootCmd.PersistentFlags().StringP("key", "k", "", "`path` to chitin.key file")
	rootCmd.MarkFlagsMutuallyExclusive("config", "key")
	rootCmd.MarkFlagsMutuallyExclusive("key", "game")
}

func startProfiling(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetBool("profile")
	if profile {
		var err error
		profilingFile, err = os.Create("cpu.pprof")
		if err != nil {
			return err
		}
		pprof.StartCPUProfile(profilingFile)
	}
	return nil
}

func stopProfiling(cmd *cobra.Command, args []string) error {
	pprof.StopCPUProfile()
	profilingFile.Close()
	return nil
}
