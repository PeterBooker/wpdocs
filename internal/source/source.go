package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Source represents a resolved WordPress source tree.
type Source struct {
	Path    string
	Version string
}

// Resolve either uses an existing local path or clones from git.
func Resolve(localPath, tag string) (*Source, error) {
	if localPath != "" {
		// Validate it looks like a WP source tree
		if _, err := os.Stat(filepath.Join(localPath, "wp-includes")); err != nil {
			return nil, fmt.Errorf("%s doesn't look like a WordPress source tree: missing wp-includes", localPath)
		}
		version := tag
		if version == "latest" {
			version = detectVersion(localPath)
		}
		return &Source{Path: localPath, Version: version}, nil
	}

	// Clone from GitHub
	return cloneFromGit(tag)
}

func cloneFromGit(tag string) (*Source, error) {
	tmpDir, err := os.MkdirTemp("", "wp-src-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}

	repo := "https://github.com/WordPress/WordPress.git"
	args := []string{"clone", "--depth", "1"}
	if tag != "latest" {
		args = append(args, "--branch", tag)
	}
	args = append(args, repo, tmpDir)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("cloning WordPress: %w", err)
	}

	version := tag
	if tag == "latest" {
		version = detectVersion(tmpDir)
	}

	return &Source{Path: tmpDir, Version: version}, nil
}

func detectVersion(wpPath string) string {
	data, err := os.ReadFile(filepath.Join(wpPath, "wp-includes", "version.php"))
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "$wp_version =") {
			// Extract version from: $wp_version = '6.7.1';
			parts := strings.Split(line, "'")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "unknown"
}

// FindFiles returns all files matching the given glob patterns under the source tree.
// It automatically skips vendor/, node_modules/, and test directories.
func (s *Source) FindFiles(patterns ...string) ([]string, error) {
	skipDirs := map[string]bool{
		"vendor":       true,
		"node_modules": true,
		".git":         true,
		"tests":        true,
		"test":         true,
	}

	var files []string
	err := filepath.Walk(s.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		for _, pattern := range patterns {
			matched, _ := filepath.Match(pattern, info.Name())
			if matched {
				relPath, _ := filepath.Rel(s.Path, path)
				files = append(files, relPath)
				break
			}
		}
		return nil
	})
	return files, err
}
