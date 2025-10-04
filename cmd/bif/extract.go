// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package bif

import (
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
		Use:     "extract path-to-chitin.key [-o output-dir][-t resource-type][-f regex-filter]",
		Aliases: []string{"ex"},
		Short:   "Extract game engine resources from BIF files",
		Long: `Extract game engine resources from BIF files.
	Structure of resources is read from chitin.key,
	so all related .bif files picked automatically.

	Additional filter may be passed to unpack only specific resources
	`,
		Run:  runExtractBif,
		Args: cobra.MinimumNArgs(0),
	}

	cmd.Flags().StringSliceP("type", "t", nil, "Resourse type filter. Comma separated integers (dec or hex) or extension names (like DLG). Take type number from https://gibberlings3.github.io/iesdp/file_formats/general.htm")
	cmd.Flags().StringP("filter", "f", "", "Regex for resourse name filtering")

	cmd.Flags().StringP("output", "o", "", "Output directory for resource files (default: current directory)")

	return cmd
}

// runExtractBif handles the `bif ex` command execution
func runExtractBif(cmd *cobra.Command, args []string) {
	typeRawInput, _ := cmd.Flags().GetStringSlice("type")
	filterRawInput, _ := cmd.Flags().GetString("filter")
	outputDir, _ := cmd.Flags().GetString("output")

	keyFilePath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		log.Fatalf("Error with .key path: %v\n", err)
	}

	// Get output directory flag, with fallback to config, then to default
	if outputDir == "" {
		outputDir = "." // Current directory as default
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v\n", err)
		return
	}

	contentFilter := getContentFilter(filterRawInput)

	resFs := fs.NewInfinityFs(keyFilePath, getFileTypeFilter(typeRawInput)...)

	for _, v := range resFs.ListResourses(contentFilter) {
		fullName := v.FullName
		file, err := resFs.Open(fullName)
		if err != nil {
			log.Fatalf("failed to extract file %s: %v", fullName, err)
		}
		defer file.Close()

		outputPath := filepath.Join(outputDir, fullName)

		err = saveFileToFile(file, outputPath)
		if err != nil {
			log.Fatalf("Error saving %s file: %v\n", outputPath, err)
			return
		}
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
