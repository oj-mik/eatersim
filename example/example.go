package main

import (
	"fmt"

	"github.com/oj-mik/eatersim"
	"github.com/oj-mik/eatersim/assembler"
)

const (
	nop = 0x00
	lda = 0x10
	add = 0x20
	sub = 0x30
	sta = 0x40
	ldi = 0x50
	jmp = 0x60
	jc  = 0x70
	jz  = 0x80

	out = 0xe0
	hlt = 0xf0
)

func main() {
	c := eatersim.NewBBCpu()

	src1 := "start:\n" +
		"  ADD adder\n" +
		"  JC complete\n" +
		"  JMP start\n" +
		"complete:\n" +
		"  OUT\n" +
		"  HLT\n" +
		"  .org 14\n" +
		"adder:\n" +
		"  .byte 33"

	bin1, err := assembler.Assemble(src1)
	if err != nil {
		fmt.Println(err)
	}

	src2 := "start:\n" +
		" lda 15\n" +
		" add 14\n" +
		" jc exit\n" +
		" jmp start\n" +
		"exit:\n" +
		" out\n" +
		" hlt\n" +
		" .org 14\n" +
		" .byte $f2\n" +
		" .byte $0f"

	bin2, err := assembler.Assemble(src2)
	if err != nil {
		fmt.Println(err)
	}

	c.RAM.Write(bin1)
	c.Run()
	fmt.Println("src1")
	fmt.Println(c.Oreg)
	fmt.Println()

	c.RAM.Write(bin2)
	c.Reset()
	c.Run()
	fmt.Println("src2")
	fmt.Println(c.Oreg)

}
