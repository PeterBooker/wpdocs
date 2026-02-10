package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/peter/wpdocs/internal/model"
	"github.com/peter/wpdocs/internal/output"
	"github.com/peter/wpdocs/internal/parser"
	"github.com/peter/wpdocs/internal/resolver"
	"github.com/peter/wpdocs/internal/source"
)

func main() {
	var (
		wpPath  string
		outDir  string
		wpTag   string
		skipJS  bool
		skipPHP bool
		workers int
	)

	root := &cobra.Command{
		Use:   "wpdocs",
		Short: "Generate WordPress developer documentation from source",
		Long: `Parses WordPress PHP and JS/TS source code, extracts functions,
classes, hooks, and their documentation, then generates a Hugo static site
suitable for developer.wordpress.org.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			// Step 1: Resolve WordPress source
			src, err := source.Resolve(wpPath, wpTag)
			if err != nil {
				return fmt.Errorf("resolving source: %w", err)
			}
			log.Printf("Using WordPress source: %s (tag: %s)", src.Path, src.Version)

			registry := model.NewRegistry()
			p := parser.New(workers)
			p.SetSrcRoot(src.Path)

			// Step 2: Parse PHP
			if !skipPHP {
				log.Println("Parsing PHP files...")
				phpFiles, err := src.FindFiles("*.php")
				if err != nil {
					return fmt.Errorf("finding PHP files: %w", err)
				}
				log.Printf("Found %d PHP files", len(phpFiles))

				if err := p.ParseFiles(phpFiles, registry); err != nil {
					return fmt.Errorf("parsing PHP: %w", err)
				}
				log.Printf("Extracted %d PHP symbols", registry.CountByLanguage("php"))
			}

			// Step 3: Parse JS/TS
			if !skipJS {
				log.Println("Parsing JS/TS files...")
				jsFiles, err := src.FindFiles("*.js", "*.ts", "*.jsx", "*.tsx")
				if err != nil {
					return fmt.Errorf("finding JS files: %w", err)
				}
				log.Printf("Found %d JS/TS files", len(jsFiles))

				if err := p.ParseFiles(jsFiles, registry); err != nil {
					return fmt.Errorf("parsing JS/TS: %w", err)
				}
				log.Printf("Extracted %d JS/TS symbols", registry.CountByLanguage("js"))
			}

			// Step 4: Resolve cross-references
			log.Println("Resolving cross-references...")
			res := resolver.New(registry)
			res.ResolveAll()
			log.Printf("Resolved %d cross-references", res.Stats().Resolved)

			// Step 5: Generate Hugo site
			log.Printf("Generating Hugo site in %s", outDir)
			gen := output.NewHugo(outDir, src.Path, src.Version)
			if err := gen.Generate(registry); err != nil {
				return fmt.Errorf("generating output: %w", err)
			}

			log.Printf("Done in %s. Total symbols: %d",
				time.Since(start).Round(time.Millisecond),
				registry.Count())
			return nil
		},
	}

	root.Flags().StringVarP(&wpPath, "source", "s", "", "Path to WordPress source (or auto-downloads if empty)")
	root.Flags().StringVarP(&outDir, "output", "o", "./docs", "Output directory for Hugo site")
	root.Flags().StringVarP(&wpTag, "tag", "t", "latest", "WordPress version tag (e.g., 6.7.1)")
	root.Flags().BoolVar(&skipJS, "skip-js", false, "Skip JS/TS parsing")
	root.Flags().BoolVar(&skipPHP, "skip-php", false, "Skip PHP parsing")
	root.Flags().IntVarP(&workers, "workers", "w", 8, "Number of parallel workers")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
