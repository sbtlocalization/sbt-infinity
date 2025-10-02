// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dialog",
		Short: "Dialog utilities",
		Long:  `Utilities for working with Infinity Engine dialogs.`,
	}

	config.AddGameFlag(cmd)

	cmd.AddCommand(NewLsCommand())
	cmd.AddCommand(NewExportCommand())
	return cmd
}
