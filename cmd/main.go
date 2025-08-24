package main

import (
	"fmt"
	"os"

	"historiadorgo/internal/presentation/cli"
)

func main() {
	rootCmd := cli.SetupCommands()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
