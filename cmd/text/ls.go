// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	p "github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [ID...]",
		Aliases: []string{"ls"},
		Short:   "List textual resources from the game",
		Long: `List all textual resources or specific IDs from the game.
Reads the texts from dialog.tlk file, and optionally lists only specified 
text IDs (e.g., 1234, 5678).`,
		Example: `  List all text entries:

      sbt-inf text list

  List specific text entries by IDs:

      sbt-inf text list 1234 5678

  List a range of text entries:

      sbt-inf text list 0..1000

  List specific and range of text entries:

      sbt-inf text list 0..100 2000 3000..4000`,
		Args: cobra.MinimumNArgs(0),
		RunE: runLs,
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	return cmd
}

func runLs(cmd *cobra.Command, args []string) error {
	tlkPath, _ := cmd.Flags().GetString("tlk")
	lang, _ := cmd.Flags().GetString("lang")
	feminine, _ := cmd.Flags().GetBool("feminine")

	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Parse text IDs from args
	var textIds []string
	if len(args) > 0 {
		textIds = args
	}

	// Resolve the key path using flags
	keyPath, err := config.ResolveKeyPath(cmd)
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

	var ids []int
	width := len(fmt.Sprintf("%d", len(tlk.Entries)-1))
	if len(textIds) > 0 {
		var err error
		ids, width, err = splitIds(textIds)
		if err != nil {
			return err
		}
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
		for _, id := range ids {
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

func splitIds(input []string) ([]int, int, error) {
	var ids []int
	var width int
	for _, id := range input {
		if strings.Contains(id, "..") {
			parts := strings.Split(id, "..")
			if len(parts) == 2 {
				start, end := parts[0], parts[1]
				var s, e int
				_, err1 := fmt.Sscanf(start, "%d", &s)
				_, err2 := fmt.Sscanf(end, "%d", &e)
				if err1 == nil && err2 == nil && s <= e {
					ids = append(ids, makeRange(s, e)...)
				}
				width = max(width, len(end))
			} else {
				return nil, 0, fmt.Errorf("invalid range format: %s", id)
			}
		} else {
			var idNum int
			_, err := fmt.Sscanf(id, "%d", &idNum)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid ID format: %s", id)
			}
			ids = append(ids, idNum)
			width = max(width, len(id))
		}
	}
	ids = unique(ids)
	slices.Sort(ids)
	return ids, width, nil
}

func makeRange(min, max int) []int {
	result := make([]int, max-min+1)
	for i := range result {
		result[i] = min + i
	}
	return result
}

func unique[T comparable](in []T) []T {
	seen := make(map[T]struct{}, len(in))
	out := make([]T, 0, len(in))
	for _, v := range in {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
