package parser

// FrontMatter is a named map type for YAML-style metadata.
type FrontMatter map[string]any

// EmptyFrontMatter returns a new, initialized FrontMatter.
func EmptyFrontMatter() FrontMatter { return make(FrontMatter) }
