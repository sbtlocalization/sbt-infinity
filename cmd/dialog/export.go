// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/supersonicpineapple/go-jsoncanvas"
)

// exportDialogsCmd represents the export-dialogs command
func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [path to chitin.key] [tlk files...] [dlg files...]",
		Short: "Export dialogs as JSON Canvas files",
		Long: `Export dialogs from DLG files (with texts from TLK file) as JSON Canvas files.
Creates a visual representation of dialog structures.

If no key file path is provided, uses the first game from .sbt-inf.toml config.`,
		Args: cobra.MinimumNArgs(0),
		RunE: runExportDialogs,
	}

	cmd.Flags().StringP("output", "o", "", "Output directory")
	cmd.Flags().StringP("tlk", "t", "", "Path to dialog.tlk file")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")
	config.AddGameFlag(cmd)

	return cmd
}

func runExportDialogs(cmd *cobra.Command, args []string) error {
	gameName, _ := cmd.Flags().GetString("game")
	tlkPath, _ := cmd.Flags().GetString("tlk")
	outputDir, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")

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

	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("unable to create output directory %s: %v", outputDir, err)
		}
	}

	for _, df := range dialogFiles {
		dlg, err := dc.LoadAllDialogs(tlkPath, df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		if verbose {
			fmt.Printf("%s: loaded %d dialogs\n", df, len(dlg.Dialogs))
		}
		for _, d := range dlg.Dialogs {
			canvas := d.ToJsonCanvas()
			dialogName := strings.TrimSuffix(d.Id.DlgName, filepath.Ext(d.Id.DlgName))
			fileName := filepath.Join(outputDir, fmt.Sprintf("%s-%d.canvas", dialogName, d.Id.Index))
			file, err := os.Create(fileName)
			if verbose {
				fmt.Printf("  exporting dialog %s to %s\n", d.Id, fileName)
			}
			if err != nil {
				fmt.Printf("error saving canvas for dialog %s: %v\n", d.Id, err)
				continue
			}
			defer file.Close()

			err = jsoncanvas.Encode(canvas, file)
			if err != nil {
				return fmt.Errorf("error encoding canvas for dialog %s: %v", d.Id, err)
			}
		}
	}
	return nil
}
