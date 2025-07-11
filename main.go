package main

import (
	"fmt"
	"image"
	"log"
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
	// 標記是否明確設定了尺寸
	widthSet  bool
	heightSet bool
)

func main() {
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

	// 使用自訂的 flag 設定來追蹤哪些參數被明確設定
	rootCmd.Flags().IntVarP(&width, "width", "w", 0, "Output width (pixels, 0=auto based on height)")
	rootCmd.Flags().IntVarP(&height, "height", "", 0, "Output height (pixels, 0=auto based on width)")
	rootCmd.Flags().IntVarP(&quality, "quality", "q", 95, "JPEG quality (1-100)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: same as input)")
	rootCmd.Flags().BoolVarP(&keepRatio, "keep-ratio", "k", false, "Keep aspect ratio when both width and height are specified")
	rootCmd.Flags().BoolVarP(&batchMode, "batch", "b", false, "Batch process all images in directory")
	rootCmd.Flags().IntVarP(&workers, "workers", "", 4, "Number of worker goroutines for batch processing")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// 設定預設值和檢查參數
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// 檢查是否明確設定了尺寸
		widthSet = cmd.Flags().Changed("width")
		heightSet = cmd.Flags().Changed("height")

		// 如果都沒設定，使用預設值
		if !widthSet && !heightSet {
			width = 800
			height = 600
			widthSet = true
			heightSet = true
		}

		// 驗證參數
		if width < 0 || height < 0 {
			log.Fatal("Width and height must be positive numbers")
		}
		if width == 0 && height == 0 {
			log.Fatal("At least one of width or height must be specified")
		}
		if quality < 1 || quality > 100 {
			log.Fatal("Quality must be between 1 and 100")
		}
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func processImages(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	// 檢查輸入路徑
	info, err := os.Stat(inputPath)
	if os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %s", inputPath)
	}

	if info.IsDir() || batchMode {
		// 目錄處理
		processBatch(inputPath)
	} else {
		// 單檔處理
		if err := resizeImage(inputPath); err != nil {
			log.Fatalf("Failed to process image: %v", err)
		}
	}
}

func processBatch(dirPath string) {
	if verbose {
		fmt.Printf("Processing directory: %s\n", dirPath)
		fmt.Printf("Using %d workers\n", workers)
	}

	// 收集所有圖片檔案
	imageFiles, err := collectImageFiles(dirPath)
	if err != nil {
		log.Fatalf("Failed to collect image files: %v", err)
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

	// 發送工作
	go func() {
		defer close(jobs)
		for _, filePath := range imageFiles {
			jobs <- filePath
		}
	}()

	// 等待完成並收集結果
	go func() {
		wg.Wait()
		close(results)
	}()

	// 統計結果
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

	// 開啟並解碼圖片
	src, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %v", inputPath, err)
	}

	// 取得原始圖片尺寸
	originalBounds := src.Bounds()
	originalWidth := originalBounds.Max.X
	originalHeight := originalBounds.Max.Y

	if verbose {
		fmt.Printf("  Original size: %dx%d\n", originalWidth, originalHeight)
	}

	// 計算目標尺寸
	targetWidth, targetHeight := calculateTargetSize(originalWidth, originalHeight)

	if verbose {
		fmt.Printf("  Target size: %dx%d\n", targetWidth, targetHeight)
	}

	var resized image.Image

	// 決定調整方式
	if (widthSet && heightSet && keepRatio) || (!widthSet && heightSet) || (widthSet && !heightSet) {
		// 保持比例調整
		if widthSet && !heightSet {
			// 只設定寬度，高度自動計算
			resized = imaging.Resize(src, targetWidth, 0, imaging.Lanczos)
		} else if !widthSet && heightSet {
			// 只設定高度，寬度自動計算
			resized = imaging.Resize(src, 0, targetHeight, imaging.Lanczos)
		} else {
			// 兩者都設定，但要保持比例
			resized = imaging.Fit(src, targetWidth, targetHeight, imaging.Lanczos)
		}
	} else {
		// 強制調整到指定尺寸（可能會變形）
		resized = imaging.Resize(src, targetWidth, targetHeight, imaging.Lanczos)
	}

	// 取得實際調整後的尺寸（用於檔名）
	actualBounds := resized.Bounds()
	actualWidth := actualBounds.Max.X
	actualHeight := actualBounds.Max.Y

	// 確定輸出路徑
	outputPath := generateOutputPath(inputPath, outputDir, actualWidth, actualHeight)

	// 確保輸出目錄存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 儲存調整後的圖片
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

	// 顯示結果資訊
	if verbose || !batchMode {
		fmt.Printf("Resized %s: %dx%d -> %dx%d\n",
			filepath.Base(inputPath), originalWidth, originalHeight, actualWidth, actualHeight)
		fmt.Printf("Output: %s\n", outputPath)

		// 顯示檔案大小資訊
		originalInfo, _ := os.Stat(inputPath)
		newInfo, _ := os.Stat(outputPath)

		fmt.Printf("File size: %s -> %s\n",
			formatFileSize(originalInfo.Size()), formatFileSize(newInfo.Size()))
	}

	return nil
}

func calculateTargetSize(originalWidth, originalHeight int) (int, int) {
	// 如果兩個尺寸都明確設定了，直接使用
	if widthSet && heightSet {
		return width, height
	}

	// 如果只設定了寬度，按比例計算高度
	if widthSet && !heightSet {
		ratio := float64(originalHeight) / float64(originalWidth)
		calculatedHeight := int(float64(width) * ratio)
		return width, calculatedHeight
	}

	// 如果只設定了高度，按比例計算寬度
	if !widthSet && heightSet {
		ratio := float64(originalWidth) / float64(originalHeight)
		calculatedWidth := int(float64(height) * ratio)
		return calculatedWidth, height
	}

	// 這種情況不應該發生，因為在 PreRun 中已經處理了
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
