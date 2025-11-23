package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/wyvernzora/chunky/internal/cli"
)

var version = "dev"

func main() {
	var c cli.CLI

	ctx := kong.Parse(&c,
		kong.Name("chunky"),
		kong.Description("Intelligent markdown document chunking for embedding pipelines"),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

	// Execute the selected command
	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
