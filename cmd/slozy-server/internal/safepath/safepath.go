// Package safepath provides utilities for safe filesystem path handling.
package safepath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Windows reserved filenames (case-insensitive).
var reservedFilenames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true,
	"COM5": true, "COM6": true, "COM7": true, "COM8": true,
	"COM9": true, "LPT1": true, "LPT2": true, "LPT3": true,
	"LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true,
	"LPT8": true, "LPT9": true,
}

// SanitizeFilename creates a safe filename from user-provided components.
// It normalizes Unicode, removes path separators, filters unsafe characters,
// checks reserved Windows names, and truncates to maxLen.
func SanitizeFilename(components ...string) string {
	parts := make([]string, 0, len(components))
	for _, c := range components {
		// Normalize Unicode
		normalized := norm.NFC.String(c)
		// Clean and get base name (removes any path components)
		clean := filepath.Base(filepath.Clean(normalized))

		// Check Windows reserved names
		if reservedFilenames[strings.ToUpper(clean)] {
			clean = "_" + clean
		}

		// Filter to only safe characters
		safe := strings.Map(func(r rune) rune {
			if r == 0 {
				return -1
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
				return r
			}
			return '_'
		}, clean)

		// Apply length limit
		if len(safe) > 100 {
			safe = safe[:100]
		}

		parts = append(parts, safe)
	}
	return strings.Join(parts, "-")
}

// ValidateInDir checks that a file path is within the given directory.
// Returns the resolved absolute path and an error if validation fails.
func ValidateInDir(dir, filePath string) (string, error) {
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		resolvedDir = filepath.Clean(dir)
	}

	cleanPath := filepath.Clean(filePath)
	absPath := filepath.Join(resolvedDir, cleanPath)

	// Ensure the resolved path is within the base directory
	if !strings.HasPrefix(absPath, resolvedDir+string(os.PathSeparator)) && absPath != resolvedDir {
		return "", fmt.Errorf("attempted to access file outside directory: %s", filePath)
	}

	// Check for path traversal sequences
	if strings.Contains(filePath, "..") || strings.Contains(filePath, "\\") {
		return "", fmt.Errorf("invalid file path: %s", filePath)
	}

	return absPath, nil
}

// HasYAMLExtension checks if the filename has a .yaml or .yml extension.
func HasYAMLExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".yaml" || ext == ".yml"
}
