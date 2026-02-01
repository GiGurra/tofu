package table

import (
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// TruncateWithEllipsis truncates a string to maxLen characters, adding "…" if truncated.
// Returns the original string if it fits within maxLen.
// Handles unicode correctly by counting runes, not bytes.
func TruncateWithEllipsis(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	runeCount := utf8.RuneCountInString(s)
	if runeCount <= maxLen {
		return s
	}

	if maxLen == 1 {
		return "…"
	}

	// Convert to runes to handle unicode correctly
	runes := []rune(s)
	return string(runes[:maxLen-1]) + "…"
}

// ShortenPath shortens a file path to fit within maxLen characters.
// It prioritizes keeping the filename visible and shortens directory components.
// Strategy:
// 1. If path fits, return as-is
// 2. Try shortening to just filename
// 3. If filename is too long, truncate it with ellipsis
func ShortenPath(path string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	// Normalize path separators
	path = filepath.ToSlash(path)

	runeCount := utf8.RuneCountInString(path)
	if runeCount <= maxLen {
		return path
	}

	// Get just the filename
	filename := filepath.Base(path)
	filenameLen := utf8.RuneCountInString(filename)

	// If filename fits, try to add some path context
	if filenameLen <= maxLen {
		// Try to add as much directory context as possible
		dir := filepath.Dir(path)
		if dir == "." || dir == "/" {
			return filename
		}

		// Calculate available space for directory prefix
		// Format: "…/dirname/filename" or just "…/filename"
		available := maxLen - filenameLen - 2 // -2 for "…/"

		if available <= 0 {
			return filename
		}

		// Try to fit last directory component
		parts := strings.Split(dir, "/")
		lastDir := parts[len(parts)-1]
		lastDirLen := utf8.RuneCountInString(lastDir)

		if lastDirLen+1 <= available { // +1 for "/"
			return "…/" + lastDir + "/" + filename
		}

		// Just use ellipsis prefix
		return "…/" + filename
	}

	// Filename is too long, truncate it
	return TruncateWithEllipsis(filename, maxLen)
}

// PadRight pads a string to the specified width with spaces on the right.
// If the string is longer than width, it is truncated.
func PadRight(s string, width int) string {
	if width <= 0 {
		return ""
	}

	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		return TruncateWithEllipsis(s, width)
	}

	return s + strings.Repeat(" ", width-runeCount)
}

// PadLeft pads a string to the specified width with spaces on the left.
// If the string is longer than width, it is truncated.
func PadLeft(s string, width int) string {
	if width <= 0 {
		return ""
	}

	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		return TruncateWithEllipsis(s, width)
	}

	return strings.Repeat(" ", width-runeCount) + s
}

// PadCenter centers a string within the specified width.
// If the string is longer than width, it is truncated.
func PadCenter(s string, width int) string {
	if width <= 0 {
		return ""
	}

	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		return TruncateWithEllipsis(s, width)
	}

	totalPadding := width - runeCount
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	return strings.Repeat(" ", leftPadding) + s + strings.Repeat(" ", rightPadding)
}
