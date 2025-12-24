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
	if verbose {
		fmt.Print("extracting context from dialogs... ")
	}
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

	for _, df := range dialogFiles {
		dc, err := dlgBuilder.LoadAllDialogs("", df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		collection.LoadContextFromDialogs(baseUrl, dc)
	}

	if verbose {
		fmt.Println("done.")
	}

	return nil
}

func processCreatures(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	if verbose {
		fmt.Print("extracting context from creatures... ")
	}

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
		fmt.Println("done.")
	}

	return nil
}

func processUiScreens(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	if verbose {
		fmt.Print("extracting context from UI screens... ")
	}

	dir, err := infFs.Open("CHU")
	if err != nil {
		return fmt.Errorf("unable to list existing CHU files: %v", err)
	}
	defer dir.Close()

	uiFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read CHU directory names: %v", err)
	}

	for _, uf := range uiFiles {
		uiFile, err := infFs.Open(uf)
		if err != nil {
			return fmt.Errorf("unable to open CHU file %q: %v", uf, err)
		}
		defer uiFile.Close()

		chu := p.NewChu()
		stream := kaitai.NewStream(uiFile)
		err = chu.Read(stream, nil, chu)
		if err != nil {
			return fmt.Errorf("unable to parse CHU file %q: %v", uf, err)
		}

		collection.LoadContextFromUiScreens(uf, chu)
	}

	if verbose {
		fmt.Println("done.")
	}

	return nil
}

func processWorldMaps(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	if verbose {
		fmt.Print("extracting context from world maps... ")
	}

	dir, err := infFs.Open("WMP")
	if err != nil {
		return fmt.Errorf("unable to list existing WMP files: %v", err)
	}
	defer dir.Close()

	wmpFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read WMP directory names: %v", err)
	}

	for _, wf := range wmpFiles {
		wmpFile, err := infFs.Open(wf)
		if err != nil {
			return fmt.Errorf("unable to open WMP file %q: %v", wf, err)
		}
		defer wmpFile.Close()

		wmp := p.NewWmp()
		stream := kaitai.NewStream(wmpFile)
		err = wmp.Read(stream, nil, wmp)
		if err != nil {
			return fmt.Errorf("unable to parse WMP file %q: %v", wf, err)
		}

		collection.LoadContextFromWorldMaps(wf, wmp)
	}

	if verbose {
		fmt.Println("done.")
	}

	return nil
}

func processAreas(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	if verbose {
		fmt.Print("extracting context from areas... ")
	}

	dir, err := infFs.Open("ARE")
	if err != nil {
		return fmt.Errorf("unable to list existing ARE files: %v", err)
	}
	defer dir.Close()

	areFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read ARE directory names: %v", err)
	}

	for _, af := range areFiles {
		areFile, err := infFs.Open(af)
		if err != nil {
			return fmt.Errorf("unable to open ARE file %q: %v", af, err)
		}
		defer areFile.Close()

		are := p.NewAre()
		stream := kaitai.NewStream(areFile)
		err = are.Read(stream, nil, are)
		if err != nil {
			return fmt.Errorf("unable to parse ARE file %q: %v", af, err)
		}

		collection.LoadContextFromAreas(af, are)
	}

	if verbose {
		fmt.Println("done.")
	}

	return nil
}
