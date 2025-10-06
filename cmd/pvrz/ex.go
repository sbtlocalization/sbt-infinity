// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package pvrz

import (
	"fmt"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/fs"
	"github.com/spf13/cobra"
)

func NewExCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [PVRZ-file...]",
		Aliases: []string{"ex"},
		Short:   "Export PVRZ as image files",
		Long:    `Export contents of PVRZ archive as image files.`,
		Args:    cobra.MinimumNArgs(0),
		RunE:    runEx,
	}

	cmd.Flags().StringP("output", "o", "", "Output directory")
	cmd.Flags().StringP("format", "f", "png", "Image format: png, jpg")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().StringSliceP("exclude", "x", []string{}, "Exclude specific PVRZ files (e.g., mos2000.PVRZ)")

	return cmd
}

func runEx(cmd *cobra.Command, args []string) error {
	outputDir, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")
	format, _ := cmd.Flags().GetString("format")
	excludeFiles, _ := cmd.Flags().GetStringSlice("exclude")

	keyPath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		return err
	}

	format = strings.ToLower(format)

	var pvrzFiles []string
	if len(args) > 0 {
		pvrzFiles = args
		for i, pf := range pvrzFiles {
			if !strings.HasSuffix(strings.ToLower(pf), ".pvrz") {
				pvrzFiles[i] = pf + ".PVRZ"
			}
		}
	}

	infFs := fs.NewInfinityFs(keyPath, fs.FileType_PVRZ)

	if len(pvrzFiles) == 0 {
		dir, err := infFs.Open("PVRZ")
		if err != nil {
			return fmt.Errorf("unable to list existing PVRZ files: %v", err)
		}
		defer dir.Close()
		pvrzFiles, err = dir.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("unable to read PVRZ directory names: %v", err)
		}
	}

	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("unable to create output directory %s: %v", outputDir, err)
		}
	}

	excludeMap := make(map[string]bool)
	for _, ef := range excludeFiles {
		ef = strings.ToUpper(ef)
		if !strings.HasSuffix(ef, ".PVRZ") {
			ef = ef + ".PVRZ"
		}
		excludeMap[ef] = true
	}

	for _, pf := range pvrzFiles {
		if excludeMap[pf] {
			if verbose {
				fmt.Printf("skipping excluded file %s\n", pf)
			}
			continue
		}

		pvrzFile, err := infFs.Open(pf)
		if err != nil {
			fmt.Printf("warning: failed to open PVRZ file %s: %v\n", pf, err)
			continue
		}
		defer pvrzFile.Close()

		if verbose {
			fmt.Printf("processing PVRZ file %s...\n", pf)
		}

		// [TODO] @GooRoo: implement PVRZ loading
		image := draw.Image(nil)

		if image != nil {
			outputPath := filepath.Join(outputDir, strings.TrimSuffix(strings.ToLower(pf), ".pvrz")+"."+format)

			file, err := os.Create(outputPath)
			if err != nil {
				fmt.Printf("warning: failed to create output file %s: %v\n", outputPath, err)
				continue
			}
			defer file.Close()

			var encodeErr error
			switch format {
			case "png":
				encodeErr = png.Encode(file, image)
			case "jpg", "jpeg":
				encodeErr = jpeg.Encode(file, image, &jpeg.Options{Quality: 90})
			default:
				encodeErr = fmt.Errorf("unsupported image format: %s", format)
			}

			if encodeErr != nil {
				fmt.Printf("warning: failed to encode image %s: %v\n", outputPath, encodeErr)
				continue
			}

			if verbose {
				fmt.Printf("saved image: %s\n", outputPath)
			}
		} else {
			fmt.Printf("warning: no image data found in PVRZ file %s\n", pf)
		}
	}

	return nil
}
