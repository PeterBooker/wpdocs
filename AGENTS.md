# WPDocs

Go application that parses WordPress source code and generates a Hugo documentation site. The goal is to produce developer documentation with the approachability and completeness of the Laravel docs.

## Tech Stack

- **Go 1.25.0** — module: `github.com/peter/wpdocs`
- **Hugo v0.155.3** — static site generator (config: `hugo.toml`, `[pagination]` table, `--minify` flag)
- **tree-sitter** (`github.com/smacker/go-tree-sitter`) — PHP and JS/TS AST parsing
- **Cobra** (`github.com/spf13/cobra`) — CLI framework

## Pipeline

The app runs a five-stage pipeline:

1. **Source resolution** (`internal/source/`) — accepts a local WordPress checkout or auto-clones a tag from GitHub
2. **PHP parsing** (`internal/parser/php.go`, `php_hooks.go`) — extracts functions, classes, interfaces, traits, hooks (`do_action`/`apply_filters`), and docblocks via tree-sitter
3. **JS/TS parsing** (`internal/parser/js.go`) — extracts functions, classes, interfaces, exports, and JSDoc
4. **Cross-reference resolution** (`internal/resolver/`) — links inheritance chains, hook call sites, `@see` references, method overrides
5. **Hugo generation** (`internal/output/hugo.go`) — writes versioned markdown + Hugo layouts/CSS, merges guides and overrides, runs `hugo build`

## Project Structure

```
cmd/wpdocs/main.go          CLI entry point
internal/
  model/model.go             Symbol data model + thread-safe Registry
  source/source.go           WordPress source tree resolution
  parser/                    tree-sitter extraction (PHP + JS/TS)
  resolver/resolver.go       Cross-reference linking
  output/
    output.go                Generator interface
    hugo.go                  Hugo site generation (~1,800 lines, embeds all templates)
content/
  guides/
    _shared/                 Guides that apply to all WP versions
    6.4/ … 6.8/             Version-specific guides (block-editor.md, whats-new.md)
  overrides/
    _shared/                 Symbol overrides that apply to all versions
    6.6/ … 6.8/             Version-specific overrides
docs/                        Hugo project root (generated output)
  hugo.toml                  Hugo config
  layouts/                   Hugo templates (also embedded in hugo.go)
  static/css/                Stylesheet
  data/versions.json         Tracks available WP versions
  content/{version}/         Generated per-version markdown
  public/                    Built static site
build.sh                     Builds binary, clones WP tags, generates all versions
```

## Building & Running

```bash
# Build and generate all versions defined in build.sh
bash build.sh

# Or build and run manually for a single version
go build -o ./wpdocs ./cmd/wpdocs
./wpdocs --source /path/to/wordpress --tag 6.8.1 --output ./docs
```

### CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--source` | `-s` | (auto-clone) | Local WordPress checkout path |
| `--output` | `-o` | `./docs` | Hugo output directory |
| `--tag` | `-t` | `latest` | WordPress version tag |
| `--guides` | `-g` | `./content/guides` | Guide markdown directory |
| `--overrides` | | `./content/overrides` | Override markdown directory |
| `--skip-js` | | `false` | Skip JS/TS parsing |
| `--skip-php` | | `false` | Skip PHP parsing |
| `--workers` | `-w` | `8` | Parallel parser workers |

## Content System

### Guides

Hand-written documentation pages in `content/guides/`. Files in `_shared/` apply to all versions. Version directories (e.g., `6.8/`) override or supplement shared guides.

Existing shared guides: getting-started, hooks, theme-development, database, rest-api, plugins, security, custom-post-types, caching, http-api, internationalization, users-roles, cron, wp-cli.

Version-specific: block-editor, whats-new (per version).

### Overrides

Curated content that replaces auto-generated documentation for specific symbols. Stored in `content/overrides/` with the same `_shared/` + version directory pattern. Path structure: `{_shared|version}/{kind}/{symbol-slug}.md` (e.g., `_shared/functions/add_action.md`).

## Conventions

- Follow existing directory structure; don't create new top-level directories without approval.
- Don't change dependencies without approval.
- Check sibling files for naming, structure, and patterns before creating or editing.
- Hugo layouts are embedded as Go string constants in `hugo.go` — edit there, not in `docs/layouts/` directly (they get overwritten on generate).
- Parser workers each get their own tree-sitter instance (not thread-safe). The `model.Registry` uses `sync.RWMutex` for concurrent writes.
