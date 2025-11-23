package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wyvernzora/chunky/pkg/chunker"
	"github.com/wyvernzora/chunky/pkg/header"
	headerBuiltin "github.com/wyvernzora/chunky/pkg/header/builtin"
	"github.com/wyvernzora/chunky/pkg/tokenizer"
	tokenizerBuiltin "github.com/wyvernzora/chunky/pkg/tokenizer/builtin"
)

// createTokenizer creates a tokenizer based on the tokenizer option.
func createTokenizer(tokenizerName string) (tokenizer.Tokenizer, error) {
	switch tokenizerName {
	case "char":
		return tokenizerBuiltin.NewCharCountTokenizer(), nil
	case "word":
		return tokenizerBuiltin.NewWordCountTokenizer(), nil
	default:
		// Assume it's a tiktoken encoding name
		tok, err := tokenizerBuiltin.NewTiktokenTokenizer(tokenizerBuiltin.WithEncoding(tokenizerName))
		if err != nil {
			return nil, fmt.Errorf("failed to create tiktoken tokenizer with encoding %q: %w", tokenizerName, err)
		}
		return tok, nil
	}
}

// createHeaderGenerator creates a header generator based on the headers option.
func createHeaderGenerator(headers []HeaderField) header.ChunkHeader {
	if len(headers) == 0 {
		// No headers specified, use YAML frontmatter
		return headerBuiltin.FrontMatterYamlHeader()
	}

	// Use key-value header with specified fields
	var opts []headerBuiltin.KeyValueHeaderOption
	for _, h := range headers {
		if h.Required {
			opts = append(opts, headerBuiltin.RequiredField(h.Path, h.Label))
		} else {
			opts = append(opts, headerBuiltin.OptionalField(h.Path, h.Label))
		}
	}
	return headerBuiltin.KeyValueHeader(opts...)
}

// processFile processes a single markdown file and returns the chunks.
func processFile(ctx context.Context, projectRoot, filePath string, c chunker.Chunker) error {
	// Construct absolute path
	absPath := filepath.Join(projectRoot, filePath)

	// Read file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Extract title from filename (without extension)
	title := filepath.Base(filePath)
	if ext := filepath.Ext(title); ext != "" {
		title = title[:len(title)-len(ext)]
	}

	// Push to chunker
	input := chunker.Input{
		Path:     filePath,
		Title:    title,
		Markdown: string(content),
	}

	if err := c.Push(ctx, input); err != nil {
		return fmt.Errorf("failed to process file: %w", err)
	}

	return nil
}
