package header

import (
	"context"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

// ChunkHeader generates the header section for each chunk.
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
type ChunkHeader func(ctx context.Context, frontmatter fm.FrontMatterView) (string, error)
