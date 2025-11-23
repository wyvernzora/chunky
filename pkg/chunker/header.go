package chunker

import (
	"context"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

// ChunkHeaderGenerator generates the header section for each chunk.
// The header typically contains serialized frontmatter but can be customized
// to include additional metadata or use different formatting.
//
// The generator receives a read-only view of the frontmatter and returns
// the header text to prepend to each chunk's body content.
//
// Example implementations:
//   - YAML frontmatter block: "---\nkey: value\n---\n\n"
//   - JSON frontmatter block: "```json\n{\"key\": \"value\"}\n```\n\n"
//   - Custom metadata format
type ChunkHeaderGenerator func(ctx context.Context, frontmatter fm.FrontMatterView) (string, error)

// defaultHeaderGenerator is the default ChunkHeaderGenerator that serializes
// frontmatter as YAML with --- delimiters.
func defaultHeaderGenerator(ctx context.Context, frontmatter fm.FrontMatterView) (string, error) {
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
