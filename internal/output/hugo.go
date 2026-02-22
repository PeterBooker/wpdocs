package output

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/peter/wpdocs/internal/model"
)

// Hugo generates a complete Hugo static site from the symbol registry.
type Hugo struct {
	outDir       string
	srcRoot      string // path to WordPress source tree
	wpVersion    string // full version e.g. "6.7.1"
	version      string // normalized major.minor e.g. "6.7"
	guidesDir    string // optional path to hand-written guide markdown files
	overridesDir string // optional path to override markdown files
}

// NewHugo creates a Hugo site generator that writes to outDir.
func NewHugo(outDir, srcRoot, wpVersion, guidesDir, overridesDir string) *Hugo {
	return &Hugo{
		outDir:       outDir,
		srcRoot:      srcRoot,
		wpVersion:    wpVersion,
		version:      normalizeVersion(wpVersion),
		guidesDir:    guidesDir,
		overridesDir: overridesDir,
	}
}

// normalizeVersion extracts major.minor from a full version string like "6.7.1".
func normalizeVersion(v string) string {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

func (h *Hugo) Generate(reg *model.Registry) error {
	// Clean only this version's content directory (preserves other versions)
	versionDir := filepath.Join(h.outDir, "content", h.version)
	_ = os.RemoveAll(versionDir)

	// Create directory structure
	dirs := []string{
		filepath.Join(h.outDir, "content", h.version),
		filepath.Join(h.outDir, "data"),
		filepath.Join(h.outDir, "layouts", "_default"),
		filepath.Join(h.outDir, "layouts", "guides"),
		filepath.Join(h.outDir, "layouts", "partials"),
		filepath.Join(h.outDir, "static", "css"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	// Write Hugo config
	if err := h.writeFile("hugo.toml", hugoConfig); err != nil {
		return fmt.Errorf("writing hugo.toml: %w", err)
	}

	// Write layouts
	layoutFiles := map[string]string{
		filepath.Join("layouts", "_default", "baseof.html"):  layoutBaseof,
		filepath.Join("layouts", "_default", "list.html"):    layoutList,
		filepath.Join("layouts", "_default", "single.html"):  layoutSingle,
		filepath.Join("layouts", "index.html"):               layoutIndex,
		filepath.Join("layouts", "guides", "list.html"):      layoutGuideList,
		filepath.Join("layouts", "guides", "single.html"):    layoutGuideSingle,
		filepath.Join("layouts", "partials", "nav.html"):     partialNav,
		filepath.Join("layouts", "partials", "meta.html"):    partialMeta,
	}
	for path, content := range layoutFiles {
		if err := h.writeFile(path, content); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}

	// Write CSS
	if err := h.writeFile(filepath.Join("static", "css", "style.css"), styleCSS); err != nil {
		return fmt.Errorf("writing style.css: %w", err)
	}

	// Update versions data file and write homepage
	if err := h.updateVersionsData(); err != nil {
		return fmt.Errorf("updating versions data: %w", err)
	}
	if err := h.writeFile(filepath.Join("content", "_index.md"), "---\ntitle: WordPress Developer Reference\n---\n"); err != nil {
		return fmt.Errorf("writing homepage: %w", err)
	}

	// Write version landing page
	versionIndex := fmt.Sprintf("---\ntitle: \"WordPress %s Reference\"\nversion: %q\n---\n", h.wpVersion, h.version)
	if err := h.writeFile(filepath.Join("content", h.version, "_index.md"), versionIndex); err != nil {
		return fmt.Errorf("writing version index: %w", err)
	}

	// Generate content by kind (under versioned path)
	kindSections := []struct {
		kind    model.SymbolKind
		section string
		title   string
	}{
		{model.KindFunction, "functions", "Functions"},
		{model.KindClass, "classes", "Classes"},
		{model.KindMethod, "methods", "Methods"},
		{model.KindHook, "hooks", "Hooks"},
		{model.KindInterface, "interfaces", "Interfaces"},
		{model.KindTrait, "traits", "Traits"},
		{model.KindEnum, "enums", "Enums"},
		{model.KindComponent, "components", "Components"},
	}

	for _, ks := range kindSections {
		symbols := reg.ByKind(ks.kind)
		if len(symbols) == 0 {
			continue
		}

		sectionDir := filepath.Join(h.outDir, "content", h.version, ks.section)
		if err := os.MkdirAll(sectionDir, 0o755); err != nil {
			return fmt.Errorf("creating section dir %s: %w", ks.section, err)
		}

		// Section index
		sectionIndex := fmt.Sprintf("---\ntitle: %q\n---\n", ks.title)
		if err := h.writeFile(filepath.Join("content", h.version, ks.section, "_index.md"), sectionIndex); err != nil {
			return fmt.Errorf("writing %s section index: %w", ks.section, err)
		}

		// Sort symbols alphabetically
		sorted := make([]*model.Symbol, len(symbols))
		copy(sorted, symbols)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})

		// Individual symbol pages
		for _, sym := range sorted {
			if err := h.writeSymbolPage(ks.section, sym); err != nil {
				return fmt.Errorf("writing symbol %s: %w", sym.ID, err)
			}
		}
	}

	// Write guides (if guides directory provided)
	if err := h.writeGuides(); err != nil {
		return fmt.Errorf("writing guides: %w", err)
	}

	// Run hugo build
	h.runHugoBuild()

	return nil
}

func (h *Hugo) writeFile(relPath, content string) error {
	absPath := filepath.Join(h.outDir, relPath)
	return os.WriteFile(absPath, []byte(content), 0o644)
}

func (h *Hugo) writeSymbolPage(section string, sym *model.Symbol) error {
	slug := symbolSlug(sym.ID)
	relPath := filepath.Join("content", h.version, section, slug+".md")
	absPath := filepath.Join(h.outDir, relPath)

	f, err := os.Create(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := symbolPageData{
		Symbol:      sym,
		Signature:   buildSignature(sym),
		Changelog:   parseChangelog(sym),
		SourceCode:  h.readSourceContext(sym.Location.File, sym.Location.StartLine),
		GitHubURL:   h.buildGitHubURL(sym.Location.File, sym.Location.StartLine, sym.Location.EndLine),
		TracURL:     h.buildTracURL(sym.Location.File, sym.Location.StartLine),
		OverrideContent: h.readOverride(section, slug),
	}

	tmpl := template.Must(template.New("symbol").Funcs(template.FuncMap{
		"yamlEscape":    yamlEscape,
		"yamlMultiline": yamlMultiline,
		"join":          strings.Join,
		"safeContent":   safeContent,
	}).Parse(symbolContentTemplate))

	return tmpl.Execute(f, data)
}

func (h *Hugo) runHugoBuild() {
	hugoPath, err := exec.LookPath("hugo")
	if err != nil {
		log.Printf("Hugo not found in PATH; skipping build. Install Hugo and run: hugo --source %s", h.outDir)
		return
	}

	absDir, err := filepath.Abs(h.outDir)
	if err != nil {
		log.Printf("Warning: could not resolve absolute path: %v", err)
		absDir = h.outDir
	}

	cmd := exec.Command(hugoPath, "--source", absDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: Hugo build failed: %v", err)
		log.Printf("You can manually build with: hugo --source %s", absDir)
	}
}

// versionsData represents the data/versions.json file.
type versionsData struct {
	All    []string `json:"all"`
	Latest string   `json:"latest"`
}

// updateVersionsData reads the existing versions.json, adds the current version, and writes it back.
func (h *Hugo) updateVersionsData() error {
	dataPath := filepath.Join(h.outDir, "data", "versions.json")

	var data versionsData
	if raw, err := os.ReadFile(dataPath); err == nil {
		_ = json.Unmarshal(raw, &data)
	}

	// Add current version if not already present
	found := false
	for _, v := range data.All {
		if v == h.version {
			found = true
			break
		}
	}
	if !found {
		data.All = append(data.All, h.version)
	}

	// Sort versions descending (newest first) using simple string compare
	sort.Slice(data.All, func(i, j int) bool {
		return data.All[i] > data.All[j]
	})

	// Latest is always the highest version
	if len(data.All) > 0 {
		data.Latest = data.All[0]
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataPath, raw, 0o644)
}

// writeGuides merges guide markdown files from _shared/ and {version}/ into the
// versioned content directory. Version-specific files override _shared/ files with
// the same name.
func (h *Hugo) writeGuides() error {
	if h.guidesDir == "" {
		return nil
	}

	// Collect guides: _shared first, then version-specific overrides
	guides := h.collectContentFiles(h.guidesDir)
	if len(guides) == 0 {
		return nil
	}

	guidesContentDir := filepath.Join(h.outDir, "content", h.version, "guides")
	if err := os.MkdirAll(guidesContentDir, 0o755); err != nil {
		return err
	}

	// Write guides section index with cascade type
	guidesIndex := "---\ntitle: \"Guides\"\ncascade:\n  type: guides\n---\n"
	if err := h.writeFile(filepath.Join("content", h.version, "guides", "_index.md"), guidesIndex); err != nil {
		return err
	}

	weight := 1
	// Sort by filename for deterministic order
	names := make([]string, 0, len(guides))
	for name := range guides {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		srcPath := guides[name]

		content, err := os.ReadFile(srcPath)
		if err != nil {
			log.Printf("Warning: could not read guide %s: %v", name, err)
			continue
		}

		body := string(content)

		// If the file already has front matter, use it as-is (cascade type applies)
		if !strings.HasPrefix(body, "---") {
			title := strings.TrimSuffix(name, ".md")
			title = strings.ReplaceAll(title, "-", " ")
			words := strings.Fields(title)
			for i, w := range words {
				if len(w) > 0 {
					words[i] = strings.ToUpper(w[:1]) + w[1:]
				}
			}
			title = strings.Join(words, " ")
			body = fmt.Sprintf("---\ntitle: %q\nweight: %d\n---\n\n%s", title, weight, body)
		}

		outPath := filepath.Join("content", h.version, "guides", name)
		if err := h.writeFile(outPath, body); err != nil {
			log.Printf("Warning: could not write guide %s: %v", name, err)
			continue
		}
		weight++
	}

	return nil
}

// collectContentFiles builds a map of filename → absolute path by reading _shared/
// first, then overlaying version-specific files. Returns only .md files.
func (h *Hugo) collectContentFiles(baseDir string) map[string]string {
	result := make(map[string]string)

	// Read _shared/ first
	sharedDir := filepath.Join(baseDir, "_shared")
	if entries, err := os.ReadDir(sharedDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				result[e.Name()] = filepath.Join(sharedDir, e.Name())
			}
		}
	}

	// Overlay version-specific (wins over _shared)
	versionDir := filepath.Join(baseDir, h.version)
	if entries, err := os.ReadDir(versionDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				result[e.Name()] = filepath.Join(versionDir, e.Name())
			}
		}
	}

	return result
}

// readOverride reads an optional override markdown file for a symbol page.
// Checks version-specific directory first, then falls back to _shared/.
func (h *Hugo) readOverride(section, slug string) string {
	if h.overridesDir == "" {
		return ""
	}

	// Version-specific override wins
	versionPath := filepath.Join(h.overridesDir, h.version, section, slug+".md")
	if data, err := os.ReadFile(versionPath); err == nil {
		return string(data)
	}

	// Fall back to _shared
	sharedPath := filepath.Join(h.overridesDir, "_shared", section, slug+".md")
	if data, err := os.ReadFile(sharedPath); err == nil {
		return string(data)
	}

	return ""
}

func symbolSlug(id string) string {
	r := strings.NewReplacer(
		"::", ".",
		"\\", ".",
		"/", ".",
		" ", "-",
		"$", "",
		"(", "",
		")", "",
		"{", "",
		"}", "",
	)
	slug := r.Replace(strings.ToLower(id))
	// Remove any remaining characters Hugo can't handle in filenames
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return -1
	}, slug)
	if slug == "" {
		slug = "unnamed"
	}
	return slug
}

// safeContent escapes HTML script tags and Hugo template delimiters in content
// that will be written into the markdown body of a Hugo page. Without this,
// literal <script> tags from WordPress docblocks get rendered as real HTML
// (due to Goldmark unsafe mode) and {{ }} from Underscore/Mustache templates
// get interpreted as Hugo template expressions.
func safeContent(s string) string {
	// Escape <script> and </script> (case-insensitive) to prevent raw HTML injection
	r := strings.NewReplacer(
		"<script", "&lt;script",
		"</script", "&lt;/script",
		"<SCRIPT", "&lt;SCRIPT",
		"</SCRIPT", "&lt;/SCRIPT",
		"<Script", "&lt;Script",
		"</Script", "&lt;/Script",
	)
	s = r.Replace(s)

	// Escape {{ and }} to prevent Hugo template evaluation
	s = strings.ReplaceAll(s, "{{", "&#123;&#123;")
	s = strings.ReplaceAll(s, "}}", "&#125;&#125;")

	return s
}

// yamlEscape quotes a string for safe YAML embedding.
func yamlEscape(s string) string {
	if s == "" {
		return `""`
	}
	// Quote if contains special characters
	if strings.ContainsAny(s, ":#{}[]|>&*!%@`'\"\n\\") {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return `"` + s + `"`
}

// changelogEntry represents one row in the Changelog table.
type changelogEntry struct {
	Version     string
	Description string
}

// symbolPageData wraps a Symbol with computed fields for the content template.
type symbolPageData struct {
	*model.Symbol
	Signature       string
	Changelog       []changelogEntry
	SourceCode      string
	GitHubURL       string
	TracURL         string
	OverrideContent string
}

// buildSignature constructs a code signature string like the WP developer reference.
func buildSignature(sym *model.Symbol) string {
	switch sym.Kind {
	case model.KindFunction, model.KindMethod:
		var b strings.Builder
		b.WriteString(sym.Name)
		b.WriteString("( ")
		for i, p := range sym.Params {
			if i > 0 {
				b.WriteString(", ")
			}
			if p.Type != "" {
				b.WriteString(p.Type)
				b.WriteString(" ")
			}
			if p.IsPassByRef {
				b.WriteString("&")
			}
			b.WriteString("$")
			b.WriteString(p.Name)
			if p.Default != "" {
				b.WriteString(" = ")
				b.WriteString(p.Default)
			}
		}
		b.WriteString(" )")
		if sym.Returns != nil && sym.Returns.Type != "" {
			b.WriteString(": ")
			b.WriteString(sym.Returns.Type)
		}
		return b.String()

	case model.KindHook:
		var b strings.Builder
		if sym.HookType == model.HookAction {
			b.WriteString("do_action( '")
		} else {
			b.WriteString("apply_filters( '")
		}
		b.WriteString(sym.HookTag)
		b.WriteString("'")
		for _, p := range sym.Params {
			b.WriteString(", ")
			if p.Type != "" {
				b.WriteString(p.Type)
				b.WriteString(" ")
			}
			b.WriteString("$")
			b.WriteString(p.Name)
		}
		b.WriteString(" )")
		return b.String()

	case model.KindClass, model.KindInterface, model.KindTrait, model.KindEnum:
		var b strings.Builder
		b.WriteString(string(sym.Kind))
		b.WriteString(" ")
		b.WriteString(sym.Name)
		if len(sym.Extends) > 0 {
			b.WriteString(" extends ")
			b.WriteString(strings.Join(sym.Extends, ", "))
		}
		if len(sym.Implements) > 0 {
			b.WriteString(" implements ")
			b.WriteString(strings.Join(sym.Implements, ", "))
		}
		return b.String()

	default:
		return sym.Name
	}
}

// parseChangelog extracts changelog entries from @since tags.
func parseChangelog(sym *model.Symbol) []changelogEntry {
	sinceEntries := sym.Doc.Tags["since"]
	if len(sinceEntries) == 0 && sym.Doc.Since != "" {
		return []changelogEntry{{Version: sym.Doc.Since, Description: "Introduced."}}
	}
	var entries []changelogEntry
	for _, entry := range sinceEntries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, " ", 2)
		ce := changelogEntry{Version: parts[0]}
		if len(parts) > 1 {
			ce.Description = parts[1]
		} else {
			ce.Description = "Introduced."
		}
		entries = append(entries, ce)
	}
	if len(entries) == 0 && sym.Doc.Since != "" {
		entries = []changelogEntry{{Version: sym.Doc.Since, Description: "Introduced."}}
	}
	return entries
}

// readSourceContext reads ±5 lines around startLine from the source file.
func (h *Hugo) readSourceContext(file string, startLine int) string {
	if h.srcRoot == "" {
		return ""
	}
	absPath := filepath.Join(h.srcRoot, file)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	start := max(startLine-6, 0)   // 5 lines before (0-indexed)
	end := min(startLine+5, len(lines)) // 5 lines after
	snippet := strings.Join(lines[start:end], "\n")
	// YAML literal blocks forbid tab characters; convert to spaces
	return strings.ReplaceAll(snippet, "\t", "    ")
}

// buildGitHubURL returns a GitHub source link for the given file and line range.
func (h *Hugo) buildGitHubURL(file string, startLine, endLine int) string {
	tag := h.wpVersion
	if tag == "" || tag == "unknown" {
		tag = "master"
	}
	return fmt.Sprintf("https://github.com/WordPress/WordPress/blob/%s/%s#L%d-L%d",
		tag, file, startLine, endLine)
}

// buildTracURL returns a Trac browser link for the given file and line.
func (h *Hugo) buildTracURL(file string, startLine int) string {
	tag := h.wpVersion
	if tag == "" || tag == "unknown" {
		return fmt.Sprintf("https://core.trac.wordpress.org/browser/trunk/%s#L%d", file, startLine)
	}
	return fmt.Sprintf("https://core.trac.wordpress.org/browser/tags/%s/%s#L%d",
		tag, file, startLine)
}

// yamlMultiline formats a multi-line string as a YAML double-quoted scalar
// with newlines escaped as \n. This avoids YAML parsing issues that literal
// block scalars (|) can trigger with source code containing #, {, : etc.
func yamlMultiline(s string) string {
	if s == "" {
		return `""`
	}
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// --- Hugo config ---

const hugoConfig = `baseURL = "/"
languageCode = "en-us"
title = "WordPress Developer Reference"

[pagination]
  pagerSize = 200

[markup]
  [markup.goldmark]
    [markup.goldmark.renderer]
      unsafe = true

[markup.tableOfContents]
  startLevel = 2
  endLevel = 4
`

// --- Layout templates ---

const layoutBaseof = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ if not .IsHome }}{{ .Title }} &ndash; {{ end }}{{ .Site.Title }}</title>
  <link rel="stylesheet" href="{{ "css/style.css" | relURL }}">
</head>
<body>
  <div class="layout">
    <aside class="sidebar">
      {{ partial "nav.html" . }}
    </aside>
    <main class="content">
      {{ block "main" . }}{{ end }}
    </main>
  </div>
</body>
</html>
`

const layoutIndex = `{{ define "main" }}
{{ with .Site.Data.versions }}
  {{ with .latest }}
  <script>window.location.href = '/{{ . }}/';</script>
  {{ end }}
{{ end }}

<h1>{{ .Site.Title }}</h1>
<p>Select a version:</p>
<ul class="version-list">
{{ with .Site.Data.versions }}
  {{ range .all }}
  <li><a href="/{{ . }}/">WordPress {{ . }}</a></li>
  {{ end }}
{{ end }}
</ul>
{{ end }}
`

const layoutList = `{{ define "main" }}
{{ $isVersionRoot := and .Params.version (eq (len .Sections) (len .Sections)) }}
{{ $parts := split (strings.TrimPrefix "/" .RelPermalink) "/" }}
{{ $depth := len (where $parts "." "!=" "") }}

{{ if le $depth 1 }}
{{/* Version landing page */}}
<h1>{{ .Title }}</h1>

{{ $guidesSection := .GetPage "guides" }}
{{ with $guidesSection }}
  {{ if .Pages }}
  <section class="guides-overview">
    <h2>Guides</h2>
    <div class="guide-cards">
      {{ range .Pages.ByWeight }}
      <a href="{{ .RelPermalink }}" class="guide-card">
        <h3>{{ .Title }}</h3>
        {{ with .Params.summary }}<p>{{ . }}</p>{{ end }}
      </a>
      {{ end }}
    </div>
  </section>
  {{ end }}
{{ end }}

<section class="reference-overview">
  <h2>Reference</h2>
  <div class="stats-grid">
    {{ $refSections := slice "functions" "classes" "methods" "hooks" "interfaces" "traits" "enums" "components" }}
    {{ range $refSections }}
      {{ $sec := $.GetPage . }}
      {{ with $sec }}
        {{ if .Pages }}
        <div class="stat-card">
          <div class="stat-number">{{ len .Pages }}</div>
          <div class="stat-label"><a href="{{ .RelPermalink }}">{{ .Title }}</a></div>
        </div>
        {{ end }}
      {{ end }}
    {{ end }}
  </div>
</section>

{{ else }}
{{/* Section listing page (functions, classes, etc.) */}}
<h1>{{ .Title }}</h1>
<p class="count">{{ len .Pages }} items</p>

<table class="listing">
  <thead>
    <tr>
      <th>Name</th>
      <th>Summary</th>
      <th>Since</th>
    </tr>
  </thead>
  <tbody>
    {{ range .Pages.ByTitle }}
    <tr{{ if .Params.deprecated }} class="deprecated-row"{{ end }}>
      <td>
        <a href="{{ .RelPermalink }}">{{ .Title }}</a>
        {{ with .Params.deprecated }}<span class="badge deprecated">Deprecated</span>{{ end }}
      </td>
      <td>{{ .Params.summary }}</td>
      <td class="since">{{ .Params.since }}</td>
    </tr>
    {{ end }}
  </tbody>
</table>
{{ end }}
{{ end }}
`

const layoutSingle = `{{ define "main" }}
<article class="wp-reference">

<h1>{{ .Title }}</h1>

{{ partial "meta.html" . }}

{{ with .Params.deprecated }}
<div class="deprecated-notice">
  <strong>This {{ $.Params.symbol_kind }} has been deprecated.</strong> {{ . }}
</div>
{{ end }}

{{ with .Params.signature }}
<section class="signature-section">
  <pre class="signature-block"><code>{{ . }}</code></pre>
</section>
{{ end }}

<section class="description-section">
  <h2>Description</h2>
  {{ with .Params.summary }}<p class="summary">{{ . }}</p>{{ end }}
  {{ with .Content }}<div class="long-description">{{ . }}</div>{{ end }}
  {{ with .Params.see_also }}
  <h3>See also</h3>
  <ul>{{ range . }}<li><code>{{ . }}</code></li>{{ end }}</ul>
  {{ end }}
  {{ with .Params.links }}
  <ul class="doc-links">{{ range . }}<li><a href="{{ . }}">{{ . }}</a></li>{{ end }}</ul>
  {{ end }}
</section>

{{ with .Params.parameters }}
<section class="parameters-section">
  <h2>Parameters</h2>
  <dl class="param-list">
    {{ range . }}
    <dt>
      <code>${{ .name }}</code>
      <span class="param-type"><code>{{ .type }}</code></span>
      {{ if .variadic }}<span class="param-tag">variadic</span>{{ end }}
      {{ if .pass_by_ref }}<span class="param-tag">by&nbsp;ref</span>{{ end }}
    </dt>
    <dd>
      {{ .description }}
      {{ with .default }}<p class="param-default">Default: <code>{{ . }}</code></p>{{ end }}
    </dd>
    {{ end }}
  </dl>
</section>
{{ end }}

{{ with .Params.returns }}{{ if .type }}
<section class="return-section">
  <h2>Return</h2>
  <p><code class="return-type">{{ .type }}</code> {{ .description }}</p>
</section>
{{ end }}{{ end }}

{{ if .Params.hook_tag }}
<section class="hook-section">
  <h2>Hook Details</h2>
  <p>Type: <strong>{{ .Params.hook_type }}</strong></p>
  <p>Tag: <code>{{ .Params.hook_tag }}</code></p>
  {{ with .Params.call_sites }}
  <h3>Fired from</h3>
  <ul>{{ range . }}<li><code>{{ . }}</code></li>{{ end }}</ul>
  {{ end }}
</section>
{{ end }}

{{ with .Params.members }}
<section>
  <h2>Members</h2>
  <ul class="member-list">{{ range . }}<li><code>{{ . }}</code></li>{{ end }}</ul>
</section>
{{ end }}

{{ with .Params.extends }}
<section>
  <h2>Extends</h2>
  <ul>{{ range . }}<li><code>{{ . }}</code></li>{{ end }}</ul>
</section>
{{ end }}

{{ with .Params.implements }}
<section>
  <h2>Implements</h2>
  <ul>{{ range . }}<li><code>{{ . }}</code></li>{{ end }}</ul>
</section>
{{ end }}

{{ if .Params.file }}
<section class="source-section">
  <h2>Source</h2>
  <p class="source-file">File: <code>{{ .Params.file }}</code>, lines {{ .Params.start_line }}&ndash;{{ .Params.end_line }}</p>
  <div class="source-links">
    {{ with .Params.github_url }}<a href="{{ . }}">View on GitHub</a>{{ end }}
    {{ with .Params.trac_url }}<a href="{{ . }}">View on Trac</a>{{ end }}
  </div>
  {{ with .Params.source_code }}
  <details class="source-code-details">
    <summary>Show source</summary>
    <pre class="source-code"><code>{{ . }}</code></pre>
  </details>
  {{ end }}
</section>
{{ end }}

{{ if or .Params.uses .Params.used_by }}
<section class="related-section">
  <h2>Related</h2>
  {{ with .Params.uses }}
  <h3>Uses</h3>
  <table class="related-table">
    <thead><tr><th>Function</th></tr></thead>
    <tbody>
      {{ range . }}<tr><td><code>{{ . }}</code></td></tr>{{ end }}
    </tbody>
  </table>
  {{ end }}
  {{ with .Params.used_by }}
  <h3>Used By</h3>
  <table class="related-table">
    <thead><tr><th>Function</th></tr></thead>
    <tbody>
      {{ range . }}<tr><td><code>{{ . }}</code></td></tr>{{ end }}
    </tbody>
  </table>
  {{ end }}
</section>
{{ end }}

{{ with .Params.changelog }}
<section class="changelog-section">
  <h2>Changelog</h2>
  <table class="changelog-table">
    <thead><tr><th>Version</th><th>Description</th></tr></thead>
    <tbody>
      {{ range . }}
      <tr><td>{{ .version }}</td><td>{{ .description }}</td></tr>
      {{ end }}
    </tbody>
  </table>
</section>
{{ end }}

</article>
{{ end }}
`

const partialNav = `{{ $pathParts := split (strings.TrimPrefix "/" .RelPermalink) "/" }}
{{ $currentVersion := index $pathParts 0 }}
{{ $versionPage := $.Site.GetPage (printf "/%s" $currentVersion) }}

<div class="nav-header">
  <a href="{{ "/" | relURL }}">{{ .Site.Title }}</a>
</div>

{{/* Build version list from actual content sections, not just data file */}}
{{ $versions := slice }}
{{ range .Site.Home.Sections }}
  {{ $versions = $versions | append .Section }}
{{ end }}
{{ if gt (len $versions) 0 }}
<div class="version-select">
  <select id="version-switcher" onchange="switchVersion(this.value)">
    {{ range $versions }}
    <option value="{{ . }}"{{ if eq . $currentVersion }} selected{{ end }}>{{ . }}</option>
    {{ end }}
  </select>
</div>
{{ end }}

<nav>
  {{ with $versionPage }}
    {{ $guidesSection := .GetPage "guides" }}
    {{ with $guidesSection }}
      {{ if .Pages }}
      <div class="nav-section-label">Guides</div>
      {{ range .Pages.ByWeight }}
      <a href="{{ .RelPermalink }}" class="nav-guide{{ if eq $.RelPermalink .RelPermalink }} active{{ end }}">
        {{ .Title }}
      </a>
      {{ end }}
      <div class="nav-divider"></div>
      {{ end }}
    {{ end }}

    <div class="nav-section-label">Reference</div>
    {{ $refSections := slice "functions" "classes" "methods" "hooks" "interfaces" "traits" "enums" "components" }}
    {{ range $refSections }}
      {{ $sec := $versionPage.GetPage . }}
      {{ with $sec }}
        {{ if .Pages }}
        <a href="{{ .RelPermalink }}"{{ if eq $.Section . }} class="active"{{ end }}>
          {{ .Title }} <span class="count">({{ len .Pages }})</span>
        </a>
        {{ end }}
      {{ end }}
    {{ end }}
  {{ end }}
</nav>

<script>
function switchVersion(v) {
  var parts = window.location.pathname.split('/').filter(Boolean);
  if (parts.length > 0) { parts[0] = v; }
  var target = '/' + parts.join('/');
  // Try the exact page first; fall back to version root if it 404s
  fetch(target, { method: 'HEAD' }).then(function(r) {
    window.location.pathname = r.ok ? target : '/' + v + '/';
  }).catch(function() {
    window.location.pathname = '/' + v + '/';
  });
}
</script>
`

const partialMeta = `<div class="meta-bar">
  <span class="badge kind">{{ .Params.symbol_kind }}</span>
  <span class="badge lang">{{ .Params.language }}</span>
  {{ with .Params.access }}<span class="badge access">{{ . }}</span>{{ end }}
  {{ with .Params.since }}<span class="badge since">Since {{ . }}</span>{{ end }}
  {{ with .Params.deprecated }}<span class="badge deprecated">Deprecated</span>{{ end }}
</div>
`

// --- Guide layout templates ---

const layoutGuideList = `{{ define "main" }}
<h1>{{ .Title }}</h1>
<div class="guide-cards">
  {{ range .Pages.ByWeight }}
  <a href="{{ .RelPermalink }}" class="guide-card">
    <h3>{{ .Title }}</h3>
    {{ with .Params.summary }}<p>{{ . }}</p>{{ end }}
  </a>
  {{ end }}
</div>
{{ end }}
`

const layoutGuideSingle = `{{ define "main" }}
<article class="guide-article">

<h1>{{ .Title }}</h1>

{{ if .TableOfContents }}
<aside class="guide-toc">
  <h4>On this page</h4>
  {{ .TableOfContents }}
</aside>
{{ end }}

<div class="guide-body">
  {{ .Content }}
</div>

<nav class="guide-pager">
  {{ with .PrevInSection }}
  <a href="{{ .RelPermalink }}" class="pager-prev">&larr; {{ .Title }}</a>
  {{ end }}
  {{ with .NextInSection }}
  <a href="{{ .RelPermalink }}" class="pager-next">{{ .Title }} &rarr;</a>
  {{ end }}
</nav>

</article>
{{ end }}
`

// --- CSS ---

const styleCSS = `/* WordPress Developer Reference */
:root {
  --wp-blue: #0073aa;
  --wp-dark: #23282d;
  --wp-light: #f0f0f1;
  --wp-border: #c3c4c7;
  --wp-red: #d63638;
  --wp-green: #00a32a;
  --sidebar-width: 240px;
  --content-max: 900px;
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif;
  --font-mono: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: var(--font-sans);
  color: #1d2327;
  line-height: 1.6;
  background: #fff;
  font-size: 15px;
}

a { color: var(--wp-blue); text-decoration: none; }
a:hover { text-decoration: underline; color: #005177; }

.layout {
  display: flex;
  min-height: 100vh;
}

/* Sidebar */
.sidebar {
  width: var(--sidebar-width);
  background: var(--wp-dark);
  color: #fff;
  padding: 1rem 0;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  overflow-y: auto;
}

.nav-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #464b50;
  margin-bottom: 0.5rem;
}

.nav-header a {
  color: #fff;
  text-decoration: none;
  font-weight: 600;
  font-size: 0.9rem;
}

.sidebar nav a {
  display: block;
  padding: 0.4rem 1rem;
  color: #b4b9be;
  text-decoration: none;
  font-size: 0.85rem;
  transition: background 0.15s, color 0.15s;
}

.sidebar nav a:hover {
  background: #32373c;
  color: #fff;
}

.sidebar nav a.active {
  background: var(--wp-blue);
  color: #fff;
}

.sidebar nav .count {
  opacity: 0.6;
  font-size: 0.8rem;
}

/* Main content */
.content {
  flex: 1;
  margin-left: var(--sidebar-width);
  padding: 2rem 3rem;
  max-width: calc(var(--content-max) + var(--sidebar-width) + 6rem);
}

h1 {
  font-size: 1.6rem;
  border-bottom: 2px solid var(--wp-blue);
  padding-bottom: 0.5rem;
  margin-bottom: 0.75rem;
  font-weight: 600;
}

h2 {
  font-size: 1.25rem;
  margin-top: 2rem;
  margin-bottom: 0.75rem;
  color: var(--wp-dark);
  padding-bottom: 0.3rem;
  border-bottom: 1px solid #e0e0e0;
}

h3 {
  font-size: 1rem;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
  color: #50575e;
}

/* Stats grid (homepage) */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
  margin: 1.5rem 0;
}

.stat-card {
  background: var(--wp-light);
  padding: 1.25rem;
  border-radius: 4px;
  text-align: center;
}

.stat-card.stat-total {
  background: var(--wp-blue);
  color: #fff;
}

.stat-card.stat-total .stat-number { color: #fff; }
.stat-card.stat-total .stat-label a { color: #fff; }

.stat-number {
  font-size: 2rem;
  font-weight: 700;
  color: var(--wp-blue);
}

.stat-label { font-size: 0.9rem; color: #50575e; }
.stat-label a { color: var(--wp-blue); text-decoration: none; }
.stat-label a:hover { text-decoration: underline; }

/* Listing table (section pages) */
.listing {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
}

.listing th, .listing td {
  text-align: left;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid #e0e0e0;
}

.listing th { background: var(--wp-light); font-weight: 600; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.03em; }

.listing a { font-weight: 500; }

.listing .since { color: #787c82; font-size: 0.85rem; white-space: nowrap; }

.deprecated-row { opacity: 0.65; }

p.count { color: #787c82; margin-bottom: 0.5rem; }

/* Badges */
.badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 3px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.badge.kind { background: var(--wp-light); color: var(--wp-dark); }
.badge.lang { background: #dce8f0; color: var(--wp-blue); }
.badge.since { background: #e7f5e7; color: #1e7e1e; }
.badge.access { background: #fef3cd; color: #856404; }
.badge.deprecated { background: #fcf0f1; color: var(--wp-red); }

.meta-bar {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
  align-items: center;
}

/* Reference article */
.wp-reference section {
  margin-bottom: 0.5rem;
}

/* Deprecated notice */
.deprecated-notice {
  background: #fcf0f1;
  border-left: 4px solid var(--wp-red);
  padding: 0.75rem 1rem;
  margin: 1rem 0;
  border-radius: 0 3px 3px 0;
}

/* Signature block */
.signature-section {
  margin: 1rem 0;
}

.signature-block {
  background: #23282d;
  color: #eee;
  padding: 1rem 1.25rem;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 0.9rem;
  line-height: 1.5;
}

.signature-block code {
  background: none;
  color: inherit;
  padding: 0;
  font-size: inherit;
}

/* Description */
.description-section .summary {
  font-size: 1.05rem;
  margin-bottom: 0.75rem;
}

.long-description {
  line-height: 1.7;
  margin-top: 0.5rem;
}

.long-description p {
  margin-bottom: 0.75rem;
}

.doc-links {
  list-style: none;
  margin-left: 0;
}

/* Parameters (definition list) */
.param-list {
  margin: 0.5rem 0 1rem;
}

.param-list dt {
  padding: 0.6rem 0 0.2rem;
  border-top: 1px solid #e0e0e0;
  font-weight: 600;
}

.param-list dt:first-child {
  border-top: none;
}

.param-list dt code {
  font-size: 0.95em;
  background: none;
  padding: 0;
  color: var(--wp-dark);
}

.param-type {
  margin-left: 0.5rem;
  font-weight: 400;
}

.param-type code {
  color: var(--wp-blue);
  background: none;
  padding: 0;
}

.param-tag {
  display: inline-block;
  margin-left: 0.4rem;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  background: #fef3cd;
  color: #856404;
}

.param-list dd {
  padding: 0.2rem 0 0.6rem 1.25rem;
  color: #50575e;
  line-height: 1.6;
}

.param-default {
  margin-top: 0.25rem;
  font-size: 0.9rem;
  color: #787c82;
}

/* Return */
.return-section p {
  line-height: 1.6;
}

.return-type {
  color: var(--wp-blue);
  background: none;
  padding: 0;
  font-weight: 600;
}

/* Source */
.source-section {
  background: var(--wp-light);
  padding: 1rem 1.25rem;
  border-radius: 4px;
  margin-top: 2rem;
}

.source-section h2 {
  margin-top: 0;
  border-bottom: none;
  padding-bottom: 0;
}

.source-file {
  margin: 0.25rem 0;
  font-size: 0.9rem;
}

.source-file code {
  background: none;
  padding: 0;
}

.source-links {
  display: flex;
  gap: 1rem;
  margin: 0.5rem 0;
  font-size: 0.9rem;
}

.source-links a {
  font-weight: 500;
}

.source-code-details {
  margin-top: 0.75rem;
}

.source-code-details summary {
  cursor: pointer;
  font-size: 0.85rem;
  color: var(--wp-blue);
  font-weight: 500;
}

.source-code-details summary:hover {
  text-decoration: underline;
}

.source-code {
  background: #23282d;
  color: #eee;
  padding: 1rem 1.25rem;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 0.85rem;
  line-height: 1.5;
  margin-top: 0.5rem;
}

.source-code code {
  background: none;
  color: inherit;
  padding: 0;
  font-size: inherit;
}

/* Related tables */
.related-table {
  width: 100%;
  border-collapse: collapse;
  margin: 0.5rem 0 1rem;
  font-size: 0.9rem;
}

.related-table th, .related-table td {
  text-align: left;
  padding: 0.4rem 0.75rem;
  border-bottom: 1px solid #e0e0e0;
}

.related-table th {
  background: var(--wp-light);
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.related-table code {
  background: none;
  padding: 0;
}

/* Changelog table */
.changelog-table {
  width: 100%;
  border-collapse: collapse;
  margin: 0.5rem 0 1rem;
  font-size: 0.9rem;
}

.changelog-table th, .changelog-table td {
  text-align: left;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid #e0e0e0;
}

.changelog-table th {
  background: var(--wp-light);
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.changelog-table td:first-child {
  white-space: nowrap;
  font-weight: 500;
  width: 120px;
}

/* Hook details */
.hook-section p {
  margin: 0.25rem 0;
}

/* Members list */
.member-list {
  column-count: 2;
  column-gap: 2rem;
}

/* General */
code {
  font-family: var(--font-mono);
  font-size: 0.9em;
  background: var(--wp-light);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
}

ul { margin-left: 1.5rem; margin-bottom: 0.5rem; }
li { margin: 0.25rem 0; }

/* Version selector */
.version-select {
  padding: 0.5rem 1rem;
  border-bottom: 1px solid #464b50;
  margin-bottom: 0.5rem;
}

.version-select select {
  width: 100%;
  padding: 0.4rem 0.5rem;
  background: #32373c;
  color: #fff;
  border: 1px solid #464b50;
  border-radius: 3px;
  font-size: 0.85rem;
  cursor: pointer;
  appearance: auto;
}

.version-select select:hover {
  border-color: var(--wp-blue);
}

/* Nav section labels */
.nav-section-label {
  padding: 0.5rem 1rem 0.2rem;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #787c82;
}

.nav-divider {
  border-bottom: 1px solid #464b50;
  margin: 0.4rem 1rem;
}

.nav-guide {
  padding-left: 1.5rem !important;
  font-size: 0.82rem !important;
}

/* Version list (homepage fallback) */
.version-list {
  list-style: none;
  margin-left: 0;
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1rem;
}

.version-list li a {
  display: block;
  padding: 0.5rem 1.25rem;
  background: var(--wp-light);
  border-radius: 4px;
  font-weight: 500;
}

.version-list li a:hover {
  background: var(--wp-blue);
  color: #fff;
  text-decoration: none;
}

/* Guide cards */
.guide-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
  margin: 1rem 0;
}

.guide-card {
  display: block;
  padding: 1.25rem;
  background: var(--wp-light);
  border-radius: 6px;
  border: 1px solid transparent;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.guide-card:hover {
  border-color: var(--wp-blue);
  box-shadow: 0 2px 8px rgba(0, 115, 170, 0.1);
  text-decoration: none;
}

.guide-card h3 {
  margin: 0 0 0.25rem;
  color: var(--wp-dark);
  font-size: 1rem;
}

.guide-card p {
  margin: 0;
  color: #50575e;
  font-size: 0.85rem;
  line-height: 1.5;
}

/* Guide overview section */
.guides-overview {
  margin-bottom: 2rem;
}

.reference-overview {
  margin-bottom: 2rem;
}

/* Guide article */
.guide-article {
  position: relative;
}

.guide-article h1 {
  margin-bottom: 1.5rem;
}

.guide-toc {
  float: right;
  width: 220px;
  margin: 0 0 1rem 2rem;
  padding: 1rem;
  background: var(--wp-light);
  border-radius: 6px;
  font-size: 0.82rem;
}

.guide-toc h4 {
  margin: 0 0 0.5rem;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: #50575e;
}

.guide-toc nav ul {
  margin: 0;
  list-style: none;
}

.guide-toc nav ul ul {
  margin-left: 0.75rem;
}

.guide-toc nav a {
  display: block;
  padding: 0.15rem 0;
  color: #50575e;
  font-size: 0.82rem;
}

.guide-toc nav a:hover {
  color: var(--wp-blue);
}

/* Guide body prose */
.guide-body {
  line-height: 1.8;
}

.guide-body h2 {
  margin-top: 2.5rem;
}

.guide-body h3 {
  margin-top: 1.5rem;
}

.guide-body p {
  margin-bottom: 1rem;
}

.guide-body pre {
  background: #23282d;
  color: #eee;
  padding: 1rem 1.25rem;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 0.85rem;
  line-height: 1.5;
  margin: 1rem 0;
}

.guide-body pre code {
  background: none;
  color: inherit;
  padding: 0;
  font-size: inherit;
}

.guide-body blockquote {
  border-left: 4px solid var(--wp-blue);
  padding: 0.5rem 1rem;
  margin: 1rem 0;
  background: #f8f9fa;
  border-radius: 0 4px 4px 0;
}

/* Guide pager */
.guide-pager {
  display: flex;
  justify-content: space-between;
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e0e0e0;
}

.guide-pager a {
  padding: 0.5rem 1rem;
  background: var(--wp-light);
  border-radius: 4px;
  font-weight: 500;
  font-size: 0.9rem;
}

.guide-pager a:hover {
  background: var(--wp-blue);
  color: #fff;
  text-decoration: none;
}

.pager-next {
  margin-left: auto;
}

/* Override content section */
.override-content {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 2px solid var(--wp-blue);
}

.override-content h2 {
  color: var(--wp-blue);
}

/* Responsive */
@media (max-width: 768px) {
  .sidebar { display: none; }
  .content { margin-left: 0; padding: 1rem; max-width: 100%; }
  .member-list { column-count: 1; }
  .guide-toc { float: none; width: 100%; margin: 0 0 1.5rem 0; }
  .guide-cards { grid-template-columns: 1fr; }
}
`

// --- Content template ---

const symbolContentTemplate = `---
title: {{ yamlEscape .Name }}
linkTitle: {{ yamlEscape .Name }}
symbol_kind: {{ yamlEscape (printf "%s" .Kind) }}
language: {{ yamlEscape .Language }}
since: {{ yamlEscape .Doc.Since }}
deprecated: {{ yamlEscape .Doc.Deprecated }}
access: {{ yamlEscape .Doc.Access }}
summary: {{ yamlEscape .Doc.Summary }}
signature: {{ yamlEscape .Signature }}
{{- if .Symbol.Params }}
parameters:
{{- range .Symbol.Params }}
  - name: {{ yamlEscape .Name }}
    type: {{ yamlEscape .Type }}
    description: {{ yamlEscape .Description }}
    default: {{ yamlEscape .Default }}
    variadic: {{ .IsVariadic }}
    pass_by_ref: {{ .IsPassByRef }}
{{- end }}
{{- end }}
{{- if .Returns }}
returns:
  type: {{ yamlEscape .Returns.Type }}
  description: {{ yamlEscape .Returns.Description }}
{{- end }}
hook_type: {{ yamlEscape (printf "%s" .HookType) }}
hook_tag: {{ yamlEscape .HookTag }}
{{- if .CallSites }}
call_sites:
{{- range .CallSites }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
{{- if .Extends }}
extends:
{{- range .Extends }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
{{- if .Implements }}
implements:
{{- range .Implements }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
{{- if .Members }}
members:
{{- range .Members }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
parent_id: {{ yamlEscape .ParentID }}
{{- if .UsedBy }}
used_by:
{{- range .UsedBy }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
{{- if .Uses }}
uses:
{{- range .Uses }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
overrides: {{ yamlEscape .Overrides }}
{{- if .Doc.SeeAlso }}
see_also:
{{- range .Doc.SeeAlso }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
{{- if .Doc.Links }}
links:
{{- range .Doc.Links }}
  - {{ yamlEscape . }}
{{- end }}
{{- end }}
{{- if .Changelog }}
changelog:
{{- range .Changelog }}
  - version: {{ yamlEscape .Version }}
    description: {{ yamlEscape .Description }}
{{- end }}
{{- end }}
file: {{ yamlEscape .Location.File }}
start_line: {{ .Location.StartLine }}
end_line: {{ .Location.EndLine }}
github_url: {{ yamlEscape .GitHubURL }}
trac_url: {{ yamlEscape .TracURL }}
{{- if .SourceCode }}
source_code: {{ yamlMultiline .SourceCode }}
{{- end }}
---

{{ safeContent .Doc.Description }}
{{- if .OverrideContent }}

<div class="override-content">

{{ safeContent .OverrideContent }}

</div>
{{- end }}
`
