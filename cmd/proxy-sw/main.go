package main

import (
	"fmt"
	"os"

	"github.com/kelaocai/proxy-sw/internal/cli"
)

var version = "dev"

func main() {
	app := cli.New(version, os.Stdout, os.Stderr)
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCode(err))
	}
}
