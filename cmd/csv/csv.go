// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package csv

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv",
		Short: "CSV utilities",
		Long:  `Utilities for working with CSV files.`,
	}

	cmd.AddCommand(NewDiffCommand())
	return cmd
}
