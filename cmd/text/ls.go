// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"path/filepath"

	"github.com/sbtlocalization/sbt-infinity/config"
	p "github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [path to chitin.key] [ID...]",
		Aliases: []string{"ls"},
		Short:   "List textual resources from the game",
		Long: `List all textual resources or specific IDs from the game.
		Reads the game structure from chitin.key file and text.tlk file, and optionally lists
		only specified text IDs (e.g., 1234, 5678).
		
		If no key file path is provided, uses the first game from .sbt-inf.toml config.`,
		Args: cobra.MinimumNArgs(0),
		RunE: runLs,
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	return cmd
}

func runLs(cmd *cobra.Command, args []string) error {
	gameName, _ := cmd.Flags().GetString("game")
	tlkPath, _ := cmd.Flags().GetString("tlk")

	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Resolve the key path and parse other files using the common helper
	keyPath, textIds, err := config.ResolveKeyPathFromArgs(args, gameName)
	if err != nil {
		return err
	}

	osFs := afero.NewOsFs()

	var tlkFs afero.Fs
	if tlkPath == "" {
		tlkFs = afero.NewBasePathFs(osFs, filepath.Join(filepath.Dir(keyPath)))
		tlkPath = "lang/en_US/dialog.tlk"
	} else {
		tlkFs = osFs
	}

	tlkFile, err := p.ReadTlkFile(tlkFs, tlkPath)
	if err != nil {
		return err
	}
	defer tlkFile.Close()

	// [TODO] @GooRoo: filter by textIds if provided
	_ = textIds

	// [TODO] @GooRoo: support json output
	_ = jsonOutput

	tlk := tlkFile.Tlk
	for i, entry := range tlk.Entries {
		printEntry(i, entry)
	}

	return nil
}

func printEntry(id int, entry *p.Tlk_StringEntry) {
	text, err := entry.Text()
	if err != nil {
		fmt.Printf("%d: error reading text: %v\n", id, err)
	} else {
		fmt.Printf("%d: %s\n", id, text)
	}
}
