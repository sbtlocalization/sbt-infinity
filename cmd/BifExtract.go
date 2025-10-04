// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

// runExtractBif handles the `bif ex` command execution
func runExtractBif(cmd *cobra.Command, args []string) {
	keyFilePath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		log.Fatalf("Error with .key path: %v\n", err)
	}

	// Get output directory flag, with fallback to config, then to default
	outputDir, _ := cmd.Flags().GetString(Bif_Flag_Output_Dir)
	if outputDir == "" {
		outputDir = "." // Current directory as default
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v\n", err)
		return
	}

	filterBifContent(cmd, keyFilePath, func(index int, name string, bifPath string, resType parser.Key_ResType) {
		resourseFound(outputDir, keyFilePath, index, name, bifPath, resType)
	})
}

func resourseFound(outputDir string, keyFilePath string, index int, name string, bifPath string, resType parser.Key_ResType) {
	fType := fs.FileTypeFromParserType(resType)
	resFs := fs.NewInfinityFs(keyFilePath, fType)

	fullName := name
	suffix := "." + fType.String()
	if !strings.HasSuffix(strings.ToUpper(fullName), suffix) {
		fullName = fullName + suffix
	}
	printLogF("Extract %s which has index %d and located in %s\n", fullName, index, bifPath)

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
	printLogF("Extracted %s\n", outputPath)
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
