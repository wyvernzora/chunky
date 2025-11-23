package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jwalton/gchalk"
	"github.com/wyvernzora/chunky/pkg/chunker"
)

// sanitizeFilename replaces special characters with underscores and collapses consecutive underscores.
func sanitizeFilename(filename string) string {
	// Replace all special characters (anything that's not alphanumeric) with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := re.ReplaceAllString(filename, "_")

	// Collapse multiple consecutive underscores into one
	re = regexp.MustCompile(`_+`)
	sanitized = re.ReplaceAllString(sanitized, "_")

	// Trim leading/trailing underscores
	sanitized = strings.Trim(sanitized, "_")

	return sanitized
}

// generateChunkFilename creates a filename for a chunk based on the spec:
// - First 8 characters of SHA256 of the source file path (relative to project root, no file name)
// - File name of the source file (sanitized)
// - XXX where XXX is the 3-digit zero-padded index of chunk in the source file
// - Extension (.md)
func generateChunkFilename(chunk chunker.Chunk) string {
	// Get directory path (without filename)
	dirPath := filepath.Dir(chunk.FilePath)

	// Compute SHA256 hash of directory path
	hash := sha256.Sum256([]byte(dirPath))
	hashStr := hex.EncodeToString(hash[:])
	hashPrefix := hashStr[:8]

	// Get filename (without extension)
	filename := filepath.Base(chunk.FilePath)
	if ext := filepath.Ext(filename); ext != "" {
		filename = filename[:len(filename)-len(ext)]
	}

	// Sanitize filename
	filename = sanitizeFilename(filename)

	// Build final filename
	// Format: {hash}_{filename}.{index}.md
	return fmt.Sprintf("%s_%s.%03d.md", hashPrefix, filename, chunk.ChunkIndex)
}

// groupChunksByFile groups chunks by their source file path, preserving order.
func groupChunksByFile(chunks []chunker.Chunk) ([]string, map[string][]chunker.Chunk) {
	grouped := make(map[string][]chunker.Chunk)
	var order []string
	seen := make(map[string]bool)

	for _, chunk := range chunks {
		if !seen[chunk.FilePath] {
			order = append(order, chunk.FilePath)
			seen[chunk.FilePath] = true
		}
		grouped[chunk.FilePath] = append(grouped[chunk.FilePath], chunk)
	}

	return order, grouped
}

// printChunkOutput prints colored output to stderr showing files and their chunks.
func printChunkOutput(chunks []chunker.Chunk, effectiveBudget int) {
	order, grouped := groupChunksByFile(chunks)

	for _, filePath := range order {
		fileChunks := grouped[filePath]

		// Print source file path with inverted colors
		fmt.Fprintf(os.Stderr, " %s \n", gchalk.Bold(filePath))

		for _, chunk := range fileChunks {
			outputFilename := generateChunkFilename(chunk)

			// Determine if chunk is jumbo
			isJumbo := chunk.Tokens > effectiveBudget

			// Choose marker and color based on status
			var marker, tokenStr string
			if isJumbo {
				marker = gchalk.WithRed().WithBold().Paint("!")
				tokenStr = gchalk.WithRed().WithBold().Paint(fmt.Sprintf("%d", chunk.Tokens))
			} else {
				marker = gchalk.Green("✓")
				tokenStr = gchalk.Green(fmt.Sprintf("%d", chunk.Tokens))
			}

			// Print the line: marker (tokens) filename
			fmt.Fprintf(os.Stderr, "    %s (%s) %s\n",
				marker,
				tokenStr,
				gchalk.Dim(outputFilename),
			)
		}

		fmt.Println()
	}
}
