package output

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed embed
var embeddedFiles embed.FS

// writeEmbeddedFiles copies Hugo layouts, config, and static assets from the
// embedded filesystem into the output directory. The templates/ subdirectory is
// skipped — those are Go text/templates used by the content generator, not Hugo
// layout files.
func writeEmbeddedFiles(outDir string) error {
	return fs.WalkDir(embeddedFiles, "embed", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Strip the "embed/" prefix to get the relative output path.
		rel := strings.TrimPrefix(path, "embed/")
		if rel == "" {
			return nil
		}

		// Skip the Go text/template directory — those are not Hugo files.
		if strings.HasPrefix(rel, "templates/") {
			return nil
		}

		dest := filepath.Join(outDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}

// readEmbeddedTemplate returns the contents of a file under the embedded
// templates/ directory. Used to load the symbol content template at startup.
func readEmbeddedTemplate(name string) (string, error) {
	data, err := embeddedFiles.ReadFile("embed/templates/" + name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
