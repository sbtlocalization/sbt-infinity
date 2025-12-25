// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/samber/lo"
	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	p "github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/sbtlocalization/sbt-infinity/text"
	"github.com/sbtlocalization/sbt-infinity/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewExCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [ID...]",
		Aliases: []string{"ex"},
		Short:   "Export textual resources from the game as xlsx",
		Long: `Export all textual resources or specific IDs from the game.
Reads the texts from dialog.tlk file, and optionally extracts only specified
text IDs (e.g., 1234, 5678).`,
		Args: cobra.MinimumNArgs(0),
		RunE: runEx,
	}

	cmd.Flags().StringP("output", "o", "dialog.xlsx", "output xlsx file `path`")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose output")
	cmd.Flags().String("dlg-base-url", "", "base `URL` for dialog references (overrides config)")
	cmd.Flags().StringSlice("context-from", []string{}, "load context from types of files. Use 'all' to include all types.\nUse 'bif `types`' command to see all types.")

	return cmd
}

func runEx(cmd *cobra.Command, args []string) error {
	tlkPath, _ := cmd.Flags().GetString("tlk")
	lang, _ := cmd.Flags().GetString("lang")
	feminine, _ := cmd.Flags().GetBool("feminine")
	verbose, _ := cmd.Flags().GetBool("verbose")
	baseUrl, _ := config.ResolveDialogBaseUrl(cmd)
	contextFrom, _ := cmd.Flags().GetStringSlice("context-from")

	outputPath, _ := cmd.Flags().GetString("output")
	if cmd.Flags().Changed("output") && !strings.HasSuffix(strings.ToLower(outputPath), ".xlsx") {
		outputPath = outputPath + ".xlsx"
	}

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

	if verbose {
		fmt.Print("loading TLK file... ")
	}
	tlkFile, err := p.ReadTlkFile(tlkFs, tlkPath)
	if err != nil {
		return err
	}
	collection := text.NewTextCollection(tlkFile.Tlk)
	tlkFile.Close()
	if verbose {
		fmt.Println("done.")
	}

	contextTypes := []fs.FileType{
		fs.FileType_ARE,
		fs.FileType_CHU,
		fs.FileType_CRE,
		fs.FileType_DLG,
		fs.FileType_ITM,
		fs.FileType_PRO,
		fs.FileType_SPL,
		fs.FileType_STO,
		fs.FileType_WMP,
	}
	if !slices.Contains(contextFrom, "all") {
		contextTypes = lo.UniqMap(contextFrom, utils.Iteratee(fs.FileTypeFromExtension))
	}

	infFs := fs.NewInfinityFs(keyPath)

	for _, t := range contextTypes {
		switch t {
		case fs.FileType_ARE:
			err = processAreas(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process areas:", err)
			}
		case fs.FileType_CHU:
			err = processUiScreens(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process UI screens:", err)
			}
		case fs.FileType_CRE:
			err = processCreatures(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process creatures:", err)
			}
		case fs.FileType_DLG:
			err = processDialogs(collection, infFs, baseUrl, verbose)
			if err != nil {
				fmt.Println("warning: unable to process dialogs:", err)
			}
		case fs.FileType_ITM:
			err = processItems(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process items:", err)
			}
		case fs.FileType_PRO:
			err = processProjectiles(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process projectiles:", err)
			}
		case fs.FileType_SPL:
			err = processSpells(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process spells:", err)
			}
		case fs.FileType_STO:
			err = processStores(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process stores:", err)
			}
		case fs.FileType_WMP:
			err = processWorldMaps(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process world maps:", err)
			}
		default:
			continue
		}
	}

	err = collection.ExportToXlsx(outputPath)
	if err != nil {
		return err
	}

	return nil
}

func processDialogs(collection *text.TextCollection, infFs afero.Fs, baseUrl string, verbose bool) error {
	dlgBuilder := dialog.NewDialogBuilder(infFs, nil, false, verbose)
	dir, err := infFs.Open("DLG")
	if err != nil {
		return fmt.Errorf("unable to list existing DLG files: %v", err)
	}
	defer dir.Close()
	dialogFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read dialog directory names: %v", err)
	}

	total := len(dialogFiles)

	if verbose {
		fmt.Print("extracting context from dialogs...")
	}

	for _, df := range dialogFiles {
		dc, err := dlgBuilder.LoadAllDialogs("", df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		collection.LoadContextFromDialogs(baseUrl, dc)
	}

	if verbose {
		fmt.Printf(" done (%d files).\n", total)
	}

	return nil
}

func processCreatures(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	dir, err := infFs.Open("CRE")
	if err != nil {
		return fmt.Errorf("unable to list existing CRE files: %v", err)
	}
	defer dir.Close()

	creFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read CRE directory names: %v", err)
	}

	var ids *p.Ids

	sndslot, err := infFs.Open("SNDSLOT.IDS")
	if err != nil {
		fmt.Println("warning: unable to open SNDSLOT.IDS:", err)
	} else {
		ids, err = p.ParseIds(sndslot)
		sndslot.Close()
		if err != nil {
			fmt.Println("warning: unable to parse SNDSLOT.IDS:", err)
		}
	}

	total := len(creFiles)

	if verbose {
		fmt.Print("extracting context from creatures...")
	}

	for _, cf := range creFiles {
		creFile, err := infFs.Open(cf)
		if err != nil {
			return fmt.Errorf("unable to open CRE file %q: %v", cf, err)
		}
		defer creFile.Close()

		cre := p.NewCre()
		stream := kaitai.NewStream(creFile)
		err = cre.Read(stream, nil, cre)
		if err != nil {
			return fmt.Errorf("unable to parse CRE file %q: %v", cf, err)
		}

		collection.LoadContextFromCreature(cf, cre, ids)
	}

	if verbose {
		fmt.Printf(" done (%d files).\n", total)
	}

	return nil
}

func processFiles(
	infFs afero.Fs,
	verbose bool,
	dirName string,
	entityName string,
	processFile func(filename string, stream *kaitai.Stream) error,
) error {
	dir, err := infFs.Open(dirName)
	if err != nil {
		return fmt.Errorf("unable to list existing %s files: %v", dirName, err)
	}
	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read %s directory names: %v", dirName, err)
	}

	total := len(files)
	processed := 0
	hasWarnings := false

	if verbose {
		fmt.Printf("extracting context from %s...", entityName)
	}

	for _, f := range files {
		file, err := infFs.Open(f)
		if err != nil {
			if verbose {
				if !hasWarnings {
					fmt.Println()
					hasWarnings = true
				}
				fmt.Printf("  warning: unable to open %s file %q: %v. skipping...\n", dirName, f, err)
			}
			continue
		}

		stream := kaitai.NewStream(file)
		if err := processFile(f, stream); err != nil {
			if verbose {
				if !hasWarnings {
					fmt.Println()
					hasWarnings = true
				}
				fmt.Printf("  warning: unable to parse %s file %q: %v. skipping...\n", dirName, f, err)
			}
		} else {
			processed++
		}
		file.Close()
	}

	if verbose {
		if processed == total {
			fmt.Printf(" done (%d files).\n", total)
		} else {
			fmt.Printf("done (%d/%d files).\n", processed, total)
		}
	}

	return nil
}

func processUiScreens(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "CHU", "UI screens", func(filename string, stream *kaitai.Stream) error {
		chu := p.NewChu()
		if err := chu.Read(stream, nil, chu); err != nil {
			return err
		}
		collection.LoadContextFromUiScreens(filename, chu)
		return nil
	})
}

func processWorldMaps(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "WMP", "world maps", func(filename string, stream *kaitai.Stream) error {
		wmp := p.NewWmp()
		if err := wmp.Read(stream, nil, wmp); err != nil {
			return err
		}
		collection.LoadContextFromWorldMaps(filename, wmp)
		return nil
	})
}

func processAreas(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "ARE", "areas", func(filename string, stream *kaitai.Stream) error {
		are := p.NewAre()
		if err := are.Read(stream, nil, are); err != nil {
			return err
		}
		collection.LoadContextFromArea(filename, are)
		return nil
	})
}

func processItems(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "ITM", "items", func(filename string, stream *kaitai.Stream) error {
		itm := p.NewItm()
		if err := itm.Read(stream, nil, itm); err != nil {
			return err
		}
		collection.LoadContextFromItem(filename, itm)
		return nil
	})
}

func processProjectiles(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "PRO", "projectiles", func(filename string, stream *kaitai.Stream) error {
		pro := p.NewPro()
		if err := pro.Read(stream, nil, pro); err != nil {
			return err
		}
		collection.LoadContextFromProjectile(filename, pro)
		return nil
	})
}

func processSpells(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "SPL", "spells", func(filename string, stream *kaitai.Stream) error {
		spl := p.NewSpl()
		if err := spl.Read(stream, nil, spl); err != nil {
			return err
		}
		collection.LoadContextFromSpell(filename, spl)
		return nil
	})
}

func processStores(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "STO", "stores", func(filename string, stream *kaitai.Stream) error {
		sto := p.NewSto()
		if err := sto.Read(stream, nil, sto); err != nil {
			return err
		}
		collection.LoadContextFromStore(filename, sto)
		return nil
	})
}
