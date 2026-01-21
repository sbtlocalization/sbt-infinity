// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Work with TLK files",
		Long:  `Utilities for working with TLK files.`,
	}

	cmd.PersistentFlags().StringP("lang", "l", "en_US", "language `code` for TLK file")
	cmd.PersistentFlags().StringP("tlk", "t", "<KEY_DIR>/lang/<LANG>/dialog.tlk", "`path` to dialog.tlk file")
	cmd.PersistentFlags().BoolP("feminine", "f", false, "open dialogf.tlk instead of dialog.tlk")

	cmd.MarkFlagsMutuallyExclusive("tlk", "lang")
	cmd.MarkFlagsMutuallyExclusive("tlk", "feminine")

	cmd.MarkFlagFilename("tlk", "tlk")

	cmd.AddCommand(NewLsCommand())
	cmd.AddCommand(NewExCommand())

	return cmd
}
