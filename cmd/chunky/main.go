package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var version = "dev"

// CLI represents the top-level command structure.
type CLI struct {
	Run  RunCmd  `cmd:"" help:"Run chunking on files"`
	Init InitCmd `cmd:"init" help:"Initialize a .chunkyrc configuration file"`
}

func main() {
	var c CLI

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
