# Getting Started with the Chunky Library

Chunky can be embedded directly into Go services whenever you need to chunk markdown before sending it to an embedding or RAG pipeline. This guide shows how to install the library, configure a chunker, and collect the resulting chunks.

## Install

```sh
go get github.com/wyvernzora/chunky@latest
```

Import the packages you need in your Go module:

```go
import (
    "context"

    "github.com/wyvernzora/chunky/pkg/chunker"
    cctx "github.com/wyvernzora/chunky/pkg/context"
    tokenizerbuiltin "github.com/wyvernzora/chunky/pkg/tokenizer/builtin"
)
```

## Quick Chunking Example

```go
ctx := cctx.WithFileInfo(
    context.Background(),
    cctx.FileInfo{
        Path:  "docs/getting-started.md",
        Title: "Getting Started",
    },
)

tok, _ := tokenizerbuiltin.NewTiktokenTokenizer(
    tokenizerbuiltin.WithEncoding("o200k_base"),
)

c, err := chunker.New(
    chunker.WithChunkTokenBudget(1200),
    chunker.WithReservedOverheadRatio(0.1), // leave 10% for downstream usage
    chunker.WithTokenizer(tok),
)
if err != nil {
    log.Fatal(err)
}

err = c.Push(ctx, chunker.Input{
    Path:     "docs/getting-started.md",
    Title:    "Getting Started",
    Markdown: string(content), // raw markdown (front matter optional)
})
if err != nil {
    log.Fatal(err)
}

for _, chunk := range c.Chunks() {
    fmt.Printf("%s chunk %d => %d tokens\n", chunk.FilePath, chunk.ChunkIndex, chunk.Tokens)
}
```

`Chunker.Push` may be called repeatedly for multiple files. Chunks accumulate until you call `Chunker.Reset()`.

## Inputs and Context

- `chunker.Input` only needs a logical path, friendly title, and markdown string. Titles are used in generated chunk headers.
- A context flows through parsing, transforms, tokenization, and header generation. Use `pkg/context` helpers to attach structured metadata (e.g., `context.WithFileInfo`) or to provide a logger for transforms.
- Documents that contain `do_not_embed: true` in front matter are ignored automatically.

## Core Options

Configure the chunker through `chunker.With*` options:

- `WithChunkTokenBudget(int)`: hard limit (front matter + body) per chunk; required.
- `WithReservedOverheadRatio(float64)`: reserve a percentage of the budget for downstream use, effectively reducing the chunk body budget.
- `WithTokenizer(tokenizer.Tokenizer)`: swap in a word, character, or custom tokenizer (see `docs/tokenizers.md`).
- `WithParser(parser.Parser)`: use a bespoke markdown parser if the built-in AST walker does not fit.
- `WithChunkHeader(header.ChunkHeader)`: inject custom metadata/header formatting per chunk.
- `WithFrontMatterTransform` / `WithSectionTransform`: append custom transforms (see dedicated docs).

Every option can be provided multiple times; transforms run in the order they are registered. When left unspecified, Chunky defaults to tiktoken (o200k_base), YAML headers, an AST parser, and a suite of normalization transforms.

## Working with Results

`Chunker.Chunks()` returns `[]chunker.Chunk` with:

- `FilePath`, `FileTitle`, and `ChunkIndex` for routing.
- `Text`, which already contains the header plus the chunk body.
- `Tokens`, the token count used when enforcing budgets.

The `Chunker.EffectiveBudget()` helper reveals the post-overhead limit, which is useful for logging jumbo chunks.
