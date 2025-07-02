// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package cmd

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/sbtlocalization/infinity-tools/parser"
	"github.com/spf13/cobra"
)

// runUpdateBam handles the update-bam command execution
func runUpdateBam(cmd *cobra.Command, args []string) {
	tomlFilePath := args[0]
	fmt.Printf("update-bam called with toml file: %s\n", tomlFilePath)

	// Load configuration from file first to get potential output directory
	config, err := LoadConfig(tomlFilePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully loaded configuration from: %s\n", tomlFilePath)
	// config.PrintConfig()

	// Get flags
	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		if config.UpdateDir != "" {
			outputDir = config.UpdateDir
		} else {
			outputDir = "override" // Default output directory
		}
	}
	
	enableTrace, _ := cmd.Flags().GetBool("trace")

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Create trace directory for debugging if trace is enabled
	var traceDir string
	if enableTrace {
		traceDir = "trace"
		if err := os.MkdirAll(traceDir, 0755); err != nil {
			fmt.Printf("Error creating trace directory: %v\n", err)
			enableTrace = false // Disable trace if directory creation fails
		}
		fmt.Printf("Trace mode enabled. Trace images will be saved to: %s\n", traceDir)
	}


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

	fmt.Printf("BAM file parsed successfully!\n")
	fmt.Printf("Frame count: %d\n", bam.Header.FrameCount)

	// Get frame entries and data blocks
	frameEntries, err := bam.FrameEntries()
	if err != nil {
		fmt.Printf("Error getting frame entries: %v\n", err)
		return
	}

	dataBlocks, err := bam.DataBlocks()
	if err != nil {
		fmt.Printf("Error getting data blocks: %v\n", err)
		return
	}

	// Load original atlas images (MOS files)
	atlasImages := make(map[uint32]*image.RGBA)
	for pageNum, mosPath := range config.InputMos {
		img, err := loadImageAsRGBA(mosPath)
		if err != nil {
			fmt.Printf("Error loading atlas image %s: %v\n", mosPath, err)
			return
		}
		atlasImages[pageNum] = img
		fmt.Printf("Loaded atlas image %d: %s\n", pageNum, mosPath)
	}

	// Load new frame images
	newFrameImages := make(map[uint32]image.Image)
	for frameIndex, framePath := range config.NewFrames {
		img, err := loadImage(framePath)
		if err != nil {
			fmt.Printf("Warning: Could not load new frame %d from %s: %v\n", frameIndex, framePath, err)
			continue
		}
		newFrameImages[frameIndex] = img
		fmt.Printf("Loaded new frame %d: %s\n", frameIndex, framePath)
	}

	// Process each frame and update atlas images
	fmt.Printf("\nProcessing frames...\n")
	for frameIndex, frame := range frameEntries.Entry {
		// Check if we have a new frame for this index
		if newFrameImg, hasNewFrame := newFrameImages[uint32(frameIndex)]; hasNewFrame {
			fmt.Printf("Updating frame %d with new image\n", frameIndex)
			

			
			err := updateAtlasWithNewFrame(frame, dataBlocks.Block, atlasImages, newFrameImg, frameIndex, enableTrace)
			if err != nil {
				fmt.Printf("Error updating frame %d: %v\n", frameIndex, err)
			}
		}
		// If no new frame, the original blocks remain in the atlas (no action needed)
	}

	// Save updated atlas images
	fmt.Printf("\nSaving updated atlas images to %s...\n", outputDir)
	for pageNum, atlasImg := range atlasImages {
		// By this moment we know for sure that the element in the map exists
		originalFilename := filepath.Base(config.InputMos[pageNum])

		outputPath := filepath.Join(outputDir, originalFilename)
		err := saveImageAsPNG(atlasImg, outputPath)
		if err != nil {
			fmt.Printf("Error saving atlas image %d: %v\n", pageNum, err)
			continue
		}
		fmt.Printf("Saved atlas image: %s\n", outputPath)
	}

	fmt.Printf("BAM update completed!\n")
}

// loadImageAsRGBA loads an image from file and converts it to *image.RGBA
func loadImageAsRGBA(path string) (*image.RGBA, error) {
	img, err := loadImage(path)
	if err != nil {
		return nil, err
	}

	// Convert to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba, nil
}

// updateAtlasWithNewFrame updates atlas images with data from a new frame
func updateAtlasWithNewFrame(frame *parser.Bam_FrameEntry, dataBlocks []*parser.Bam_DataBlock, atlasImages map[uint32]*image.RGBA, newFrameImg image.Image, frameIndex int, enableTrace bool) error {
	// Process each data block for this frame
	startIdx := frame.DataBlocksStartIndex
	endIdx := startIdx + frame.DataBlocksCount

	blockCounter := 0
	for blockIdx := startIdx; blockIdx < endIdx; blockIdx++ {
		if int(blockIdx) >= len(dataBlocks) {
			fmt.Println("Warning: Data block index out of range for frame", frameIndex)
			continue
		}

		block := dataBlocks[blockIdx]

		// Get the target atlas image
		atlasImg, exists := atlasImages[block.PrvzPage]
		if !exists {
			fmt.Printf("Warning: Atlas image for page %d not found for frame %d\n", block.PrvzPage, frameIndex)
			continue
		}

		// Define extended destination rectangle in the atlas (where to copy TO)
		// Extend by 1 pixel in each direction
		extendedDstRect := image.Rect(
			int(block.SourceX)-1,
			int(block.SourceY)-1,
			int(block.SourceX+block.Width)+1,
			int(block.SourceY+block.Height)+1,
		)

		// Create an extended image that handles out-of-bounds areas as transparent
		extendedWidth := int(block.Width) + 2
		extendedHeight := int(block.Height) + 2
		extendedImg := image.NewRGBA(image.Rect(0, 0, extendedWidth, extendedHeight))

		// Fill the extended image with data from newFrameImg, using transparent for out-of-bounds
		newFrameBounds := newFrameImg.Bounds()
		for y := 0; y < extendedHeight; y++ {
			for x := 0; x < extendedWidth; x++ {
				sourceX := int(block.TargetX) - 1 + x
				sourceY := int(block.TargetY) - 1 + y
				
				if sourceX >= newFrameBounds.Min.X && sourceX < newFrameBounds.Max.X &&
				   sourceY >= newFrameBounds.Min.Y && sourceY < newFrameBounds.Max.Y {
					// Source pixel is within bounds, copy it
					extendedImg.Set(x, y, newFrameImg.At(sourceX, sourceY))
				} else {
					// Source pixel is out of bounds, use transparent
					extendedImg.Set(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
		}

		// TRACE: Save individual block as separate image
		if enableTrace {
			blockPath := fmt.Sprintf("trace/Frame%d_Block%03d.png", frameIndex, blockCounter)
			if err := saveImageAsPNG(extendedImg, blockPath); err != nil {
				fmt.Printf("Warning: Could not save block trace %s: %v\n", blockPath, err)
			}
		}

		// TRACE: Print block details
		if enableTrace { 
			fmt.Printf("  Block %d: src(%d,%d %dx%d) -> dst(%d,%d %dx%d) page=%d\n", 
				blockCounter, 
				block.TargetX, block.TargetY, block.Width, block.Height,
				block.SourceX, block.SourceY, block.Width, block.Height,
				block.PrvzPage)
		}

		// Copy the extended image data to atlas
		draw.Draw(atlasImg, extendedDstRect, extendedImg, image.Point{0, 0}, draw.Src)

		// TRACE: Save intermediate atlas state
		if enableTrace {
			atlasPath := fmt.Sprintf("trace/Frame%d_Atlas_Page%d_After%d.png", frameIndex, block.PrvzPage, blockCounter+1)
			if err := saveImageAsPNG(atlasImg, atlasPath); err != nil {
				fmt.Printf("Warning: Could not save atlas trace %s: %v\n", atlasPath, err)
			}
		}

		blockCounter++
	}

	return nil
}

// updateBamCmd represents the updateBam command
var updateBamCmd = &cobra.Command{
	Use:   "update-bam <toml-file>",
	Short: "Update BAM atlas images with new frame images",
	Long: `Update BAM atlas images with new frame images specified in the configuration.

This command reads the BAM file and original atlas images, then replaces specific
frame regions in the atlas with new frame images specified in the NewFrames section
of the configuration file.

For each frame that has a corresponding entry in NewFrames, the blocks that make up
that frame will be updated in the atlas images with data from the new frame image.
Frames without new replacements will remain unchanged in the atlas.

The updated atlas images are saved to the output directory with the same filenames
as the original atlas images.

Use the --trace flag to enable debug mode, which saves intermediate images to a 
'trace' directory showing the step-by-step atlas modification process.`,
	Args: cobra.ExactArgs(1),
	Run:  runUpdateBam,
}

func init() {
	rootCmd.AddCommand(updateBamCmd)

	// Add flags
	updateBamCmd.Flags().StringP("output", "o", "", "Output directory for updated MOS files (default: from config file or 'override')")
	updateBamCmd.Flags().BoolP("trace", "t", false, "Enable trace mode to generate debug images")
}
