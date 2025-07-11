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

var (
	width     int
	height    int
	quality   int
	outputDir string
	keepRatio bool
	batchMode bool
	workers   int
	verbose   bool
	// Flag to track if dimensions were explicitly set
	widthSet  bool
	heightSet bool
)

func main() {
	// Set up slog logger based on verbose flag
	var handler slog.Handler
	if verbose {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})
	}
	slog.SetDefault(slog.New(handler))

	rootCmd := &cobra.Command{
		Use:   "resize-tool [image-file-or-directory]",
		Short: "A powerful image resizing tool",
		Long: `A command-line tool to resize images with various options.
Supports JPEG, PNG, GIF, TIFF, and BMP formats.
Can process single files or batch process directories.

By default, if only width or height is specified, the other dimension
will be calculated automatically to maintain aspect ratio.`,
		Args: cobra.ExactArgs(1),
		Run:  processImages,
	}

	// Use custom flag settings to track which parameters were explicitly set
	rootCmd.Flags().IntVarP(&width, "width", "w", 0, "Output width (pixels, 0=auto based on height)")
	rootCmd.Flags().IntVarP(&height, "height", "", 0, "Output height (pixels, 0=auto based on width)")
	rootCmd.Flags().IntVarP(&quality, "quality", "q", 95, "JPEG quality (1-100)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: same as input)")
	rootCmd.Flags().BoolVarP(&keepRatio, "keep-ratio", "k", false, "Keep aspect ratio when both width and height are specified")
	rootCmd.Flags().BoolVarP(&batchMode, "batch", "b", false, "Batch process all images in directory")
	rootCmd.Flags().IntVarP(&workers, "workers", "", 4, "Number of worker goroutines for batch processing")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Set default values and check parameters
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Check if dimensions were explicitly set
		widthSet = cmd.Flags().Changed("width")
		heightSet = cmd.Flags().Changed("height")

		// If neither is set, use default values
		if !widthSet && !heightSet {
			width = 800
			height = 600
			widthSet = true
			heightSet = true
		}

		// Validate parameters
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

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute command", "err", err)
		os.Exit(1)
	}
}

func processImages(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// Check input path
	info, err := os.Stat(inputPath)
	if os.IsNotExist(err) {
		slog.Error(fmt.Sprintf("Path does not exist: %s", inputPath))
		os.Exit(1)
	}

	if info.IsDir() || batchMode {
		// Directory processing
		processBatch(inputPath)
	} else {
		// Single file processing
		if err := resizeImage(inputPath); err != nil {
			slog.Error(fmt.Sprintf("Failed to process image: %v", err))
			os.Exit(1)
		}
	}
}

func processBatch(dirPath string) {
	if verbose {
		fmt.Printf("Processing directory: %s\n", dirPath)
		fmt.Printf("Using %d workers\n", workers)
	}

	// Collect all image files
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

	// 使用 worker pool 進行並行處理
	jobs := make(chan string, len(imageFiles))
	results := make(chan error, len(imageFiles))

	// 啟動 workers
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

	// Send jobs
	go func() {
		defer close(jobs)
		for _, filePath := range imageFiles {
			jobs <- filePath
		}
	}()

	// Wait for completion and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect statistics
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

func resizeImage(inputPath string) error {
	if verbose {
		fmt.Printf("Processing: %s\n", inputPath)
	}

	// Open and decode image
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

	// Calculate target dimensions
	targetWidth, targetHeight := calculateTargetSize(originalWidth, originalHeight)

	if verbose {
		fmt.Printf("  Target size: %dx%d\n", targetWidth, targetHeight)
	}

	var resized image.Image

	// Determine resize method
	if (widthSet && heightSet && keepRatio) || (!widthSet && heightSet) || (widthSet && !heightSet) {
		// Keep aspect ratio
		if widthSet && !heightSet {
			// Only width set, height auto-calculated
			resized = imaging.Resize(src, targetWidth, 0, imaging.Lanczos)
		} else if !widthSet && heightSet {
			// Only height set, width auto-calculated
			resized = imaging.Resize(src, 0, targetHeight, imaging.Lanczos)
		} else {
			// Both set, but keep ratio (fit within bounds)
			resized = imaging.Fit(src, targetWidth, targetHeight, imaging.Lanczos)
		}
	} else {
		// Force resize to exact dimensions (may distort)
		resized = imaging.Resize(src, targetWidth, targetHeight, imaging.Lanczos)
	}

	// Get actual resized dimensions (for filename)
	actualBounds := resized.Bounds()
	actualWidth := actualBounds.Max.X
	actualHeight := actualBounds.Max.Y

	// Determine output path
	outputPath := generateOutputPath(inputPath, outputDir, actualWidth, actualHeight)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Save resized image
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

	// Display result information
	if verbose || !batchMode {
		fmt.Printf("Resized %s: %dx%d -> %dx%d\n",
			filepath.Base(inputPath), originalWidth, originalHeight, actualWidth, actualHeight)
		fmt.Printf("Output: %s\n", outputPath)

		// Display file size information
		originalInfo, _ := os.Stat(inputPath)
		newInfo, _ := os.Stat(outputPath)

		fmt.Printf("File size: %s -> %s\n",
			formatFileSize(originalInfo.Size()), formatFileSize(newInfo.Size()))
	}

	return nil
}

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
