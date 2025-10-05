// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	p "github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/sbtlocalization/sbt-infinity/text"
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

	cmd.Flags().StringP("output", "o", "dialog.xlsx", "Output xlsx file path")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().String("dlg-base-url", "", "Base URL for dialog references (overrides config)")

	return cmd
}

func runEx(cmd *cobra.Command, args []string) error {
	tlkPath, _ := cmd.Flags().GetString("tlk")
	lang, _ := cmd.Flags().GetString("lang")
	feminine, _ := cmd.Flags().GetBool("feminine")
	verbose, _ := cmd.Flags().GetBool("verbose")
	baseUrl, _ := config.ResolveDialogBaseUrl(cmd)

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

	tlkFile, err := p.ReadTlkFile(tlkFs, tlkPath)
	if err != nil {
		return err
	}
	collection := text.NewTextCollection(tlkFile.Tlk)
	tlkFile.Close()

	infFs := fs.NewInfinityFs(keyPath, fs.FileType_DLG, fs.FileType_CRE)

	// process dialogs
	err = processDialogs(collection, infFs, baseUrl, verbose)
	if err != nil {
		fmt.Println("warning: unable to process dialogs:", err)
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

	for _, df := range dialogFiles {
		dc, err := dlgBuilder.LoadAllDialogs("", df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		collection.LoadContextFromDialogs(baseUrl, dc)
	}

	return nil
}
