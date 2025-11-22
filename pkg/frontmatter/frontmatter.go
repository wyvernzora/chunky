package frontmatter

import "encoding/json"

// FrontMatter is a map of YAML frontmatter key-value pairs extracted from
// a Markdown document header.
type FrontMatter map[string]any

// EmptyFrontMatter creates a new empty FrontMatter map.
func EmptyFrontMatter() FrontMatter { return make(FrontMatter) }

// Clone performs a deep clone of the FrontMatter using JSON serialization.
func (fm FrontMatter) Clone() FrontMatter {
	if fm == nil {
		return EmptyFrontMatter()
	}
	b, _ := json.Marshal(fm)
	var out FrontMatter
	_ = json.Unmarshal(b, &out)
	if out == nil {
		out = EmptyFrontMatter()
	}
	return out
}

// ---- Read-only View ----

// FrontMatterView provides a read-only interface to frontmatter data.
// All returned values are deep copies to prevent mutation of the underlying data.
type FrontMatterView interface {
	// Get retrieves a value by key, returning a deep copy.
	// The second return value indicates whether the key exists.
	Get(key string) (any, bool)

	// Keys returns a slice of all frontmatter keys.
	Keys() []string

	// AsMap returns a deep copy of the entire frontmatter as a map.
	AsMap() map[string]any
}

// roFrontMatter is the wrapper type that implements FrontMatterView.
type roFrontMatter struct {
	m FrontMatter
}

func (ro roFrontMatter) Get(key string) (any, bool) {
	v, ok := ro.m[key]
	return deepCopyJSON(v), ok
}

func (ro roFrontMatter) Keys() []string {
	keys := make([]string, 0, len(ro.m))
	for k := range ro.m {
		keys = append(keys, k)
	}
	return keys
}

func (ro roFrontMatter) AsMap() map[string]any {
	return ro.m.Clone()
}

// View returns a read-only interface over a deep-copied snapshot of the frontmatter.
func (fm FrontMatter) View() FrontMatterView {
	return roFrontMatter{m: fm.Clone()}
}

// deepCopyJSON is a helper that deep copies a value using JSON serialization.
func deepCopyJSON(v any) any {
	b, _ := json.Marshal(v)
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}
