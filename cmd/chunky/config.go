package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".chunkyrc"

// FindProjectRoot searches for .chunkyrc starting from the current directory
// and walking up the directory tree. Returns the directory containing .chunkyrc,
// or the current directory if not found.
func FindProjectRoot() (string, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := cwd
	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			// Found .chunkyrc
			return dir, true, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding config
			return cwd, false, nil
		}
		dir = parent
	}
}

// LoadConfig loads the .chunkyrc file from the given directory.
// Returns nil if the file doesn't exist.
func LoadConfig(projectRoot string) (*ChunkyOptions, error) {
	configPath := filepath.Join(projectRoot, ConfigFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var opts ChunkyOptions
	if err := yaml.Unmarshal(data, &opts); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &opts, nil
}

// SaveConfig writes a ChunkyOptions struct to a .chunkyrc file in the given directory.
func SaveConfig(projectRoot string, opts *ChunkyOptions) error {
	configPath := filepath.Join(projectRoot, ConfigFileName)

	// Use the opts directly - Files are now included in config
	data, err := yaml.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Add a comment header
	header := "# Chunky configuration file\n# See https://github.com/wyvernzora/chunky for documentation\n\n"
	data = append([]byte(header), data...)

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// MergeOptions merges CLI options into config options.
// CLI options take precedence over config options.
// The merging logic is:
//   - If a CLI value is explicitly set (non-zero/non-default), it overrides config
//   - For slices, CLI values are appended to config values
func MergeOptions(config, cli *ChunkyOptions) *ChunkyOptions {
	result := &ChunkyOptions{}

	// Files: append CLI files to config files
	result.Files = append(result.Files, config.Files...)
	result.Files = append(result.Files, cli.Files...)

	// OutDir: CLI takes precedence if set
	if cli.OutDir != "" && cli.OutDir != "." {
		result.OutDir = cli.OutDir
	} else if config.OutDir != "" {
		result.OutDir = config.OutDir
	} else {
		result.OutDir = "."
	}

	// Budget: CLI takes precedence if not default
	if cli.Budget != 0 && cli.Budget != 1000 {
		result.Budget = cli.Budget
	} else if config.Budget != 0 {
		result.Budget = config.Budget
	} else {
		result.Budget = 1000
	}

	// Overhead: CLI takes precedence if not default
	if cli.Overhead != 0 && cli.Overhead != 0.05 {
		result.Overhead = cli.Overhead
	} else if config.Overhead != 0 {
		result.Overhead = config.Overhead
	} else {
		result.Overhead = 0.05
	}

	// Strict: CLI takes precedence if set
	if cli.Strict {
		result.Strict = true
	} else {
		result.Strict = config.Strict
	}

	// Tokenizer: CLI takes precedence if not default
	if cli.Tokenizer != "" && cli.Tokenizer != "o200k_base" {
		result.Tokenizer = cli.Tokenizer
	} else if config.Tokenizer != "" {
		result.Tokenizer = config.Tokenizer
	} else {
		result.Tokenizer = "o200k_base"
	}

	// Headers: append CLI headers to config headers
	result.Headers = append(result.Headers, config.Headers...)
	result.Headers = append(result.Headers, cli.Headers...)

	return result
}
