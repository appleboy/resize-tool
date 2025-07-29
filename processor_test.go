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
				if _, err := file.WriteString("not an image"); err != nil {
					t.Fatalf("Failed to write to fake file: %v", err)
				}
				file.Close()
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
				if _, err := file.WriteString("not valid image data"); err != nil {
					t.Fatalf("Failed to write to invalid file: %v", err)
				}
				file.Close()
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
		expectExit  bool
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
			expectExit: false,
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
			expectExit: false,
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
