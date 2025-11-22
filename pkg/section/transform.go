package section

import (
	"context"

	fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

// Transform is a function that performs in-place modifications on a Section.
// It receives a context for cancellation and metadata, a read-only view of
// frontmatter, and the section to transform.
//
// Transforms may modify the section's content and recursively affect children.
type Transform func(ctx context.Context, fm fm.FrontMatterView, s *Section) error

// ApplyTransform walks the section tree depth-first and applies the given transforms
// to every section in order. Stops and returns on the first error.
//
// Parameters:
//   - ctx: Context for cancellation and metadata propagation
//   - fm: FrontMatter to provide as read-only view to transforms
//   - root: Root section to start the walk from
//   - ts: Variable number of transforms to apply in order
//
// Example:
//
//	ctx := context.Background()
//	err := section.ApplyTransform(ctx, frontmatter, root, transform1, transform2)
func ApplyTransform(ctx context.Context, fm fm.FrontMatter, root *Section, ts ...Transform) error {
	return walk(ctx, fm.View(), root, ts)
}

func walk(ctx context.Context, fm fm.FrontMatterView, s *Section, ts []Transform) error {
	// Apply transforms in declared order
	for _, t := range ts {
		if err := t(ctx, fm, s); err != nil {
			return err
		}
	}

	// Recurse into children
	for _, c := range s.Children() {
		if err := walk(ctx, fm, c, ts); err != nil {
			return err
		}
	}

	return nil
}
