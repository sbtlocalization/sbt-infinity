// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"encoding/json"
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
Reads the texts from dialog.tlk file, and optionally lists only specified 
text IDs (e.g., 1234, 5678).

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
	lang, _ := cmd.Flags().GetString("lang")
	feminine, _ := cmd.Flags().GetBool("feminine")

	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Resolve the key path and parse other files using the common helper
	keyPath, textIds, err := config.ResolveKeyPathFromArgs(args, gameName)
	if err != nil {
		return err
	}

	osFs := afero.NewOsFs()

	var tlkFs afero.Fs
	if !cmd.Flags().Changed("tlk") {
		tlkFs = afero.NewBasePathFs(osFs, filepath.Dir(keyPath))
		if feminine {
			tlkPath = filepath.Join("lang", lang, "dialogf.tlk")
		} else {
			tlkPath = filepath.Join("lang", lang, "dialog.tlk")
		}
	} else {
		tlkFs = osFs
	}

	tlkFile, err := p.ReadTlkFile(tlkFs, tlkPath)
	if err != nil {
		return err
	}
	defer tlkFile.Close()

	tlk := tlkFile.Tlk

	width := len(fmt.Sprintf("%d", len(tlk.Entries)-1))
	if len(textIds) > 0 {
		maxIdLen := 0
		for _, idStr := range textIds {
			if len(idStr) > maxIdLen {
				maxIdLen = len(idStr)
			}
		}
		width = maxIdLen
	}

	var printFunc func(int, *p.Tlk_StringEntry)
	if jsonOutput {
		printFunc = jsonifyEntry
	} else {
		printFunc = func(id int, entry *p.Tlk_StringEntry) {
			printEntry(width, id, entry)
		}
	}

	if len(textIds) > 0 {
		for _, idStr := range textIds {
			var id int
			_, err := fmt.Sscanf(idStr, "%d", &id)
			if err != nil {
				fmt.Printf("Invalid ID format: %s\n", idStr)
				continue
			}
			entry := tlk.Entries[id]
			printFunc(id, entry)
		}
	} else {
		for i, entry := range tlk.Entries {
			printFunc(i, entry)
		}
	}

	return nil
}

type tlkEntry struct {
	Id       int    `json:"id"`
	HasText  bool   `json:"has_text"`
	HasSound bool   `json:"has_sound"`
	HasToken bool   `json:"has_token"`
	Text     string `json:"text,omitempty"`
	Sound    string `json:"sound,omitempty"`
}

func jsonifyEntry(id int, entry *p.Tlk_StringEntry) {
	hasText := entry.Flags.TextExists
	text := ""
	if hasText {
		t, err := entry.Text()
		if err != nil {
			t = fmt.Sprintf("error reading text: %v", err)
		}
		text = t
	}
	hasSound := entry.Flags.SoundExists
	sound := ""
	if hasSound {
		sound = entry.AudioName
	}
	jsonEntry := tlkEntry{
		Id:       id,
		HasText:  hasText,
		HasSound: hasSound,
		HasToken: entry.Flags.TokenExists,
		Sound:    sound,
		Text:     text,
	}
	jsonData, _ := json.Marshal(jsonEntry)
	fmt.Println(string(jsonData))
}

func printEntry(width, id int, entry *p.Tlk_StringEntry) {
	text, err := entry.Text()
	if err != nil {
		fmt.Printf("#%d: error reading text: %v\n", id, err)
	} else {
		fmt.Printf("#%-*d %s\n", width, id, text)
	}
}
