package output

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	symbolTmpl   *template.Template
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

// Generate writes versioned markdown content, Hugo layouts, config, and static
// assets to the output directory. It does NOT invoke the Hugo binary — call
// Build() separately when you are ready to produce the final static site.
func (h *Hugo) Generate(reg *model.Registry) error {
	// Parse the symbol content template once for reuse across all pages.
	if err := h.initTemplate(); err != nil {
		return fmt.Errorf("parsing symbol template: %w", err)
	}

	// Clean only this version's content directory (preserves other versions)
	versionDir := filepath.Join(h.outDir, "content", h.version)
	_ = os.RemoveAll(versionDir)

	// Create directory structure
	dirs := []string{
		filepath.Join(h.outDir, "content", h.version),
		filepath.Join(h.outDir, "data"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	// Write Hugo config, layouts, and static assets from embedded files
	if err := writeEmbeddedFiles(h.outDir); err != nil {
		return fmt.Errorf("writing embedded files: %w", err)
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

	return nil
}

// Build invokes the Hugo binary to produce the final static site from the
// generated content. It is safe to call after one or more Generate() runs.
// When minify is true, Hugo's --minify flag is passed.
func (h *Hugo) Build(minify bool) {
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

	args := []string{"--source", absDir}
	if minify {
		args = append(args, "--minify")
	}

	cmd := exec.Command(hugoPath, args...)
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

	// Sort versions descending (newest first) using semantic comparison
	sort.Slice(data.All, func(i, j int) bool {
		return compareVersions(data.All[i], data.All[j]) > 0
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
