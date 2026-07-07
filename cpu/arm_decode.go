package cpu

func (c *CPU) decodeARM(opcode uint32) {
	// Instruction decoding based on ARM Architecture Reference Manual
	
	// Check for branch
	if (opcode >> 25) & 7 == 0b101 {
		c.handleBranch(opcode)
		return
	}

	// Check for SWI
	if (opcode >> 24) & 0xF == 0xF {
		c.handleSWI(opcode)
		return
	}

	// Check for Coprocessor
	if (opcode >> 24) & 0xE == 0xE {
		if (opcode >> 24) & 0xF == 0xE && (opcode >> 4) & 1 == 1 {
			c.handleCoprocessorRegisterTransfer(opcode)
		} else {
			c.unimplementedARM(opcode)
		}
		return
	}

	// Check for BX / BLX
	if (opcode & 0x0FFFFFF0) == 0x012FFF10 {
		c.handleBX(opcode)
		return
	}

	// Multiplies / DSP / SWP / Extra load/stores
	if (opcode >> 25) & 7 == 0b000 {
		if (opcode & 0x00000090) == 0x00000090 {
			// Multiply and Multiply Long
			op := (opcode >> 21) & 0xF
			switch op {
			case 0x0, 0x1:
				c.handleMultiply(opcode)
				return
			case 0x4, 0x5, 0x6, 0x7:
				c.handleMultiplyLong(opcode)
				return
			}

			// SWP
			if (opcode & 0x0FB00FF0) == 0x01000090 {
				c.handleSWP(opcode)
				return
			}

			// Extra load/stores (LDRH, STRH, LDRSB, LDRSH, LDRD, STRD)
			if (opcode & 0x0E400F90) == 0x00000090 {
				c.handleHalfwordSignedTransfer(opcode)
				return
			}
			
			// DSP Multiplies and Q-math
			if (opcode & 0x0F900090) == 0x01000080 { // SMLAxy / SMLAWy
				if (opcode & (1<<5)) == 0 {
					c.handleSMLAxy(opcode)
				} else {
					c.handleSMLAWy(opcode)
				}
				return
			}
			if (opcode & 0x0F900090) == 0x01400080 { // SMLALxy
				c.handleSMLALxy(opcode)
				return
			}
			if (opcode & 0x0F900090) == 0x01600080 { // SMULxy / SMULWy
				if (opcode & (1<<5)) == 0 {
					c.handleSMULxy(opcode)
				} else {
					c.handleSMULWy(opcode)
				}
				return
			}
			if (opcode & 0x0FF000F0) == 0x01200070 {
				c.handleQADD(opcode) // Note: Need to check opcodes exactly, but let's fall back if needed
				return
			}
			
			c.unimplementedARM(opcode)
			return
		}
		
		// CLZ
		if (opcode & 0x0FFF0FF0) == 0x016F0F10 {
			c.handleCLZ(opcode)
			return
		}

		// QADD, QSUB, QDADD, QDSUB
		if (opcode & 0x0FF00FF0) == 0x01000050 {
			op := (opcode >> 21) & 3
			switch op {
			case 0: c.handleQADD(opcode)
			case 1: c.handleQSUB(opcode)
			case 2: c.handleQDADD(opcode)
			case 3: c.handleQDSUB(opcode)
			}
			return
		}

		// Standard Data Processing
		c.handleDataProcessing(opcode)
		return
	}

	// Data Processing with Immediate
	if (opcode >> 25) & 7 == 0b001 {
		c.handleDataProcessing(opcode)
		return
	}

	// Single Data Transfer (LDR/STR)
	if (opcode >> 26) & 3 == 0b01 {
		c.handleSingleDataTransfer(opcode)
		return
	}

	// Block Data Transfer (LDM/STM)
	if (opcode >> 25) & 7 == 0b100 {
		isLdr := (opcode & (1 << 20)) != 0
		if isLdr {
			c.handleLDMMelonDS(opcode)
		} else {
			c.handleSTMMelonDS(opcode)
		}
		return
	}

	c.unimplementedARM(opcode)
}
