# Chunk Headers

Every chunk emitted by Chunky begins with a header that serializes front matter so downstream systems can attribute each chunk to its source. The header is produced by a `header.ChunkHeader` generator.

## Default Behavior

The library and CLI default to `builtin.FrontMatterYamlHeader()`, which renders:

```yaml
---
title: User Guide
file_path: docs/guide.md
tags:
  - how-to
---

<chunk body...>
```

Token counts include both the header and body, so wide headers eat into your budget.

## Header Generator Interface

```go
type ChunkHeader func(ctx context.Context, fm frontmatter.FrontMatterView) (string, error)
```

- Receives read-only front matter (`FrontMatterView`).
- Can look at context metadata (file info, logger).
- Returns the text prepended to every chunk body.

## Built-In Generators

Located in `pkg/header/builtin`:

- `FrontMatterYamlHeader()`: canonical YAML block with `---` delimiters.
- `KeyValueHeader(opts...)`: plain-text key/value pairs; combine with `RequiredField` / `OptionalField` helpers to control which fields appear.

CLI flag examples:

- `chunky -H title:Title -H tags:Tags` → key/value header with Title and Tags.
- `chunky -H file_path!:Source` → same header but `file_path` is required; missing data aborts processing.

In `.chunkyrc`:

```yaml
headers:
  - path: file_path
    label: Document
    required: true
  - path: tags
    label: Tags
```

Leaving `headers` empty reverts to YAML serialization.

## Custom Generators

```go
import (
    "bytes"
    "text/template"

    "github.com/wyvernzora/chunky/pkg/chunker"
    fm "github.com/wyvernzora/chunky/pkg/frontmatter"
)

tmpl, err := template.New("chunk-header").Parse(`+++ 
title = "{{ .FM.title }}"
path  = "{{ .FM.file_path }}"
+++

`)
if err != nil {
    log.Fatal(err)
}

generator := func(ctx context.Context, view fm.FrontMatterView) (string, error) {
    buf := &bytes.Buffer{}
    data := struct{ FM fm.FrontMatterView }{FM: view}
    if err := tmpl.Execute(buf, data); err != nil {
        return "", err
    }
    return buf.String(), nil
}

c, err := chunker.New(
    chunker.WithChunkTokenBudget(1000),
    chunker.WithChunkHeader(generator),
)
```

Use `frontmatter.ViewCopy` if you need to convert the read-only view to a map. Keep the header short and deterministic; token estimations rely on being able to count it through the selected tokenizer.
