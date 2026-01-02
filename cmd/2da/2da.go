// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package twoda

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "2da",
		Short: "2DA file utilities",
		Long:  `Utilities for working with Infinity Engine 2DA files.`,
	}

	cmd.AddCommand(NewShowCommand())
	return cmd
}
