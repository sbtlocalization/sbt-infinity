// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/supersonicpineapple/go-jsoncanvas"
	_ "golang.org/x/image/bmp"
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
	cmd.Flags().StringP("tlk", "t", "", "Path to dialog.tlk file (default: <key_dir>/lang/en_US/dialog.tlk)")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolP("speakers", "s", true, "Load and export information about characters from CRE files")
	config.AddGameFlag(cmd)

	return cmd
}

func runExportDialogs(cmd *cobra.Command, args []string) error {
	gameName, _ := cmd.Flags().GetString("game")
	tlkPath, _ := cmd.Flags().GetString("tlk")
	outputDir, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")
	withCreatures, _ := cmd.Flags().GetBool("speakers")

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

	typesToLoad := []fs.FileType{fs.FileType_DLG}
	if withCreatures {
		typesToLoad = append(typesToLoad, fs.FileType_CRE, fs.FileType_BMP)
	}
	dlgFs := fs.NewInfinityFs(keyPath, typesToLoad...)

	dc := dialog.NewDialogBuilder(dlgFs, tlkFs, withCreatures)

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

			for _, creature := range d.AllCreatures {
				if creature.Portrait != "" {
					if picFile, err := dlgFs.Open(creature.Portrait); err == nil {
						img, _, err := image.Decode(picFile)
						picFile.Close()
						if err != nil {
							fmt.Printf("error converting portrait %s for %s: %v\n", creature.Portrait, creature.LongName, err)
							continue
						}

						picName := strings.TrimSuffix(creature.Portrait, filepath.Ext(creature.Portrait)) + ".png"

						err = os.MkdirAll(filepath.Join(outputDir, "portraits"), 0755)
						if err != nil {
							fmt.Printf("error creating portraits directory for %s: %v\n", creature.LongName, err)
							continue
						}

						outFile, err := os.Create(filepath.Join(outputDir, "portraits", picName))
						if err != nil {
							fmt.Printf("error creating portrait file %s for %s: %v\n", picName, creature.LongName, err)
							continue
						}
						defer outFile.Close()

						if err := png.Encode(outFile, img); err != nil {
							fmt.Printf("error writing portrait file %s for %s: %v\n", picName, creature.LongName, err)
							continue
						}
					}
				}
			}
		}
	}
	return nil
}
