// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package sound

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sound",
		Short: "Work with game audio files",
		Long:  `Utilities for converting between Infinity Engine audio formats (WAVC/ACM) and standard WAV/FLAC.`,
	}

	cmd.PersistentFlags().StringP("filter", "f", "", "Wildcard for resourse name filtering. Case insensitive.")
	cmd.PersistentFlags().StringP("bif-filter", "b", "", "Wildcard for filtering by BIF names. Case insensitive.")

	cmd.AddCommand(NewExportCommand())

	return cmd
}
