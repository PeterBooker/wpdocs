# Skill: Hugo Templates

Activate this skill when modifying the site's HTML structure, layout, navigation, or styling. All templates are embedded as Go string constants in `internal/output/hugo.go` — they are **not** edited in `docs/layouts/` directly (those files get overwritten on every generate).

## Where Templates Live

Every template is a `const` string in `hugo.go`. The `Generate()` method writes them to disk at build time.

| Constant | Written to | Purpose |
|----------|-----------|---------|
| `hugoConfig` | `hugo.toml` | Site config (pagination, markup settings) |
| `layoutBaseof` | `layouts/_default/baseof.html` | HTML shell: head, sidebar + main layout |
| `layoutIndex` | `layouts/index.html` | Homepage: redirects to latest version |
| `layoutList` | `layouts/_default/list.html` | Symbol listing pages (functions, classes, etc.) |
| `layoutSingle` | `layouts/_default/single.html` | Individual symbol reference page |
| `layoutGuideList` | `layouts/guides/list.html` | Guide listing page |
| `layoutGuideSingle` | `layouts/guides/single.html` | Individual guide page |
| `partialNav` | `layouts/partials/nav.html` | Sidebar navigation + version switcher |
| `partialMeta` | `layouts/partials/meta.html` | Symbol metadata bar (since, deprecated) |
| `styleCSS` | `static/css/style.css` | All site styles |
| `symbolContentTemplate` | *(used by Go's `text/template`)* | Generates per-symbol `.md` files with YAML front matter |

## Editing Workflow

1. Find the relevant `const` in `hugo.go`
2. Edit the Go string literal (backtick-delimited raw strings)
3. Rebuild and regenerate: `go build -o ./wpdocs ./cmd/wpdocs && ./wpdocs --source <path> --tag <version> --output ./docs`
4. Preview: `hugo server --source ./docs`

## Hugo Templating Basics

Templates use Go's `html/template` syntax within Hugo's framework:

```
{{ .Title }}                       — Page title from front matter
{{ .Content }}                     — Rendered markdown body
{{ .Params.signature }}            — Custom front matter field
{{ with .Params.field }}...{{ end }}  — Conditional block (skips if empty/zero)
{{ range .Params.list }}...{{ end }}  — Iterate over a list
{{ partial "nav.html" . }}         — Include a partial template
{{ "path" | relURL }}              — Hugo function piping
```

## Front Matter Schema

The `symbolContentTemplate` generates YAML front matter for every symbol page. These fields are available in `layoutSingle` via `.Params`:

**Identity:** `symbol_kind`, `language`, `since`, `deprecated`, `access`, `summary`, `signature`

**Parameters:** `parameters` (list of `{name, type, description, default, variadic, pass_by_ref}`)

**Return:** `returns` (`{type, description}`)

**Hooks:** `hook_type`, `hook_tag`, `call_sites`

**Relationships:** `extends`, `implements`, `members`, `parent_id`, `used_by`, `uses`, `overrides`, `see_also`, `links`

**Source:** `file`, `start_line`, `end_line`, `github_url`, `trac_url`, `source_code`

**History:** `changelog` (list of `{version, description}`)

## Architecture Notes

- **Versioned content.** All content lives under `content/{version}/` (e.g., `content/6.8/functions/`). The `partialNav` resolves the current version from the URL path.
- **Version switcher.** `partialNav` builds the version list from actual content sections (`{{ range .Site.Home.Sections }}`), not just `data/versions.json`. The JS `switchVersion()` function replaces the version segment in the current URL.
- **Guides use a cascade type.** The guides section index sets `cascade: type: guides`, which routes all children through `layouts/guides/single.html` instead of `_default/single.html`.
- **Goldmark unsafe mode is on** (`hugo.toml`). Override content can include raw HTML, but `safeContent()` in Go escapes `<script>` tags and `{{ }}` delimiters before writing.
- **CSS is a single file.** All styles in `styleCSS`. Key class names: `.layout`, `.sidebar`, `.content`, `.wp-reference`, `.signature-block`, `.param-list`, `.source-code-details`, `.override-content`.
