// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package bif

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extract [-o output-dir] [-t type][flags]...",
		Aliases: []string{"ex"},
		Short:   "Extract game engine resources from BIF files",
		Long: `Extract game engine resources from BIF files.
Structure of resources is read from chitin.key,
so all related .bif files picked automatically.

Additional filter may be passed to unpack only specific resources.`,
		Example: `  Extract resourses with 'ttb' in name into 'tmp' folder from data/ITEMS.BIF file only:

      sbt-inf bif ex -k 'D:\Games\Baldur''s Gate - Enhanced Edition\chitin.key' -b items -f ttb* -o tmp

  Extract only LUA and DLG files into 'tmp' folder (the path to chitin.key is taken from sbt-inf.toml):
  
      sbt-inf bif ex -t LUA,DLG -o tmp

  Extract files into 'out' folder, each type in its own subfolder:

      sbt-inf bif ex -t WAV -t DLG -o out --folders`,
		Run:  runExtractBif,
		Args: cobra.MaximumNArgs(0),
	}

	cmd.Flags().StringP("output", "o", ".", "Output directory for resource files (default: current directory)")
	cmd.Flags().Bool("folders", false, "Create a separate folder for each type")

	cmd.MarkFlagDirname("output")

	return cmd
}

// runExtractBif handles the `bif ex` command execution
func runExtractBif(cmd *cobra.Command, args []string) {
	typeRawInput, _ := cmd.Flags().GetStringSlice("type")
	bifFilterRawInput, _ := cmd.Flags().GetString("bif-filter")
	filterRawInput, _ := cmd.Flags().GetString("filter")
	outputDir, _ := cmd.Flags().GetString("output")
	createFolders, _ := cmd.Flags().GetBool("folders")

	keyFilePath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		log.Fatalf("Error with .key path: %v\n", err)
	}

	resFs := fs.NewInfinityFs(keyFilePath,
		fs.WithTypeFilter(getFileTypeFilter(typeRawInput)...),
		fs.WithBifFilter(bifFilterRawInput),
		fs.WithContentFilter(filterRawInput),
	)

	for _, v := range resFs.ListResources() {
		fullName := v.FullName
		file, err := resFs.Open(fullName)
		if err != nil {
			log.Fatalf("failed to extract file %s: %v", fullName, err)
		}
		defer file.Close()

		// Create output directory if it doesn't exist
		currentOutputDir := outputDir
		if createFolders {
			currentOutputDir = filepath.Join(outputDir, v.Type.String())
		}
		if err := os.MkdirAll(currentOutputDir, 0755); err != nil {
			log.Fatalf("Error creating output directory: %v\n", err)
			return
		}

		outputPath := filepath.Join(currentOutputDir, fullName)

		err = saveFileToFile(file, outputPath)
		if err != nil {
			log.Fatalf("Error saving %s file: %v\n", outputPath, err)
			return
		}
		fmt.Printf("Extracted: %s\n", outputPath)
	}
}

func saveFileToFile(src io.Reader, path string) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, src)

	return err
}
