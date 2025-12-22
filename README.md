# Go Image Resize Tool

[English](./README.md) | [繁體中文](./README.zh-tw.md) | [简体中文](./README.zh-cn.md)

[![Lint and Testing](https://github.com/appleboy/resize-tool/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/resize-tool/actions/workflows/testing.yml)
[![Trivy Security Scan](https://github.com/appleboy/resize-tool/actions/workflows/trivy.yml/badge.svg)](https://github.com/appleboy/resize-tool/actions/workflows/trivy.yml)

[![Slides](https://img.shields.io/badge/Slides-SpeakerDeck-009287?logo=speakerdeck)](https://speakerdeck.com/appleboy/the-smart-choice-for-command-line-image-resizing)

![Go Image Resize Tool](./images/resize-tool.png)

A simple yet powerful image resizing tool built with Go.

## Table of Contents

- [Go Image Resize Tool](#go-image-resize-tool)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
    - [Install via Script](#install-via-script)
      - [Script Customization](#script-customization)
    - [Build from Source](#build-from-source)
    - [Direct Usage](#direct-usage)
  - [Usage](#usage)
    - [Show Help](#show-help)
    - [Basic Usage](#basic-usage)
    - [CLI Advanced Usage](#cli-advanced-usage)
  - [Parameters](#parameters)
  - [Output Filename Format](#output-filename-format)
  - [Examples](#examples)
    - [1. Batch Process Multiple Images](#1-batch-process-multiple-images)
    - [2. Website Image Optimization](#2-website-image-optimization)
    - [3. Create Thumbnails](#3-create-thumbnails)
    - [4. Other Useful Examples](#4-other-useful-examples)
  - [Supported Image Formats](#supported-image-formats)
  - [Build Instructions](#build-instructions)
  - [Performance Tips](#performance-tips)
  - [Error Handling](#error-handling)
  - [Technical Details](#technical-details)
    - [Libraries Used](#libraries-used)
    - [Image Processing Algorithms](#image-processing-algorithms)
  - [License](#license)

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

### Install via Script

You can install the latest prebuilt binary for your platform using the provided install script:

```bash
curl -fsSL https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh | bash
```

Or with wget:

```bash
wget -qO- https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh | bash
```

By default, the binary will be installed to `~/.resize-tool/bin/resize-tool` and added to your shell PATH automatically.

#### Script Customization

You can customize the installation by setting environment variables:

- `VERSION`: Install a specific version (default: latest release)
- `INSTALL_DIR`: Change the install directory (default: `~/.resize-tool/bin`)
- `CURL_INSECURE=true`: Allow insecure SSL download (not recommended)

Example:

```bash
INSTALL_DIR="$HOME/bin" VERSION="1.2.3" bash <(curl -fsSL https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh)
```

After installation, restart your terminal or run `source ~/.bashrc` (or your shell config) to update your PATH.

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
resize-tool [options] <image-file>
```

## Usage

### Show Help

```bash
resize-tool --help
```

### Basic Usage

```bash
# Default resize to 800x600 pixels
resize-tool image.jpg

# 🎯 Specify only width, height auto-calculated proportionally (recommended)
resize-tool -w 1200 image.jpg

# 🎯 Specify only height, width auto-calculated proportionally (recommended)
resize-tool --height 800 image.jpg

# Specify exact dimensions (may distort)
resize-tool -w 1200 --height 800 image.jpg

# Specify dimensions but maintain aspect ratio (fit within bounds)
resize-tool -k -w 1200 --height 800 image.jpg
```

### CLI Advanced Usage

```bash
# Set JPEG quality (1-100)
resize-tool -q 85 -w 1000 image.jpg

# Specify output directory
resize-tool -w 800 -o ./resized/ image.jpg

# Overwrite original files (no new filename with dimensions)
resize-tool -w 800 --overwrite image.jpg

# Batch process all images in directory
resize-tool -b -w 1200 /path/to/image/directory

# Batch process and overwrite original files
resize-tool -b -w 1200 --overwrite /path/to/image/directory

# Use multiple threads for batch processing
resize-tool -b --workers 8 -w 1920 /path/to/image/directory

# Verbose output mode
resize-tool -v -w 800 image.jpg

# Combine multiple options
resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## Parameters

| Parameter      | Short | Default | Description                                             |
| -------------- | ----- | ------- | ------------------------------------------------------- |
| `--width`      | `-w`  | 0       | Output width (pixels, 0=auto-calculate based on height) |
| `--height`     |       | 0       | Output height (pixels, 0=auto-calculate based on width) |
| `--quality`    | `-q`  | 95      | JPEG quality (1-100)                                    |
| `--output`     | `-o`  | same    | Output directory (default: same as input)               |
| `--keep-ratio` | `-k`  | false   | Keep aspect ratio when both width and height specified  |
| `--batch`      | `-b`  | false   | Batch process all images in directory                   |
| `--workers`    |       | 4       | Number of parallel workers for batch processing         |
| `--verbose`    | `-v`  | false   | Enable verbose output                                   |
| `--overwrite`  |       | false   | Overwrite original files instead of creating new ones   |
| `--help`       | `-h`  |         | Show help message                                       |

## Output Filename Format

Resized files will automatically include dimension information:

- Original file: `photo.jpg`
- Output file: `photo_800x600.jpg`

**Note**: When using `--overwrite`, the original file is replaced and no dimension suffix is added.

## Examples

### 1. Batch Process Multiple Images

```bash
# Process all jpg files in current directory
for img in *.jpg; do
    ./resize-tool -w 1200 "$img"
done
```

```bash
# Process all png files in current directory (height only)
for img in *.png; do
    ./resize-tool --height 800 "$img"
done
```

### 2. Website Image Optimization

```bash
# Create three different sizes (smart aspect ratio)
./resize-tool -w 1920 -q 85 -o ./large/ image.jpg
./resize-tool -w 1200 -q 85 -o ./medium/ image.jpg
./resize-tool -w 600 -q 80 -o ./small/ image.jpg
```

### 3. Create Thumbnails

```bash
# Create square thumbnails (fixed size, may crop)
./resize-tool -w 300 --height 300 -o ./thumbnails/ image.jpg

# Create thumbnails (keep aspect ratio, max 300x300)
./resize-tool -w 300 --height 300 -k -o ./thumbnails/ image.jpg
```

### 4. Other Useful Examples

```bash
# Specify only width, height auto-calculated
./resize-tool -w 1200 image.jpg

# Specify only height, width auto-calculated
./resize-tool --height 800 image.jpg

# Specify both width and height (may distort)
./resize-tool -w 1200 --height 800 image.jpg

# Specify both width and height, keep aspect ratio (fit within bounds)
./resize-tool -k -w 1200 --height 800 image.jpg

# Set JPEG quality
./resize-tool -q 85 -w 1000 image.jpg

# Specify output directory
./resize-tool -w 800 -o ./resized/ image.jpg

# Overwrite original file (replace in place)
./resize-tool -w 800 --overwrite image.jpg

# Batch process all images in directory
./resize-tool -b -w 1200 /path/to/image/directory

# Batch process and overwrite original files
./resize-tool -b -w 1200 --overwrite /path/to/image/directory

# Batch process with multiple threads
./resize-tool -b --workers 8 -w 1920 /path/to/image/directory

# Verbose output mode
./resize-tool -v -w 800 image.jpg

# Combine multiple options (Note: --overwrite cannot be used with --output)
./resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## Supported Image Formats

- **Input formats**: JPEG, PNG, GIF, TIFF, BMP
- **Output formats**: Same as input format

## Build Instructions

To build the tool from source:

```bash
go build -o resize-tool .
```

For cross-compiling (multi-platform builds):

```bash
make release  # Build for multiple platforms
```

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
