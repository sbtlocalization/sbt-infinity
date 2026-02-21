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
	"github.com/sbtlocalization/sbt-infinity/dcanvas"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	_ "golang.org/x/image/bmp"
)

// exportDialogsCmd represents the export-dialogs command
func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [DLG-file...]",
		Aliases: []string{"ex"},
		Short:   "Export dialogs as dCanvas files",
		Long: `Export dialogs from DLG files (with texts from TLK file) as dCanvas files.
Creates a visual representation of dialog structures.`,
		Args: cobra.MinimumNArgs(0),
		RunE: runExportDialogs,
	}

	cmd.Flags().StringP("lang", "l", "en_US", "Language code for TLK file")
	cmd.Flags().StringP("tlk", "t", "<KEY_DIR>/lang/<LANG>/dialog.tlk", "Path to dialog.tlk file")
	cmd.Flags().BoolP("feminine", "f", false, "Open dialogf.tlk instead of dialog.tlk")

	cmd.Flags().StringP("output", "o", "", "Output directory")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolP("speakers", "s", true, "Load and export information about characters from CRE files")
	cmd.Flags().StringSliceP("exclude", "x", []string{}, "Exclude specific dialog files (e.g., ABISHAB.DLG)")

	cmd.MarkFlagDirname("output")

	return cmd
}

func runExportDialogs(cmd *cobra.Command, args []string) error {
	tlkPath, _ := cmd.Flags().GetString("tlk")
	lang, _ := cmd.Flags().GetString("lang")
	feminine, _ := cmd.Flags().GetBool("feminine")
	outputDir, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")
	withCreatures, _ := cmd.Flags().GetBool("speakers")
	excludeFiles, _ := cmd.Flags().GetStringSlice("exclude")

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

	typesToLoad := []fs.FileType{fs.FileType_DLG}
	if withCreatures {
		typesToLoad = append(typesToLoad, fs.FileType_CRE, fs.FileType_BMP)
	}
	dlgFs := fs.NewInfinityFs(keyPath, fs.WithTypeFilter(typesToLoad...))

	dc := dialog.NewDialogBuilder(dlgFs, tlkFs, withCreatures, verbose)

	if len(dialogFiles) == 0 {
		dir, err := dlgFs.Open("DLG")
		if err != nil {
			return fmt.Errorf("unable to list existing DLG files: %v", err)
		}
		defer dir.Close()
		dialogFiles, err = dir.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("unable to read dialog directory names: %v", err)
		}
	}

	excludeMap := make(map[string]bool)
	for _, ef := range excludeFiles {
		ef = strings.ToUpper(ef)
		if !strings.HasSuffix(ef, ".DLG") {
			ef = ef + ".DLG"
		}
		excludeMap[ef] = true
	}

	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("unable to create output directory %s: %v", outputDir, err)
		}
	}

	for _, df := range dialogFiles {
		if excludeMap[df] {
			if verbose {
				fmt.Printf("Skipping excluded dialog file %s\n", df)
			}
			continue
		}

		dlg, err := dc.LoadAllDialogs(tlkPath, df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		if verbose {
			fmt.Printf("%s: loaded %d dialogs\n", df, len(dlg.Dialogs))
		}
		for _, d := range dlg.Dialogs {
			canvas := d.ToDCanvas()
			dialogName := strings.TrimSuffix(d.Id.DlgName, filepath.Ext(d.Id.DlgName))
			fileName := filepath.Join(outputDir, fmt.Sprintf("%s-%d.d.canvas", dialogName, d.Id.Index))
			file, err := os.Create(fileName)
			if verbose {
				fmt.Printf("  exporting dialog %s to %s\n", d.Id, fileName)
			}
			if err != nil {
				fmt.Printf("error saving canvas for dialog %s: %v\n", d.Id, err)
				continue
			}
			defer file.Close()

			err = dcanvas.Encode(canvas, file)
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
