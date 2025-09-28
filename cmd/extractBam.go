// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/spf13/cobra"
)

// runExtractBam handles the extract-bam command execution
func runExtractBam(cmd *cobra.Command, args []string) {
	tomlFilePath := args[0]
	fmt.Printf("extract-bam called with toml file: %s\n", tomlFilePath)

	// Load configuration from file first to get potential output directory
	config, err := LoadConfig(tomlFilePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully loaded configuration from: %s\n", tomlFilePath)

	// Get output directory flag, with fallback to config, then to default
	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		if config.ExtractDir != "" {
			outputDir = config.ExtractDir
		} else {
			outputDir = "." // Current directory as default
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Print all configuration details
	config.PrintConfig()

	// Parse the BAM file
	fmt.Printf("\nParsing BAM file: %s\n", config.BamPath)

	// Open the BAM file
	file, err := os.Open(config.BamPath)
	if err != nil {
		fmt.Printf("Error opening BAM file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a Kaitai stream from the file
	stream := kaitai.NewStream(file)

	// Parse the BAM file
	bam := parser.NewBam()
	err = bam.Read(stream, nil, bam)
	if err != nil {
		fmt.Printf("Error parsing BAM file: %v\n", err)
		return
	}

	// Display BAM file information
	fmt.Printf("BAM file parsed successfully!\n")
	fmt.Printf("Frame count: %d\n", bam.Header.FrameCount)
	fmt.Printf("Cycle count: %d\n", bam.Header.CycleCount)
	fmt.Printf("Data blocks count: %d\n", bam.Header.DataBlocksCount)

	// Get frame entries
	frameEntries, err := bam.FrameEntries()
	if err != nil {
		fmt.Printf("Error getting frame entries: %v\n", err)
		return
	}

	// Get data blocks
	dataBlocks, err := bam.DataBlocks()
	if err != nil {
		fmt.Printf("Error getting data blocks: %v\n", err)
		return
	}

	// Load atlas images (MOS files) into memory
	atlasImages := make(map[uint32]image.Image)
	for pageNum, mosPath := range config.InputMos {
		img, err := loadImage(mosPath)
		if err != nil {
			fmt.Printf("Warning: Could not load atlas image %s: %v\n", mosPath, err)
			continue
		}
		atlasImages[pageNum] = img
		fmt.Printf("Loaded atlas image %d: %s\n", pageNum, mosPath)
	}

	// Export each frame as PNG
	fmt.Printf("\nExporting frames to %s...\n", outputDir)
	for i, frame := range frameEntries.Entry {
		err := exportFrameAsPNG(frame, dataBlocks.Block, atlasImages, outputDir, i)
		if err != nil {
			fmt.Printf("Error exporting frame %d: %v\n", i, err)
			continue
		}
		if i%10 == 0 || i == len(frameEntries.Entry)-1 {
			fmt.Printf("Exported frame %d/%d\n", i+1, len(frameEntries.Entry))
		}
	}

	fmt.Printf("Frame extraction completed!\n")
}

// loadImage loads an image from file (supports PNG format)
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// saveImageAsPNG saves an image as PNG file
func saveImageAsPNG(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// exportFrameAsPNG exports a single frame as a PNG file
func exportFrameAsPNG(frame *parser.Bam_FrameEntry, dataBlocks []*parser.Bam_DataBlock, atlasImages map[uint32]image.Image, outputDir string, frameIndex int) error {
	// Create a new RGBA image with the frame's dimensions
	img := image.NewRGBA(image.Rect(0, 0, int(frame.Width), int(frame.Height)))

	// Process each data block for this frame
	startIdx := frame.DataBlocksStartIndex
	endIdx := startIdx + frame.DataBlocksCount

	for blockIdx := startIdx; blockIdx < endIdx; blockIdx++ {
		if int(blockIdx) >= len(dataBlocks) {
			continue
		}

		block := dataBlocks[blockIdx]

		// Get the source atlas image
		atlasImg, exists := atlasImages[block.PrvzPage]
		if !exists {
			fmt.Printf("Warning: Atlas image for page %d not found for frame %d\n", block.PrvzPage, frameIndex)
			continue
		}

		// Define source rectangle in the atlas
		srcRect := image.Rect(
			int(block.SourceX),
			int(block.SourceY),
			int(block.SourceX+block.Width),
			int(block.SourceY+block.Height),
		)

		// Define destination rectangle in the frame
		dstRect := image.Rect(
			int(block.TargetX),
			int(block.TargetY),
			int(block.TargetX+block.Width),
			int(block.TargetY+block.Height),
		)

		// Copy the image data from atlas to frame
		draw.Draw(img, dstRect, atlasImg, srcRect.Min, draw.Src)
	}

	// Save the frame as PNG
	filename := fmt.Sprintf("Frame%d.png", frameIndex)
	outputPath := filepath.Join(outputDir, filename)

	err := saveImageAsPNG(img, outputPath)
	if err != nil {
		return fmt.Errorf("failed to encode PNG for frame %d: %v", frameIndex, err)
	}

	return nil
}

// extractBamCmd represents the extract-bam command
var extractBamCmd = &cobra.Command{
	Use:   "extract-bam <toml-file>",
	Short: "Extract BAM file frames as PNG images",
	Long: `Extract BAM file frames as individual PNG images using the configuration specified 
in the provided toml file.

Each frame will be exported as FrameXXX.png where XXX is the frame index.
The frames are composed from image atlas files (MOS) specified in the configuration.

The toml-file argument is required and should point to a valid configuration file
that contains the necessary parameters for BAM extraction.`,
	Args: cobra.ExactArgs(1),
	Run:  runExtractBam,
}

func init() {
	rootCmd.AddCommand(extractBamCmd)

	// Add output directory flag
	extractBamCmd.Flags().StringP("output", "o", "", "Output directory for PNG files (default: from config file or current directory)")
}
