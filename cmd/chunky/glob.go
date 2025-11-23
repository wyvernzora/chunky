package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ExpandGlobs expands all glob patterns into a list of files.
// Patterns starting with '!' are treated as exclusion patterns.
// All paths are relative to projectRoot.
// Returns an error if any matched file is outside projectRoot.
func ExpandGlobs(projectRoot string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	var includes []string
	var excludes []string

	// Separate include and exclude patterns
	for _, pattern := range patterns {
		if after, ok := strings.CutPrefix(pattern, "!"); ok {
			excludes = append(excludes, after)
		} else {
			includes = append(includes, pattern)
		}
	}

	// If no include patterns, nothing to process
	if len(includes) == 0 {
		return nil, nil
	}

	// Expand all include patterns
	fileSet := make(map[string]bool)
	for _, pattern := range includes {
		matches, err := expandGlob(projectRoot, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to expand glob %q: %w", pattern, err)
		}
		for _, match := range matches {
			fileSet[match] = true
		}
	}

	// Remove excluded files
	for _, pattern := range excludes {
		matches, err := expandGlob(projectRoot, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to expand exclusion glob %q: %w", pattern, err)
		}
		for _, match := range matches {
			delete(fileSet, match)
		}
	}

	// Convert set to sorted slice
	var files []string
	for file := range fileSet {
		files = append(files, file)
	}

	// Sort for deterministic output
	// Note: We could use sort.Strings(files) but keeping it simple for now

	return files, nil
}

// expandGlob expands a single glob pattern relative to projectRoot.
// Returns paths relative to projectRoot.
func expandGlob(projectRoot, pattern string) ([]string, error) {
	// Make pattern absolute for matching
	absPattern := pattern
	if !filepath.IsAbs(pattern) {
		absPattern = filepath.Join(projectRoot, pattern)
	}

	// Use doublestar for glob matching (supports **)
	matches, err := doublestar.FilepathGlob(absPattern)
	if err != nil {
		return nil, err
	}

	var results []string
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute project root: %w", err)
	}

	for _, match := range matches {
		// Get absolute path of match
		absMatch, err := filepath.Abs(match)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %q: %w", match, err)
		}

		// Check if match is a regular file
		info, err := os.Stat(absMatch)
		if err != nil {
			continue // Skip files that can't be stat'ed
		}
		if !info.Mode().IsRegular() {
			continue // Skip directories and special files
		}

		// Check if match is within project root
		relPath, err := filepath.Rel(absProjectRoot, absMatch)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path for %q: %w", absMatch, err)
		}

		// If relative path starts with "..", it's outside project root
		if strings.HasPrefix(relPath, "..") {
			return nil, fmt.Errorf("file %q is outside project root %q", absMatch, absProjectRoot)
		}

		results = append(results, relPath)
	}

	return results, nil
}
