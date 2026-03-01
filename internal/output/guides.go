package output

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

// collectContentFiles builds a map of filename -> absolute path by reading _shared/
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
