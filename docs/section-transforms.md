# Custom Section Transforms

Section transforms operate on the parsed markdown tree after front matter is settled but before tokenization. They are perfect for normalizing whitespace, annotating headings, or injecting synthetic content into specific sections.

## Interface

```go
type Transform func(
    ctx context.Context,
    fm frontmatter.FrontMatter,
    s *section.Section,
) error
```

- `s` is a mutable node in the section tree. Use getters/setters from `pkg/section` to inspect or alter content.
- `fm` lets you coordinate with metadata (e.g., skip sections when `do_not_embed: true`).
- Return an error to halt processing of the current document.

Transforms are run depth-first on the tree. Built-ins normalize newlines, collapse blank lines, prefix headings, and append HTML comments describing section paths.

## Building a Transform

```go
func InjectBreadcrumbs() section.Transform {
    return func(ctx context.Context, fm frontmatter.FrontMatter, s *section.Section) error {
        if s.Level() == 0 {
            return nil // skip the synthetic root
        }
        breadcrumb := strings.Join(s.Path(), " › ")
        annotated := fmt.Sprintf("<!-- %s -->\n%s", breadcrumb, strings.TrimLeft(s.Content(), "\n"))
        s.SetContent(annotated)
        return nil
    }
}

func StripHiddenSections() section.Transform {
    return func(ctx context.Context, fm frontmatter.FrontMatter, s *section.Section) error {
        if strings.Contains(s.Title(), "[hidden]") {
            s.SetContent("")  // remove body
            s.ClearChildren() // drop nested sections
        }
        return nil
    }
}
```

Use helpers such as `s.Children()`, `s.CreateChild`, `s.ReplaceContent`, and `s.Path()` to walk or mutate the tree. When you need to inspect file metadata, pull it from `pkg/context` via `cctx.MustFileInfo(ctx)`.

## Registering Transforms

```go
c, err := chunker.New(
    chunker.WithChunkTokenBudget(1200),
    chunker.WithSectionTransforms(
        builtin.NormalizeNewlinesTransform(),
        StripHiddenSections(),
        InjectBreadcrumbs(),
    ),
)
```

Ordering is crucial: run normalization transforms first, then structural transforms, then any annotation or logging steps. The CLI registers a rich default set (normalize, trim, heading annotations); when embedding the library you may append your own or replace the list entirely.
