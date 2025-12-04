package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// Global variables for command-line flags and internal state
var (
	width     int    // Output image width in pixels
	height    int    // Output image height in pixels
	quality   int    // JPEG quality (1-100)
	outputDir string // Output directory for resized images
	keepRatio bool   // Whether to keep aspect ratio when both width and height are set
	batchMode bool   // Whether to process all images in a directory
	workers   int    // Number of worker goroutines for batch processing
	verbose   bool   // Enable verbose output
	overwrite bool   // Whether to overwrite original files

	// Flags to track if dimensions were explicitly set by the user
	widthSet  bool
	heightSet bool
)

// setupConfig initializes the CLI configuration, flags, and validation
func setupConfig() {
	// Set up logger
	setupLogger()

	// Add version command
	rootCmd.AddCommand(createVersionCommand())

	// Register command-line flags and bind them to variables
	rootCmd.Flags().
		IntVarP(&width, "width", "w", 0, "Output width (pixels, 0=auto based on height)")
	rootCmd.Flags().
		IntVarP(&height, "height", "", 0, "Output height (pixels, 0=auto based on width)")
	rootCmd.Flags().IntVarP(&quality, "quality", "q", 95, "JPEG quality (1-100)")
	rootCmd.Flags().
		StringVarP(&outputDir, "output", "o", "", "Output directory (default: same as input)")
	rootCmd.Flags().
		BoolVarP(&keepRatio, "keep-ratio", "k", false, "Keep aspect ratio when both width and height are specified")
	rootCmd.Flags().
		BoolVarP(&batchMode, "batch", "b", false, "Batch process all images in directory")
	rootCmd.Flags().
		IntVarP(&workers, "workers", "", 4, "Number of worker goroutines for batch processing")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().
		BoolVar(&overwrite, "overwrite", false, "Overwrite original files instead of creating new ones")

	// PreRun: Validate and set up parameters before running the main command
	rootCmd.PreRun = validateConfig
}

// validateConfig validates the configuration parameters
func validateConfig(cmd *cobra.Command, args []string) {
	// Check if dimensions were explicitly set by the user
	widthSet = cmd.Flags().Changed("width")
	heightSet = cmd.Flags().Changed("height")

	// If neither width nor height is set, use default values
	if !widthSet && !heightSet {
		width = 800
		height = 600
		widthSet = true
		heightSet = true
	}

	// Validate input parameters
	if width < 0 || height < 0 {
		slog.Error("Width and height must be positive numbers")
		os.Exit(1)
	}
	if width == 0 && height == 0 {
		slog.Error("At least one of width or height must be specified")
		os.Exit(1)
	}
	if quality < 1 || quality > 100 {
		slog.Error("Quality must be between 1 and 100")
		os.Exit(1)
	}

	// Validate overwrite and output flags combination
	if overwrite && outputDir != "" {
		slog.Error(
			"Cannot use --overwrite with --output: --overwrite replaces original files in place",
		)
		os.Exit(1)
	}
}
