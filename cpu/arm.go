package cpu

import (
	"fmt"
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

	CP15 *CP15
	SWICallback func(uint32)
}

func NewCPU(isARM9 bool, bus MemoryInterface) *CPU {
	c := &CPU{
		IsARM9: isARM9,
		Bus:    bus,
	}
	if isARM9 {
		c.CP15 = &CP15{}
	}
	c.initLUTs()
	return c
}

func (c *CPU) initLUTs() {
	c.populateThumbLUT()
}

func (c *CPU) handleBranch(opcode uint32) {
	link := (opcode & (1 << 24)) != 0
	offset := opcode & 0x00FFFFFF
	if (offset & 0x00800000) != 0 {
		offset |= 0xFF000000
	}
	if link {
		c.R[14] = c.R[15]
	}
	c.R[15] = c.R[15] + (offset << 2)
}

func (c *CPU) handleDataProcessing(opcode uint32) {
	isImmediate := (opcode & (1 << 25)) != 0
	op := (opcode >> 21) & 0xF
	setFlags := (opcode & (1 << 20)) != 0
	rn := (opcode >> 16) & 0xF
	rd := (opcode >> 12) & 0xF

	var op2 uint32
	if isImmediate {
		imm := opcode & 0xFF
		rotate := (opcode >> 8) & 0xF
		shift := rotate * 2
		op2 = (imm >> shift) | (imm << (32 - shift))
	} else {
		rm := opcode & 0xF
		op2 = c.R[rm]
	}

	op1 := c.R[rn]
	var res uint32

	switch op {
	case 0x0:
		res = op1 & op2 // AND
	case 0x1:
		res = op1 ^ op2 // EOR
	case 0x2:
		res = op1 - op2 // SUB
	case 0x3:
		res = op2 - op1 // RSB
	case 0x4:
		res = op1 + op2 // ADD
	case 0x8:
		res = op1 & op2 // TST
	case 0x9:
		res = op1 ^ op2 // TEQ
	case 0xA:
		res = op1 - op2 // CMP
	case 0xB:
		res = op1 + op2 // CMN
	case 0xC:
		res = op1 | op2 // ORR
	case 0xD:
		res = op2 // MOV
	case 0xE:
		res = op1 & (^op2) // BIC
	case 0xF:
		res = ^op2 // MVN
	default:
		res = 0
	}

	if op != 0x8 && op != 0x9 && op != 0xA && op != 0xB {
		c.R[rd] = res
	}

	if setFlags {
		if res == 0 {
			c.CPSR |= 0x40000000 // Z
		} else {
			c.CPSR &^= 0x40000000
		}
		if (res & 0x80000000) != 0 {
			c.CPSR |= 0x80000000 // N
		} else {
			c.CPSR &^= 0x80000000
		}
	}
}

func (c *CPU) unimplementedARM(opcode uint32) {
	panic(fmt.Sprintf("Unimplemented ARM opcode: %08X at PC: %08X", opcode, c.R[15]-8))
}

func (c *CPU) unimplementedTHUMB(opcode uint16) {
	panic(fmt.Sprintf("Unimplemented THUMB opcode: %04X at PC: %08X", opcode, c.R[15]-4))
}

func (c *CPU) handleThumbBranch(opcode uint16) {
	// Format 18: B offset
	offset := uint32(opcode & 0x7FF)
	if (offset & 0x400) != 0 {
		offset |= 0xFFFFF800 // sign extend
	}
	offset <<= 1
	c.R[15] = c.R[15] + offset
}

func (c *CPU) handleSingleDataTransfer(opcode uint32) {
	iBit := (opcode & (1 << 25)) != 0
	pBit := (opcode & (1 << 24)) != 0
	uBit := (opcode & (1 << 23)) != 0
	bBit := (opcode & (1 << 22)) != 0
	wBit := (opcode & (1 << 21)) != 0
	lBit := (opcode & (1 << 20)) != 0

	rn := (opcode >> 16) & 0xF
	rd := (opcode >> 12) & 0xF

	var offset uint32
	if !iBit {
		offset = opcode & 0xFFF
	} else {
		rm := opcode & 0xF
		offset = c.R[rm]
	}

	address := c.R[rn]
	if rn == 15 {
		address += 8
	}

	var effAddr uint32
	if pBit {
		if uBit {
			effAddr = address + offset
		} else {
			effAddr = address - offset
		}
	} else {
		effAddr = address
	}

	if lBit { // Load
		if bBit {
			c.R[rd] = uint32(c.Bus.Read8(effAddr))
		} else {
			c.R[rd] = c.Bus.Read32(effAddr & 0xFFFFFFFC)
		}
	} else { // Store
		val := c.R[rd]
		if rd == 15 {
			val += 12
		}
		if bBit {
			c.Bus.Write8(effAddr, byte(val))
		} else {
			c.Bus.Write32(effAddr&0xFFFFFFFC, val)
		}
	}

	if (!pBit) || wBit {
		if uBit {
			c.R[rn] = address + offset
		} else {
			c.R[rn] = address - offset
		}
	}
}

func (c *CPU) handleBlockDataTransfer(opcode uint32) {
	pBit := (opcode & (1 << 24)) != 0
	uBit := (opcode & (1 << 23)) != 0
	wBit := (opcode & (1 << 21)) != 0
	lBit := (opcode & (1 << 20)) != 0

	rn := (opcode >> 16) & 0xF
	regList := opcode & 0xFFFF

	address := c.R[rn]

	numBits := 0
	for i := 0; i < 16; i++ {
		if (regList & (1 << i)) != 0 {
			numBits++
		}
	}

	var startAddr uint32
	if uBit {
		startAddr = address
		if pBit {
			startAddr += 4
		}
	} else {
		startAddr = address - uint32(numBits*4)
		if !pBit {
			startAddr += 4
		}
	}

	addr := startAddr
	for i := 0; i < 16; i++ {
		if (regList & (1 << i)) != 0 {
			if lBit { // LDM
				c.R[i] = c.Bus.Read32(addr & 0xFFFFFFFC)
			} else { // STM
				val := c.R[i]
				if i == 15 {
					val += 12
				}
				c.Bus.Write32(addr&0xFFFFFFFC, val)
			}
			addr += 4
		}
	}

	if wBit {
		if uBit {
			c.R[rn] = address + uint32(numBits*4)
		} else {
			c.R[rn] = address - uint32(numBits*4)
		}
	}
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

		c.decodeARM(opcode)
	}
}

func (c *CPU) checkCondition(cond uint32) bool {
	// N, Z, C, V flags in CPSR
	n := (c.CPSR & 0x80000000) != 0
	z := (c.CPSR & 0x40000000) != 0
	c_flag := (c.CPSR & 0x20000000) != 0
	v := (c.CPSR & 0x10000000) != 0

	switch cond {
	case 0x0:
		return z
	case 0x1:
		return !z
	case 0x2:
		return c_flag
	case 0x3:
		return !c_flag
	case 0x4:
		return n
	case 0x5:
		return !n
	case 0x6:
		return v
	case 0x7:
		return !v
	case 0x8:
		return c_flag && !z
	case 0x9:
		return !c_flag || z
	case 0xA:
		return n == v
	case 0xB:
		return n != v
	case 0xC:
		return !z && (n == v)
	case 0xD:
		return z || (n != v)
	case 0xE:
		return true
	case 0xF:
		return false // Actually depends on ARM version, mostly false for execution
	}
	return false
}

func (c *CPU) handleCoprocessorRegisterTransfer(opcode uint32) {
	if c.CP15 == nil {
		return // Ignore on ARM7
	}
	isMRC := (opcode & (1 << 20)) != 0
	rd := (opcode >> 12) & 0xF
	crn := (opcode >> 16) & 0xF
	crm := opcode & 0xF
	opc2 := (opcode >> 5) & 0x7

	if !isMRC {
		// MCR
		val := c.R[rd]
		if crn == 1 && crm == 0 && opc2 == 0 {
			c.CP15.ControlRegister = val
		} else if crn == 9 && crm == 1 && opc2 == 0 {
			c.CP15.DTCMRegion = val
		} else if crn == 9 && crm == 1 && opc2 == 1 {
			c.CP15.ITCMRegion = val
		}
	} else {
		// MRC (stubbed for now)
		c.R[rd] = 0
	}
}

func (c *CPU) handleSWI(opcode uint32) {
	swiNum := opcode & 0x00FFFFFF
	if c.SWICallback != nil {
		c.SWICallback(swiNum)
	}
}

func (c *CPU) handleBX(opcode uint32) {
	rn := opcode & 0xF
	addr := c.R[rn]
	
	if (addr & 1) != 0 {
		c.CPSR |= 0x20 // Switch to Thumb
		c.R[15] = addr & 0xFFFFFFFE
	} else {
		c.CPSR &^= 0x20 // Switch to ARM
		c.R[15] = addr & 0xFFFFFFFC
	}
}


