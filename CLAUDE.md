# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build the application
make build

# Run tests
make test

# Run a single test
go test -v -run TestFunctionName ./...

# Run linter
make lint

# Format code
make format

# Install dependencies
make deps

# Build release versions (cross-platform)
make release

# Clean build files
make clean
```

## Architecture Overview

This is a Go CLI tool for resizing images. The codebase is organized as follows:

- **main.go**: Entry point, sets up Cobra root command
- **config.go**: CLI flag definitions and validation logic (`setupConfig`, `validateConfig`)
- **processor.go**: Core image processing (`resizeImage`, `calculateTargetSize`, `generateOutputPath`)
- **batch.go**: Directory batch processing with worker pool (`processBatch`, `collectImageFiles`)
- **utils.go**: Logger setup
- **version.go**: Version information (injected at build time via ldflags)

### Key Dependencies

- `github.com/disintegration/imaging` - Image processing (Lanczos algorithm)
- `github.com/spf13/cobra` - CLI framework
- `github.com/appleboy/com` - Utility library (file size formatting)

### Processing Flow

1. CLI parses flags in `config.go`, validates in `validateConfig`
2. `processImages` determines single file vs batch mode
3. For batch: `processBatch` uses a worker pool with goroutines
4. For single: `resizeImage` handles opening, resizing, and saving
5. Aspect ratio logic is handled by `widthSet`/`heightSet` flags tracking which dimensions were explicitly set

### Supported Image Formats

Input/Output: JPEG, PNG, GIF, TIFF, BMP
