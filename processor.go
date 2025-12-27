package main

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

		if len(imageFiles) == 1 {
			// Single image file
			if err := resizeImage(imageFiles[0]); err != nil {
				slog.Error(fmt.Sprintf("Failed to process image: %v", err))
				os.Exit(1)
			}
		} else {
			// Multiple image files
			processMultipleFiles(imageFiles)
		}
		return
	}

	// Single argument - process as before
	inputPath := args[0]

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

		if len(files) == 1 {
			// Single file matched, process it normally
			if err := resizeImage(files[0]); err != nil {
				slog.Error(fmt.Sprintf("Failed to process image: %v", err))
				os.Exit(1)
			}
		} else {
			// Multiple files matched, process them in batch
			processMultipleFiles(files)
		}
		return
	}

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
	if (widthSet && heightSet && keepRatio) || (!widthSet && heightSet) ||
		(widthSet && !heightSet) {
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
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".tiff": true,
		".tif":  true,
		".bmp":  true,
	}
	ext := strings.ToLower(filepath.Ext(path))
	return supportedExts[ext]
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
processMultipleFiles processes multiple image files using a worker pool.
Similar to processBatch but works with a pre-defined list of files.
*/
func processMultipleFiles(files []string) {
	if verbose {
		fmt.Printf("Processing %d files from glob pattern\n", len(files))
		fmt.Printf("Using %d workers\n", workers)
	}

	fmt.Printf("Found %d image files\n", len(files))

	// Create channels for jobs and results for the worker pool
	jobs := make(chan string, len(files))
	results := make(chan error, len(files))

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
		for _, filePath := range files {
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
