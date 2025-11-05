package main

import (
	"fmt"
	"os"

	"github.com/zot/p2p-webapp/internal/bundle"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <source-binary> <site-dir> <output>\n", os.Args[0])
		os.Exit(1)
	}

	sourceBinary := os.Args[1]
	siteDir := os.Args[2]
	output := os.Args[3]

	if err := bundle.CreateBundle(sourceBinary, siteDir, output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	info, _ := os.Stat(output)
	sizeMB := float64(info.Size()) / 1024 / 1024
	fmt.Printf("✓ Created %s (%.2f MB)\n", output, sizeMB)
}
