package main

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/spf13/cobra"
)

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
resizeImage resizes a single image file according to the specified parameters.
It preserves aspect ratio if required, saves the output, and prints information if verbose is enabled.
*/
func resizeImage(inputPath string) error {
	if verbose {
		fmt.Printf("Processing: %s\n", inputPath)
		if overwrite {
			fmt.Printf("  Warning: Will overwrite original file\n")
		}
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
If overwrite is enabled, returns the original file path (ignoring outputDir).
*/
func generateOutputPath(inputPath, outputDir string, width, height int) string {
	// If overwrite mode is enabled, always return original file path
	if overwrite {
		return inputPath
	}

	// Original logic: generate new filename with dimensions
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
