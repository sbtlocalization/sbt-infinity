// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Work with TLK files",
		Long:  `Utilities for working with TLK files.`,
	}

	cmd.PersistentFlags().StringP("lang", "l", "en_US", "Language code for TLK file")
	cmd.PersistentFlags().StringP("tlk", "t", "<KEY_DIR>/lang/<LANG>/dialog.tlk", "Path to dialog.tlk file")
	cmd.PersistentFlags().BoolP("feminine", "f", false, "Open dialogf.tlk instead of dialog.tlk")
	config.AddGameFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive("tlk", "lang")
	cmd.MarkFlagsMutuallyExclusive("tlk", "feminine")

	cmd.AddCommand(NewLsCommand())

	return cmd
}
