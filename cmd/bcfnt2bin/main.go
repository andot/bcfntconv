package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andot/bcfntconv/pkg/convert"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <input.bcfnt> [output_shared_font.bin]\n", filepath.Base(os.Args[0]))
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 || len(args) > 2 {
		usage()
		os.Exit(2)
	}

	in := ""
	out := "shared_font.bin"
	for _, a := range args {
		ext := strings.ToLower(filepath.Ext(a))
		switch ext {
		case ".bcfnt":
			in = a
		case ".bin":
			out = a
		default:
			// if no extension and we don't have input yet, treat as input and append .bcfnt
			if ext == "" && in == "" {
				in = a + ".bcfnt"
			}
		}
	}
	if in == "" {
		usage()
		os.Exit(2)
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
