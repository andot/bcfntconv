package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andot/shared-font-converter/pkg/convert"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <input.bcfnt> [output_shared_font.bin]\n", filepath.Base(os.Args[0]))
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		usage()
		os.Exit(2)
	}
	in := os.Args[1]
	out := "shared_font.bin"
	if len(os.Args) == 3 {
		out = os.Args[2]
	}

	data, err := os.ReadFile(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	outData, err := convert.BcfntToShared(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "convert error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(out, outData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}
}
