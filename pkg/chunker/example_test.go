package chunker_test

import (
	"context"
	"fmt"
	"log"

	"github.com/wyvernzora/chunky/pkg/chunker"
	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	fmbuiltin "github.com/wyvernzora/chunky/pkg/frontmatter/builtin"
)

func ExampleChunker() {
	// Create a new chunker with a 1000 token budget
	c, err := chunker.New(
		chunker.WithChunkTokenBudget(1000),
		chunker.WithReservedOverheadRatio(0.1), // Reserve 10% for overhead
	)
	if err != nil {
		log.Fatal(err)
	}

	// Push a markdown document
	err = c.Push(context.Background(), chunker.Input{
		Path:  "docs/guide.md",
		Title: "Getting Started Guide",
		Markdown: `---
author: John Doe
---

# Introduction

Welcome to our comprehensive guide on getting started with the platform.

## Prerequisites

Before you begin, ensure you have the following:

- A valid account
- Basic knowledge of markdown
- Internet connection

## Installation

Follow these steps to install the software:

1. Download the installer
2. Run the installation wizard
3. Configure your settings

## Configuration

After installation, you'll need to configure:

### Database Settings

Set up your database connection string in the config file.

### API Keys

Generate API keys from the admin panel.

## Conclusion

You're now ready to start using the platform!
`,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Get all chunks
	chunks := c.Chunks()

	// Display chunk information
	fmt.Printf("Generated %d chunk(s)\n", len(chunks))
	for _, chunk := range chunks {
		fmt.Printf("\nChunk %d (%s):\n", chunk.ChunkIndex, chunk.FilePath)
		fmt.Printf("  Title: %s\n", chunk.FileTitle)
		fmt.Printf("  Tokens: %d\n", chunk.Tokens)
		fmt.Printf("  Text length: %d bytes\n", len(chunk.Text))
	}

	// Output:
	// Generated 1 chunk(s)
	//
	// Chunk 1 (docs/guide.md):
	//   Title: Getting Started Guide
	//   Tokens: 244
	//   Text length: 1235 bytes
}

func ExampleChunker_withTransforms() {
	// Create a frontmatter transform that adds default metadata
	addMetadata := fmbuiltin.MergeFrontMatter(fm.FrontMatter{
		"timestamp":   "2025-11-22",
		"environment": "production",
	})

	// Create chunker with transforms
	c, err := chunker.New(
		chunker.WithChunkTokenBudget(500),
		chunker.WithFrontMatterTransform(addMetadata),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Push(context.Background(), chunker.Input{
		Path:     "docs/api.md",
		Title:    "API Reference",
		Markdown: "# API\n\nOur API provides endpoints for managing resources.",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated %d chunk(s)\n", len(c.Chunks()))
	// Output:
	// Generated 1 chunk(s)
}

func ExampleChunker_Reset() {
	c, err := chunker.New(
		chunker.WithChunkTokenBudget(1000),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Process first batch
	c.Push(context.Background(), chunker.Input{
		Path:     "doc1.md",
		Title:    "Document 1",
		Markdown: "# Document 1\n\nContent here.",
	})

	fmt.Printf("Batch 1: %d chunks\n", len(c.Chunks()))

	// Reset and process second batch
	c.Reset()

	c.Push(context.Background(), chunker.Input{
		Path:     "doc2.md",
		Title:    "Document 2",
		Markdown: "# Document 2\n\nDifferent content.",
	})

	fmt.Printf("Batch 2: %d chunks\n", len(c.Chunks()))

	// Output:
	// Batch 1: 1 chunks
	// Batch 2: 1 chunks
}
