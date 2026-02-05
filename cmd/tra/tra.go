// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package tra

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tra",
		Short: "Work with TRA files",
		Long:  `Utilities for working with TRA (WeiDU translation) files.`,
	}

	cmd.AddCommand(NewExportCommand())
	cmd.AddCommand(NewImportCommand())
	cmd.AddCommand(NewUpdateCommand())

	return cmd
}
