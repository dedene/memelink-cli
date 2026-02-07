// Package main is the entry point for the memelink CLI.
package main

import (
	"fmt"
	"os"

	"github.com/dedene/memelink-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cmd.ExitCode(err))
	}
}
