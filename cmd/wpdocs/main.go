package main

import (
	"os"

	"github.com/alecthomas/kong"
)

// CLI defines the top-level wpdocs command with shared flags and subcommands.
type CLI struct {
	Output string `help:"Hugo output directory." short:"o" default:"./docs"`

	Generate    GenerateCmd    `cmd:"" help:"Parse WordPress source and generate Hugo content for one version."`
	GenerateAll GenerateAllCmd `cmd:"" name:"generate-all" help:"Generate docs for all configured WordPress versions."`
	Build       BuildCmd       `cmd:"" help:"Run Hugo to produce the final static site."`
	Serve       ServeCmd       `cmd:"" help:"Start Hugo dev server for preview."`
	Clean       CleanCmd       `cmd:"" help:"Remove generated content and built site."`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("wpdocs"),
		kong.Description("Generate WordPress developer documentation from source."),
		kong.UsageOnError(),
	)
	err := ctx.Run(&cli)
	if err != nil {
		os.Exit(1)
	}
}
