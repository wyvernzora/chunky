# Custom Front Matter Transforms

Front matter transforms allow you to inject, validate, or reshape document metadata before chunk headers are generated. They run after the markdown parser extracts YAML front matter and before section transforms are applied.

## Anatomy of a Transform

The function signature defined in `pkg/frontmatter` is:

```go
type Transform func(ctx context.Context, fm frontmatter.FrontMatter) error
```

- `ctx` carries `pkg/context` metadata (file info, logger, etc.).
- `fm` is a mutable `map[string]any`. Mutations are reflected in downstream steps.
- Returning an error aborts processing for the current document; the CLI will surface the failure.

## Writing a Transform

```go
func RequireTags(min int) frontmatter.Transform {
    return func(ctx context.Context, fm frontmatter.FrontMatter) error {
        tags, _ := fm["tags"].([]any)
        if len(tags) < min {
            info, _ := cctx.FileInfoFrom(ctx)
            return fmt.Errorf("%s needs at least %d tags", info.Path, min)
        }
        return nil
    }
}

func InjectCollection(name string) frontmatter.Transform {
    return func(ctx context.Context, fm frontmatter.FrontMatter) error {
        if _, ok := fm["collection"]; !ok {
            fm["collection"] = name
        }
        return nil
    }
}
```

Transforms can also inspect the original file contents (available through `cctx.MustFileInfo(ctx).Content`) or log warnings with `cctx.Logger(ctx)`.

## Registering Transforms

Pass transforms when constructing the chunker:

```go
c, err := chunker.New(
    chunker.WithChunkTokenBudget(1024),
    chunker.WithFrontMatterTransforms(
        builtin.InjectFilePath("file_path"), // keep defaults if you need them
        RequireTags(2),
        InjectCollection("guides"),
    ),
)
```

Transforms are executed in the order provided. The CLI ships with `InjectFilePath` by default; if you build your own CLI or service, make sure you re-register any defaults you rely on.

## Testing Transforms

- Create standalone unit tests by invoking the transform with a fake context and in-memory front matter map.
- Use `frontmatter.Serialize` during debugging to inspect the full metadata payload.
- Combine transforms cautiously: a late-stage transform may override metadata produced earlier; keep ordering predictable and document assumptions.
