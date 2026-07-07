package cpu

// Handle Halfword and Signed Byte Data Transfers (LDRH, STRH, LDRSB, LDRSH)
func (c *CPU) handleHalfwordSignedTransfer(opcode uint32) {
	pBit := (opcode & (1 << 24)) != 0
	uBit := (opcode & (1 << 23)) != 0
	wBit := (opcode & (1 << 21)) != 0
	lBit := (opcode & (1 << 20)) != 0

	rn := (opcode >> 16) & 0xF
	rd := (opcode >> 12) & 0xF

	isImm := (opcode & (1 << 22)) != 0
	var offset uint32
	if isImm {
		offset = (opcode & 0xF) | ((opcode >> 4) & 0xF0)
	} else {
		rm := opcode & 0xF
		offset = c.R[rm]
	}

	address := c.R[rn]
	if rn == 15 {
		address += 4 // PC+8
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

	op := (opcode >> 5) & 0x3

	switch op {
	case 1: // STRH / LDRH
		if lBit { // LDRH
			val := c.Bus.Read16(effAddr)
			c.R[rd] = uint32(val)
		} else { // STRH
			val := c.R[rd]
			if rd == 15 {
				val += 8 // PC+12 for stores usually
			}
			c.Bus.Write16(effAddr, uint16(val))
		}
	case 2: // LDRD / LDRSB
		if lBit { // LDRSB
			val := c.Bus.Read8(effAddr)
			c.R[rd] = uint32(int32(int8(val)))
		} else { // LDRD
			if !c.IsARM9 {
				return
			}
			if (rd & 1) != 0 {
				rd--
			}
			c.R[rd] = c.Bus.Read32(effAddr)
			c.R[rd+1] = c.Bus.Read32(effAddr + 4)
		}
	case 3: // STRD / LDRSH
		if lBit { // LDRSH
			val := c.Bus.Read16(effAddr)
			c.R[rd] = uint32(int32(int16(val)))
		} else { // STRD
			if !c.IsARM9 {
				return
			}
			if (rd & 1) != 0 {
				rd--
			}
			c.Bus.Write32(effAddr, c.R[rd])
			c.Bus.Write32(effAddr+4, c.R[rd+1])
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

// Handle SWP and SWPB
func (c *CPU) handleSWP(opcode uint32) {
	base := c.R[(opcode>>16)&0xF]
	rm := c.R[opcode&0xF]
	rd := (opcode >> 12) & 0xF

	isByte := (opcode & (1 << 22)) != 0
	if isByte { // SWPB
		val := c.Bus.Read8(base)
		c.Bus.Write8(base, byte(rm))
		c.R[rd] = uint32(val)
	} else { // SWP
		val := c.Bus.Read32(base)
		c.Bus.Write32(base, rm)
		shift := 8 * (base & 3)
		c.R[rd] = (val >> shift) | (val << (32 - shift))
	}
}

// LDM with melonDS edge cases
func (c *CPU) handleLDMMelonDS(opcode uint32) {
	baseid := (opcode >> 16) & 0xF
	base := c.R[baseid]
	var wbbase uint32
	preinc := (opcode & (1 << 24)) != 0

	if (opcode & (1 << 23)) == 0 {
		var numBits uint32
		for i := 0; i < 16; i++ {
			if (opcode & (1 << i)) != 0 {
				numBits++
			}
		}
		base -= 4 * numBits

		if (opcode & (1 << 21)) != 0 {
			wbbase = base
		}
		preinc = !preinc
	}

	isUserMode := false
	if (opcode&(1<<22)) != 0 && (opcode&(1<<15)) == 0 {
		isUserMode = true
	}

	for i := 0; i < 15; i++ {
		if (opcode & (1 << i)) != 0 {
			if preinc {
				base += 4
			}
			val := c.Bus.Read32(base &^ 3)
			if isUserMode {
				c.setRegUser(i, val)
			} else {
				c.R[i] = val
			}
			if !preinc {
				base += 4
			}
		}
	}

	var pc uint32
	if (opcode & (1 << 15)) != 0 {
		if preinc {
			base += 4
		}
		pc = c.Bus.Read32(base &^ 3)
		if !preinc {
			base += 4
		}
		if !c.IsARM9 {
			pc &^= 1
		}
	}

	if (opcode & (1 << 21)) != 0 {
		if (opcode & (1 << 23)) != 0 {
			wbbase = base
		}
		if (opcode & (1 << baseid)) != 0 {
			if !c.IsARM9 {
				rlist := opcode & 0xFFFF
				if (rlist &^ (1 << baseid)) == 0 || (rlist &^ ((2 << baseid) - 1)) != 0 {
					c.R[baseid] = wbbase
				}
			}
		} else {
			c.R[baseid] = wbbase
		}
	}

	if (opcode & (1 << 15)) != 0 {
		if (opcode & (1 << 22)) != 0 {
			mode := c.CPSR & 0x1F
			if mode == ModeFIQ {
				c.CPSR = c.SPSR_fiq
			} else if mode == ModeSVC {
				c.CPSR = c.SPSR_svc
			} else if mode == ModeAbt {
				c.CPSR = c.SPSR_abt
			} else if mode == ModeIRQ {
				c.CPSR = c.SPSR_irq
			} else if mode == ModeUnd {
				c.CPSR = c.SPSR_und
			}
		}
		c.R[15] = pc
	}
}

// STM with melonDS edge cases
func (c *CPU) handleSTMMelonDS(opcode uint32) {
	baseid := (opcode >> 16) & 0xF
	base := c.R[baseid]
	oldbase := base
	preinc := (opcode & (1 << 24)) != 0

	if (opcode & (1 << 23)) == 0 {
		var numBits uint32
		for i := 0; i < 16; i++ {
			if (opcode & (1 << i)) != 0 {
				numBits++
			}
		}
		base -= 4 * numBits

		if (opcode & (1 << 21)) != 0 {
			c.R[baseid] = base
		}
		preinc = !preinc
	}

	isBanked := false
	if (opcode & (1 << 22)) != 0 {
		mode := c.CPSR & 0x1F
		if mode == 0x11 { // ModeFIQ
			isBanked = (baseid >= 8 && baseid < 15)
		} else if mode != 0x10 && mode != 0x1F { // Not User or System
			isBanked = (baseid >= 13 && baseid < 15)
		}
	}

	for i := uint32(0); i < 16; i++ {
		if (opcode & (1 << i)) != 0 {
			if preinc {
				base += 4
			}

			if i == baseid && !isBanked {
				if !c.IsARM9 || (opcode & ((1 << i) - 1)) == 0 {
					c.Bus.Write32(base&^3, oldbase)
				} else {
					c.Bus.Write32(base&^3, base)
				}
			} else {
				val := uint32(0)
				if (opcode & (1 << 22)) != 0 {
					val = c.getRegUser(int(i))
				} else {
					val = c.R[i]
					if i == 15 {
						val += 8 // PC+12 for STM
					}
				}
				c.Bus.Write32(base&^3, val)
			}

			if !preinc {
				base += 4
			}
		}
	}

	if (opcode & (1 << 23)) != 0 && (opcode & (1 << 21)) != 0 {
		c.R[baseid] = base
	}
}

func (c *CPU) getRegUser(i int) uint32 {
	if i < 8 || i == 15 {
		return c.R[i]
	}
	if i >= 8 && i <= 12 {
		mode := c.CPSR & 0x1F
		if mode == ModeFIQ {
			switch i {
			case 8:
				return c.R8_usr
			case 9:
				return c.R9_usr
			case 10:
				return c.R10_usr
			case 11:
				return c.R11_usr
			case 12:
				return c.R12_usr
			}
		}
		return c.R[i]
	}
	if i == 13 {
		return c.R13_usr
	}
	if i == 14 {
		return c.R14_usr
	}
	return c.R[i]
}

func (c *CPU) setRegUser(i int, val uint32) {
	if i < 8 || i == 15 {
		c.R[i] = val
		return
	}
	if i >= 8 && i <= 12 {
		mode := c.CPSR & 0x1F
		if mode == ModeFIQ {
			switch i {
			case 8:
				c.R8_usr = val
			case 9:
				c.R9_usr = val
			case 10:
				c.R10_usr = val
			case 11:
				c.R11_usr = val
			case 12:
				c.R12_usr = val
			}
			return
		}
		c.R[i] = val
		return
	}
	if i == 13 {
		c.R13_usr = val
		return
	}
	if i == 14 {
		c.R14_usr = val
		return
	}
}
