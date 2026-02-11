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

## Usage

```bash
# Generate docs from a local WordPress source tree
wpdocs --source /path/to/wordpress

# Auto-download the latest WordPress and generate docs
wpdocs

# Target a specific WordPress version
wpdocs --tag 6.7.1

# Specify output directory
wpdocs --source /path/to/wordpress --output ./my-docs

# Skip JS/TS or PHP parsing
wpdocs --source /path/to/wordpress --skip-js
wpdocs --source /path/to/wordpress --skip-php

# Control parallelism
wpdocs --source /path/to/wordpress --workers 16
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--source` | `-s` | *(auto-download)* | Path to a local WordPress source tree |
| `--output` | `-o` | `./docs` | Output directory for the generated Hugo site |
| `--tag` | `-t` | `latest` | WordPress version tag (e.g. `6.7.1`) |
| `--skip-js` | | `false` | Skip JavaScript/TypeScript parsing |
| `--skip-php` | | `false` | Skip PHP parsing |
| `--workers` | `-w` | `8` | Number of parallel parser workers |

## Building and Serving the Site

After running `wpdocs`, a Hugo site is generated in the output directory (`./docs` by default). If Hugo is installed, the site is built automatically. You can also build and serve it manually:

```bash
# Build the static site
hugo --source ./docs

# Serve locally with live reload
hugo server --source ./docs
```

The built site will be in `./docs/public/`.

## Project Structure

```bash
cmd/wpdocs/          CLI entry point
internal/
  model/             Symbol data model and thread-safe registry
  source/            WordPress source resolution and file discovery
  parser/            Tree-sitter based PHP and JS/TS extraction
  resolver/          Cross-reference resolution (inheritance, hooks, overrides)
  output/            Hugo site generator (templates, CSS, content)
```
