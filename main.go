package main

import (
	"fmt"
	"os"

	"github.com/rh-amarin/hyperfleet-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		os.Exit(1)
	}
}
