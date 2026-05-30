package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

/*
processBatch processes all supported image files in the specified directory using a worker pool.
It collects image files, distributes them to worker goroutines, and prints a summary of results.
*/
func processBatch(dirPath string) {
	if verbose {
		fmt.Printf("Processing directory: %s\n", dirPath)
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

	runWorkerPool(imageFiles)
}

/*
runWorkerPool resizes the given image files concurrently using a pool of worker
goroutines and prints a summary of the results.
*/
func runWorkerPool(files []string) {
	fmt.Printf("Found %d image files\n", len(files))

	// Never start more workers than there are files to process.
	workerCount := min(workers, len(files))
	if verbose {
		fmt.Printf("Using %d workers\n", workerCount)
	}

	// Create channels for jobs and results for the worker pool
	jobs := make(chan string, len(files))
	results := make(chan error, len(files))

	// Start worker goroutines to process images concurrently
	var wg sync.WaitGroup
	for range workerCount {
		wg.Go(func() {
			for filePath := range jobs {
				results <- resizeImage(filePath, false)
			}
		})
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

/*
collectImageFiles recursively collects all supported image files from the given directory.
Returns a slice of file paths and any error encountered.
*/
func collectImageFiles(dirPath string) ([]string, error) {
	var imageFiles []string

	// Walk through the directory and collect files with supported extensions.
	// WalkDir avoids an lstat per entry (it only needs the dir-entry type here).
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && isImageFile(path) {
			imageFiles = append(imageFiles, path)
		}

		return nil
	})

	return imageFiles, err
}
