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

	cmd.PersistentFlags().StringP("tlk", "t", "", "Path to dialog.tlk file (default: <key_dir>/lang/en_US/dialog.tlk)")
	config.AddGameFlag(cmd)

	cmd.AddCommand(NewLsCommand())

	return cmd
}
