package main

import (
	"fmt"
	"os"

	"github.com/foden/cdc/internal/infrastructure"
)

func main() {
	if err := infrastructure.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cdc failed: %v\n", err)
		os.Exit(1)
	}
}
