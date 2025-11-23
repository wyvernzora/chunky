package cli

import (
	"fmt"
	"strings"
)

// ChunkyOptions represents the unified configuration for both CLI and .chunkyrc.
// This struct is used by Kong for CLI parsing and YAML for config file parsing.
// Note: Files is separate to avoid Kong's restriction on mixing positional args with subcommands.
type ChunkyOptions struct {
	OutDir    string        `yaml:"outDir" help:"Output directory for chunks" short:"o" default:"."`
	Budget    int           `yaml:"budget" help:"Token budget per chunk" short:"b" default:"1000"`
	Overhead  float64       `yaml:"overhead" help:"Overhead fraction (0.01-0.5)" short:"e" default:"0.05"`
	Strict    bool          `yaml:"strict" help:"Fail on jumbo chunks" short:"s"`
	Tokenizer string        `yaml:"tokenizer" help:"Tokenizer (e.g., o200k_base, char, word, cl100k_base, etc.)" short:"t" default:"o200k_base"`
	Headers   []HeaderField `yaml:"headers" help:"Header fields to include" short:"H"`
	Files     []string      `yaml:"files,omitempty" json:"-" kong:"-"` // Not a CLI flag, only in config
}

// HeaderField represents a frontmatter field to include in chunk headers.
type HeaderField struct {
	Path     string `yaml:"path" help:"Frontmatter key path"`
	Label    string `yaml:"label" help:"Display label (defaults to Path if empty)"`
	Required bool   `yaml:"required" help:"If true, fail when missing"`
}

// UnmarshalText implements encoding.TextUnmarshaler for CLI flag parsing.
// Supports formats:
//   - "path" -> {Path: "path", Label: "path", Required: false}
//   - "path!" -> {Path: "path", Label: "path", Required: true}
//   - "path:Label" -> {Path: "path", Label: "Label", Required: false}
//   - "path!:Label" -> {Path: "path", Label: "Label", Required: true}
func (h *HeaderField) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return fmt.Errorf("empty header field specification")
	}

	// Check for required marker (!)
	required := false
	if strings.HasSuffix(s, "!") {
		required = true
		s = strings.TrimSuffix(s, "!")
	}

	// Check if there's a label after '!'
	// Handle cases like "path!:Label"
	if idx := strings.Index(s, "!:"); idx != -1 {
		required = true
		h.Path = s[:idx]
		h.Label = s[idx+2:]
		h.Required = required
		return nil
	}

	// Split by ':' to separate path and label
	parts := strings.SplitN(s, ":", 2)
	h.Path = strings.TrimSpace(parts[0])

	if h.Path == "" {
		return fmt.Errorf("empty path in header field specification")
	}

	if len(parts) == 2 {
		h.Label = strings.TrimSpace(parts[1])
	} else {
		h.Label = h.Path
	}

	h.Required = required

	return nil
}

// String returns a string representation of the HeaderField.
func (h HeaderField) String() string {
	req := ""
	if h.Required {
		req = "!"
	}
	if h.Label != h.Path {
		return fmt.Sprintf("%s%s:%s", h.Path, req, h.Label)
	}
	return fmt.Sprintf("%s%s", h.Path, req)
}
