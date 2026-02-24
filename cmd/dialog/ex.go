// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package dialog

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
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
	cmd.Flags().String("sound-prefix", "sounds/", "Prefix for sound names in dCanvas output")
	cmd.Flags().String("sound-suffix", ".wav", "Suffix for sound names in dCanvas output")
	cmd.Flags().String("report", "", "Path to write a JSON report of all sounds and strrefs used in dialogs")
	cmd.Flags().Bool("scan-overlaps", false, "Scan dCanvas output for overlapping nodes and include affected file paths in the report")

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
	soundPrefix, _ := cmd.Flags().GetString("sound-prefix")
	soundSuffix, _ := cmd.Flags().GetString("sound-suffix")
	reportPath, _ := cmd.Flags().GetString("report")
	scanOverlaps, _ := cmd.Flags().GetBool("scan-overlaps")

	fmtOpts := dialog.FormatOptions{
		SoundPrefix: soundPrefix,
		SoundSuffix: soundSuffix,
	}

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

	soundSet := make(map[string]struct{})
	strrefSet := make(map[uint32]struct{})
	overlapFiles := make([]string, 0)

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
			canvas := d.ToDCanvas(fmtOpts)

			if reportPath != "" {
				for _, node := range d.All() {
					switch node.Type {
					case dialog.StateNodeType:
						strrefSet[node.State.TextRef] = struct{}{}
						if node.State.Sound != "" {
							soundSet[node.State.Sound] = struct{}{}
						}
					case dialog.TransitionNodeType:
						if node.Transition.HasText {
							strrefSet[node.Transition.TextRef] = struct{}{}
							if node.Transition.Sound != "" {
								soundSet[node.Transition.Sound] = struct{}{}
							}
						}
						if node.Transition.HasJournalText {
							strrefSet[node.Transition.JournalTextRef] = struct{}{}
							if node.Transition.JournalSound != "" {
								soundSet[node.Transition.JournalSound] = struct{}{}
							}
						}
					}
				}
			}
			dialogName := strings.TrimSuffix(d.Id.DlgName, filepath.Ext(d.Id.DlgName))
			fileName := filepath.Join(outputDir, fmt.Sprintf("%s-%d.d.canvas", dialogName, d.Id.Index))
			if scanOverlaps && canvas.HasOverlappingNodes() {
				overlapFiles = append(overlapFiles, fileName)
				if verbose {
					fmt.Printf("  warning: overlapping nodes detected in %s\n", fileName)
				}
			}
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

	if reportPath != "" {
		sounds := make([]string, 0, len(soundSet))
		for s := range soundSet {
			sounds = append(sounds, s)
		}
		slices.Sort(sounds)

		strrefs := make([]uint32, 0, len(strrefSet))
		for s := range strrefSet {
			strrefs = append(strrefs, s)
		}
		slices.Sort(strrefs)

		slices.Sort(overlapFiles)

		report := struct {
			Sounds      []string `json:"sounds"`
			Strrefs     []uint32 `json:"strrefs"`
			OverlapFiles []string `json:"overlap_files,omitempty"`
		}{
			Sounds:      sounds,
			Strrefs:     strrefs,
			OverlapFiles: overlapFiles,
		}

		reportFile, err := os.Create(reportPath)
		if err != nil {
			return fmt.Errorf("error creating report file %s: %v", reportPath, err)
		}
		defer reportFile.Close()

		enc := json.NewEncoder(reportFile)
		enc.SetIndent("", "\t")
		if err := enc.Encode(report); err != nil {
			return fmt.Errorf("error writing report to %s: %v", reportPath, err)
		}

		if verbose {
			fmt.Printf("Report written to %s: %d sounds, %d strrefs\n", reportPath, len(sounds), len(strrefs))
		}
	}

	return nil
}
