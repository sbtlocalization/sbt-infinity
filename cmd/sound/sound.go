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

	cmd.PersistentFlags().StringP("filter", "f", "", "Regex for resource name filtering")

	cmd.AddCommand(NewExportCommand())

	return cmd
}
