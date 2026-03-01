# Contributing to WPDocs

## Quick Start

```bash
# Build the binary
go build -o ./wpdocs ./cmd/wpdocs

# Generate docs for a single WordPress version
./wpdocs generate --tag 6.8.1

# Or with a local WordPress checkout
./wpdocs generate --source /path/to/wordpress --tag 6.8.1

# Build the Hugo site
./wpdocs build

# Preview with the dev server
./wpdocs serve

# Generate all versions and build (reads from wpdocs.toml)
./wpdocs generate-all --minify

# Clean generated files
./wpdocs clean

# Run tests
go test ./...
```

If you don't have a local WordPress checkout, omit `--source` and the tool will auto-clone the tag from GitHub.

## CLI Commands

| Command | Description |
|---------|-------------|
| `wpdocs generate` | Parse WP source + write Hugo content for one version |
| `wpdocs generate-all` | Generate docs for all versions in `wpdocs.toml` |
| `wpdocs build` | Run Hugo to produce the final static site |
| `wpdocs serve` | Start Hugo dev server for preview |
| `wpdocs clean` | Remove generated content and `public/` |

All commands accept `--output` (`-o`) to specify the Hugo output directory (default: `./docs`).

Run `wpdocs <command> --help` for full flag details.

## Architecture

The app runs a five-stage pipeline:

1. **Source resolution** (`internal/source/`) — accepts a local WordPress checkout or auto-clones a tag from GitHub.
2. **PHP parsing** (`internal/parser/php.go`, `php_hooks.go`) — extracts functions, classes, interfaces, traits, hooks, and docblocks via tree-sitter.
3. **JS/TS parsing** (`internal/parser/js.go`) — extracts functions, classes, interfaces, exports, and JSDoc.
4. **Cross-reference resolution** (`internal/resolver/`) — links inheritance chains, hook call sites, `@see` references, method overrides.
5. **Hugo generation** (`internal/output/`) — writes versioned markdown + Hugo layouts/CSS, merges guides and overrides.

The `generate` command runs steps 1-5. The `build` command invokes Hugo separately. The `generate-all` command loops through all configured versions, then runs a final Hugo build.

## Hugo Templates

Templates, CSS, config, and other Hugo assets live in `internal/output/embed/`:

```
internal/output/embed/
  hugo.toml                          Hugo site config
  layouts/_default/baseof.html       HTML shell (sidebar + content)
  layouts/_default/list.html         Section listing pages
  layouts/_default/single.html       Individual symbol reference page
  layouts/index.html                 Homepage (redirects to latest version)
  layouts/guides/list.html           Guide cards grid
  layouts/guides/single.html         Guide page with TOC and pager
  layouts/partials/nav.html          Sidebar navigation + version switcher
  layouts/partials/meta.html         Metadata badges (kind, language, etc.)
  static/css/style.css               Stylesheet
  templates/symbol.md.tmpl           Go text/template for symbol markdown
```

These files are embedded into the binary at compile time via Go's `embed.FS` (see `internal/output/embed.go`). When you edit a template, you must rebuild the binary for changes to take effect.

For rapid iteration on templates:
1. Edit files in `internal/output/embed/` directly
2. Run `go build -o ./wpdocs ./cmd/wpdocs && ./wpdocs generate --tag 6.8.1`
3. Preview with `./wpdocs serve`

The `symbol.md.tmpl` template uses Go's `text/template` syntax with custom functions (`yamlEscape`, `yamlMultiline`, `safeContent`) to generate per-symbol markdown with YAML front matter.

## Output Package Structure

The Hugo output code is split by responsibility:

| File | Purpose |
|------|---------|
| `hugo.go` | `Hugo` struct, `Generate()` orchestrator, `Build()`, version data management |
| `embed.go` | `embed.FS` declaration, scaffold writer that copies embedded files to output |
| `content.go` | Symbol page generation, signature builder, changelog parser |
| `guides.go` | Guide merging (`_shared/` + version overlay), override reader |
| `helpers.go` | Pure utilities: YAML escaping, slug generation, version comparison, URL builders |

## Content System

### Guides

Hand-written documentation in `content/guides/`. Files in `_shared/` apply to all WordPress versions. Version-specific directories (e.g., `6.8/`) override or supplement shared guides with the same filename.

### Overrides

Curated content that replaces auto-generated documentation for specific symbols. Stored in `content/overrides/` with the same `_shared/` + version directory pattern.

Path structure: `{_shared|version}/{kind}/{symbol-slug}.md`

Example: `_shared/functions/add_action.md` replaces the auto-generated description for `add_action` across all versions.

## Adding a New Symbol Kind

1. Add the constant in `internal/model/model.go` (e.g., `KindWidget SymbolKind = "widget"`).
2. Extract it in the parser (`internal/parser/`).
3. Add an entry to the `kindSections` slice in `internal/output/hugo.go`'s `Generate()` method.
4. The existing layouts and CSS handle new kinds automatically via the generic list/single templates.

## Running Tests

```bash
# All tests
go test ./...

# Output package only (helpers + content builders)
go test ./internal/output/...

# With verbose output
go test -v ./internal/output/...
```

## Multi-Version Builds

Configure versions in `wpdocs.toml`:

```toml
versions = ["6.8.1", "6.7.2", "6.6.2", "6.5.5", "6.4.5"]
```

Then run:

```bash
./wpdocs generate-all --minify
```

This clones each WordPress tag (cached in `.wp-cache/`), generates content for each version, then runs `hugo --minify`. The final static site is in `docs/public/`.

You can also override versions on the command line:

```bash
./wpdocs generate-all --versions 6.8.1,6.7.2
```

## Conventions

- Follow existing directory structure; don't create new top-level directories without approval.
- Don't change dependencies without approval.
- Check sibling files for naming, structure, and patterns before creating or editing.
- Hugo layouts are embedded in `internal/output/embed/` — edit there, not in `docs/layouts/` directly (they get overwritten on generate).
- Parser workers each get their own tree-sitter instance (not thread-safe). The `model.Registry` uses `sync.RWMutex` for concurrent writes.
