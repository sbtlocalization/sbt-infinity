// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package ids

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ids",
		Short: "IDS file utilities",
		Long:  `Utilities for working with Infinity Engine IDS files.`,
	}

	cmd.AddCommand(NewShowCommand())
	return cmd
}
