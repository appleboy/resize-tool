package main

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/appleboy/com/file"
	"github.com/disintegration/imaging"
	"github.com/spf13/cobra"
)

/*
processImages determines whether the input path is a file or directory,
and processes it accordingly. If batch mode is enabled or the input is a directory,
it processes all images in the directory. Otherwise, it processes a single image file.
Supports glob patterns like images/*.png or photos/**\/*.jpg.
Also handles multiple arguments when shell expands glob patterns.
*/
func processImages(cmd *cobra.Command, args []string) {
	// Handle multiple arguments (shell-expanded glob or multiple files)
	if len(args) > 1 {
		// Filter out non-image files
		var imageFiles []string
		for _, path := range args {
			info, err := os.Stat(path)
			if err != nil {
				slog.Error(fmt.Sprintf("Cannot access file: %s, error: %v", path, err))
				continue
			}
			if !info.IsDir() && isImageFile(path) {
				imageFiles = append(imageFiles, path)
			}
		}

		if len(imageFiles) == 0 {
			slog.Error("No valid image files found in arguments")
			os.Exit(1)
		}

		resizeFiles(imageFiles)
		return
	}

	// Single argument - process as before
	inputPath := args[0] // #nosec G602 -- cobra.MinimumNArgs(1) ensures args is not empty

	// Check if the input contains a glob pattern
	if containsGlobPattern(inputPath) {
		// Expand glob pattern and process matching files
		files, err := expandGlobPattern(inputPath)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to expand glob pattern: %v", err))
			os.Exit(1)
		}

		if len(files) == 0 {
			slog.Error(fmt.Sprintf("No image files match pattern: %s", inputPath))
			os.Exit(1)
		}

		resizeFiles(files)
		return
	}

	// Check if the input path exists and is accessible
	info, err := statInputPath(inputPath)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if info.IsDir() || batchMode {
		// Process all images in the directory
		processBatch(inputPath)
	} else {
		// Process a single image file
		if err := resizeImage(inputPath, true); err != nil {
			slog.Error(fmt.Sprintf("Failed to process image: %v", err))
			os.Exit(1)
		}
	}
}

/*
statInputPath stats a single input path, returning a descriptive error if it
does not exist or cannot be accessed. Keeping this separate from processImages
(which calls os.Exit) makes the not-exist vs other-error handling unit-testable
and guarantees a nil FileInfo is never paired with a nil error.
*/
func statInputPath(inputPath string) (os.FileInfo, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", inputPath)
		}
		return nil, fmt.Errorf("cannot access path %s: %w", inputPath, err)
	}
	return info, nil
}

/*
resizeFiles dispatches a list of image files: a single file is resized directly
(with rich per-file output), while multiple files go through the worker pool.
*/
func resizeFiles(files []string) {
	if len(files) == 1 {
		if err := resizeImage(files[0], true); err != nil {
			slog.Error(fmt.Sprintf("Failed to process image: %v", err))
			os.Exit(1)
		}
		return
	}
	processMultipleFiles(files)
}

/*
resizeImage resizes a single image file according to the specified parameters.
It preserves aspect ratio if required and saves the output. When detailed is
true (single-file runs) it prints the per-file result block; worker-pool calls
pass detailed=false so that, without --verbose, the pool prints only its summary
instead of per-file blocks. With --verbose every call still prints its progress
and result lines, so concurrent pool output may interleave.
*/
func resizeImage(inputPath string, detailed bool) error {
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

	// Get original image dimensions (Dx/Dy account for a non-zero bounds origin)
	originalBounds := src.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	if verbose {
		fmt.Printf("  Original size: %dx%d\n", originalWidth, originalHeight)
	}

	// Calculate target dimensions based on flags and original size
	targetWidth, targetHeight := calculateTargetSize(originalWidth, originalHeight)

	if verbose {
		fmt.Printf("  Target size: %dx%d\n", targetWidth, targetHeight)
	}

	var resized image.Image

	// Choose resizing method based on which dimensions were set and the keep-ratio flag
	switch {
	case widthSet && !heightSet:
		// Only width set, height is auto-calculated
		resized = imaging.Resize(src, targetWidth, 0, imaging.Lanczos)
	case !widthSet && heightSet:
		// Only height set, width is auto-calculated
		resized = imaging.Resize(src, 0, targetHeight, imaging.Lanczos)
	case keepRatio && targetWidth > 0 && targetHeight > 0:
		// Both set to positive values, keep ratio (fit within bounds)
		resized = imaging.Fit(src, targetWidth, targetHeight, imaging.Lanczos)
	default:
		// Force resize to the given dimensions (a 0 dimension is auto-derived
		// by imaging; with both set and no keep-ratio this may distort).
		resized = imaging.Resize(src, targetWidth, targetHeight, imaging.Lanczos)
	}

	// Get actual resized dimensions (used for output filename)
	actualBounds := resized.Bounds()
	actualWidth := actualBounds.Dx()
	actualHeight := actualBounds.Dy()

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
	case extJPG, extJPEG:
		saveErr = imaging.Save(resized, outputPath, imaging.JPEGQuality(quality))
	case extPNG, extGIF, extTIFF, extTIF, extBMP:
		saveErr = imaging.Save(resized, outputPath)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	if saveErr != nil {
		return fmt.Errorf("failed to save image: %v", saveErr)
	}

	// Print the per-file result block for single-file runs or in verbose mode.
	// Worker-pool calls pass detailed=false to avoid interleaved concurrent output.
	if verbose || detailed {
		fmt.Printf("Resized %s: %dx%d -> %dx%d\n",
			filepath.Base(inputPath), originalWidth, originalHeight, actualWidth, actualHeight)
		fmt.Printf("Output: %s\n", outputPath)

		// Print file size information
		originalInfo, err := os.Stat(inputPath)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to get original file info: %v", err))
		} else {
			newInfo, err := os.Stat(outputPath)
			if err != nil {
				slog.Error(fmt.Sprintf("Failed to get output file info: %v", err))
			} else {
				fmt.Printf("File size: %s -> %s\n",
					file.FormatSize(originalInfo.Size()), file.FormatSize(newInfo.Size()))
			}
		}
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

/*
containsGlobPattern checks if the input path contains glob pattern characters.
Returns true if the path contains *, ?, or [ characters.
*/
func containsGlobPattern(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

/*
isImageFile checks if the file has a supported image extension.
Returns true if the file extension is one of: jpg, jpeg, png, gif, tiff, tif, bmp.
*/
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return supportedImageExts[ext]
}

/*
expandGlobPattern expands a glob pattern and filters the results to only include image files.
Returns a slice of matching image file paths and any error encountered.
*/
func expandGlobPattern(pattern string) ([]string, error) {
	// Use filepath.Glob to expand the pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %v", err)
	}

	// Filter to only include image files
	var imageFiles []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			// Skip files that can't be accessed
			continue
		}

		// Only include regular files (not directories) that are image files
		if !info.IsDir() && isImageFile(match) {
			imageFiles = append(imageFiles, match)
		}
	}

	return imageFiles, nil
}

/*
processMultipleFiles processes a pre-defined list of image files using a worker pool.
Similar to processBatch but works with an explicit list rather than a directory walk.
*/
func processMultipleFiles(files []string) {
	if verbose {
		fmt.Printf("Processing %d files\n", len(files))
	}

	runWorkerPool(files)
}
