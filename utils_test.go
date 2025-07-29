//go:build !integration

package main

import (
	"testing"
)

// TestFormatFileSize covers various byte sizes and expected text.
func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1500, "1.5 KB"},
		{10 * 1024, "10.0 KB"},
		{1024*1024 - 1, "1024.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{5*1024*1024 + 123, "5.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 5, "5.0 GB"},
		{1234567890, "1.1 GB"},
	}
	for _, tt := range tests {
		got := formatFileSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("formatFileSize(%d) = %q; want %q", tt.bytes, got, tt.expected)
		}
	}
}

// Note: setupLogger depends on global 'verbose' and changes global slog config,
// so is not tested here. Test it only if refactored to make side effects testable.
