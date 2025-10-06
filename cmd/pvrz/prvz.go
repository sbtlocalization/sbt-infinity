// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package pvrz

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pvrz",
		Short: "PVRZ utilities",
		Long:  `Utilities for working with Infinity Engine PVRZ files.`,
	}

	cmd.AddCommand(NewExCommand())

	return cmd
}
