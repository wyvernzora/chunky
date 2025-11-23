package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// CLI represents the top-level command structure.
type CLI struct {
	Run  RunCmd  `cmd:"" help:"Run chunking on files"`
	Init InitCmd `cmd:"init" help:"Initialize a .chunkyrc configuration file"`
}

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
	if err := validateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Expand globs to get file list
	files, err := ExpandGlobs(projectRoot, opts.Files)
	if err != nil {
		return fmt.Errorf("failed to expand globs: %w", err)
	}

	// Sort files for deterministic output
	sort.Strings(files)

	// Print effective configuration
	printEffectiveConfig(projectRoot, opts, files)

	// TODO: Actual chunking will be implemented later
	fmt.Println("\n⚠ Chunking not yet implemented - this is a dry run")

	return nil
}

// InitCmd creates a new .chunkyrc file.
type InitCmd struct {
	ChunkyOptions

	Files []string `arg:"" optional:"" help:"File globs to include in config"`
	Force bool     `help:"Overwrite existing .chunkyrc" short:"f"`
}

// Run executes the init command.
func (i *InitCmd) Run() error {
	// Find project root (or use current directory)
	projectRoot, foundConfig, err := FindProjectRoot()
	if err != nil {
		return err
	}

	// Check if config already exists
	if foundConfig && !i.Force {
		configPath := filepath.Join(projectRoot, ConfigFileName)
		return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
	}

	// Use current directory if no config found
	if !foundConfig {
		projectRoot, err = filepath.Abs(".")
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Copy Files into ChunkyOptions for saving
	i.ChunkyOptions.Files = i.Files

	// Validate options before saving
	if err := validateOptions(&i.ChunkyOptions); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Save config
	if err := SaveConfig(projectRoot, &i.ChunkyOptions); err != nil {
		return err
	}

	configPath := filepath.Join(projectRoot, ConfigFileName)
	fmt.Printf("✓ Created configuration file at %s\n", configPath)

	return nil
}

// validateOptions validates the ChunkyOptions struct.
func validateOptions(opts *ChunkyOptions) error {
	if opts.Budget <= 0 {
		return fmt.Errorf("budget must be positive, got %d", opts.Budget)
	}

	if opts.Overhead < 0.01 || opts.Overhead > 0.5 {
		return fmt.Errorf("overhead must be in range [0.01, 0.5], got %.2f", opts.Overhead)
	}

	// Tokenizer validation removed - any string is acceptable
	// Invalid tokenizers will be rejected by tiktoken at runtime

	return nil
}

// printEffectiveConfig prints the effective configuration and file list.
func printEffectiveConfig(projectRoot string, opts *ChunkyOptions, files []string) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("EFFECTIVE CONFIGURATION")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("Project Root:  %s\n", projectRoot)
	fmt.Printf("Output Dir:    %s\n", opts.OutDir)
	fmt.Printf("Token Budget:  %d\n", opts.Budget)
	fmt.Printf("Overhead:      %.2f (%.0f%%)\n", opts.Overhead, opts.Overhead*100)
	fmt.Printf("Strict Mode:   %t\n", opts.Strict)
	fmt.Printf("Tokenizer:     %s\n", opts.Tokenizer)

	fmt.Println("\nHeader Fields:")
	if len(opts.Headers) == 0 {
		fmt.Println("  (none)")
	} else {
		for i, h := range opts.Headers {
			req := ""
			if h.Required {
				req = " [REQUIRED]"
			}
			label := h.Label
			if label == "" {
				label = h.Path
			}
			fmt.Printf("  %d. %s → %s%s\n", i+1, h.Path, label, req)
		}
	}

	fmt.Printf("\nFiles (%d total):\n", len(files))
	if len(files) == 0 {
		fmt.Println("  (none matched)")
	} else {
		for _, f := range files {
			fmt.Printf("  - %s\n", f)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}
