// assembler is an assembler for Ben Eater's 8-bit breadboard CPU.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oj-mik/eatersim/assembler"
)

var in, out string

func init() {
	flag.StringVar(&in, "i", "in.asm", "path of file to assemble")
	flag.StringVar(&out, "o", "a.out", "path of where to store assembled output file")
}

func main() {
	flag.Parse()

	infile, err := os.Open(in)
	if err != nil {
		fmt.Printf("Could not open input file: %s\n", err)
		return
	}
	defer infile.Close()

	outfile, err := os.Create(out)
	if err != nil {
		fmt.Printf("Could not create output file: %s\n", err)
		return
	}
	defer outfile.Close()

	bin, err := assembler.AssembleFrom(infile)
	if err != nil {
		fmt.Printf("Could not assemble input file: %s\n", err)
		return
	}

	_, err = outfile.Write(bin)
	if err != nil {
		fmt.Printf("Could not write to output file: %s\n", err)
		return
	}

}
