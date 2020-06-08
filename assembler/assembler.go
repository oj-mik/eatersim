// Package assembler implements an assembler for Ben Eaters 8-bit breadboard CPU.
//
// The assembler is inspired by the assembler used in Ben's 6502 computer series,
// and attempts to implement similar functionality where apliccable to the 8-bit
// CPU.
//
// For further details, see Ben's web page at eater.net/8bit
//
// The assembler has the following features:
//  - support for all instructions detailed in Ben's tutorial videos.
//    * NOP         - No operation
//    * LDA regaddr - Load value from memory address 'regaddr' to A register
//    * ADD regaddr - Add value from memory address 'regaddr' to current value in A register
//    * SUB regaddr - Subtract value from memory address 'regaddr' to current value in A register
//    * STA regaddr - Store value from A register to memory address 'regaddr'
//    * LDI value   - Load value to A register
//    * JMP regaddr - Jump to instruction in memory address 'regaddr'
//    * JC  regaddr - Jump on carry to instruction in memory address 'regaddr'
//    * JZ  regaddr - Jump on zero to instruction in memory address 'regaddr'
//    * OUT         - Output A register to Output register
//    * HLT         - Halt the execution
//  - support for .org and .byte directives.
//    * .org  - instruct the assembler to move to register address passed as parameter.
//    * .byte - instruct the assembler to store raw value to register.
//  - support for symbols and labels which may be passed as parameters by name to instructions.
//    * symbol=value
//    * label:
//  - support for comments.
//    *   ADD 15 ;comment after semicolon
//  - decimal, hexadecimal or binary representation of values.
//    * 15 is decimal representation of 15
//    * $0f is hexadecimal representation of 15
//    * %1111 is binary representation of 15
//
//
// Instructions and dot directives must be preceeded by a whitespace character.
// Symbol names and label names must start at the first character of the line.
// Symbol names and label names may contain any graphic unicode character as
// defined by go's unicode.IsGraphic(), except reserved characters '$', '%', '#', '.', ';' and '='.
// Instructions and directives are not case sensitive. Symbols and labels are.

package assembler

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

const (
	dotOrg  = 0x01
	dotByte = 0x02

	label  = 0x09
	symbol = 0x0a

	noCode = 0xff

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

type codeline struct {
	instr byte
	value byte

	label string
}

// AssembleFrom reads assembly code from reader r and returns the assembled
// binary as a byte slice. If errors are encountered, an empty byte slice will
// be returned, together with the error.
func AssembleFrom(r io.Reader) ([]byte, error) {
	src := make([]byte, 2048, 2048)

	var e error
	var n int
	for e != io.EOF {
		var i int
		i, e = r.Read(src[n:cap(src)])
		n += i
		if i == 0 && e != io.EOF {
			b := make([]byte, n+2048)
			copy(b, src)
			src = b[:cap(b)]
		}

	}
	bin, e := Assemble(string(src[:n]))

	return bin, e
}

// Assemble parses assembly code passed as src and returns the assembled
// binary as a byte slice. If errors are encountered, an empty byte slice will
// be returned, together with the error.
func Assemble(src string) ([]byte, error) {
	cls, err := decode(src)
	if err != nil {
		return nil, err
	}
	bin, err := assemble(cls)
	if err != nil {
		return nil, err
	}
	return bin, nil
}

func assemble(cls []codeline) ([]byte, error) {
	labels, e := mapLabels(cls)
	if e != nil {
		return nil, e
	}

	var raddr int
	bin := make([]byte, 16)
	used := make([]bool, 16)
	for i := range cls {
		// must add check for raddr out of bounds (panic) and overwriting of already
		// written register
		e = cls[i].assembleLn(bin, &raddr, used, labels)
		if e != nil {
			return nil, e
		}
	}
	return bin, nil
}

func (cl codeline) assembleLn(reg []byte, raddr *int, used []bool, labels map[string]byte) error {
	switch cl.instr {
	case noCode, label, symbol:
	case nop, out, hlt:
		if *raddr > 15 {
			return errors.New("program exceeds registry size of 16 bytes")
		}
		if used[*raddr] {
			return fmt.Errorf("registry address conflict at address %v, check .org directives", *raddr)
		}
		used[*raddr] = true
		reg[*raddr] = cl.instr
		*raddr++
	case lda, add, sub, sta, ldi, jmp, jc, jz:
		if *raddr > 15 {
			return errors.New("program exceeds registry size of 16 bytes")
		}
		if used[*raddr] {
			return fmt.Errorf("registry address conflict at address %v, check .org directives", *raddr)
		}
		used[*raddr] = true
		if cl.label == "" {
			reg[*raddr] = cl.instr | (cl.value & 0x0f)
		} else {
			if _, ok := labels[cl.label]; !ok {
				return errors.New("Unknown symbol: " + cl.label)
			}
			if labels[cl.label] > 0x0f {
				return fmt.Errorf("symbol %s holds value greater than 15 while used as parameter in instruction.", cl.label)
			}
			reg[*raddr] = cl.instr | (labels[cl.label] & 0x0f)
		}
		*raddr++
	case dotOrg:
		*raddr = int(cl.value)
	case dotByte:
		if *raddr > 15 {
			return errors.New("program exceeds registry size of 16 bytes")
		}
		if used[*raddr] {
			return fmt.Errorf("registry address conflict at address %v, check .org directives", *raddr)
		}
		used[*raddr] = true
		reg[*raddr] = cl.value
		*raddr++
	}
	return nil
}

func (cl codeline) mapLabel(raddr *int, labels *map[string]byte) error {
	switch cl.instr {
	case label:
		if _, ok := (*labels)[cl.label]; ok {
			return errors.New("duplicate label: " + cl.label)
		}
		(*labels)[cl.label] = byte(*raddr)

	case symbol:
		if _, ok := (*labels)[cl.label]; ok {
			return errors.New("duplicate label: " + cl.label)
		}
		(*labels)[cl.label] = cl.value

	case noCode:
	case nop, lda, add, sub, sta, ldi, jmp, jc, jz, out, hlt:
		*raddr++
	case dotOrg:
		*raddr = int(cl.value)
	case dotByte:
		*raddr++
	}
	return nil
}

func mapLabels(cls []codeline) (map[string]byte, error) {
	var regaddr int
	labels := make(map[string]byte)

	for i := range cls {
		e := cls[i].mapLabel(&regaddr, &labels)
		if e != nil {
			return nil, e
		}
	}
	return labels, nil
}

func trimcomments(lns string) string {
	n := strings.IndexRune(lns, ';')
	if n != -1 {
		return lns[:n]
	}
	return lns
}

func decode(src string) ([]codeline, error) {
	lns := strings.Split(src, "\n")

	var err error
	cls := make([]codeline, len(lns))
	cnt := 0

	for i := range lns {
		cls[cnt], err = decodeln(lns[i])
		if err != nil {
			return nil, err
		}
		if cls[cnt].instr != noCode {
			cnt++
		}
	}

	return cls[:cnt], nil
}

func decodeln(ln string) (codeline, error) {

	s := trimcomments(ln) // remove comments
	s = toSingleSpace(s)  // convert all sequences of whitespace characters to a single space

	switch len(s) { // must return if len(s) < 2 to avoid panic
	case 0:
		return codeline{instr: noCode}, nil

	case 1:
		if s[0] != ' ' {
			return codeline{instr: noCode}, errors.New("error decoding: " + ln)
		}
		return codeline{instr: noCode}, nil
	}

	// the following decode*(s) functions might panic if they are passed an
	// empty string or a string containing only whitespace, however above we have
	// verified that s is at least two characters long and that there are no
	// double spaces, so we should be safe.
	var cl codeline
	var err error

	switch {
	case s[0] == ' ' && s[1] == '.':
		// is dotdirective
		s = strings.TrimSpace(s)
		cl, err = decodeDotDir(s)

	case s[0] == ' ' && s[1] != '.':
		// is instruction
		s = strings.TrimSpace(s)
		cl, err = decodeInstr(s)

	case s[0] != ' ':
		// is symbol or label
		s = strings.TrimSpace(s)
		cl, err = decodeSymbol(s)
	}

	if err != nil {
		return cl, fmt.Errorf("error decoding: \"%s\": %s", ln, err)
	}
	return cl, err
}

func decodeDotDir(ln string) (codeline, error) {
	ss := strings.Split(ln, " ")
	if len(ss) != 2 {
		return codeline{instr: noCode}, errors.New("incorrect number of parameters")
	}

	var cl codeline
	var err error

	switch strings.ToLower(ss[0]) {
	case ".org":
		cl.instr = dotOrg
	case ".byte":
		cl.instr = dotByte
	default:
		cl.instr = noCode
		return cl, errors.New("unknown dot-directive")
	}

	cl.value, err = decodeVal(ss[1], 8)
	if err != nil {
		cl.instr = noCode
		cl.value = 0
	}

	return cl, err
}

func decodeInstr(ln string) (codeline, error) {
	ss := strings.Split(ln, " ")

	var cl codeline
	var err error

	switch strings.ToLower(ss[0]) {
	case "nop":
		cl.instr = nop
	case "lda":
		cl.instr = lda
	case "add":
		cl.instr = add
	case "sub":
		cl.instr = sub
	case "sta":
		cl.instr = sta
	case "ldi":
		cl.instr = ldi
	case "jmp":
		cl.instr = jmp
	case "jc":
		cl.instr = jc
	case "jz":
		cl.instr = jz
	case "out":
		cl.instr = out
	case "hlt":
		cl.instr = hlt
	default:
		cl.instr = noCode
		err = fmt.Errorf("unknown instruction %s", ss[0])
		return cl, err
	}

	switch cl.instr {
	case nop, out, hlt:
		if len(ss) > 1 {
			cl.instr = noCode
			err = fmt.Errorf("unexpected parameters after instruction %s", ss[0])
			return cl, err
		}
	case lda, add, sub, sta, ldi, jmp, jc, jz:
		if len(ss) != 2 {
			cl.instr = noCode
			err = fmt.Errorf("expecting 1 parameter after instruction %s, got %v", ss[0], len(ss)-1)
		}

		if ss[1][0] == '$' || ss[1][0] == '%' || unicode.IsDigit([]rune(ss[1])[0]) {
			cl.value, err = decodeVal(ss[1], 4)
			if err != nil {
				cl.instr = noCode
				cl.value = 0
				return cl, err
			}
		} else {
			if r := checkSymbol(ss[1]); r != "" {
				cl.instr = noCode
				err = fmt.Errorf("illegal character '%s' in value '%s'", r, ss[1])
				return cl, err
			}
			cl.label = ss[1]
		}

	}

	return cl, err
}

func decodeSymbol(ln string) (codeline, error) {
	var cl codeline
	var err error

	switch {
	case strings.Contains(ln, "="):
		// is symbol
		ss := strings.Split(ln, "=")
		if len(ss) != 2 {
			cl = codeline{instr: noCode}
			err = fmt.Errorf("found more than one equal sign in symbol statement")
			return cl, err
		}
		ss[0] = strings.Trim(ss[0], " ")
		ss[1] = strings.Trim(ss[1], " ")

		if r := checkSymbol(ss[0]); r != "" {
			cl.instr = noCode
			err = fmt.Errorf("illegal character '%s' in symbol %s", r, ss[1])
			return cl, err
		}
		cl = codeline{instr: symbol}
		cl.value, err = decodeVal(ss[1], 8)
		if err != nil {
			cl.instr = noCode
			return cl, err
		}
		return cl, err
	case strings.Contains(ln, ":"):
		// is label
		if n := strings.Count(ln, ":"); n != 1 {
			cl = codeline{instr: noCode}
			err = fmt.Errorf("found more than one colon in label statement")
			return cl, err
		}
		s := strings.Trim(ln, ": ")
		if r := checkSymbol(s); r != "" {
			cl = codeline{instr: noCode}
			err = fmt.Errorf("illegal character '%s' in label %s", r, s)
			return cl, err
		}
		cl = codeline{instr: label}
		cl.label = s
		return cl, err
	default:
		// is illegal
		cl = codeline{instr: noCode}
		err = fmt.Errorf("left justified text must be 'symbol=value' or 'label:', leading whitespace missing?")
		return cl, err
	}

	return cl, err
}

func decodeVal(s string, bitSize int) (byte, error) {
	var r uint64
	var e error
	switch {
	case s[0] == '$':
		r, e = strconv.ParseUint(s[1:], 16, bitSize)
	case s[0] == '%':
		r, e = strconv.ParseUint(s[1:], 2, bitSize)
	default:
		r, e = strconv.ParseUint(s, 10, bitSize)
	}
	return byte(r), e
}

func checkSymbol(s string) string {
	for _, r := range []rune(s) {
		if !unicode.IsGraphic(r) {
			return strconv.QuoteRune(r)
		}
		switch r {
		case '$', '#', '.', ';', '=', ' ', '%':
			return string(r)
		}
	}
	return ""
}

// toSingleSpace converts all sequences of whitespace characters within s into
// a single space
func toSingleSpace(s string) string {
	r := strings.Replace(s, "\t", " ", -1)
	r = strings.Replace(r, "\n", " ", -1)
	r = strings.Replace(r, "\v", " ", -1)
	r = strings.Replace(r, "\f", " ", -1)
	r = strings.Replace(r, "\r", " ", -1)
	r = strings.Replace(r, "\u0085", " ", -1) // next line-character
	r = strings.Replace(r, "\u00a0", " ", -1) // no brake space-character
	for strings.Contains(r, "  ") {
		r = strings.Replace(r, "  ", " ", -1)
	}
	return r
}
