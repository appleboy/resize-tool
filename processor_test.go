package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Helper function to create a test image file
func createTestImage(path string, width, height int) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Create a simple image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern to make it interesting
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)  // #nosec G115 - Safe conversion for test data
			g := uint8((y * 255) / height) // #nosec G115 - Safe conversion for test data
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// Helper function to reset global variables to default state
func resetGlobals() {
	width = 0
	height = 0
	quality = 95
	outputDir = ""
	keepRatio = false
	batchMode = false
	workers = 4
	verbose = false
	overwrite = false
	widthSet = false
	heightSet = false
}

func TestCalculateTargetSize(t *testing.T) {
	tests := []struct {
		name           string
		originalWidth  int
		originalHeight int
		setupFlags     func()
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "Both width and height set",
			originalWidth:  800,
			originalHeight: 600,
			setupFlags: func() {
				resetGlobals()
				width = 400
				height = 300
				widthSet = true
				heightSet = true
			},
			expectedWidth:  400,
			expectedHeight: 300,
		},
		{
			name:           "Only width set - calculate height proportionally",
			originalWidth:  800,
			originalHeight: 600,
			setupFlags: func() {
				resetGlobals()
				width = 400
				widthSet = true
				heightSet = false
			},
			expectedWidth:  400,
			expectedHeight: 300,
		},
		{
			name:           "Only height set - calculate width proportionally",
			originalWidth:  800,
			originalHeight: 600,
			setupFlags: func() {
				resetGlobals()
				height = 300
				widthSet = false
				heightSet = true
			},
			expectedWidth:  400,
			expectedHeight: 300,
		},
		{
			name:           "Square image - width only",
			originalWidth:  100,
			originalHeight: 100,
			setupFlags: func() {
				resetGlobals()
				width = 50
				widthSet = true
				heightSet = false
			},
			expectedWidth:  50,
			expectedHeight: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()
			gotWidth, gotHeight := calculateTargetSize(tt.originalWidth, tt.originalHeight)
			if gotWidth != tt.expectedWidth || gotHeight != tt.expectedHeight {
				t.Errorf("calculateTargetSize() = (%d, %d), want (%d, %d)",
					gotWidth, gotHeight, tt.expectedWidth, tt.expectedHeight)
			}
		})
	}
}

func TestGenerateOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		outputDir  string
		width      int
		height     int
		setupFlags func()
		expected   string
	}{
		{
			name:      "Normal mode - same directory",
			inputPath: "/path/to/image.jpg",
			outputDir: "",
			width:     800,
			height:    600,
			setupFlags: func() {
				resetGlobals()
				overwrite = false
			},
			expected: "/path/to/image_800x600.jpg",
		},
		{
			name:      "Normal mode - custom output directory",
			inputPath: "/path/to/image.png",
			outputDir: "/output/dir",
			width:     400,
			height:    300,
			setupFlags: func() {
				resetGlobals()
				overwrite = false
			},
			expected: "/output/dir/image_400x300.png",
		},
		{
			name:      "Overwrite mode - ignores output directory",
			inputPath: "/path/to/image.jpg",
			outputDir: "/output/dir",
			width:     800,
			height:    600,
			setupFlags: func() {
				resetGlobals()
				overwrite = true
			},
			expected: "/path/to/image.jpg",
		},
		{
			name:      "Complex filename with multiple dots",
			inputPath: "/path/to/image.backup.jpg",
			outputDir: "",
			width:     1920,
			height:    1080,
			setupFlags: func() {
				resetGlobals()
				overwrite = false
			},
			expected: "/path/to/image.backup_1920x1080.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()
			got := generateOutputPath(tt.inputPath, tt.outputDir, tt.width, tt.height)
			if got != tt.expected {
				t.Errorf("generateOutputPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResizeImage(t *testing.T) {
	// Create a temporary directory for tests
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupFlags  func()
		setupImage  func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid PNG image - width only",
			setupFlags: func() {
				resetGlobals()
				width = 200
				widthSet = true
				heightSet = false
				verbose = false
			},
			setupImage: func() string {
				imagePath := filepath.Join(tempDir, "test.png")
				if err := createTestImage(imagePath, 400, 300); err != nil {
					t.Fatalf("Failed to create test image: %v", err)
				}
				return imagePath
			},
			expectError: false,
		},
		{
			name: "valid PNG image - height only",
			setupFlags: func() {
				resetGlobals()
				height = 150
				widthSet = false
				heightSet = true
				verbose = false
			},
			setupImage: func() string {
				imagePath := filepath.Join(tempDir, "test2.png")
				if err := createTestImage(imagePath, 400, 300); err != nil {
					t.Fatalf("Failed to create test image: %v", err)
				}
				return imagePath
			},
			expectError: false,
		},
		{
			name: "valid PNG image - both dimensions",
			setupFlags: func() {
				resetGlobals()
				width = 100
				height = 100
				widthSet = true
				heightSet = true
				verbose = false
			},
			setupImage: func() string {
				imagePath := filepath.Join(tempDir, "test3.png")
				if err := createTestImage(imagePath, 400, 300); err != nil {
					t.Fatalf("Failed to create test image: %v", err)
				}
				return imagePath
			},
			expectError: false,
		},
		{
			name: "non-existent image file",
			setupFlags: func() {
				resetGlobals()
				width = 200
				widthSet = true
				verbose = false
			},
			setupImage: func() string {
				return filepath.Join(tempDir, "nonexistent.png")
			},
			expectError: true,
			errorMsg:    "failed to open image",
		},
		{
			name: "unsupported image format",
			setupFlags: func() {
				resetGlobals()
				width = 200
				widthSet = true
				verbose = false
			},
			setupImage: func() string {
				// Create a text file with unsupported extension
				filePath := filepath.Join(tempDir, "fake.xyz")
				file, err := os.Create(filePath)
				if err != nil {
					t.Fatalf("Failed to create fake file: %v", err)
				}
				defer func() {
					if cerr := file.Close(); cerr != nil {
						t.Fatalf("Failed to close fake file: %v", cerr)
					}
				}()
				if _, err := file.WriteString("not an image"); err != nil {
					t.Fatalf("Failed to write to fake file: %v", err)
				}
				return filePath
			},
			expectError: true,
			errorMsg:    "image: unknown format",
		},
		{
			name: "invalid image data",
			setupFlags: func() {
				resetGlobals()
				width = 200
				widthSet = true
				verbose = false
			},
			setupImage: func() string {
				// Create a file with image extension but invalid content
				filePath := filepath.Join(tempDir, "invalid.png")
				file, err := os.Create(filePath)
				if err != nil {
					t.Fatalf("Failed to create invalid file: %v", err)
				}
				defer func() {
					if cerr := file.Close(); cerr != nil {
						t.Fatalf("Failed to close invalid file: %v", cerr)
					}
				}()
				if _, err := file.WriteString("not valid image data"); err != nil {
					t.Fatalf("Failed to write to invalid file: %v", err)
				}
				return filePath
			},
			expectError: true,
			errorMsg:    "failed to open image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()
			imagePath := tt.setupImage()

			err := resizeImage(imagePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}

				// Check if output file was created (for successful cases)
				if !overwrite {
					// Get the actual dimensions after resizing to match the real output path
					actualWidth, actualHeight := calculateTargetSize(400, 300) // Original test image size
					expectedOutput := generateOutputPath(imagePath, outputDir, actualWidth, actualHeight)
					if _, statErr := os.Stat(expectedOutput); os.IsNotExist(statErr) {
						t.Errorf("Expected output file %q was not created", expectedOutput)
					}
				}
			}
		})
	}
}

// Helper function to check if string contains substring

func TestProcessImages(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupFlags  func()
		setupInput  func() string
		expectPanic bool
	}{
		{
			name: "single image file",
			setupFlags: func() {
				resetGlobals()
				width = 200
				widthSet = true
				verbose = false
			},
			setupInput: func() string {
				imagePath := filepath.Join(tempDir, "single.png")
				if err := createTestImage(imagePath, 400, 300); err != nil {
					t.Fatalf("Failed to create test image: %v", err)
				}
				return imagePath
			},
		},
		{
			name: "directory with batch mode",
			setupFlags: func() {
				resetGlobals()
				width = 100
				widthSet = true
				batchMode = true
				verbose = false
			},
			setupInput: func() string {
				// Create a directory with test images
				dirPath := filepath.Join(tempDir, "batch_test")
				if err := os.MkdirAll(dirPath, 0o755); err != nil {
					t.Fatalf("Failed to create batch test directory: %v", err)
				}

				// Create multiple test images
				if err := createTestImage(filepath.Join(dirPath, "img1.png"), 200, 150); err != nil {
					t.Fatalf("Failed to create test image 1: %v", err)
				}
				if err := createTestImage(filepath.Join(dirPath, "img2.png"), 300, 200); err != nil {
					t.Fatalf("Failed to create test image 2: %v", err)
				}

				return dirPath
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()
			inputPath := tt.setupInput()

			// Create mock cobra command and args
			cmd := &cobra.Command{}
			args := []string{inputPath}

			// This test primarily checks that the function doesn't panic
			// and properly dispatches to the correct processing function
			defer func() {
				if r := recover(); r != nil && !tt.expectPanic {
					t.Errorf("processImages() caused panic: %v", r)
				}
			}()

			processImages(cmd, args)
		})
	}
}

func TestCalculateTargetSizeEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		originalWidth  int
		originalHeight int
		setupFlags     func()
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "Very small image proportional resize",
			originalWidth:  10,
			originalHeight: 5,
			setupFlags: func() {
				resetGlobals()
				width = 20
				widthSet = true
				heightSet = false
			},
			expectedWidth:  20,
			expectedHeight: 10,
		},
		{
			name:           "Large aspect ratio - width only",
			originalWidth:  1000,
			originalHeight: 100,
			setupFlags: func() {
				resetGlobals()
				width = 500
				widthSet = true
				heightSet = false
			},
			expectedWidth:  500,
			expectedHeight: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()
			gotWidth, gotHeight := calculateTargetSize(tt.originalWidth, tt.originalHeight)
			if gotWidth != tt.expectedWidth || gotHeight != tt.expectedHeight {
				t.Errorf("calculateTargetSize() = (%d, %d), want (%d, %d)",
					gotWidth, gotHeight, tt.expectedWidth, tt.expectedHeight)
			}
		})
	}
}

func TestGenerateOutputPathEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		outputDir  string
		width      int
		height     int
		setupFlags func()
		expected   string
	}{
		{
			name:      "Windows path style",
			inputPath: filepath.Join("C:", "path", "to", "image.jpg"),
			outputDir: "",
			width:     800,
			height:    600,
			setupFlags: func() {
				resetGlobals()
				overwrite = false
			},
			expected: filepath.Join("C:", "path", "to", "image_800x600.jpg"),
		},
		{
			name:      "No extension",
			inputPath: "/path/to/imagefile",
			outputDir: "",
			width:     100,
			height:    100,
			setupFlags: func() {
				resetGlobals()
				overwrite = false
			},
			expected: "/path/to/imagefile_100x100",
		},
		{
			name:      "Very long filename",
			inputPath: "/path/to/very_long_filename_with_many_characters_and_details.jpeg",
			outputDir: "/output",
			width:     1920,
			height:    1080,
			setupFlags: func() {
				resetGlobals()
				overwrite = false
			},
			expected: "/output/very_long_filename_with_many_characters_and_details_1920x1080.jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()
			got := generateOutputPath(tt.inputPath, tt.outputDir, tt.width, tt.height)
			if got != tt.expected {
				t.Errorf("generateOutputPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func BenchmarkCalculateTargetSize(b *testing.B) {
	resetGlobals()
	width = 800
	widthSet = true
	heightSet = false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateTargetSize(1920, 1080)
	}
}

func BenchmarkGenerateOutputPath(b *testing.B) {
	resetGlobals()
	overwrite = false

	inputPath := "/path/to/test/image.jpg"
	outputDir := "/output/dir"
	width := 1920
	height := 1080

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateOutputPath(inputPath, outputDir, width, height)
	}
}

// Tests for glob pattern support

func TestContainsGlobPattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "pattern with asterisk",
			path:     "images/*.png",
			expected: true,
		},
		{
			name:     "pattern with question mark",
			path:     "image?.png",
			expected: true,
		},
		{
			name:     "pattern with brackets",
			path:     "image[123].png",
			expected: true,
		},
		{
			name:     "pattern with double asterisk",
			path:     "photos/**/*.jpg",
			expected: true,
		},
		{
			name:     "normal file path without pattern",
			path:     "images/photo.png",
			expected: false,
		},
		{
			name:     "directory path without pattern",
			path:     "/path/to/images/",
			expected: false,
		},
		{
			name:     "empty string",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsGlobPattern(tt.path)
			if got != tt.expected {
				t.Errorf("containsGlobPattern(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "JPEG file with jpg extension",
			path:     "photo.jpg",
			expected: true,
		},
		{
			name:     "JPEG file with jpeg extension",
			path:     "photo.jpeg",
			expected: true,
		},
		{
			name:     "PNG file",
			path:     "photo.png",
			expected: true,
		},
		{
			name:     "GIF file",
			path:     "photo.gif",
			expected: true,
		},
		{
			name:     "TIFF file with tiff extension",
			path:     "photo.tiff",
			expected: true,
		},
		{
			name:     "TIFF file with tif extension",
			path:     "photo.tif",
			expected: true,
		},
		{
			name:     "BMP file",
			path:     "photo.bmp",
			expected: true,
		},
		{
			name:     "uppercase extension",
			path:     "photo.PNG",
			expected: true,
		},
		{
			name:     "mixed case extension",
			path:     "photo.JpG",
			expected: true,
		},
		{
			name:     "text file",
			path:     "document.txt",
			expected: false,
		},
		{
			name:     "PDF file",
			path:     "document.pdf",
			expected: false,
		},
		{
			name:     "no extension",
			path:     "photo",
			expected: false,
		},
		{
			name:     "path with directory",
			path:     "/path/to/photo.jpg",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isImageFile(tt.path)
			if got != tt.expected {
				t.Errorf("isImageFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestExpandGlobPattern(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := []struct {
		path    string
		isImage bool
	}{
		{filepath.Join(tempDir, "image1.png"), true},
		{filepath.Join(tempDir, "image2.jpg"), true},
		{filepath.Join(tempDir, "image3.jpeg"), true},
		{filepath.Join(tempDir, "document.txt"), false},
		{filepath.Join(tempDir, "photo1.gif"), true},
		{filepath.Join(tempDir, "photo2.bmp"), true},
	}

	// Create the test files
	for _, tf := range testFiles {
		if tf.isImage {
			if err := createTestImage(tf.path, 100, 100); err != nil {
				t.Fatalf("Failed to create test image %s: %v", tf.path, err)
			}
		} else {
			file, err := os.Create(tf.path)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", tf.path, err)
			}
			file.Close()
		}
	}

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "match all PNG files",
			pattern:       filepath.Join(tempDir, "*.png"),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "match all JPG/JPEG files",
			pattern:       filepath.Join(tempDir, "image*.jpg") + "*",
			expectedCount: 1, // Only .jpg, not .jpeg with this pattern
			expectError:   false,
		},
		{
			name:          "match all image files",
			pattern:       filepath.Join(tempDir, "*"),
			expectedCount: 5, // All image files, excluding .txt
			expectError:   false,
		},
		{
			name:          "no matches",
			pattern:       filepath.Join(tempDir, "nonexistent*.png"),
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "match specific prefix",
			pattern:       filepath.Join(tempDir, "photo*"),
			expectedCount: 2, // photo1.gif and photo2.bmp
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := expandGlobPattern(tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("expandGlobPattern() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expandGlobPattern() unexpected error: %v", err)
				}
				if len(files) != tt.expectedCount {
					t.Errorf("expandGlobPattern() returned %d files, want %d. Files: %v",
						len(files), tt.expectedCount, files)
				}
				// Verify all returned files are image files
				for _, file := range files {
					if !isImageFile(file) {
						t.Errorf("expandGlobPattern() returned non-image file: %s", file)
					}
				}
			}
		})
	}
}

func TestProcessMultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test images
	testFiles := []string{
		filepath.Join(tempDir, "test1.png"),
		filepath.Join(tempDir, "test2.png"),
		filepath.Join(tempDir, "test3.png"),
	}

	for _, path := range testFiles {
		if err := createTestImage(path, 200, 150); err != nil {
			t.Fatalf("Failed to create test image %s: %v", path, err)
		}
	}

	tests := []struct {
		name       string
		setupFlags func()
		files      []string
	}{
		{
			name: "process multiple files with width only",
			setupFlags: func() {
				resetGlobals()
				width = 100
				widthSet = true
				heightSet = false
				verbose = false
			},
			files: testFiles,
		},
		{
			name: "process multiple files with custom output",
			setupFlags: func() {
				resetGlobals()
				width = 150
				widthSet = true
				heightSet = false
				outputDir = filepath.Join(tempDir, "output")
				verbose = false
				// Pre-create output directory to avoid race conditions in test
				_ = os.MkdirAll(filepath.Join(tempDir, "output"), 0o755)
			},
			files: testFiles[:2], // Only first two files
		},
		{
			name: "process multiple files with verbose",
			setupFlags: func() {
				resetGlobals()
				height = 100
				widthSet = false
				heightSet = true
				verbose = true
			},
			files: testFiles[1:], // Skip first file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()

			// This test primarily checks that the function doesn't panic
			// and processes files without errors
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("processMultipleFiles() caused panic: %v", r)
				}
			}()

			processMultipleFiles(tt.files)

			// Verify output files were created (if not in overwrite mode)
			if !overwrite {
				for _, inputFile := range tt.files {
					// Check that at least one output file with the input file base name exists
					baseName := strings.TrimSuffix(
						filepath.Base(inputFile),
						filepath.Ext(inputFile),
					)
					outputPath := filepath.Dir(inputFile)
					if outputDir != "" {
						outputPath = outputDir
					}

					// Find files with matching base name in output directory
					pattern := filepath.Join(outputPath, baseName+"_*.png")
					matches, err := filepath.Glob(pattern)
					if err != nil {
						t.Errorf("Error searching for output files: %v", err)
						continue
					}
					if len(matches) == 0 {
						t.Errorf("No output file found matching pattern %q", pattern)
					}
				}
			}
		})
	}
}

func TestProcessImagesWithGlobPattern(t *testing.T) {
	tempDir := t.TempDir()

	// Create test images
	testImages := []string{
		filepath.Join(tempDir, "photo1.png"),
		filepath.Join(tempDir, "photo2.png"),
		filepath.Join(tempDir, "image1.jpg"),
	}

	for _, path := range testImages {
		if err := createTestImage(path, 300, 200); err != nil {
			t.Fatalf("Failed to create test image %s: %v", path, err)
		}
	}

	tests := []struct {
		name       string
		setupFlags func()
		pattern    string
	}{
		{
			name: "glob pattern matching multiple PNG files",
			setupFlags: func() {
				resetGlobals()
				width = 150
				widthSet = true
				heightSet = false
				verbose = false
			},
			pattern: filepath.Join(tempDir, "photo*.png"),
		},
		{
			name: "glob pattern matching single file",
			setupFlags: func() {
				resetGlobals()
				height = 100
				widthSet = false
				heightSet = true
				verbose = false
			},
			pattern: filepath.Join(tempDir, "image1.jpg"),
		},
		{
			name: "glob pattern matching all images",
			setupFlags: func() {
				resetGlobals()
				width = 200
				widthSet = true
				heightSet = false
				verbose = true
			},
			pattern: filepath.Join(tempDir, "*"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFlags()

			// Create mock cobra command and args
			cmd := &cobra.Command{}
			args := []string{tt.pattern}

			// This test checks that processImages handles glob patterns correctly
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("processImages() with glob pattern caused panic: %v", r)
				}
			}()

			processImages(cmd, args)
		})
	}
}
