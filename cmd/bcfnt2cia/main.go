package main

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const TitleID = "0004009B00014002"

// ncchHeader stored as base64 to keep single-file embedding
const ncchHeaderB64 = `My7n5mvi9dmDk1nJf2l7Y5NReYj2tur5QWOdzYCk33ipxX5UHj1teYK5JqA0shkuXySoYwCL5R1uvfiHBZUOU9TPUyk2mDCa69dPh89UC4b/to8p3wtBxmUqCfKV89ZA8dWJHvo+yIllLjMsKIyV6q5vyUNPzBYVBXblTgoMOdK8nKZFFsP+IsRCFWNFTphITt3IPD7YlGvdE+/tSrjXxUEC+6PCW44+B7NTvz4j7Pudl8oTmKbwtmG7hPVczwBVDidsXOWh3jf75NvfeEk+HW+ozxDkJjTu/W46KfM7pMgQO+wmDFwHUE912NpxqDx0ZwN68ooZsb9ewM0qzNdkrk5DQ0gxCwAAAkABAJsABAAwMAAAAAAAAAJAAQCbAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQ1RSLVAtQ1RBUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQEABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAADALAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOpiDT/0YhpjTCPEb880qpb7R2Cqv2I5ab+HfciO7e3I=`

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// track generated cfa path so cleanup can remove it
var generatedCFA string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <input.bcfnt> [output.cia]\n", filepath.Base(os.Args[0]))
}

func exitWithCleanup(code int) {
	cleanup()
	os.Exit(code)
}

func main() {
	// parse positional args (order-insensitive by extension)
	args := os.Args[1:]
	if len(args) < 1 || len(args) > 2 {
		usage()
		exitWithCleanup(2)
	}

	input := ""
	var output string
	for _, a := range args {
		ext := strings.ToLower(filepath.Ext(a))
		switch ext {
		case ".bcfnt":
			input = a
		case ".cia":
			output = a
		default:
			// treat extension-less as input (.bcfnt)
			if ext == "" && input == "" {
				input = a + ".bcfnt"
			}
		}
	}
	if input == "" {
		usage()
		exitWithCleanup(2)
	}
	if output == "" {
		base := stringsTrimExt(filepath.Base(input))
		output = base + ".cia"
	}

	// prompt if output exists
	if _, err := os.Stat(output); err == nil {
		fmt.Printf("%s exists. Overwrite? (Y/n) [Enter=overwrite]: ", output)
		r := bufio.NewReader(os.Stdin)
		resp, _ := r.ReadString('\n')
		resp = strings.TrimSpace(resp)
		if strings.EqualFold(resp, "n") {
			// find next available
			base := stringsTrimExt(output)
			for i := 1; ; i++ {
				cand := fmt.Sprintf("%s-%d.cia", base, i)
				if _, err := os.Stat(cand); os.IsNotExist(err) {
					output = cand
					break
				}
			}
		}
	}

	// ensure input exists
	if _, err := os.Stat(input); err != nil {
		fmt.Fprintf(os.Stderr, "Font file does not exist: %v\n", err)
		exitWithCleanup(1)
	}

	// prepare romfs dir
	if err := os.MkdirAll("romfs", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create romfs dir: %v\n", err)
		exitWithCleanup(1)
	}

	// compress using 3dstool (use separate flags for compatibility)
	fmt.Println("Compressing font with 3dstool...")
	if err := runCmd("3dstool", "-z", "-v", "-f", input, "--compress-type", "lzex", "--compress-out", "romfs/cbf_std.bcfnt.lz"); err != nil {
		fmt.Fprintf(os.Stderr, "3dstool compress failed: %v\n", err)
		exitWithCleanup(1)
	}

	// build romfs-mod.bin
	fmt.Println("Building romfs-mod.bin with 3dstool...")
	if err := runCmd("3dstool", "-c", "-v", "-t", "romfs", "-f", "romfs-mod.bin", "--romfs-dir", "romfs"); err != nil {
		fmt.Fprintf(os.Stderr, "3dstool romfs build failed: %v\n", err)
		exitWithCleanup(1)
	}

	// write embedded ncchheader to disk
	hdr, err := base64.StdEncoding.DecodeString(ncchHeaderB64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode embedded ncch header: %v\n", err)
		exitWithCleanup(1)
	}
	if err := os.WriteFile("ncchheader.bin", hdr, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write embedded ncchheader: %v\n", err)
		exitWithCleanup(1)
	}

	// build cfa (prefer xor if xorpad present)
	fmt.Println("Building .cfa with 3dstool...")
	cfaArgs := []string{"-c", "-v", "-t", "cfa", "-f", fmt.Sprintf("%s.cfa", stringsTrimExt(output)), "--header", "ncchheader.bin", "--romfs", "romfs-mod.bin"}
	xorpad := TitleID + ".Main.romfs.xorpad"
	if _, err := os.Stat(xorpad); err == nil {
		cfaArgs = append(cfaArgs, "--xor", xorpad)
	} else {
		fmt.Fprintf(os.Stderr, "Warning: xorpad %s not found; building unencrypted CFA\n", xorpad)
	}

	if err := runCmd("3dstool", cfaArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "3dstool cfa build failed: %v\n", err)
		exitWithCleanup(1)
	}
	generatedCFA = fmt.Sprintf("%s.cfa", stringsTrimExt(output))

	// adjust romfs size inside cfa
	if err := fixCFA(fmt.Sprintf("%s.cfa", stringsTrimExt(output))); err != nil {
		fmt.Fprintf(os.Stderr, "failed to fix cfa: %v\n", err)
		exitWithCleanup(1)
	}

	// make_cia
	fmt.Println("Running make_cia to produce .cia...")
	if err := runCmd("make_cia", "-v", "-o", output, fmt.Sprintf("--content0=%s.cfa", stringsTrimExt(output)), "--index_0=0"); err != nil {
		fmt.Fprintf(os.Stderr, "make_cia failed: %v\n", err)
		exitWithCleanup(1)
	}

	fmt.Printf("Created %s\n", output)
	exitWithCleanup(0)
}

func stringsTrimExt(s string) string {
	ext := filepath.Ext(s)
	return s[:len(s)-len(ext)]
}

func fixCFA(cfaPath string) error {
	fi, err := os.Stat("romfs-mod.bin")
	if err != nil {
		return err
	}
	romfsSize := uint32(fi.Size() / 0x200)

	f, err := os.OpenFile(cfaPath, os.O_RDWR, fs.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	// always write xor flag byte
	if _, err := f.Seek(0x18F, 0); err != nil {
		return err
	}
	if _, err := f.Write([]byte{0x00}); err != nil {
		return err
	}

	// write romfs size at 0x1B4 (little endian)
	if _, err := f.Seek(0x1B4, 0); err != nil {
		return err
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, romfsSize)
	if _, err := f.Write(buf); err != nil {
		return err
	}
	return nil
}

func cleanup() {
	_ = os.Remove("romfs/cbf_std.bcfnt.lz")
	_ = os.RemoveAll("romfs")
	_ = os.Remove("romfs-mod.bin")
	_ = os.Remove("ncchheader.bin")
	if generatedCFA != "" {
		_ = os.Remove(generatedCFA)
	}
}
