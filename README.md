# WPDocs

Note: This is currently just for fun testing and idea generation.

A command-line tool that parses WordPress source code (PHP and JS/TS) and generates a static developer reference site using Hugo. It extracts functions, classes, hooks, interfaces, traits, methods, and their documentation to produce a browsable site similar to developer.wordpress.org.

## How It Works

wpdocs follows a five-step pipeline:

1. **Source Resolution** — Uses a local WordPress checkout or clones a specific version from GitHub.
2. **PHP Parsing** — Extracts functions, classes, interfaces, traits, hooks, and docblocks from PHP files using tree-sitter.
3. **JS/TS Parsing** — Extracts functions, classes, interfaces, and JSDoc documentation from JavaScript and TypeScript files.
4. **Cross-Reference Resolution** — Connects symbols through inheritance chains, method overrides, hook bindings, and `@see` references.
5. **Hugo Site Generation** — Renders a complete static site with per-symbol pages, parameter tables, source context, changelog, and links to GitHub/Trac.

All parsing is done via [tree-sitter](https://tree-sitter.github.io/) for syntax-aware AST analysis rather than regex matching.

## Prerequisites

- **Go** 1.25+
- **GCC** — Required for CGo. The tree-sitter parsing library is a C library with Go bindings, so a C compiler must be installed and CGo must be enabled.
- **Hugo** — Used to build the generated static site. If Hugo is not installed, wpdocs will still generate all the Hugo source files but skip the build step.
- **PHP** — A local WordPress source tree (PHP files) is required as input. You can either point to an existing checkout or let wpdocs clone one from GitHub automatically.
- **Git** — Required if you want wpdocs to automatically clone the WordPress source.

### Installing dependencies

**Ubuntu/Debian:**

```bash
# Install GCC (required for CGo / tree-sitter)
sudo apt-get update && sudo apt-get install -y gcc

# Install Hugo (latest release from GitHub)
HUGO_VERSION=$(curl -s https://api.github.com/repos/gohugoio/hugo/releases/latest | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/')
curl -Lo hugo.deb "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-amd64.deb"
sudo dpkg -i hugo.deb
rm hugo.deb
```

**macOS:**

```bash
# Install GCC (Xcode command line tools)
xcode-select --install

# Install Hugo
brew install hugo
```

Verify that CGo is enabled (Go enables it automatically when a C compiler is found):

```bash
go env CGO_ENABLED
# Should output: 1
```

If it outputs `0`, ensure `gcc` is installed and on your `PATH`, or explicitly enable it:

```bash
export CGO_ENABLED=1
```

## Installation

```bash
go install github.com/peter/wpdocs/cmd/wpdocs@latest
```

Or build from source:

```bash
git clone https://github.com/peter/wpdocs.git
cd wpdocs
go build -o wpdocs ./cmd/wpdocs
```

## Commands

wpdocs is organized as a multi-command CLI. The general syntax is:

```
wpdocs [global flags] <command> [command flags]
```

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `./docs` | Hugo output directory |

### `generate`

Parse WordPress source and generate Hugo content for a single version.

```bash
# Generate docs from a local WordPress source tree
wpdocs generate --source /path/to/wordpress

# Auto-download the latest WordPress and generate docs
wpdocs generate

# Target a specific WordPress version
wpdocs generate --tag 6.7.1

# Specify output directory
wpdocs -o ./my-docs generate --source /path/to/wordpress

# Skip JS/TS or PHP parsing
wpdocs generate --source /path/to/wordpress --skip-js
wpdocs generate --source /path/to/wordpress --skip-php

# Control parallelism
wpdocs generate --source /path/to/wordpress --workers 16
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--source` | `-s` | *(auto-clone)* | Path to a local WordPress source tree |
| `--tag` | `-t` | `latest` | WordPress version tag (e.g. `6.7.1`) |
| `--guides` | `-g` | `./content/guides` | Path to guide markdown files |
| `--overrides` | | `./content/overrides` | Path to override markdown files |
| `--skip-js` | | `false` | Skip JavaScript/TypeScript parsing |
| `--skip-php` | | `false` | Skip PHP parsing |
| `--workers` | `-w` | `8` | Number of parallel parser workers |

### `generate-all`

Generate docs for multiple WordPress versions in one run. Versions are read from `wpdocs.toml` or passed via `--versions`.

```bash
# Generate docs for versions defined in wpdocs.toml
wpdocs generate-all

# Override versions from the command line
wpdocs generate-all --versions 6.8.1 --versions 6.7.2 --versions 6.6.2

# Minify the final build
wpdocs generate-all --minify

# Custom cache directory and parallelism
wpdocs generate-all --cache-dir /tmp/wp-cache --workers 16
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--versions` | `-v` | *(from wpdocs.toml)* | WordPress version tags |
| `--cache-dir` | | `./.wp-cache` | Directory to cache WordPress source trees |
| `--guides` | `-g` | `./content/guides` | Path to guide markdown files |
| `--overrides` | | `./content/overrides` | Path to override markdown files |
| `--workers` | `-w` | `8` | Number of parallel parser workers |
| `--minify` | | `false` | Minify the final Hugo build output |

### `build`

Run Hugo to produce the final static site from previously generated content.

```bash
wpdocs build
wpdocs build --minify
wpdocs -o ./my-docs build --minify
```

| Flag | Default | Description |
|------|---------|-------------|
| `--minify` | `false` | Minify the output |

### `serve`

Start a Hugo development server for local preview.

```bash
wpdocs serve
wpdocs serve --port 8080
wpdocs -o ./my-docs serve
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `1313` | Server port |

### `clean`

Remove all generated content and the built site.

```bash
wpdocs clean
wpdocs -o ./my-docs clean
```

This removes versioned content directories, the root `_index.md`, the `public/` build output, and `data/versions.json`.

## Configuration

For multi-version builds, create a `wpdocs.toml` in your working directory:

```toml
versions = ["6.8.1", "6.7.2", "6.6.2"]
```

This is used by `generate-all` when no `--versions` flags are provided.

## Guides and Overrides

wpdocs supports two types of supplementary content that are merged into the generated site:

- **Guides** (`./content/guides/`) — Standalone documentation pages (tutorials, conceptual docs) added alongside the auto-generated reference. Organized into `_shared/` (applies to all versions) and version-specific directories (e.g. `6.8/`).
- **Overrides** (`./content/overrides/`) — Extra content appended to individual symbol pages to supplement the auto-generated documentation. Same directory structure as guides.

## Typical Workflow

```bash
# Single version
wpdocs generate --tag 6.8.1
wpdocs build --minify
wpdocs serve

# Multiple versions
wpdocs generate-all --minify
wpdocs serve

# Clean up generated files
wpdocs clean
```

## Project Structure

```
cmd/wpdocs/          CLI entry point and commands
content/
  guides/            Guide markdown files (_shared/ and per-version)
  overrides/         Override markdown files (_shared/ and per-version)
internal/
  model/             Symbol data model and thread-safe registry
  source/            WordPress source resolution and file discovery
  parser/            Tree-sitter based PHP and JS/TS extraction
  resolver/          Cross-reference resolution (inheritance, hooks, overrides)
  output/            Hugo site generator (templates, CSS, content)
```
