package main

import (
	"fmt"
	"os"

	"github.com/danpet-dev/keyforge/cmd/keyforge/commands"
)

var version = "dev"

func main() {
	commands.Version = version
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
