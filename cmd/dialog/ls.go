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
		Use:     "list [DLG-file...]",
		Aliases: []string{"ls"},
		Short:   "List dialogs from the game",
		Long: `List all dialogs or specific dialog files from the game.
Reads the game structure from chitin.key file and dialog.tlk file, and optionally lists
only specified dialog files (e.g., ABISHAB.DLG, DMORTE.DLG with or without extension).`,
		Args: cobra.MinimumNArgs(0),
		RunE: runLs,
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("with-length", "l", false, "Include number of nodes for each dialog (slower)")

	return cmd
}

func runLs(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	withLength, _ := cmd.Flags().GetBool("with-length")
	tlkPath, _ := cmd.Flags().GetString("tlk")

	// Resolve the key path and parse other files using the common helper
	keyPath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		return err
	}

	var dialogFiles []string
	if len(args) > 0 {
		dialogFiles = args
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

	dc := dialog.NewDialogBuilder(dlgFs, tlkFs, false, false)

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
		var dlg *dialog.DialogCollection
		var err error
		if withLength {
			dlg, err = dc.LoadAllDialogs(tlkPath, df)
		} else {
			dlg, err = dc.LoadAllRootStates(df)
		}
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
					Nodes *int   `json:"nodes,omitempty"`
				}{
					State: d.Id.Index,
					File:  fullDlgName,
					Bif:   bifPath,
				}
				if withLength {
					nodeCount := d.NodeCount()
					output.Nodes = &nodeCount
				}
				jsonData, err := json.Marshal(output)
				if err != nil {
					return fmt.Errorf("error marshaling JSON: %v", err)
				}
				fmt.Println(string(jsonData))
			} else {
				if withLength {
					fmt.Printf("%s\t%d\n", d.Id, d.NodeCount())
				} else {
					fmt.Println(d.Id)
				}
			}
		}
	}

	return nil
}
