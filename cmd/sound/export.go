// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package sound

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	snd "github.com/sbtlocalization/sbt-infinity/sound"
	"github.com/spf13/cobra"
)

func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [-o output-dir] [--format wav|flac] [flags]...",
		Aliases: []string{"ex"},
		Short:   "Export game audio files as WAV or FLAC",
		Long: `Export audio resources from BIF files, converting WAVC/ACM format
to standard WAV or FLAC. Resources are read directly from the game
via chitin.key.`,
		Example: `  Export all sound files as WAV:

      sbt-inf sound ex -g pst -o out

  Export a specific sound as FLAC:

      sbt-inf sound ex -g pst -o out -f TTB011 --format flac`,
		Run:  runExportSound,
		Args: cobra.MaximumNArgs(0),
	}

	cmd.Flags().StringP("output", "o", ".", "Output directory for exported audio files")
	cmd.Flags().String("format", "wav", "Output format: wav or flac")
	cmd.Flags().BoolP("verbose", "v", false, "Print each file being exported")

	cmd.MarkFlagDirname("output")
	cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return cobra.FixedCompletions([]cobra.Completion{"wav", "flac"}, cobra.ShellCompDirectiveNoFileComp)(cmd, args, toComplete)
	})

	return cmd
}

func runExportSound(cmd *cobra.Command, args []string) {
	filterRawInput, _ := cmd.Flags().GetString("filter")
	bifFilterRawInput, _ := cmd.Flags().GetString("bif-filter")
	outputDir, _ := cmd.Flags().GetString("output")
	format, _ := cmd.Flags().GetString("format")
	verbose, _ := cmd.Flags().GetBool("verbose")

	format = strings.ToLower(format)
	if format != "wav" && format != "flac" {
		log.Fatalf("Unsupported format: %s (use wav or flac)\n", format)
	}

	keyFilePath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		log.Fatalf("Error with .key path: %v\n", err)
	}

	resFs := fs.NewInfinityFs(keyFilePath,
		fs.WithTypeFilter(fs.FileType_WAV),
		fs.WithBifFilter(bifFilterRawInput),
		fs.WithContentFilter(filterRawInput),
	)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v\n", err)
	}

	converted := 0
	skipped := 0
	failed := 0

	for _, v := range resFs.ListResources() {
		file, err := resFs.Open(v.FullName)
		if err != nil {
			log.Printf("Failed to open %s: %v\n", v.FullName, err)
			failed++
			continue
		}

		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			log.Printf("Failed to read %s: %v\n", v.FullName, err)
			failed++
			continue
		}

		var pcm []byte
		var channels, sampleRate, bitsPerSample int
		isRIFF := snd.IsRIFF(data)

		if isRIFF {
			pcm, channels, sampleRate, bitsPerSample, err = snd.DecodeRiff(data)
		} else if snd.IsWAVC(data) {
			pcm, channels, sampleRate, bitsPerSample, err = snd.DecodeWavc(data)
		} else if snd.IsACM(data) {
			pcm, channels, sampleRate, bitsPerSample, err = snd.DecodeAcm(data)
		} else {
			if verbose {
				fmt.Printf("Skipping %s (unknown format)\n", v.FullName)
			}
			skipped++
			continue
		}

		if err != nil {
			log.Printf("Failed to decode %s: %v\n", v.FullName, err)
			failed++
			continue
		}

		baseName := strings.TrimSuffix(v.FullName, filepath.Ext(v.FullName))
		outName := baseName + "." + format
		outPath := filepath.Join(outputDir, outName)

		outFile, err := os.Create(outPath)
		if err != nil {
			log.Printf("Failed to create %s: %v\n", outPath, err)
			failed++
			continue
		}

		if isRIFF && format == "wav" {
			_, err = outFile.Write(data)
		} else if format == "flac" {
			err = snd.WriteFlac(outFile, pcm, channels, sampleRate, bitsPerSample)
		} else {
			err = snd.WriteWav(outFile, pcm, channels, sampleRate, bitsPerSample)
		}
		outFile.Close()

		if err != nil {
			log.Printf("Failed to write %s: %v\n", outPath, err)
			os.Remove(outPath)
			failed++
			continue
		}

		if verbose {
			fmt.Printf("Exported %s -> %s\n", v.FullName, outName)
		}
		converted++
	}

	fmt.Printf("Done: %d converted", converted)
	if skipped > 0 {
		fmt.Printf(", %d skipped", skipped)
	}
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Println()
}
