package main

import (
	"fmt"
	"log/slog"
	"os"
)

/*
formatFileSize returns a human-readable string for a file size in bytes (e.g., "1.2 MB").
*/
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

// setupLogger sets up slog logger based on verbose flag
func setupLogger() {
	var handler slog.Handler
	if verbose {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})
	}
	slog.SetDefault(slog.New(handler))
}
