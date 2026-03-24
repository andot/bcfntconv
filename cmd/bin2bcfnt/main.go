package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andot/bcfntconv/pkg/convert"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [input_shared_font.bin] [output.bcfnt]\n", filepath.Base(os.Args[0]))
}

func main() {
	// Defaults: input -> ./shared_font.bin, output -> ./output.bcfnt
	in := "shared_font.bin"
	out := "output.bcfnt"
	args := os.Args[1:]
	if len(args) > 2 {
		usage()
		os.Exit(2)
	}
	for _, a := range args {
		ext := strings.ToLower(filepath.Ext(a))
		switch ext {
		case ".bin":
			in = a
		case ".bcfnt":
			out = a
		default:
			// if user provided one arg without extension, assume it's the output and append .bcfnt
			if ext == "" && len(args) == 1 {
				out = a + ".bcfnt"
			}
		}
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
