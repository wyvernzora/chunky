package builtin

import (
	"context"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
	"github.com/wyvernzora/chunky/pkg/header"
)

// FrontMatterYamlHeader creates a chunk header generator that serializes
// frontmatter as YAML with --- delimiters.
//
// Output format:
//
//	---
//	title: My Document
//	author: John Doe
//	---
//
// If the frontmatter is empty, returns an empty string.
func FrontMatterYamlHeader() header.ChunkHeader {
	return func(ctx context.Context, frontmatter fm.FrontMatterView) (string, error) {
		// Convert view to map for serialization
		fmMap := frontmatter.AsMap()
		if len(fmMap) == 0 {
			return "", nil
		}

		yamlBlock, err := fm.Serialize(fmMap)
		if err != nil {
			return "", err
		}

		// Add trailing newline after closing ---
		if len(yamlBlock) > 0 && yamlBlock[len(yamlBlock)-1] != '\n' {
			yamlBlock += "\n"
		}

		return yamlBlock, nil
	}
}
