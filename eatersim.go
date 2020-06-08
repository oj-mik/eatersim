// Package eatersim implements an emulator for Ben Eaters 8-bit breadboard CPU.
//
// The components are abstracted to the breadboard level. All control signals
// and clock inputs have been implemented as noninverting, as in the software
// implementation the hardware constraints Ben faced does not apply.
//
// For further details, see Ben's web page at eater.net/8bit
package eatersim

import (
	"errors"
	"fmt"
)

// Clk represents the Clock-board
type Clk struct {
	// output signal holding the current clock state
	CLK bool

	// control signals
	// HLT stops blocks the CLK from pulsing when true
	HLT *bool
}

// NewClk creates a new clock board and initialize it's signals with the signals
// passed in the function call
func NewClk(hlt *bool) *Clk {
	c := new(Clk)
	c.HLT = hlt
	return c
}

// Executes the logic of the clock once. Updates the internal states based on
// the previous state of the clock and the current state of the signals.
func (c *Clk) Exec() {
	if !ptbool(c.HLT) {
		c.CLK = !c.CLK
	} else {
		c.CLK = false
	}
}

// Implements the Stringer-interface
func (c *Clk) String() string {
	s := fmt.Sprintf("CLK: %v\n", c.CLK)
	s += "\nactive control signals: "
	if ptbool(c.HLT) {
		s += "HLT"
	} else {
		s += "none"
	}
	return s
}

// Reg represents a generic 8-bit register board.
// Reads data from bus to buffer on positive clock edge when enable input is true.
// Writes data from register to bus when enable output is true.
type Reg struct {
	// internal register buffer
	BUF byte

	// bus signal
	// r/w
	BUS *byte

	// control signals
	// read only
	// CLK is the clock pulse
	// CLR clears the buffer
	// EI enables input from the bus to the register
	// EO enables output from the register to the bus
	CLK, CLR, EI, EO *bool

	// helper states
	clkprev, clkre bool
}

// NewReg creates a new 8-bit register board and initialize it's signals with
// the signals passed in the function call
func NewReg(bus *byte, clk, clr, ei, eo *bool) *Reg {
	r := new(Reg)
	r.BUS = bus
	r.CLK = clk
	r.CLR = clr
	r.EI = ei
	r.EO = eo
	return r
}

// Executes the logic of the register once. Updates the internal states and
// possibly the bus.
func (r *Reg) Exec() {
	r.clkre = ptbool(r.CLK) && !r.clkprev
	r.clkprev = ptbool(r.CLK)

	if ptbool(r.EI) && r.clkre {
		r.BUF = ptbyte(r.BUS)
	}
	if ptbool(r.CLR) {
		r.BUF = 0
	}
	if ptbool(r.EO) && r.BUS != nil {
		*r.BUS = r.BUF
	}
}

// Implements the Stringer-interface
func (r *Reg) String() string {
	s := fmt.Sprintf("BUF: %08b", r.BUF)
	s += "\nactive control signals: "
	f := false
	if ptbool(r.CLK) {
		s += "CLK"
		f = true
	}
	if ptbool(r.CLR) {
		if f {
			s += ", "
		}
		s += "CLR"
		f = true
	}
	if ptbool(r.EI) {
		if f {
			s += ", "
		}
		s += "EI"
		f = true
	}
	if ptbool(r.EO) {
		if f {
			s += ", "
		}
		s += "EO"
		f = true
	}
	if !f {
		s += "none"
	}
	return s
}

// Ireg represents the instruction register board.
// Reads data from bus to buffer on positive clock edge when enable input is true.
// Writes the 4 least significant bits from register to bus when enable output is true.
type Ireg struct {
	// internal instruction register buffer
	BUF byte

	// bus signal
	// r/w
	BUS *byte

	// control signals
	// read only
	// CLK is the clock pulse
	// CLR clears the buffer
	// EI enables input from the bus to the register
	// EO enables output from the register to the bus
	CLK, CLR, EI, EO *bool

	// helper states
	clkprev, clkre bool
}

// NewIreg creates a new instuction register board and initialize it's signals
// with the signals passed in the function call
func NewIreg(bus *byte, clk, clr, ei, eo *bool) *Ireg {
	r := new(Ireg)
	r.BUS = bus
	r.CLK = clk
	r.CLR = clr
	r.EI = ei
	r.EO = eo
	return r
}

// Executes the logic of the register once. Updates the internal states and the
// bus.
func (r *Ireg) Exec() {
	r.clkre = ptbool(r.CLK) && !r.clkprev
	r.clkprev = ptbool(r.CLK)

	if ptbool(r.EI) && r.clkre {
		r.BUF = ptbyte(r.BUS)
	}
	if ptbool(r.CLR) {
		r.BUF = 0
	}
	if ptbool(r.EO) && r.BUS != nil {
		*r.BUS = r.BUF & 0x0f
	}
}

// Implements the Stringer-interface
func (r *Ireg) String() string {
	s := fmt.Sprintf("BUF: %08b", r.BUF)
	s += "\nactive control signals: "
	f := false
	if ptbool(r.CLK) {
		s += "CLK"
		f = true
	}
	if ptbool(r.CLR) {
		if f {
			s += ", "
		}
		s += "CLR"
		f = true
	}
	if ptbool(r.EI) {
		if f {
			s += ", "
		}
		s += "EI"
		f = true
	}
	if ptbool(r.EO) {
		if f {
			s += ", "
		}
		s += "EO"
		f = true
	}
	if !f {
		s += "none"
	}
	return s
}

// Reg4 represents a 4 bit register used for the memory address register board.
// Reads the 4 least significant bits from bus to buffer on positive clock edge
// when enable input is true.
type Reg4 struct {
	// internal register buffer
	BUF byte

	// bus signal
	// read only
	BUS *byte

	// control signals
	// read only
	// CLK is the clock pulse
	// CLR clears the buffer
	// EI enables input from the bus to the register
	CLK, CLR, EI *bool

	// helper states
	clkprev, clkre bool
}

// NewReg4 creates a new 4-bit register board and initialize it's signals
// with the signals passed in the function call
func NewReg4(bus *byte, clk, clr, ei *bool) *Reg4 {
	r := new(Reg4)
	r.BUS = bus
	r.CLK = clk
	r.CLR = clr
	r.EI = ei
	return r
}

// Executes the logic of the register once. Updates the internal states.
func (r *Reg4) Exec() {
	r.clkre = ptbool(r.CLK) && !r.clkprev
	r.clkprev = ptbool(r.CLK)

	if ptbool(r.EI) && r.clkre {
		r.BUF = ptbyte(r.BUS) & 0x0f
	}
	if ptbool(r.CLR) {
		r.BUF = 0
	}
}

// Implements the Stringer-interface
func (r *Reg4) String() string {
	s := fmt.Sprintf("BUF: %04b", r.BUF&0x0f)
	s += "\nactive control signals: "
	f := false
	if ptbool(r.CLK) {
		s += "CLK"
		f = true
	}
	if ptbool(r.CLR) {
		if f {
			s += ", "
		}
		s += "CLR"
		f = true
	}
	if ptbool(r.EI) {
		if f {
			s += ", "
		}
		s += "EI"
		f = true
	}
	if !f {
		s += "none"
	}
	return s
}

// Mem represents the random access memory board. Reads the value from the bus
// and stores it in the memory at the address of the Address signal on positive
// clock edge if ram input signal is active. Writes the value from the the
// memory at the address of the Address signal to the bus if ram output signal
// is active.
type Mem struct {
	// memory
	MEM [0x10]byte

	// address signal
	// read only
	Addr *byte

	// bus signal
	// read/write
	BUS *byte

	// control signals
	// read only
	// CLK is the clock pulse
	// RI (ram input) enables input from the bus to the memory
	// RO (ram output) enables output from the memory to the bus
	CLK, RI, RO *bool

	// helper states
	clkprev, clkre bool
}

// NewMem creates a new random access memory board and initialize it's signals
// with the signals passed in the function call
func NewMem(addr, bus *byte, clk, ri, ro *bool) *Mem {
	m := new(Mem)
	m.Addr = addr
	m.BUS = bus
	m.CLK = clk
	m.RI = ri
	m.RO = ro
	return m
}

// Executes the logic of the memory once.
func (m *Mem) Exec() {
	m.clkre = ptbool(m.CLK) && !m.clkprev
	m.clkprev = ptbool(m.CLK)

	if ptbool(m.RI) && m.clkre {
		m.MEM[int(ptbyte(m.Addr)&0x0f)] = ptbyte(m.BUS)
	}

	if ptbool(m.RO) && m.BUS != nil {
		*m.BUS = m.MEM[int(ptbyte(m.Addr)&0x0f)]
	}
}

// Implements the Writer-interface. Overwrites the memory with the values in p.
// If p is greater than the memory, write will read the first 16 bytes of p into
// the memory and return an error. If p is shorter than 16 bytes, the remaining
// locations in the memory will be left untouched.
func (m *Mem) Write(p []byte) (n int, err error) {
	n = len(p)

	if n > 0x10 {
		n = 0x10
		err = errors.New("buffer larger than memory")
	}

	for i := 0; i < n; i++ {
		m.MEM[i] = p[i]
	}

	return
}

// Implements the Stringer-interface
func (m *Mem) String() string {
	s := fmt.Sprintf("Addr: %04b, MEM: %04b", ptbyte(m.Addr)&0x0f, m.MEM[int(ptbyte(m.Addr)&0x0f)])
	s += "\nactive control signals: "
	f := false
	if ptbool(m.CLK) {
		s += "CLK"
		f = true
	}
	if ptbool(m.RI) {
		if f {
			s += ", "
		}
		s += "RI"
		f = true
	}
	if ptbool(m.RO) {
		if f {
			s += ", "
		}
		s += "RO"
		f = true
	}
	if !f {
		s += "none"
	}
	return s
}

// Alu represents the arithmetic logic unit board. Reads the values from the
// Areg and Breg signals and calculates the sum or difference depending on the
// state of the subtract signal. Writes the calculated value to the bus if
// enable output is active. Activates the carry flag or the zero flag on
// positive clock edge, depending on whether the current calculation causes
// integer overflow or the result is zero.
type Alu struct {
	// buffer
	BUF byte

	// register inputs
	// read only
	Areg, Breg *byte

	// bus signal
	// write only
	BUS *byte

	// control signals
	// read only
	// CLK is the clock pulse
	// CLR clears the current value from the carry and zero flags
	// EO enables output from the alu to the bus
	// SU selects substitute as the arithmetic operation
	// FI (flag input) updates the carry and zero flags
	CLK, CLR, EO, SU, FI *bool

	// flags
	// write only
	// CF is the carry flag
	// ZF is the zero flag
	CF, ZF bool

	// helper variables
	clkprev, clkre bool
	bufCF, bufZF   bool
}

// NewAlu creates a new arithmetic logic unit board and initialize it's signals
// with the signals passed in the function call
func NewAlu(areg, breg, bus *byte, clk, clr, eo, su, fi *bool) *Alu {
	a := new(Alu)
	a.Areg = areg
	a.Breg = breg
	a.BUS = bus
	a.CLK = clk
	a.CLR = clr
	a.EO = eo
	a.SU = su
	a.FI = fi
	return a
}

// Executes the logic of the ALU once.
func (a *Alu) Exec() {
	a.clkre = ptbool(a.CLK) && !a.clkprev
	a.clkprev = ptbool(a.CLK)

	if ptbool(a.FI) && a.clkre {
		a.CF = a.bufCF
		a.ZF = a.bufZF
	}

	if ptbool(a.CLR) {
		a.CF = false
		a.ZF = false
	}

	if !ptbool(a.SU) {
		// adding
		a.BUF = ptbyte(a.Areg) + ptbyte(a.Breg)

		// set flags reg
		a.bufCF = 0xff < int(ptbyte(a.Areg))+int(ptbyte(a.Breg))
		a.bufZF = a.BUF == 0

	} else {
		// subtracting
		a.BUF = ptbyte(a.Areg) - ptbyte(a.Breg)

		// set flags reg
		a.bufCF = ptbyte(a.Areg) < ptbyte(a.Breg)
		a.bufZF = a.BUF == 0
	}

	if ptbool(a.EO) && a.BUS != nil {
		*a.BUS = a.BUF
	}
}

// Implements the Stringer-interface
func (a *Alu) String() string {
	s := fmt.Sprintf("BUF: %08b, Areg: %08b, Breg: %08b", a.BUF, ptbyte(a.Areg), ptbyte(a.Breg))
	s += "\nactive flags: "
	f := false
	if a.CF {
		if f {
			s += ", "
		}
		s += "CF"
		f = true
	}
	if a.ZF {
		if f {
			s += ", "
		}
		s += "ZF"
		f = true
	}
	if !f {
		s += "none"
	}
	s += "\nactive control signals: "
	f = false
	if ptbool(a.CLK) {
		s += "CLK"
		f = true
	}
	if ptbool(a.CLR) {
		if f {
			s += ", "
		}
		s += "CLR"
		f = true
	}
	if ptbool(a.EO) {
		if f {
			s += ", "
		}
		s += "EO"
		f = true
	}
	if ptbool(a.SU) {
		if f {
			s += ", "
		}
		s += "SU"
		f = true
	}
	if ptbool(a.FI) {
		if f {
			s += ", "
		}
		s += "FI"
		f = true
	}
	if !f {
		s += "none"
	}
	return s
}

// Ctr represents the program counter board. The counter increments its counter
// value on positive clock edge when count enable is active. The counter reads
// the four least significant bits from the bus and stores them in the counter
// value on positive clock edge when the jump signal is active. It writes the
// four least significant bits from the counter value to the bus when counter
// output is active.
type Ctr struct {
	// counter value
	// read write
	CNT byte

	// bus signals
	// read/write
	BUS *byte

	// control signals
	// read only
	// CLK is the clock pulse
	// CLR clears the current value from the counter
	// J (jump) input the value of the bus into the counter
	// CE (counter enable) increment the counter value by one
	CLK, CLR, CO, J, CE *bool

	// helper states
	clkprev, clkre bool
}

// NewCtr creates a new program counter board and initialize it's signals
// with the signals passed in the function call
func NewCtr(bus *byte, clk, clr, co, j, ce *bool) *Ctr {
	c := new(Ctr)
	c.BUS = bus
	c.CLK = clk
	c.CLR = clr
	c.CO = co
	c.J = j
	c.CE = ce
	return c
}

// Executes the logic of the counter once.
func (c *Ctr) Exec() {
	c.clkre = ptbool(c.CLK) && !c.clkprev
	c.clkprev = ptbool(c.CLK)

	if ptbool(c.CE) && c.clkre {
		c.CNT = (c.CNT + 1) & 0x0f
	}

	if ptbool(c.J) && c.clkre {
		c.CNT = ptbyte(c.BUS) & 0x0f
	}

	if ptbool(c.CLR) {
		c.CNT = 0
	}

	if ptbool(c.CO) && c.BUS != nil {
		*c.BUS = c.CNT & 0x0f
	}
}

// Implements the Stringer-interface
func (c *Ctr) String() string {
	s := fmt.Sprintf("CNT: %04b", c.CNT)
	s += "\nactive control signals: "
	f := false
	if ptbool(c.CLK) {
		s += "CLK"
		f = true
	}
	if ptbool(c.CLR) {
		if f {
			s += ", "
		}
		s += "CLR"
		f = true
	}
	if ptbool(c.CO) {
		if f {
			s += ", "
		}
		s += "CO"
		f = true
	}
	if ptbool(c.J) {
		if f {
			s += ", "
		}
		s += "J"
		f = true
	}
	if ptbool(c.CE) {
		if f {
			s += ", "
		}
		s += "CE"
		f = true
	}
	if !f {
		s += "none"
	}
	return s
}

// Ctrl represents the control logic board. It reads the current instruction
// code from Inst and uses the instruction code in combination with an internal
// micro instruction counter to determine which control signals should be set.
// The internal micro instruction counter is incremented on falling clock edge.
type Ctrl struct {
	// Instruction code signal
	// read only
	Inst *byte

	// micro instruction counter
	Cnt byte

	// clock signal
	// read only
	CLK *bool

	// status flags
	// read only
	// CF is the carry flag from the alu
	// ZF is the zero flag from the alu
	CF, ZF *bool

	// general control flags
	// CLR is the signal to clear and reset registers
	// HLT is the signal to halt the clock
	CLR, HLT bool

	// a register control flags
	// AI is the signel to read from the bus into register a
	// AO is the signal to write from register a to the bus
	AI, AO bool

	// b register control flag
	// BI is the signal to read from the bus into register b
	BI bool

	// output register control flag
	// OI is the signal to read from the bus into the output register
	OI bool

	// memory address register control flag
	// MI is the signal to read from the bus into the memory address register
	MI bool

	// instruction register control flag
	// II is the signal to read from the bus into the instruction register
	// IO is the signal to write from the instruction register into the bus
	II, IO bool

	// arithmetic logic unit control flag
	// EO is the signal to write the current calculation value into the bus
	// SU is the signal to select subtraction
	// FI is the signal to update the carry and zero flag
	EO, SU, FI bool

	// program counter control flag
	// CO is the signal to output the counter value to the bus
	// J is the signal to read the counter value from the bus
	// CE is the signal to increment the counter value
	CO, J, CE bool

	// random access memory control flag
	// RI is the signal to read value from the bus and store in ram
	// RO is the signal to read value from ram and write to bus
	RI, RO bool

	// helper states
	clkprev, clkfe bool
	clrrst         int
}

// NewCtrl creates a new control logic board and initialize it's signals
// with the signals passed in the function call
func NewCtrl(inst *byte, clk, clr, cf, zf *bool) *Ctrl {
	c := new(Ctrl)
	c.Inst = inst
	c.CLK = clk
	c.CF = cf
	c.ZF = zf
	return c
}

// Exec executes the logic of the control logic board once.
func (c *Ctrl) Exec() {
	c.clkfe = !ptbool(c.CLK) && c.clkprev
	c.clkprev = ptbool(c.CLK)

	if c.clkfe {
		c.Cnt++
	}

	if c.Cnt == 5 {
		c.Cnt = 0
	}

	c.resetFlags()

	if c.CLR {
		c.Cnt = 0
		c.HLT = false
		if c.clrrst == 1 {
			c.CLR = false
			c.clrrst = 0
		}
		if c.clrrst > 0 {
			c.clrrst -= 1
		}
		return
	}

	switch c.Cnt {
	case 0:
		// fetch 1
		c.CO, c.MI = true, true

	case 1:
		// fetch 2
		c.RO, c.II, c.CE = true, true, true

	default:
		switch ptbyte(c.Inst) >> 4 {
		case 0x0:
			// nop

		case 0x1:
			// lda
			switch c.Cnt {
			case 2:
				c.IO, c.MI = true, true
			case 3:
				c.RO, c.AI = true, true
			}

		case 0x2:
			// add
			switch c.Cnt {
			case 2:
				c.IO, c.MI = true, true
			case 3:
				c.RO, c.BI = true, true
			case 4:
				c.EO, c.AI, c.FI = true, true, true
			}

		case 0x3:
			// sub
			switch c.Cnt {
			case 2:
				c.IO, c.MI = true, true
			case 3:
				c.RO, c.BI = true, true
			case 4:
				c.EO, c.AI, c.SU, c.FI = true, true, true, true
			}

		case 0x4:
			// sta
			switch c.Cnt {
			case 2:
				c.IO, c.MI = true, true
			case 3:
				c.AO, c.RI = true, true
			}

		case 0x5:
			// ldi
			switch c.Cnt {
			case 2:
				c.IO, c.AI = true, true
			}

		case 0x6:
			// jmp
			switch c.Cnt {
			case 2:
				c.IO, c.J = true, true
			}

		case 0x7:
			// jc
			switch c.Cnt {
			case 2:
				if ptbool(c.CF) {
					c.IO, c.J = true, true
				}
			}

		case 0x8:
			// jz
			switch c.Cnt {
			case 2:
				if ptbool(c.ZF) {
					c.IO, c.J = true, true
				}
			}

		case 0xe:
			// out
			switch c.Cnt {
			case 2:
				c.AO, c.OI = true, true
			}

		case 0xf:
			// hlt
			switch c.Cnt {
			case 2:
				c.HLT = true
			}
		}
	}
}

// Reset activates the CLR flag and keeps it active until the second call to Exec.
func (c *Ctrl) Reset() {
	c.CLR = true
	c.clrrst = 2
}

// Implements the Stringer-interface
func (c *Ctrl) String() string {
	s := fmt.Sprintf("Inst: %04b, CNT: %04b", ptbyte(c.Inst)>>4, c.Cnt&0x0f)

	s += "\nactive status flags: "
	f := false
	if ptbool(c.CF) {
		if f {
			s += ", "
		}
		s += "CF"
		f = true
	}
	if ptbool(c.ZF) {
		if f {
			s += ", "
		}
		s += "ZF"
		f = true
	}
	if !f {
		s += "none"
	}

	s += "\nactive control flags: "
	f = false
	if c.AI {
		s += "AI"
		f = true
	}
	if c.AO {
		if f {
			s += ", "
		}
		s += "AO"
		f = true
	}
	if c.BI {
		if f {
			s += ", "
		}
		s += "BI"
		f = true
	}
	if c.OI {
		if f {
			s += ", "
		}
		s += "OI"
		f = true
	}
	if c.MI {
		if f {
			s += ", "
		}
		s += "MI"
		f = true
	}
	if c.II {
		if f {
			s += ", "
		}
		s += "II"
		f = true
	}
	if c.IO {
		if f {
			s += ", "
		}
		s += "IO"
		f = true
	}
	if c.EO {
		if f {
			s += ", "
		}
		s += "EO"
		f = true
	}
	if c.SU {
		if f {
			s += ", "
		}
		s += "SU"
		f = true
	}
	if c.FI {
		if f {
			s += ", "
		}
		s += "FI"
		f = true
	}
	if c.CO {
		if f {
			s += ", "
		}
		s += "CO"
		f = true
	}
	if c.J {
		if f {
			s += ", "
		}
		s += "J"
		f = true
	}
	if c.CE {
		if f {
			s += ", "
		}
		s += "CE"
		f = true
	}
	if c.RI {
		if f {
			s += ", "
		}
		s += "RI"
		f = true
	}
	if c.RO {
		if f {
			s += ", "
		}
		s += "RO"
		f = true
	}
	if !f {
		s += "none"
	}

	s += "\nactive control signals: "
	f = false
	if ptbool(c.CLK) {
		s += "CLK"
		f = true
	}
	if !f {
		s += "none"
	}

	return s
}

func (c *Ctrl) resetFlags() {
	// a register control flags
	c.AI, c.AO = false, false

	// b register control flag
	c.BI = false

	// output register control flag
	c.OI = false

	// memory address register control flag
	c.MI = false

	// instruction register control flag
	c.II, c.IO = false, false

	// arithmetic logic unit control flag
	c.EO, c.SU, c.FI = false, false, false

	// program counter control flag
	c.CO, c.J, c.CE = false, false, false

	// random access memory control flag
	c.RI, c.RO = false, false
}

// BBCpu represents a complete default setup of the Ben Eater 8 bit breadbord
// cpu.
type BBCpu struct {
	// Clock board
	CLK *Clk

	// Control logics board
	CL *Ctrl

	// A register, B register and output register board
	Areg, Breg, Oreg *Reg

	// Memory Address Register board
	MAR *Reg4

	// Instruction Register board
	IR *Ireg

	// Arithmetic Logic Unit board
	ALU *Alu

	// Program Counter board
	PC *Ctr

	// Random Access Memory board
	RAM *Mem

	// Data Bus
	BUS byte
}

// NewBBCpu creates a new 8-bit breadboard CPU and initialize the interface
// between all the boards according to Ben Eaters instructions.
func NewBBCpu() *BBCpu {
	cpu := new(BBCpu)

	cpu.CL = new(Ctrl)

	cpu.CLK = NewClk(&cpu.CL.HLT)

	cpu.Areg = NewReg(&cpu.BUS, &cpu.CLK.CLK, &cpu.CL.CLR, &cpu.CL.AI, &cpu.CL.AO)
	cpu.Breg = NewReg(&cpu.BUS, &cpu.CLK.CLK, &cpu.CL.CLR, &cpu.CL.BI, nil)
	cpu.Oreg = NewReg(&cpu.BUS, &cpu.CLK.CLK, &cpu.CL.CLR, &cpu.CL.OI, nil)

	cpu.ALU = NewAlu(&cpu.Areg.BUF, &cpu.Breg.BUF, &cpu.BUS, &cpu.CLK.CLK, &cpu.CL.CLR, &cpu.CL.EO, &cpu.CL.SU, &cpu.CL.FI)

	cpu.MAR = NewReg4(&cpu.BUS, &cpu.CLK.CLK, &cpu.CL.CLR, &cpu.CL.MI)
	cpu.RAM = NewMem(&cpu.MAR.BUF, &cpu.BUS, &cpu.CLK.CLK, &cpu.CL.RI, &cpu.CL.RO)

	cpu.PC = NewCtr(&cpu.BUS, &cpu.CLK.CLK, &cpu.CL.CLR, &cpu.CL.CO, &cpu.CL.J, &cpu.CL.CE)

	cpu.IR = NewIreg(&cpu.BUS, &cpu.CLK.CLK, &cpu.CL.CLR, &cpu.CL.II, &cpu.CL.IO)

	cpu.CL.CLK = &cpu.CLK.CLK
	cpu.CL.Inst = &cpu.IR.BUF
	cpu.CL.CF = &cpu.ALU.CF
	cpu.CL.ZF = &cpu.ALU.ZF

	return cpu
}

// Exec executes the control logic of all the boards once.
func (c *BBCpu) Exec() {
	c.CLK.Exec()
	c.CL.Exec()
	c.Areg.Exec()
	c.Breg.Exec()
	c.Oreg.Exec()
	c.ALU.Exec()
	c.MAR.Exec()
	c.RAM.Exec()
	c.PC.Exec()
	c.IR.Exec()
}

// Run executes the logic of the breadboard cpu until it halts
func (c *BBCpu) Run() {
	for !c.CL.HLT {
		c.Exec()
	}
}

// Instruction executes the logic of the breadboard cpu until the current
// instruction is complete. Returns immediately if clr or hlt is active.
func (c *BBCpu) Instruction() {
	c.Exec()

	for !(c.CL.Cnt == 4 && c.CLK.CLK) {
		c.Exec()
	}

}

// Step executes the logic of the breadboard cpu twice, which means one full
// clock cycle if the cpu is not halted. Synchronism to rising/falling edge of
// clock must be checked manually.
func (c *BBCpu) Step() {
	c.Exec()
	c.Exec()
}

// Half step executes the logic of the breadboard cpu once.
func (c *BBCpu) HalfStep() {
	c.Exec()
}

// Reset resets the breadboard cpu.
func (c *BBCpu) Reset() {
	c.CL.Reset()
	c.Exec()
	c.Exec()
}

// String implements the Stringer-interface
func (c *BBCpu) String() string {
	s := fmt.Sprintf("bus:\nBUS: %08b\n\n", c.BUS)
	s += fmt.Sprintf("areg:\n%s\n\n", c.Areg)
	s += fmt.Sprintf("breg:\n%s\n\n", c.Breg)
	s += fmt.Sprintf("alu:\n%s\n\n", c.ALU)
	s += fmt.Sprintf("pc:\n%s\n\n", c.PC)
	s += fmt.Sprintf("mar:\n%s\n\n", c.MAR)
	s += fmt.Sprintf("ram:\n%s\n\n", c.RAM)
	s += fmt.Sprintf("ir:\n%s\n\n", c.IR)
	s += fmt.Sprintf("cl:\n%s\n\n", c.CL)
	s += fmt.Sprintf("oreg:\n%s", c.Oreg)
	return s
}

// interprets nil-pointers as false, else return the value pointed to by p
func ptbool(p *bool) bool {
	if p != nil {
		return *p
	}
	return false
}

// interprets nil-pointers as 0x00, else return the value pointed to by p
func ptbyte(p *byte) byte {
	if p != nil {
		return *p
	}
	return 0
}
