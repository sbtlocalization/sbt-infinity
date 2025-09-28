// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [path to chitin.key] [dialog files...]",
		Aliases: []string{"ls"},
		Short:   "List dialogs from the game",
		Long: `List all dialogs or specific dialog files from the game.
		Reads the game structure from chitin.key file and dialog.tlk file, and optionally lists
		only specified dialog files (e.g., ABISHAB.DLG, DMORTE.DLG).
		
		If no key file path is provided, uses the first game from .sbt-inf.toml config.`,
		Args: cobra.MinimumNArgs(0),
		RunE: runLs,
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().StringP("tlk", "t", "", "Path to dialog.tlk file")
	config.AddGameFlag(cmd)

	return cmd
}

func runLs(cmd *cobra.Command, args []string) error {
	gameName, _ := cmd.Flags().GetString("game")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	tlkPath, _ := cmd.Flags().GetString("tlk")

	// Resolve the key path and parse other files using the common helper
	keyPath, dialogFiles, err := config.ResolveKeyPathFromArgs(args, gameName)
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

	dlgFs := fs.NewInfinityFs(keyPath, fs.FileType_DLG)

	dc := dialog.NewDialogBuilder(dlgFs, tlkFs)

	if len(dialogFiles) == 0 {
		dir, err := dlgFs.Open("DLG")
		if err != nil {
			return fmt.Errorf("unable to list existing DLG files: %v", err)
		}
		dialogFiles, err = dir.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("unable to read dialog directory names: %v", err)
		}
	}

	for _, df := range dialogFiles {
		dlg, err := dc.LoadAllRootStates(df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		// Get BIF file path for this dialog file
		var bifPath string
		var fullDlgName string
		if jsonOutput {
			fullDlgName = df
			if !strings.HasSuffix(strings.ToLower(fullDlgName), ".dlg") {
				fullDlgName = fullDlgName + ".DLG"
			}

			// Get BIF file path from the InfinityFs
			if path, err := dlgFs.GetBifFilePath(fullDlgName); err == nil {
				bifPath = path
			}
		}

		for _, d := range dlg.Dialogs {
			if jsonOutput {
				output := struct {
					State uint32 `json:"root_state"`
					File  string `json:"file"`
					Bif   string `json:"bif"`
				}{
					State: d.Id.Index,
					File:  fullDlgName,
					Bif:   bifPath,
				}
				jsonData, err := json.Marshal(output)
				if err != nil {
					return fmt.Errorf("error marshaling JSON: %v", err)
				}
				fmt.Println(string(jsonData))
			} else {
				fmt.Println(d.Id)
			}
		}
	}

	return nil
}
