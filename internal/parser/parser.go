package parser

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/php"
	typescript "github.com/smacker/go-tree-sitter/typescript/typescript"
	tsx "github.com/smacker/go-tree-sitter/typescript/tsx"

	"github.com/peter/wpdocs/internal/model"
)

// Parser extracts documentation from PHP and JS/TS source files using tree-sitter.
type Parser struct {
	workers int
	srcRoot string
}

// New creates a parser with the given number of parallel workers.
func New(workers int) *Parser {
	return &Parser{workers: workers}
}

// SetSrcRoot sets the WordPress source root for resolving absolute file paths.
func (p *Parser) SetSrcRoot(root string) {
	p.srcRoot = root
}

// ParseFiles processes all given files and adds symbols to the registry.
// Each worker goroutine gets its own sitter.Parser instance (not thread-safe).
func (p *Parser) ParseFiles(files []string, reg *model.Registry) error {
	if len(files) == 0 {
		return nil
	}

	ch := make(chan string, len(files))
	for _, f := range files {
		ch <- f
	}
	close(ch)

	var wg sync.WaitGroup
	errCh := make(chan error, p.workers)

	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sp := sitter.NewParser()
			for file := range ch {
				if err := p.parseFile(sp, file, reg); err != nil {
					errCh <- fmt.Errorf("%s: %w", file, err)
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		log.Printf("Warning: %d files had parse errors", len(errs))
	}
	return nil
}

func (p *Parser) parseFile(sp *sitter.Parser, relPath string, reg *model.Registry) error {
	absPath := relPath
	if p.srcRoot != "" {
		absPath = filepath.Join(p.srcRoot, relPath)
	}

	src, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	lang, langName, err := detectLanguage(relPath)
	if err != nil {
		return err
	}

	sp.SetLanguage(lang)

	tree, err := sp.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()

	switch langName {
	case "php":
		extractPHP(root, src, relPath, reg)
	case "js":
		extractJS(root, src, relPath, reg)
	}

	return nil
}

func detectLanguage(path string) (*sitter.Language, string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".php":
		return php.GetLanguage(), "php", nil
	case ".js", ".jsx":
		return javascript.GetLanguage(), "js", nil
	case ".ts":
		return typescript.GetLanguage(), "js", nil
	case ".tsx":
		return tsx.GetLanguage(), "js", nil
	default:
		return nil, "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}
