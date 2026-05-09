package main

import (
	"os"

	"github.com/rh-amarin/hyperfleet-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
