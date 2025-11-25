<div align="center">
    <br>
    <br>
    <img width="256" src="docs/chunky.png">
    <h1 align="center">Chunky</h1>
</div>

<p align="center">
<b>Intelligent markdown document chunking for embedding pipelines </b>
</p>

<hr>
<br>
<br>

## Why Chunky Exists
Modern embedding and retrieval pipelines are fed by documentation sets that were never authored with token limits in mind. Most chunking scripts either slice text by byte count or split at headings without understanding front matter, resulting in chunks that lack attribution, smash unrelated sections together, or spend half their budget repeating boilerplate. Those rough edges turn into hallucinated answers, duplicated embeddings, and wasted compute.

Chunky treats documentation as structured content. It parses markdown (with YAML front matter), builds a section tree, and respects hierarchy as it normalizes, annotates, and reflows text. Tokenization happens before writing chunks, so each chunk leaves room for downstream overhead instead of guessing. The result is a deterministic set of chunks with consistent metadata that downstream systems can trust.

## How Chunky Solves the Problem
The core idea is that headings already organize related concepts, so Chunky treats the document as a tree of headings and their content. The chunker tries to keep entire subtrees together whenever possible, only splitting a heading’s content across chunks when it truly cannot fit. You can swap out tokenizers, customize transforms, or even replace the header generator. The CLI and library run the same pipeline: parse → front-matter transforms → section transforms → tokenize → greedily pack chunks under a target token budget. Every chunk begins with a header that serializes the front matter so you can always trace text back to its source file, title, or tags. The library surfaces these pieces through Go interfaces for teams that want to embed chunking into bespoke ingestion services, while the CLI provides a batteries-included workflow for static documentation repos.

### Jumbo Chunks, Reserved Overhead, Effective Budget, and Strict Mode
Chunky computes an **effective budget** by reserving a percentage of every chunk’s token budget for downstream manipulation: `effectiveBudget = budget * (1 - overhead)`. The reserved overhead is critical because embedding pipelines often decorate chunks with additional metadata, vector-store annotations, or wrapper formats during ingestion. Without that buffer, the final payload could exceed the model’s limit even if Chunky’s raw output did not.

A **jumbo chunk** is any chunk whose body exceeds the effective budget. Jumbo chunks usually originate from large contiguous blocks—code samples, tables, or multi-paragraph narratives that lack intervening headings. Because Chunky prioritizes keeping related information together, it refuses to split those blocks arbitrarily; instead it surfaces a warning and lets the downstream embedding pipeline decide whether to truncate, summarize, or split the chunk differently.

**Strict mode** (`-s/--strict`) elevates jumbo chunk warnings into hard errors. Enable it when you want CI to enforce disciplined documentation: each heading’s content should comfortably fit under the chunk budget, which produces cleaner, more uniform embeddings. Strict mode is also a reminder that better-organized documentation (with frequent headings and smaller sections) results in better chunking overall.

## CLI Workflow
```
# The following command will chunk this README file using default settings (o200k, 1k tokens per chunk, 10% overhead)
# Output files written to the `./chunks` directory.
$ chunky run -o chunks README.md
```

### Getting Started
1. Install the CLI: `go install github.com/wyvernzora/chunky/cmd/chunky@latest`.
2. Run `chunky init` in your documentation repo to scaffold `.chunkyrc`. This file captures default globs, token budget, tokenizer name, header fields, and other options so CI runs stay consistent.
3. Execute `chunky [flags] [globs...]` (or simply `chunky` if `files` are defined in `.chunkyrc`). Matching markdown files are parsed, chunked, and written to the configured output directory. Add `-d/--dry-run` when you only want preview output on stderr.
4. Inspect stderr output for jumbo chunk warnings, chunk counts per file, and the effective token budget. Adjust `.chunkyrc` or the CLI flags when you change documentation layout or target models.

### Commands
- `chunky` – main entry point; runs chunking with the current directory as project root.
- `chunky init` – writes a commented `.chunkyrc` populated with sensible defaults; rerun it only if you want a fresh template (existing files are not overwritten).

### Flags and `.chunkyrc` Options
Every CLI flag mirrors a key inside `.chunkyrc`. Flags override config on a per-run basis.

| Flag | `.chunkyrc` key | Description | Default |
| --- | --- | --- | --- |
| `-o, --out-dir <path>` | `outDir` | Directory to write chunk files. Relative paths resolve from the project root. | `.` |
| `-b, --budget <int>` | `budget` | Total token budget per chunk (header + body). Required for the library, configurable here. | `1000` |
| `-e, --overhead <ratio>` | `overhead` | Fraction of the budget reserved for downstream overhead. The chunk body budget becomes `budget * (1 - overhead)`. | `0.05` (5%) |
| `-s, --strict` | `strict` | When enabled, the run fails if any chunk exceeds the effective body budget (jumbo chunks). | `false` |
| `-t, --tokenizer <name>` | `tokenizer` | Tokenizer to use. `char` and `word` select the approximate tokenizers; any other value is treated as a tiktoken encoding (e.g., `o200k_base`, `cl100k_base`). | `o200k_base` |
| `-H, --header <spec>` | `headers` | Adds a key-value header field (see “Chunk Headers” below). Can be repeated. | *(YAML front matter dump)* |
| `-d, --dry-run` | `dryRun` | Skips writing files; prints chunk previews and stats only. Useful for tuning globs. | `false` |
| `-v, --verbose` | `verbose` | Shows the resolved configuration, project root, and the list of files before processing. | `false` |
| *(positional globs)* | `files` | File globs to include. Configure permanently via `.chunkyrc` or provide at the end of the CLI command. | none |

Example `.chunkyrc` snippet:

```yaml
outDir: chunks
budget: 1200
overhead: 0.1
tokenizer: cl100k_base
headers:
  - path: file_path
    label: Source
    required: true
files:
  - docs/**/*.md
  - guides/*.md
```

### Chunk Headers and the `-H` Flag
Each chunk starts with a header so downstream systems know where the text came from. By default Chunky serializes the entire front matter as YAML. When you pass `-H path[:Label][!]` you switch to a compact key/value header that only contains the fields you care about:

- `path` – dot-notation path inside front matter (e.g., `title`, `metadata.slug`).
- `Label` – optional display label; defaults to the path (`title:Title` prints `Title: ...`).
- `!` – mark the field as required. Missing data aborts the run (`file_path!:Source`).

You can repeat `-H` to include multiple fields. The equivalent `.chunkyrc` entry looks like:

```yaml
headers:
  - path: title
    label: Title
  - path: file_path
    label: Document
    required: true
  - path: tags
    label: Tags
```

Leave `headers` empty (or omit `-H`) to fall back to the YAML front matter block. See `docs/chunk-headers.md` for custom generators.

Library usage, advanced customization, and extension guides live under `docs/`.
