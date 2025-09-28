// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: @definitelythehuman
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

const (
	Bif_Flag_Type_Filter    string = "type"
	Bif_Flag_Content_Filter string = "filter"
	Bif_Flag_Verbose        string = "verbose"
	Bif_Flag_JSON           string = "json"
	Bif_Flag_Output_Dir     string = "output"
)

var bif_log_level_verbose = false

// mainBifCmd represents the bif command which has subcommands
var mainBifCmd = &cobra.Command{
	Use:   "bif ls|ex path-to-chitin.key [-o output-dir][-t resource-type][-f regex-filter]",
	Short: "unpack or list BIF files into resources",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Error: must also specify list(ls) or extract(ex) command")
	},
}

var listBifCmd = &cobra.Command{
	Use:   "ls path-to-chitin.key [-j=json][-t resource-type][-f regex-filter]",
	Short: "list all BIFF files and resources attached to KEY file",
	Long: `list all BIFF files and resources attached to KEY file.

	Additional filter may be passed to unpack only specific resources
	`,
	Args: cobra.ExactArgs(1),
	Run:  runListBif,
}

var extractBifCmd = &cobra.Command{
	Use:   "ex path-to-chitin.key [-o output-dir][-t resource-type][-f regex-filter]",
	Short: "unpack(extracts) BIF files into resources",
	Long: `unpack(extracts) BIF files into set of resources.
	Structure of resources is read from chitin.key,
	so all related .bif files picked automatically.

	Additional filter may be passed to unpack only specific resources
	`,
	Args: cobra.ExactArgs(1),
	Run:  runExtractBif,
}

func init() {
	rootCmd.AddCommand(mainBifCmd)
	mainBifCmd.AddCommand(listBifCmd)
	mainBifCmd.AddCommand(extractBifCmd)

	mainBifCmd.PersistentFlags().StringP(Bif_Flag_Type_Filter, "t", "", "Resourse type filter. Comma separated integers (dec or hex) or extension names (like DLG). Take type number from https://gibberlings3.github.io/iesdp/file_formats/general.htm")
	mainBifCmd.PersistentFlags().StringP(Bif_Flag_Content_Filter, "f", "", "Regex for resourse name filtering")
	mainBifCmd.PersistentFlags().BoolP(Bif_Flag_Verbose, "v", false, "Output more debug information")

	listBifCmd.Flags().BoolP(Bif_Flag_JSON, "j", false, "Decorate output as JSON")

	extractBifCmd.Flags().StringP(Bif_Flag_Output_Dir, "o", "", "Output directory for resource files (default: current directory)")
}

func initLogF(cmd *cobra.Command) {
	isVerbose, _ := cmd.Flags().GetBool(Bif_Flag_Verbose)
	bif_log_level_verbose = isVerbose
}

func printLogF(format string, a ...any) {
	if bif_log_level_verbose {
		fmt.Printf(format, a...)
	}
}

// Warning! API user MUST close returned file
func parseKeyFile(filepath string) (*parser.Key, *os.File) {
	// Open the KEY file
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Error opening KEY file: %v\n", err)
		return nil, nil
	}

	// Create a Kaitai stream from the file
	stream := kaitai.NewStream(file)

	// Parse the KEY file
	keyFile := parser.NewKey()
	err = keyFile.Read(stream, nil, keyFile)
	if err != nil {
		log.Fatalf("Error parsing KEY file: %v\n", err)
		file.Close()
		return nil, nil
	}

	return keyFile, file
}

// Parses argument like `-t 1011,0x409,1022,DLG,bmp` into list of Key_ResType
// TODO: remove duplicate types
func getTypeFilter(cmd *cobra.Command) (filter []parser.Key_ResType) {
	rawInput, _ := cmd.Flags().GetString(Bif_Flag_Type_Filter)
	if len(rawInput) == 0 {
		return filter
	}

	tokens := strings.Split(rawInput, ",")
	printLogF("Extracted tokens %v\n", tokens)

	filter = make([]parser.Key_ResType, len(tokens))
	if len(tokens) == 0 {
		return filter
	}

	for key, value := range tokens {
		if fType := fs.FileTypeFromExtension(value); fType.IsValid() {
			filter[key] = fType.ToParserType()
		} else if parsed, err := strconv.ParseInt(value, 0, 32); err == nil {
			resType := fs.FileType(parsed)
			if resType.IsValid() {
				filter[key] = resType.ToParserType()
			} else {
				log.Fatalf("Value 0x%x (%d) does not match known type\n", parsed, parsed)
			}
		} else {
			log.Fatalf("Value %s does not match known type\n", value)
		}
	}

	return filter
}

func getContentFilter(cmd *cobra.Command) *regexp.Regexp {
	rawInput, _ := cmd.Flags().GetString(Bif_Flag_Content_Filter)
	if len(rawInput) == 0 {
		return nil
	}

	compiled, err := regexp.Compile(rawInput)
	if err != nil {
		log.Fatalf("Value %s is not Regexp\n", rawInput)
		return nil
	}

	return compiled
}
