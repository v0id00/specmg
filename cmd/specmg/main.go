package main

import (
	"fmt"
	"os"

	"github.com/v0id00/specmg/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if cli.IsUsageErr(err) {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
