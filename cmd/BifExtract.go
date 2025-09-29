// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

// runExtractBif handles the `bif ex` command execution
func runExtractBif(cmd *cobra.Command, args []string) {
	initLogF(cmd)

	keyFilePath := args[0]
	printLogF("bif ex called with key file: %s\n", keyFilePath)

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

	keyFile, realFile := parseKeyFile(keyFilePath)
	// Close file on this level to avoid keep interface opened
	// TODO: use NewInfinityFs ?
	defer realFile.Close()

	// Display KEY file information
	printLogF("KEY file parsed successfully!\n")
	printLogF("BIF files count: %d\n", keyFile.NumBiffEntries)
	printLogF("Packed resource count: %d\n", keyFile.NumResEntries)

	biffs, _ := keyFile.BiffEntries()
	for key, value := range biffs {
		filePath, _ := value.FilePath()
		printLogF("Biff index %d, for file %s\n", key, filePath)
	}

	resEntries, _ := keyFile.ResEntries()
	for key, value := range resEntries {
		if value.Type == parser.Key_ResType__Dlg {
			si := value.Locator.BiffFileIndex
			nti := value.Locator.FileIndex
			printLogF("Found dlg res: index %d, name %s, source_index %d, ntls_index %d\n", key, value.Name, si, nti)

			targetFilePath, _ := biffs[si].FilePath()
			printLogF("Going to unpack %s\n", targetFilePath)

			p := filepath.Dir(keyFilePath)
			bFile, err := os.Open(filepath.Join(p, targetFilePath))
			if err != nil {
				log.Fatalf("Error opening BIFF file: %v\n", err)
				return
			}
			defer bFile.Close()

			// Create a Kaitai stream from the file
			stream := kaitai.NewStream(bFile)

			// Parse the BIFF file
			bifFile := parser.NewBif()
			err = bifFile.Read(stream, nil, bifFile)
			if err != nil {
				log.Fatalf("Error parsing Bif file: %v\n", err)
				return
			}

			printLogF("BIF file parsed successfully!\n")
			printLogF("BIF.NumFileEntries: %d\n", bifFile.NumFileEntries)
			printLogF("BIF.NumTilesetEntries: %d\n", bifFile.NumTilesetEntries)

			if nti < uint64(bifFile.NumFileEntries) {
				fileEntries, _ := bifFile.FileEntries()
				currentEntry := fileEntries[nti]
				targetType := currentEntry.ResType
				targetBlob, _ := currentEntry.Data()
				printLogF("Process entry with locator %d, type %d, extension %s\n", currentEntry.Locator.FileIndex, currentEntry.ResType, targetType)

				outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", value.Name, targetType))

				err := saveBlobToFile(targetBlob, outputPath)
				if err != nil {
					log.Fatalf("Error saving %s file: %v\n", outputPath, err)
					return
				}

			}
		}
	}
}

// saveBlobToFile saves extracted data into path
func saveBlobToFile(blob []byte, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(blob)

	return err
}
