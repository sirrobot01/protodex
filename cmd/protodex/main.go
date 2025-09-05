package main

import (
	"os"

	"github.com/sirrobot01/protodex/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
