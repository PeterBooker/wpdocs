package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// normalizeVersion extracts major.minor from a full version string like "6.7.1".
func normalizeVersion(v string) string {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

// compareVersions compares two dotted version strings numerically.
// Returns >0 if a > b, <0 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := range maxLen {
		var ai, bi int
		if i < len(aParts) {
			ai, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bi, _ = strconv.Atoi(bParts[i])
		}
		if ai != bi {
			return ai - bi
		}
	}
	return 0
}

// symbolSlug converts a symbol ID into a safe filename slug.
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

// writeFile writes content to a file relative to the output directory.
func (h *Hugo) writeFile(relPath, content string) error {
	absPath := filepath.Join(h.outDir, relPath)
	return os.WriteFile(absPath, []byte(content), 0o644)
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
	start := max(startLine-6, 0)        // 5 lines before (0-indexed)
	end := min(startLine+5, len(lines))  // 5 lines after
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
