package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andot/shared-font-converter/pkg/convert"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [input_shared_font.bin] [output.bcfnt]\n", filepath.Base(os.Args[0]))
}

func main() {
	// Defaults: input -> ./shared_font.bin, output -> ./output.bcfnt
	in := "shared_font.bin"
	out := "output.bcfnt"
	switch len(os.Args) {
	case 1:
		// no args: use defaults
	case 2:
		in = os.Args[1]
	case 3:
		in = os.Args[1]
		out = os.Args[2]
	default:
		usage()
		os.Exit(2)
	}

	data, err := os.ReadFile(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	outData, err := convert.SharedToBcfnt(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "convert error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(out, outData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}
}
