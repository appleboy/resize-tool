package main

import (
	"log/slog"
	"os"
)

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
