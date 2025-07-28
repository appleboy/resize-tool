package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

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
