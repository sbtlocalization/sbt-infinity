// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/sbtlocalization/infinity-tools/parser"
	"github.com/spf13/cobra"
)

// runUnpackBif handles the unpack-bif command execution
func runUnpackBif(cmd *cobra.Command, args []string) {
	keyFilePath := args[0]
	fmt.Printf("unpack-bif called with key file: %s\n", keyFilePath)

	// Get output directory flag, with fallback to config, then to default
	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		outputDir = "." // Current directory as default
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Open the KEY file
	file, err := os.Open(keyFilePath)
	if err != nil {
		fmt.Printf("Error opening KEY file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a Kaitai stream from the file
	stream := kaitai.NewStream(file)

	// Parse the KEY file
	keyFile := parser.NewKey()
	err = keyFile.Read(stream, nil, keyFile)
	if err != nil {
		fmt.Printf("Error parsing KEY file: %v\n", err)
		return
	}

	// Display KEY file information
	fmt.Printf("KEY file parsed successfully!\n")
	fmt.Printf("BIF files count: %d\n", keyFile.NumBiffEntries)
	fmt.Printf("Packed resource count: %d\n", keyFile.NumResEntries)

	biffs, _ := keyFile.BiffEntries()
	for key, value := range biffs {
		fileName, _ := value.FileNameExt()
		fmt.Printf("Biff index %d, for file %s\n", key, fileName.Name)
	}

	resEntries, _ := keyFile.ResEntries()
	for key, value := range resEntries {
		if value.Type == 1011 {
			si, _ := value.SourceIndex()
			nti, _ := value.NonTilesetIndex()
			fmt.Printf("Found dlg res: index %d, name %s, source_index %d, ntls_index %d\n", key, value.Name, si, nti)

			targetFilePath, _ := biffs[si].FileNameExt()
			fmt.Printf("Going to unpack %s\n", targetFilePath.Name)

			p := filepath.Dir(keyFilePath)
			bFile, err := os.Open(filepath.Join(p, targetFilePath.Name))
			if err != nil {
				fmt.Printf("Error opening BIFF file: %v\n", err)
				return
			}
			defer bFile.Close()

			// Create a Kaitai stream from the file
			stream := kaitai.NewStream(bFile)

			// Parse the BIFF file
			bifFile := parser.NewBif()
			err = bifFile.Read(stream, nil, bifFile)
			if err != nil {
				fmt.Printf("Error parsing Bif file: %v\n", err)
				return
			}

			fmt.Printf("BIF file parsed successfully!\n")
			fmt.Printf("BIF.NumFileEntries: %d\n", bifFile.NumFileEntries)
			fmt.Printf("BIF.NumTilesetEntries: %d\n", bifFile.NumTilesetEntries)

			if nti >= 0 && nti < int(bifFile.NumFileEntries) {
				fileEntries, _ := bifFile.FileEntries()
				currentEntry := fileEntries[nti]
				targetExtension, _ := currentEntry.FileExtension()
				targetExtension = strings.TrimSpace(targetExtension)
				targetBlob, _ := currentEntry.ResBlob()
				fmt.Printf("Process entry with locator %d, type %d, extension %s\n", currentEntry.ResLocator, currentEntry.ResType, targetExtension)

				outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", value.Name, targetExtension))

				err := saveBlobToFile(targetBlob, outputPath)
				if err != nil {
					fmt.Printf("Error saving %s file: %v\n", outputPath, err)
					return
				}

			} else {
				fmt.Printf("Tilesets not yet supported\n")
			}
		}
	}

	//TODO: implement filters later, now unpack all stuff

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

// unpackBifCmd represents the unpack-bif command
var unpackBifCmd = &cobra.Command{
	Use:   "unpack-bif <path to chitin.key>",
	Short: "unpack BIF files into resources",
	Long: `unpack BIF files into set of resources.
	Structure of resources is read from chitin.key,
	so all related .bif files picked automatically.

	Additional filter may be passed to unpack only specific resources
	`,
	Args: cobra.ExactArgs(1),
	Run:  runUnpackBif,
}

func init() {
	rootCmd.AddCommand(unpackBifCmd)

	// Add output directory flag
	unpackBifCmd.Flags().StringP("output", "o", "", "Output directory for resource files (default: current directory)")
	// TODO:
	unpackBifCmd.Flags().StringP("type", "t", "", "Type filter. Take type number from https://gibberlings3.github.io/iesdp/file_formats/general.htm")
	// TODO:
	unpackBifCmd.Flags().StringP("filter", "f", "", "Regexp for filtering")
}
