package cpu

import (
	"log"
)

const (
	ModeUser = 0x10
	ModeFIQ  = 0x11
	ModeIRQ  = 0x12
	ModeSVC  = 0x13
	ModeAbt  = 0x17
	ModeUnd  = 0x1B
	ModeSys  = 0x1F
)

type MemoryInterface interface {
	Read8(addr uint32) byte
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, val byte)
	Write16(addr uint32, val uint16)
	Write32(addr uint32, val uint32)
}

type CPU struct {
	Bus MemoryInterface

	// R0-R15
	R [16]uint32

	// Current Program Status Register
	CPSR uint32

	// Saved Program Status Registers
	SPSR_fiq uint32
	SPSR_svc uint32
	SPSR_abt uint32
	SPSR_irq uint32
	SPSR_und uint32

	// Banked Registers
	R8_fiq, R9_fiq, R10_fiq, R11_fiq, R12_fiq uint32
	R13_fiq, R14_fiq                          uint32
	R13_svc, R14_svc                          uint32
	R13_abt, R14_abt                          uint32
	R13_irq, R14_irq                          uint32
	R13_und, R14_und                          uint32

	R8_usr, R9_usr, R10_usr, R11_usr, R12_usr uint32
	R13_usr, R14_usr                          uint32

	IsARM9 bool
	Cycles uint64

	// Look-up tables for instruction decoding
	ARMLUT   [4096]func(uint32)
	THUMBLUT [1024]func(uint16)
}

func NewCPU(isARM9 bool, bus MemoryInterface) *CPU {
	c := &CPU{
		IsARM9: isARM9,
		Bus:    bus,
	}
	c.initLUTs()
	return c
}

func (c *CPU) initLUTs() {
	for i := 0; i < 4096; i++ {
		c.ARMLUT[i] = c.unimplementedARM
	}
	for i := 0; i < 1024; i++ {
		c.THUMBLUT[i] = c.unimplementedTHUMB
	}
}

func (c *CPU) unimplementedARM(opcode uint32) {
	log.Fatalf("Unimplemented ARM opcode: %08X at PC: %08X", opcode, c.R[15]-8)
}

func (c *CPU) unimplementedTHUMB(opcode uint16) {
	log.Fatalf("Unimplemented THUMB opcode: %04X at PC: %08X", opcode, c.R[15]-4)
}

func (c *CPU) Step() {
	// Check if THUMB state (T bit in CPSR)
	isThumb := (c.CPSR & 0x20) != 0

	if isThumb {
		opcode := c.Bus.Read16(c.R[15])
		c.R[15] += 2
		index := opcode >> 6 // 10-bit index for THUMB LUT
		c.THUMBLUT[index](opcode)
	} else {
		opcode := c.Bus.Read32(c.R[15])
		c.R[15] += 4
		
		// Condition code check (top 4 bits)
		if !c.checkCondition(opcode >> 28) {
			return
		}

		// Use bits 27-16 to index the 4096-entry ARM LUT
		index := (opcode >> 16) & 0xFFF
		c.ARMLUT[index](opcode)
	}
}

func (c *CPU) checkCondition(cond uint32) bool {
	// N, Z, C, V flags in CPSR
	n := (c.CPSR & 0x80000000) != 0
	z := (c.CPSR & 0x40000000) != 0
	c_flag := (c.CPSR & 0x20000000) != 0
	v := (c.CPSR & 0x10000000) != 0

	switch cond {
	case 0x0: return z
	case 0x1: return !z
	case 0x2: return c_flag
	case 0x3: return !c_flag
	case 0x4: return n
	case 0x5: return !n
	case 0x6: return v
	case 0x7: return !v
	case 0x8: return c_flag && !z
	case 0x9: return !c_flag || z
	case 0xA: return n == v
	case 0xB: return n != v
	case 0xC: return !z && (n == v)
	case 0xD: return z || (n != v)
	case 0xE: return true
	case 0xF: return false // Actually depends on ARM version, mostly false for execution
	}
	return false
}
