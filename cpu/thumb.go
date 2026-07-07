package cpu

func (c *CPU) populateThumbLUT() {
	for i := 0; i < 1024; i++ {
		op := i >> 4 // top 6 bits of opcode
		switch {
		case op < 0b010010: // 00xxxx, 010000, 010001
			c.THUMBLUT[i] = c.handleThumbALU
		case op >= 0b010010 && op <= 0b110011:
			c.THUMBLUT[i] = c.handleThumbLoadStore
		case op >= 0b110100:
			c.THUMBLUT[i] = c.handleThumbBranches
		default:
			c.THUMBLUT[i] = c.unimplementedTHUMB
		}
	}
}

func (c *CPU) handleThumbALU(opcode uint16) {
	// Step 1: Thumb ALU
	op := opcode >> 11
	switch op {
	case 0b00000, 0b00001, 0b00010: // LSL, LSR, ASR immediate
		shiftType := (opcode >> 11) & 3
		imm5 := uint32((opcode >> 6) & 0x1F)
		rm := opcode & 0x7
		rd := (opcode >> 3) & 0x7

		val := c.R[rm]
		var carryBit uint32 = c.CPSR & 0x20000000

		switch shiftType {
		case 0: // LSL
			if imm5 != 0 {
				carryBit = (val >> (32 - imm5)) & 1
				val <<= imm5
			}
		case 1: // LSR
			if imm5 == 0 {
				imm5 = 32
			}
			carryBit = (val >> (imm5 - 1)) & 1
			if imm5 == 32 {
				val = 0
			} else {
				val >>= imm5
			}
		case 2: // ASR
			if imm5 == 0 {
				imm5 = 32
			}
			carryBit = (val >> (imm5 - 1)) & 1
			if (val & 0x80000000) != 0 {
				if imm5 == 32 {
					val = 0xFFFFFFFF
				} else {
					val = uint32(int32(val) >> imm5)
				}
			} else {
				if imm5 == 32 {
					val = 0
				} else {
					val >>= imm5
				}
			}
		}
		c.R[rd] = val
		c.updateThumbFlags(val, true, carryBit != 0, false, false)

	case 0b00011: // Add/Subtract register/immediate
		isSub := (opcode & (1 << 9)) != 0
		isImm := (opcode & (1 << 10)) != 0
		rmOrImm := uint32((opcode >> 6) & 0x7)
		rn := (opcode >> 3) & 0x7
		rd := opcode & 0x7

		op1 := c.R[rn]
		op2 := rmOrImm
		if !isImm {
			op2 = c.R[rmOrImm]
		}

		if isSub {
			c.R[rd] = c.doThumbSub(op1, op2, true)
		} else {
			c.R[rd] = c.doThumbAdd(op1, op2, true)
		}

	case 0b00100, 0b00101, 0b00110, 0b00111: // MOV, CMP, ADD, SUB immediate
		opc := (opcode >> 11) & 3
		rd := (opcode >> 8) & 0x7
		imm8 := uint32(opcode & 0xFF)

		switch opc {
		case 0: // MOV
			c.R[rd] = imm8
			c.updateThumbFlags(imm8, true, false, false, false)
		case 1: // CMP
			c.doThumbSub(c.R[rd], imm8, true)
		case 2: // ADD
			c.R[rd] = c.doThumbAdd(c.R[rd], imm8, true)
		case 3: // SUB
			c.R[rd] = c.doThumbSub(c.R[rd], imm8, true)
		}

	case 0b01000:
		if (opcode & 0xFC00) == 0x4000 {
			// Format 4: ALU operations
			aluOp := (opcode >> 6) & 0xF
			rs := (opcode >> 3) & 0x7
			rd := opcode & 0x7

			op1 := c.R[rd]
			op2 := c.R[rs]

			switch aluOp {
			case 0x0: // AND
				res := op1 & op2
				c.R[rd] = res
				c.updateThumbFlags(res, true, false, false, false)
			case 0x1: // EOR
				res := op1 ^ op2
				c.R[rd] = res
				c.updateThumbFlags(res, true, false, false, false)
			case 0x2: // LSL
				shift := op2 & 0xFF
				res := op1
				carry := (c.CPSR & 0x20000000) != 0
				if shift > 0 {
					if shift < 32 {
						carry = ((res >> (32 - shift)) & 1) != 0
						res <<= shift
					} else if shift == 32 {
						carry = (res & 1) != 0
						res = 0
					} else {
						carry = false
						res = 0
					}
				}
				c.R[rd] = res
				c.updateThumbFlags(res, true, carry, true, false)
			case 0x3: // LSR
				shift := op2 & 0xFF
				res := op1
				carry := (c.CPSR & 0x20000000) != 0
				if shift > 0 {
					if shift < 32 {
						carry = ((res >> (shift - 1)) & 1) != 0
						res >>= shift
					} else if shift == 32 {
						carry = ((res >> 31) & 1) != 0
						res = 0
					} else {
						carry = false
						res = 0
					}
				}
				c.R[rd] = res
				c.updateThumbFlags(res, true, carry, true, false)
			case 0x4: // ASR
				shift := op2 & 0xFF
				res := op1
				carry := (c.CPSR & 0x20000000) != 0
				if shift > 0 {
					if shift < 32 {
						carry = ((res >> (shift - 1)) & 1) != 0
						res = uint32(int32(res) >> shift)
					} else {
						carry = ((res >> 31) & 1) != 0
						if (res & 0x80000000) != 0 {
							res = 0xFFFFFFFF
						} else {
							res = 0
						}
					}
				}
				c.R[rd] = res
				c.updateThumbFlags(res, true, carry, true, false)
			case 0x5: // ADC
				carryIn := uint32(0)
				if (c.CPSR & 0x20000000) != 0 {
					carryIn = 1
				}
				res64 := uint64(op1) + uint64(op2) + uint64(carryIn)
				res := uint32(res64)
				carry := res64 > 0xFFFFFFFF
				overflow := ((op1^op2^0xFFFFFFFF)&(op1^res))&0x80000000 != 0
				c.R[rd] = res
				c.updateThumbFlags(res, true, carry, true, overflow)
			case 0x6: // SBC
				carryIn := uint32(0)
				if (c.CPSR & 0x20000000) != 0 {
					carryIn = 1
				}
				res64 := uint64(op1) - uint64(op2) - uint64(1-carryIn)
				res := uint32(res64)
				carry := res64 <= 0xFFFFFFFF
				overflow := ((op1^op2)&(op1^res))&0x80000000 != 0
				c.R[rd] = res
				c.updateThumbFlags(res, true, carry, true, overflow)
			case 0x7: // ROR
				shift := op2 & 0xFF
				res := op1
				carry := (c.CPSR & 0x20000000) != 0
				if shift > 0 {
					shift &= 0x1F
					if shift == 0 {
						carry = ((res >> 31) & 1) != 0
					} else {
						carry = ((res >> (shift - 1)) & 1) != 0
						res = (res >> shift) | (res << (32 - shift))
					}
				}
				c.R[rd] = res
				c.updateThumbFlags(res, true, carry, true, false)
			case 0x8: // TST
				c.updateThumbFlags(op1&op2, true, false, false, false)
			case 0x9: // NEG
				c.R[rd] = c.doThumbSub(0, op2, true)
			case 0xA: // CMP
				c.doThumbSub(op1, op2, true)
			case 0xB: // CMN
				c.doThumbAdd(op1, op2, true)
			case 0xC: // ORR
				res := op1 | op2
				c.R[rd] = res
				c.updateThumbFlags(res, true, false, false, false)
			case 0xD: // MUL
				res := op1 * op2
				c.R[rd] = res
				c.updateThumbFlags(res, true, false, false, false)
			case 0xE: // BIC
				res := op1 & (^op2)
				c.R[rd] = res
				c.updateThumbFlags(res, true, false, false, false)
			case 0xF: // MVN
				res := ^op2
				c.R[rd] = res
				c.updateThumbFlags(res, true, false, false, false)
			}
		} else if (opcode & 0xFC00) == 0x4400 {
			// Format 5: Hi register operations/branch exchange
			opc := (opcode >> 8) & 3
			h1 := (opcode & 0x80) != 0
			h2 := (opcode & 0x40) != 0
			rm := (opcode >> 3) & 0x7
			rd := opcode & 0x7
			if h1 {
				rd += 8
			}
			if h2 {
				rm += 8
			}

			valRm := c.R[rm]
			if rm == 15 {
				valRm += 2 // PC is +4 in thumb
			}

			switch opc {
			case 0: // ADD
				c.R[rd] += valRm
				if rd == 15 {
					c.R[15] &= 0xFFFFFFFE
				}
			case 1: // CMP
				valRd := c.R[rd]
				if rd == 15 {
					valRd += 2
				}
				c.doThumbSub(valRd, valRm, true)
			case 2: // MOV
				c.R[rd] = valRm
				if rd == 15 {
					c.R[15] &= 0xFFFFFFFE
				}
			case 3: // BX / BLX
				addr := valRm
				if h1 { // (opcode & 0x80) != 0 is BLX
					c.R[14] = (c.R[15] - 2) | 1
				}
				if (addr & 1) != 0 {
					c.CPSR |= 0x20 // Thumb
					c.R[15] = addr & 0xFFFFFFFE
				} else {
					c.CPSR &^= 0x20 // ARM
					c.R[15] = addr & 0xFFFFFFFC
				}
			}
		}
	}
}

func (c *CPU) handleThumbLoadStore(opcode uint16) {
	// Step 2: Thumb Load/Store
	op := opcode >> 12
	switch op {
	case 0b0100: // LDR PC-rel (Format 6)
		rd := (opcode >> 8) & 0x7
		imm8 := uint32(opcode & 0xFF)
		addr := (c.R[15] & 0xFFFFFFFC) + (imm8 << 2)
		c.R[rd] = c.Bus.Read32(addr)

	case 0b0101: // Load/Store with register offset (Format 7 & 8)
		opc := (opcode >> 9) & 0x7
		rm := (opcode >> 6) & 0x7
		rn := (opcode >> 3) & 0x7
		rd := opcode & 0x7
		addr := c.R[rn] + c.R[rm]

		switch opc {
		case 0: // STR
			c.Bus.Write32(addr&0xFFFFFFFC, c.R[rd])
		case 1: // STRH
			c.Bus.Write16(addr&0xFFFFFFFE, uint16(c.R[rd]))
		case 2: // STRB
			c.Bus.Write8(addr, byte(c.R[rd]))
		case 3: // LDRSB
			val := int8(c.Bus.Read8(addr))
			c.R[rd] = uint32(int32(val))
		case 4: // LDR
			val := c.Bus.Read32(addr & 0xFFFFFFFC)
			if (addr & 3) != 0 {
				shift := (addr & 3) * 8
				val = (val >> shift) | (val << (32 - shift))
			}
			c.R[rd] = val
		case 5: // LDRH
			val := c.Bus.Read16(addr & 0xFFFFFFFE)
			c.R[rd] = uint32(val)
		case 6: // LDRB
			c.R[rd] = uint32(c.Bus.Read8(addr))
		case 7: // LDRSH
			val := int16(c.Bus.Read16(addr & 0xFFFFFFFE))
			c.R[rd] = uint32(int32(val))
		}

	case 0b0110, 0b0111: // Load/Store word/byte with immediate offset (Format 9)
		isLdr := (opcode & (1 << 11)) != 0
		isByte := (opcode & (1 << 12)) != 0
		imm5 := uint32((opcode >> 6) & 0x1F)
		rn := (opcode >> 3) & 0x7
		rd := opcode & 0x7

		addr := c.R[rn]
		if isByte {
			addr += imm5
			if isLdr {
				c.R[rd] = uint32(c.Bus.Read8(addr))
			} else {
				c.Bus.Write8(addr, byte(c.R[rd]))
			}
		} else {
			addr += imm5 << 2
			if isLdr {
				val := c.Bus.Read32(addr & 0xFFFFFFFC)
				if (addr & 3) != 0 {
					shift := (addr & 3) * 8
					val = (val >> shift) | (val << (32 - shift))
				}
				c.R[rd] = val
			} else {
				c.Bus.Write32(addr&0xFFFFFFFC, c.R[rd])
			}
		}

	case 0b1000: // Load/Store halfword immediate offset (Format 10)
		isLdr := (opcode & (1 << 11)) != 0
		imm5 := uint32((opcode >> 6) & 0x1F)
		rn := (opcode >> 3) & 0x7
		rd := opcode & 0x7
		addr := c.R[rn] + (imm5 << 1)

		if isLdr {
			val := c.Bus.Read16(addr & 0xFFFFFFFE)
			c.R[rd] = uint32(val)
		} else {
			c.Bus.Write16(addr&0xFFFFFFFE, uint16(c.R[rd]))
		}

	case 0b1001: // SP-relative Load/Store (Format 11)
		isLdr := (opcode & (1 << 11)) != 0
		rd := (opcode >> 8) & 0x7
		imm8 := uint32(opcode & 0xFF)
		addr := c.R[13] + (imm8 << 2)

		if isLdr {
			c.R[rd] = c.Bus.Read32(addr & 0xFFFFFFFC)
		} else {
			c.Bus.Write32(addr&0xFFFFFFFC, c.R[rd])
		}

	case 0b1010: // Load Address (Format 12)
		isSP := (opcode & (1 << 11)) != 0
		rd := (opcode >> 8) & 0x7
		imm8 := uint32(opcode & 0xFF)

		if isSP {
			c.R[rd] = c.R[13] + (imm8 << 2)
		} else {
			c.R[rd] = (c.R[15] & 0xFFFFFFFC) + (imm8 << 2)
		}

	case 0b1011:
		if (opcode & 0x0F00) == 0x0000 { // ADD SP, imm (Format 13)
			isSub := (opcode & 0x80) != 0
			imm7 := uint32(opcode & 0x7F)
			if isSub {
				c.R[13] -= (imm7 << 2)
			} else {
				c.R[13] += (imm7 << 2)
			}
		} else if (opcode & 0x0600) == 0x0400 { // PUSH/POP (Format 14)
			isPop := (opcode & 0x800) != 0
			isRBit := (opcode & 0x100) != 0
			regList := opcode & 0xFF

			if isPop {
				addr := c.R[13]
				for i := 0; i < 8; i++ {
					if (regList & (1 << i)) != 0 {
						c.R[i] = c.Bus.Read32(addr)
						addr += 4
					}
				}
				if isRBit {
					c.R[15] = c.Bus.Read32(addr)
					addr += 4
					c.R[15] &= 0xFFFFFFFE
				}
				c.R[13] = addr
			} else {
				numRegs := 0
				for i := 0; i < 8; i++ {
					if (regList & (1 << i)) != 0 {
						numRegs++
					}
				}
				if isRBit {
					numRegs++
				}
				addr := c.R[13] - uint32(numRegs*4)
				c.R[13] = addr
				for i := 0; i < 8; i++ {
					if (regList & (1 << i)) != 0 {
						c.Bus.Write32(addr, c.R[i])
						addr += 4
					}
				}
				if isRBit {
					c.Bus.Write32(addr, c.R[14])
				}
			}
		}

	case 0b1100: // LDM/STM (Format 15)
		isLdr := (opcode & (1 << 11)) != 0
		rn := (opcode >> 8) & 0x7
		regList := opcode & 0xFF
		addr := c.R[rn]

		numRegs := 0
		for i := 0; i < 8; i++ {
			if (regList & (1 << i)) != 0 {
				numRegs++
			}
		}

		c.R[rn] = addr + uint32(numRegs*4)
		for i := 0; i < 8; i++ {
			if (regList & (1 << i)) != 0 {
				if isLdr {
					c.R[i] = c.Bus.Read32(addr)
				} else {
					c.Bus.Write32(addr, c.R[i])
				}
				addr += 4
			}
		}
	}
}

func (c *CPU) handleThumbBranches(opcode uint16) {
	// Step 3: Thumb Branches
	op5 := opcode >> 11 // Top 5 bits
	switch op5 {
	case 0b11010, 0b11011:
		if (opcode >> 8) == 0xDF { // SWI (Format 17)
			if c.SWICallback != nil {
				c.SWICallback(uint32(opcode & 0xFF))
			}
		} else { // Bcc (Format 16)
			cond := uint32((opcode >> 8) & 0xF)
			if cond != 0xE && cond != 0xF {
				if c.checkCondition(cond) {
					offset := uint32(opcode & 0xFF)
					if (offset & 0x80) != 0 {
						offset |= 0xFFFFFF00
					}
					c.R[15] += offset << 1
				}
			}
		}
	case 0b11100, 0b11101: // B unconditional (Format 18)
		offset := uint32(opcode & 0x7FF)
		if (offset & 0x400) != 0 {
			offset |= 0xFFFFF800
		}
		c.R[15] += offset << 1
	case 0b11110: // BL/BLX prefix
		offset := uint32(opcode & 0x7FF)
		if (offset & 0x400) != 0 {
			offset |= 0xFFFFF800
		}
		c.R[14] = c.R[15] + (offset << 12)
	case 0b11111: // BL/BLX suffix
		offset := uint32(opcode&0x7FF) << 1
		pc := c.R[14] + offset

		if c.IsARM9 && (opcode&(1<<12)) == 0 { // BLX suffix
			c.R[14] = (c.R[15] - 2) | 1
			c.CPSR &^= 0x20 // Switch to ARM
			c.R[15] = pc & 0xFFFFFFFC
		} else { // BL suffix
			c.R[14] = (c.R[15] - 2) | 1
			c.R[15] = pc | 1 // maintain Thumb mode
		}
	}
}

// Helpers for ALU
func (c *CPU) doThumbAdd(op1, op2 uint32, updateFlags bool) uint32 {
	res64 := uint64(op1) + uint64(op2)
	res := uint32(res64)
	if updateFlags {
		carry := (res64 > 0xFFFFFFFF)
		overflow := ((op1^op2^0xFFFFFFFF)&(op1^res))&0x80000000 != 0
		c.updateThumbFlags(res, true, carry, true, overflow)
	}
	return res
}

func (c *CPU) doThumbSub(op1, op2 uint32, updateFlags bool) uint32 {
	res64 := uint64(op1) - uint64(op2)
	res := uint32(res64)
	if updateFlags {
		carry := (op1 >= op2)
		overflow := ((op1^op2)&(op1^res))&0x80000000 != 0
		c.updateThumbFlags(res, true, carry, true, overflow)
	}
	return res
}

func (c *CPU) updateThumbFlags(res uint32, setNZ bool, carry bool, setC bool, overflow bool) {
	if setNZ {
		if res == 0 {
			c.CPSR |= 0x40000000
		} else {
			c.CPSR &^= 0x40000000
		}
		if (res & 0x80000000) != 0 {
			c.CPSR |= 0x80000000
		} else {
			c.CPSR &^= 0x80000000
		}
	}
	if setC {
		if carry {
			c.CPSR |= 0x20000000
		} else {
			c.CPSR &^= 0x20000000
		}
		if overflow {
			c.CPSR |= 0x10000000
		} else {
			c.CPSR &^= 0x10000000
		}
	}
}
