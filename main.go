package main

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/spf13/cobra"
)

// Version is set at compile time
var Version = "dev"

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

	// Flags to track if dimensions were explicitly set by the user
	widthSet  bool
	heightSet bool
)

// Entry point of the application
func main() {
	// Set up slog logger based on verbose flag
	var handler slog.Handler
	if verbose {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})
	}
	slog.SetDefault(slog.New(handler))

	// Define the root Cobra command for the CLI
	rootCmd := &cobra.Command{
		Use:     "resize-tool [image-file-or-directory]",
		Short:   "A powerful image resizing tool",
		Version: Version,
		Long: `A command-line tool to resize images with various options.
Supports JPEG, PNG, GIF, TIFF, and BMP formats.
Can process single files or batch process directories.

By default, if only width or height is specified, the other dimension
will be calculated automatically to maintain aspect ratio.

Example usage:
		resize-tool input.jpg --width 800
		resize-tool input.jpg --height 600
		resize-tool images/ --batch --width 1024 --output resized/
`,
		Args: cobra.ExactArgs(1),
		Run:  processImages,
	}

	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Register command-line flags and bind them to variables
	rootCmd.Flags().IntVarP(&width, "width", "w", 0, "Output width (pixels, 0=auto based on height)")
	rootCmd.Flags().IntVarP(&height, "height", "", 0, "Output height (pixels, 0=auto based on width)")
	rootCmd.Flags().IntVarP(&quality, "quality", "q", 95, "JPEG quality (1-100)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: same as input)")
	rootCmd.Flags().BoolVarP(&keepRatio, "keep-ratio", "k", false, "Keep aspect ratio when both width and height are specified")
	rootCmd.Flags().BoolVarP(&batchMode, "batch", "b", false, "Batch process all images in directory")
	rootCmd.Flags().IntVarP(&workers, "workers", "", 4, "Number of worker goroutines for batch processing")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// PreRun: Validate and set up parameters before running the main command
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
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
	}

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute command", "err", err)
		os.Exit(1)
	}
}

/*
processImages determines whether the input path is a file or directory,
and processes it accordingly. If batch mode is enabled or the input is a directory,
it processes all images in the directory. Otherwise, it processes a single image file.
*/
func processImages(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Check if the input path exists
	info, err := os.Stat(inputPath)
	if os.IsNotExist(err) {
		slog.Error(fmt.Sprintf("Path does not exist: %s", inputPath))
		os.Exit(1)
	}

	if info.IsDir() || batchMode {
		// Process all images in the directory
		processBatch(inputPath)
	} else {
		// Process a single image file
		if err := resizeImage(inputPath); err != nil {
			slog.Error(fmt.Sprintf("Failed to process image: %v", err))
			os.Exit(1)
		}
	}
}

/*
processBatch processes all supported image files in the specified directory using a worker pool.
It collects image files, distributes them to worker goroutines, and prints a summary of results.
*/
func processBatch(dirPath string) {
	if verbose {
		fmt.Printf("Processing directory: %s\n", dirPath)
		fmt.Printf("Using %d workers\n", workers)
	}

	// Collect all image files in the directory
	imageFiles, err := collectImageFiles(dirPath)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to collect image files: %v", err))
		os.Exit(1)
	}

	if len(imageFiles) == 0 {
		fmt.Println("No image files found in directory")
		return
	}

	fmt.Printf("Found %d image files\n", len(imageFiles))

	// Create channels for jobs and results for the worker pool
	jobs := make(chan string, len(imageFiles))
	results := make(chan error, len(imageFiles))

	// Start worker goroutines to process images concurrently
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				err := resizeImage(filePath)
				results <- err
			}
		}()
	}

	// Send image file paths to the jobs channel
	go func() {
		defer close(jobs)
		for _, filePath := range imageFiles {
			jobs <- filePath
		}
	}()

	// Close the results channel after all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and count results from workers
	successCount := 0
	errorCount := 0
	for err := range results {
		if err != nil {
			if verbose {
				fmt.Printf("Error: %v\n", err)
			}
			errorCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("Batch processing completed: %d success, %d errors\n", successCount, errorCount)
}

/*
collectImageFiles recursively collects all supported image files from the given directory.
Returns a slice of file paths and any error encountered.
*/
func collectImageFiles(dirPath string) ([]string, error) {
	var imageFiles []string
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".tiff": true,
		".tif":  true,
		".bmp":  true,
	}

	// Walk through the directory and collect files with supported extensions
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExts[ext] {
				imageFiles = append(imageFiles, path)
			}
		}

		return nil
	})

	return imageFiles, err
}

/*
resizeImage resizes a single image file according to the specified parameters.
It preserves aspect ratio if required, saves the output, and prints information if verbose is enabled.
*/
func resizeImage(inputPath string) error {
	if verbose {
		fmt.Printf("Processing: %s\n", inputPath)
	}

	// Open and decode the input image file
	src, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %v", inputPath, err)
	}

	// Get original image dimensions
	originalBounds := src.Bounds()
	originalWidth := originalBounds.Max.X
	originalHeight := originalBounds.Max.Y

	if verbose {
		fmt.Printf("  Original size: %dx%d\n", originalWidth, originalHeight)
	}

	// Calculate target dimensions based on flags and original size
	targetWidth, targetHeight := calculateTargetSize(originalWidth, originalHeight)

	if verbose {
		fmt.Printf("  Target size: %dx%d\n", targetWidth, targetHeight)
	}

	var resized image.Image

	// Choose resizing method based on flags
	if (widthSet && heightSet && keepRatio) || (!widthSet && heightSet) || (widthSet && !heightSet) {
		// Keep aspect ratio
		switch {
		case widthSet && !heightSet:
			// Only width set, height is auto-calculated
			resized = imaging.Resize(src, targetWidth, 0, imaging.Lanczos)
		case !widthSet && heightSet:
			// Only height set, width is auto-calculated
			resized = imaging.Resize(src, 0, targetHeight, imaging.Lanczos)
		default:
			// Both set, but keep ratio (fit within bounds)
			resized = imaging.Fit(src, targetWidth, targetHeight, imaging.Lanczos)
		}
	} else {
		// Force resize to exact dimensions (may distort aspect ratio)
		resized = imaging.Resize(src, targetWidth, targetHeight, imaging.Lanczos)
	}

	// Get actual resized dimensions (used for output filename)
	actualBounds := resized.Bounds()
	actualWidth := actualBounds.Max.X
	actualHeight := actualBounds.Max.Y

	// Generate output file path
	outputPath := generateOutputPath(inputPath, outputDir, actualWidth, actualHeight)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Save the resized image in the appropriate format
	var saveErr error
	ext := strings.ToLower(filepath.Ext(inputPath))

	switch ext {
	case ".jpg", ".jpeg":
		saveErr = imaging.Save(resized, outputPath, imaging.JPEGQuality(quality))
	case ".png":
		saveErr = imaging.Save(resized, outputPath)
	case ".gif":
		saveErr = imaging.Save(resized, outputPath)
	case ".tiff", ".tif":
		saveErr = imaging.Save(resized, outputPath)
	case ".bmp":
		saveErr = imaging.Save(resized, outputPath)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	if saveErr != nil {
		return fmt.Errorf("failed to save image: %v", saveErr)
	}

	// Print result information if verbose or not in batch mode
	if verbose || !batchMode {
		fmt.Printf("Resized %s: %dx%d -> %dx%d\n",
			filepath.Base(inputPath), originalWidth, originalHeight, actualWidth, actualHeight)
		fmt.Printf("Output: %s\n", outputPath)

		// Print file size information
		originalInfo, _ := os.Stat(inputPath)
		newInfo, _ := os.Stat(outputPath)

		fmt.Printf("File size: %s -> %s\n",
			formatFileSize(originalInfo.Size()), formatFileSize(newInfo.Size()))
	}

	return nil
}

/*
calculateTargetSize computes the target width and height for resizing,
preserving aspect ratio if only one dimension is set.
*/
func calculateTargetSize(originalWidth, originalHeight int) (int, int) {
	// If both dimensions are explicitly set, use them directly
	if widthSet && heightSet {
		return width, height
	}

	// If only width is set, calculate height proportionally
	if widthSet && !heightSet {
		ratio := float64(originalHeight) / float64(originalWidth)
		calculatedHeight := int(float64(width) * ratio)
		return width, calculatedHeight
	}

	// If only height is set, calculate width proportionally
	if !widthSet && heightSet {
		ratio := float64(originalWidth) / float64(originalHeight)
		calculatedWidth := int(float64(height) * ratio)
		return calculatedWidth, height
	}

	// This case should not happen as it's handled in PreRun
	return width, height
}

/*
generateOutputPath creates the output file path for the resized image,
including the new dimensions in the filename and using the specified output directory if provided.
*/
func generateOutputPath(inputPath, outputDir string, width, height int) string {
	dir := filepath.Dir(inputPath)
	if outputDir != "" {
		dir = outputDir
	}

	filename := filepath.Base(inputPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	newFilename := fmt.Sprintf("%s_%dx%d%s", nameWithoutExt, width, height, ext)
	return filepath.Join(dir, newFilename)
}

/*
formatFileSize returns a human-readable string for a file size in bytes (e.g., "1.2 MB").
*/
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
