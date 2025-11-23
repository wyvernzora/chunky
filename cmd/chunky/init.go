package main

import (
	"fmt"
	"path/filepath"
)

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
	if err := (&i.ChunkyOptions).Validate(); err != nil {
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
