package frontmatter

import "context"

// Transform is a function that performs in-place mutations on FrontMatter.
// It receives a context for cancellation and metadata, and the frontmatter
// to transform.
//
// Transforms may add, modify, or remove keys from the frontmatter map.
type Transform func(ctx context.Context, fm FrontMatter) error

// ApplyTransform applies the given transforms to the frontmatter in order.
// Each transform is applied sequentially, and the function stops and returns
// on the first error encountered.
//
// Parameters:
//   - ctx: Context for cancellation and metadata propagation
//   - fm: FrontMatter to transform in-place
//   - ts: Variable number of transforms to apply in order
//
// Returns:
//   - error: The first error encountered, or nil if all transforms succeed
//
// Example:
//
//	ctx := context.Background()
//	fm := frontmatter.FrontMatter{"title": "My Doc"}
//	err := frontmatter.ApplyTransform(ctx, fm, transform1, transform2)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ApplyTransform(ctx context.Context, fm FrontMatter, ts ...Transform) error {
	for _, t := range ts {
		if err := t(ctx, fm); err != nil {
			return err
		}
	}
	return nil
}
