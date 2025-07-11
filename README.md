# Go Image Resize Tool

A simple yet powerful image resizing tool built with Go.

## Features

- Support for multiple image formats: JPEG, PNG, GIF, TIFF, BMP
- **🎯 Smart Aspect Ratio**: When only width or height is specified, the other dimension is automatically calculated proportionally
- Flexible resizing options
- Aspect ratio preservation options
- Adjustable JPEG quality
- Batch processing for directories
- Parallel processing for improved efficiency
- Custom output directories
- Detailed progress and size information display

## Installation

### Build from Source

```bash
git clone <your-repo>
cd resize-tool
go mod tidy
go build -o resize-tool .
```

### Direct Usage

If you have the compiled binary, you can use it directly:

```bash
./resize-tool [options] <image-file>
```

## Usage

### Basic Usage

```bash
# Default resize to 800x600 pixels
./resize-tool image.jpg

# 🎯 Specify only width, height auto-calculated proportionally (recommended)
./resize-tool -w 1200 image.jpg

# 🎯 Specify only height, width auto-calculated proportionally (recommended)
./resize-tool --height 800 image.jpg

# Specify exact dimensions (may distort)
./resize-tool -w 1200 --height 800 image.jpg

# Specify dimensions but maintain aspect ratio (fit within bounds)
./resize-tool -k -w 1200 --height 800 image.jpg
```

### Advanced Options

```bash
# Set JPEG quality (1-100)
./resize-tool -q 85 -w 1000 image.jpg

# Specify output directory
./resize-tool -w 800 -o ./resized/ image.jpg

# Batch process all images in directory
./resize-tool -b -w 1200 /path/to/images/

# Use multiple threads for batch processing
./resize-tool -b --workers 8 -w 1920 /path/to/images/

# Verbose output mode
./resize-tool -v -w 800 image.jpg

# Combine multiple options
./resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## Parameters

| Parameter      | Short | Default | Description                                               |
| -------------- | ----- | ------- | --------------------------------------------------------- |
| `--width`      | `-w`  | 0       | Output width (pixels, 0=auto-calculate based on height)  |
| `--height`     |       | 0       | Output height (pixels, 0=auto-calculate based on width)  |
| `--quality`    | `-q`  | 95      | JPEG quality (1-100)                                     |
| `--output`     | `-o`  | same    | Output directory (default: same as input)                |
| `--keep-ratio` | `-k`  | false   | Keep aspect ratio when both width and height specified   |
| `--batch`      | `-b`  | false   | Batch process all images in directory                     |
| `--workers`    |       | 4       | Number of parallel workers for batch processing          |
| `--verbose`    | `-v`  | false   | Enable verbose output                                     |
| `--help`       | `-h`  |         | Show help message                                         |

## Output Filename Format

Resized files will automatically include dimension information:

- Original file: `photo.jpg`
- Output file: `photo_800x600.jpg`

## Examples

### 1. Batch Process Multiple Images

```bash
# Process all jpg files in current directory
for img in *.jpg; do
    ./resize-tool -w 1200 --height 800 -k "$img"
done
```

### 2. 為網站最佳化圖片

```bash
# 建立三種不同尺寸
./resize-tool -w 1920 --height 1080 -q 85 -o ./large/ image.jpg
./resize-tool -w 1200 --height 800 -q 85 -o ./medium/ image.jpg
./resize-tool -w 600 --height 400 -q 80 -o ./small/ image.jpg
```

### 3. 保持比例的縮圖

```bash
# 建立不超過 300x300 的縮圖，保持原比例
./resize-tool -w 300 --height 300 -k -o ./thumbnails/ image.jpg
```

## 支援的圖片格式

- **輸入格式**：JPEG, PNG, GIF, TIFF, BMP
- **Output formats**: Same as input format

## Performance Tips

- Uses Lanczos algorithm for high-quality image resizing
- Large file processing may require more memory
- JPEG quality setting affects both file size and image quality

## Error Handling

The tool automatically handles common error conditions:

- File not found
- Unsupported image formats
- Output directory creation failure
- Out of memory

## Technical Details

### Libraries Used

- `github.com/disintegration/imaging` - Image processing
- `github.com/spf13/cobra` - CLI interface

### Image Processing Algorithms

- **Resize algorithm**: Lanczos (high quality)
- **Aspect ratio preservation**: Uses Fit method, scales image to fit within specified dimensions
- **Force dimensions**: Uses Resize method, may change aspect ratio

## License

MIT License
