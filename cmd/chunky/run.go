package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/wyvernzora/chunky/pkg/chunker"
)

// RunCmd is the main command that processes files.
type RunCmd struct {
	ChunkyOptions

	Files []string `arg:"" optional:"" help:"File globs to process"`
}

// Run executes the main chunking command.
func (r *RunCmd) Run() error {
	// Copy Files into ChunkyOptions for processing
	r.ChunkyOptions.Files = r.Files
	// Find project root
	projectRoot, foundConfig, err := FindProjectRoot()
	if err != nil {
		return err
	}

	// Load config if found
	var configOpts *ChunkyOptions
	if foundConfig {
		configOpts, err = LoadConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		fmt.Printf("✓ Loaded configuration from %s\n", filepath.Join(projectRoot, ConfigFileName))
	} else {
		configOpts = &ChunkyOptions{}
		fmt.Printf("⚠ No .chunkyrc found, using defaults and CLI flags\n")
	}

	// Merge CLI options with config
	opts := MergeOptions(configOpts, &r.ChunkyOptions)

	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Expand globs to get file list
	files, err := ExpandGlobs(projectRoot, opts.Files)
	if err != nil {
		return fmt.Errorf("failed to expand globs: %w", err)
	}

	// Sort files for deterministic output
	sort.Strings(files)

	// Print effective configuration only in verbose mode
	if opts.Verbose {
		opts.Print(projectRoot, files)
	}

	// Create tokenizer
	tok, err := createTokenizer(opts.Tokenizer)
	if err != nil {
		return fmt.Errorf("failed to create tokenizer: %w", err)
	}

	// Create header generator
	headerGen := createHeaderGenerator(opts.Headers)

	// Create chunker
	c, err := chunker.New(
		chunker.WithChunkTokenBudget(opts.Budget),
		chunker.WithReservedOverheadRatio(opts.Overhead),
		chunker.WithTokenizer(tok),
		chunker.WithChunkHeader(headerGen),
	)
	if err != nil {
		return fmt.Errorf("failed to create chunker: %w", err)
	}

	// Process all files
	ctx := context.Background()
	if opts.Verbose {
		fmt.Println("\nProcessing files...")
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
			if err := processFile(ctx, projectRoot, file, c); err != nil {
				return fmt.Errorf("error processing %s: %w", file, err)
			}
		}
	} else {
		for _, file := range files {
			if err := processFile(ctx, projectRoot, file, c); err != nil {
				return fmt.Errorf("error processing %s: %w", file, err)
			}
		}
	}

	// Get chunks
	chunks := c.Chunks()
	effectiveBudget := c.EffectiveBudget()

	// Check for jumbo chunks if strict mode is enabled
	var jumboChunks []chunker.Chunk
	for _, chunk := range chunks {
		if chunk.Tokens > effectiveBudget {
			jumboChunks = append(jumboChunks, chunk)
		}
	}

	if len(jumboChunks) > 0 {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "\n⚠ Warning: Found %d jumbo chunk(s) exceeding effective budget of %d tokens:\n", len(jumboChunks), effectiveBudget)
			for _, chunk := range jumboChunks {
				fmt.Fprintf(os.Stderr, "  - %s (chunk %d): %d tokens\n", chunk.FilePath, chunk.ChunkIndex, chunk.Tokens)
			}
		}
		if opts.Strict {
			return fmt.Errorf("strict mode enabled: aborting due to jumbo chunks")
		}
	}

	// Print chunk output to stderr
	printChunkOutput(chunks, effectiveBudget)

	// Skip file writes if in dry run mode
	if opts.DryRun {
		return nil
	}

	// Write files
	absOutDir := opts.OutDir
	if !filepath.IsAbs(absOutDir) {
		absOutDir = filepath.Join(projectRoot, absOutDir)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(absOutDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, chunk := range chunks {
		filename := generateChunkFilename(chunk)
		outPath := filepath.Join(absOutDir, filename)

		if err := os.WriteFile(outPath, []byte(chunk.Text), 0644); err != nil {
			return fmt.Errorf("failed to write chunk file %s: %w", filename, err)
		}
	}

	return nil
}
