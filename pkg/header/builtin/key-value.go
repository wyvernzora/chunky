package builtin

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/sanity-io/litter"
	"github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/header"
)

// FieldSpec defines a field to include in the key-value header.
type FieldSpec struct {
	Path     string // Frontmatter field path (e.g., "title" or "metadata.version")
	Label    string // Display label for the field
	Required bool   // Whether the field must be present and non-empty
}

// KeyValueHeaderOption configures a KeyValueHeader generator.
type KeyValueHeaderOption func(*keyValOpts)

type keyValOpts struct {
	fields []FieldSpec
}

// KeyValueHeader creates a chunk header generator that formats frontmatter
// fields as simple key-value pairs.
//
// Each field is rendered as "Label: value" on a separate line. Field values
// must be scalars (string, bool, number) or slices of scalars. Maps and
// nested structures will return an error.
//
// Required fields must be present and non-empty, or an error is returned.
// Optional fields are silently skipped if missing or empty.
//
// A field is considered empty if it's nil, an empty/whitespace string, or
// an empty slice.
//
// Example:
//
//	gen := builtin.KeyValueHeader(
//	    builtin.RequiredField("title", "Title"),
//	    builtin.OptionalField("author", "Author"),
//	    builtin.OptionalField("tags", "Tags"),
//	)
//
// Output:
//
//	Title: My Document
//	Author: John Doe
//	Tags: [tag1 tag2 tag3]
func KeyValueHeader(opts ...KeyValueHeaderOption) header.ChunkHeader {
	cfg := &keyValOpts{}
	for _, opt := range opts {
		opt(cfg)
	}
	fields := slices.Clone(cfg.fields)

	return func(_ context.Context, fm frontmatter.FrontMatterView) (string, error) {
		var b strings.Builder

		for _, f := range fields {
			val, ok := fm.Get(f.Path)

			// Validate value type (must be scalar or slice of scalars)
			if ok {
				if err := validateValue(val); err != nil {
					return "", fmt.Errorf("field %q: %w", f.Path, err)
				}
			}

			empty := isEmpty(val)

			// Check required fields
			if (!ok || empty) && f.Required {
				return "", fmt.Errorf("required field missing or empty: %s", f.Path)
			}

			// Skip missing or empty optional fields
			if !ok || empty {
				continue
			}

			// Render the field
			rendered := valueString(val)
			label := strings.TrimSpace(f.Label)
			b.WriteString(label)
			b.WriteString(": ")
			b.WriteString(rendered)
			b.WriteByte('\n')
		}

		// one blank line after header (keeps header visually distinct but cheap)
		b.WriteByte('\n')
		return b.String(), nil
	}
}

// RequiredField adds a required field to the key-value header.
// If the field is missing or empty, an error is returned.
func RequiredField(path, label string) KeyValueHeaderOption {
	return func(o *keyValOpts) {
		o.fields = append(o.fields, FieldSpec{Path: path, Label: label, Required: true})
	}
}

// OptionalField adds an optional field to the key-value header.
// If the field is missing or empty, it is silently skipped.
func OptionalField(path, label string) KeyValueHeaderOption {
	return func(o *keyValOpts) {
		o.fields = append(o.fields, FieldSpec{Path: path, Label: label, Required: false})
	}
}

var (
	// Compact, deterministic output suitable for inline headers.
	lit = litter.Options{
		Compact:           true, // one-line-ish where possible
		StripPackageNames: true, // "map[string]any" -> "map[string]any" w/o full pkgs
		HidePrivateFields: true, // keep noise down if structs slip in
		// NumWidth, DisablePointerMethods, HomePackage, etc. available if you need later
	}
	spaceRE = regexp.MustCompile(`\s+`)
)

// isScalar checks if a value is a scalar type.
func isScalar(v any) bool {
	if v == nil {
		return true
	}
	switch v.(type) {
	case string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

// validateValue ensures a value is either a scalar or a slice of scalars.
func validateValue(v any) error {
	if v == nil {
		return nil
	}

	// Check if it's a scalar
	if isScalar(v) {
		return nil
	}

	// Check if it's a slice of scalars
	if slice, ok := v.([]any); ok {
		for i, elem := range slice {
			if !isScalar(elem) {
				return fmt.Errorf("slice element at index %d is not a scalar", i)
			}
		}
		return nil
	}

	// Not a scalar or slice of scalars
	return fmt.Errorf("value must be a scalar or slice of scalars, got %T", v)
}

// isEmpty checks if a value is considered empty for header purposes.
// Assumes value has already been validated.
func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []any:
		return len(val) == 0
	default:
		return false
	}
}

// valueString returns a header-friendly string for any value.
// Strings are rendered as-is; maps/slices get a compact literal representation.
// Whitespace is collapsed so the header stays tight.
func valueString(v any) string {
	// Handle strings directly without quotes
	if s, ok := v.(string); ok {
		return s
	}

	// For other types, use litter for compact representation
	s := lit.Sdump(v)
	s = strings.TrimSpace(s)
	s = spaceRE.ReplaceAllString(s, " ")
	return s
}
