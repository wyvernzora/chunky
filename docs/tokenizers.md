# Tokenizers

Tokenizers determine how Chunky measures the size of each section and chunk. Accurate token counts are critical when you target strict embedding limits.

## Interfaces

```go
type Tokenizer interface {
    Count(text string) (int, error)
    Tokenize(ctx context.Context, root *section.Section) (*tokenizer.TokenizedSection, error)
}
```

- `Count` is used for quick estimates (e.g., header sizes).
- `Tokenize` annotates each section in the tree with `ContentTokens` and `SubtreeTokens`, enabling the greedy chunking algorithm.

## Built-In Tokenizers

All implementations live in `pkg/tokenizer/builtin`:

| Name | CLI Flag (`-t/--tokenizer`) | Notes |
| ---- | --------------------------- | ----- |
| `TiktokenTokenizer` | any tiktoken encoding (default `o200k_base`) | Uses `tiktoken-go`; best accuracy for OpenAI-style models. |
| `WordCountTokenizer` | `word` | Approximates tokens via words-per-token ratio (default 0.75). |
| `CharacterCountTokenizer` | `char` | Approximates via characters-per-token (default 4). |

When using the CLI, pass `-t char`, `-t word`, or any encoding accepted by tiktoken such as `cl100k_base` or `text-embedding-3-large`. In `.chunkyrc`, set `tokenizer: cl100k_base`.

## Custom Tokenizers

Implement the interface when you need a proprietary estimator:

```go
type MyTokenizer struct { ... }

func (t *MyTokenizer) Count(text string) (int, error) {
    return mylib.CountTokens(text), nil
}

func (t *MyTokenizer) Tokenize(ctx context.Context, root *section.Section) (*tokenizer.TokenizedSection, error) {
    return tokenizer.TokenizeTree(ctx, root, t.Count)
}

c, err := chunker.New(
    chunker.WithChunkTokenBudget(1500),
    chunker.WithTokenizer(&MyTokenizer{}),
)
```

You can wrap external APIs inside the tokenizer, but keep in mind that tokenization happens for every section in every file—favor deterministic, fast implementations.

## Troubleshooting

- Call `tok.Count` yourself to validate expectations on representative snippets before chunking entire docs.
- If tokenization fails, Chunky aborts the document. Surface the error upstream to decide whether to retry with `Strict` disabled or fall back to a simpler tokenizer.
- When approximating (`word` or `char`), leave extra overhead using `WithReservedOverheadRatio` or the CLI `--overhead` flag to reduce jumbo chunks.
