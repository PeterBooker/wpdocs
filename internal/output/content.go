package output

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/peter/wpdocs/internal/model"
)

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

// changelogEntry represents one row in the Changelog table.
type changelogEntry struct {
	Version     string
	Description string
}

// initTemplate parses the symbol content template once and stores it on the
// Hugo struct for reuse across all symbol pages.
func (h *Hugo) initTemplate() error {
	raw, err := readEmbeddedTemplate("symbol.md.tmpl")
	if err != nil {
		return err
	}

	tmpl, err := template.New("symbol").Funcs(template.FuncMap{
		"yamlEscape":    yamlEscape,
		"yamlMultiline": yamlMultiline,
		"join":          strings.Join,
		"safeContent":   safeContent,
	}).Parse(raw)
	if err != nil {
		return err
	}

	h.symbolTmpl = tmpl
	return nil
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
		Symbol:          sym,
		Signature:       buildSignature(sym),
		Changelog:       parseChangelog(sym),
		SourceCode:      h.readSourceContext(sym.Location.File, sym.Location.StartLine),
		GitHubURL:       h.buildGitHubURL(sym.Location.File, sym.Location.StartLine, sym.Location.EndLine),
		TracURL:         h.buildTracURL(sym.Location.File, sym.Location.StartLine),
		OverrideContent: h.readOverride(section, slug),
	}

	return h.symbolTmpl.Execute(f, data)
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
