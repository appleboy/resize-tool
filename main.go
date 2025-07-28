package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "resize-tool [image-file-or-directory]",
	Short:   "A powerful image resizing tool",
	Version: getVersion(),
	Long: `A command-line tool to resize images with various options.
Supports JPEG, PNG, GIF, TIFF, and BMP formats.
Can process single files or batch process directories.

By default, if only width or height is specified, the other dimension
will be calculated automatically to maintain aspect ratio.

Example usage:
		resize-tool input.jpg --width 800
		resize-tool input.jpg --height 600
		resize-tool images/ --batch --width 1024 --output resized/
		resize-tool input.jpg --width 800 --overwrite
`,
	Args: cobra.ExactArgs(1),
	Run:  processImages,
}

// Entry point of the application
func main() {
	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute command", "err", err)
		os.Exit(1)
	}
}

func init() {
	// Set up configuration
	setupConfig()
}
