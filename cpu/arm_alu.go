package cpu

func checkOverflowAdd(a, b uint32) bool {
	res := a + b
	return ((a ^ b) & 0x80000000) == 0 && ((a ^ res) & 0x80000000) != 0
}

func checkOverflowSub(a, b uint32) bool {
	res := a - b
	return ((a ^ b) & 0x80000000) != 0 && ((a ^ res) & 0x80000000) != 0
}

func (c *CPU) handleMultiply(opcode uint32) {
	rm := c.R[opcode&0xF]
	rs := c.R[(opcode>>8)&0xF]
	setFlags := (opcode & (1 << 20)) != 0
	isMLA := (opcode & (1 << 21)) != 0

	var res uint32
	if isMLA {
		rn := c.R[(opcode>>12)&0xF]
		res = (rm * rs) + rn
	} else {
		res = rm * rs
	}

	rd := (opcode >> 16) & 0xF
	c.R[rd] = res

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
		if !c.IsARM9 {
			c.CPSR &^= 0x20000000 // Clear C on ARM7
		}
	}
}

func (c *CPU) handleMultiplyLong(opcode uint32) {
	rm := c.R[opcode&0xF]
	rs := c.R[(opcode>>8)&0xF]
	setFlags := (opcode & (1 << 20)) != 0
	isSigned := (opcode & (1 << 22)) != 0
	isAccumulate := (opcode & (1 << 21)) != 0

	var res uint64
	if isSigned {
		res = uint64(int64(int32(rm)) * int64(int32(rs)))
	} else {
		res = uint64(rm) * uint64(rs)
	}

	rdLo := (opcode >> 12) & 0xF
	rdHi := (opcode >> 16) & 0xF

	if isAccumulate {
		accum := uint64(c.R[rdLo]) | (uint64(c.R[rdHi]) << 32)
		res += accum
	}

	c.R[rdLo] = uint32(res)
	c.R[rdHi] = uint32(res >> 32)

	if setFlags {
		if res == 0 {
			c.CPSR |= 0x40000000 // Z
		} else {
			c.CPSR &^= 0x40000000
		}
		if (res & (1 << 63)) != 0 {
			c.CPSR |= 0x80000000 // N
		} else {
			c.CPSR &^= 0x80000000
		}
		if !c.IsARM9 {
			c.CPSR &^= 0x20000000 // Clear C on ARM7
		}
	}
}

func (c *CPU) handleSMLAxy(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rs := c.R[(opcode>>8)&0xF]
	rn := c.R[(opcode>>12)&0xF]

	if (opcode & (1 << 5)) != 0 {
		rm >>= 16
	} else {
		rm &= 0xFFFF
	}
	if (opcode & (1 << 6)) != 0 {
		rs >>= 16
	} else {
		rs &= 0xFFFF
	}

	resMul := uint32(int32(int16(rm)) * int32(int16(rs)))
	res := resMul + rn

	c.R[(opcode>>16)&0xF] = res
	if checkOverflowAdd(resMul, rn) {
		c.CPSR |= 0x08000000 // Q flag
	}
}

func (c *CPU) handleSMLAWy(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rs := c.R[(opcode>>8)&0xF]
	rn := c.R[(opcode>>12)&0xF]

	if (opcode & (1 << 6)) != 0 {
		rs >>= 16
	} else {
		rs &= 0xFFFF
	}

	resMul := uint32((int64(int32(rm)) * int64(int16(rs))) >> 16)
	res := resMul + rn

	c.R[(opcode>>16)&0xF] = res
	if checkOverflowAdd(resMul, rn) {
		c.CPSR |= 0x08000000 // Q flag
	}
}

func (c *CPU) handleSMULxy(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rs := c.R[(opcode>>8)&0xF]

	if (opcode & (1 << 5)) != 0 {
		rm >>= 16
	} else {
		rm &= 0xFFFF
	}
	if (opcode & (1 << 6)) != 0 {
		rs >>= 16
	} else {
		rs &= 0xFFFF
	}

	res := uint32(int32(int16(rm)) * int32(int16(rs)))
	c.R[(opcode>>16)&0xF] = res
}

func (c *CPU) handleSMULWy(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rs := c.R[(opcode>>8)&0xF]

	if (opcode & (1 << 6)) != 0 {
		rs >>= 16
	} else {
		rs &= 0xFFFF
	}

	res := uint32((int64(int32(rm)) * int64(int16(rs))) >> 16)
	c.R[(opcode>>16)&0xF] = res
}

func (c *CPU) handleSMLALxy(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rs := c.R[(opcode>>8)&0xF]

	if (opcode & (1 << 5)) != 0 {
		rm >>= 16
	} else {
		rm &= 0xFFFF
	}
	if (opcode & (1 << 6)) != 0 {
		rs >>= 16
	} else {
		rs &= 0xFFFF
	}

	res := int64(int16(rm)) * int64(int16(rs))
	rd := int64(uint64(c.R[(opcode>>12)&0xF]) | (uint64(c.R[(opcode>>16)&0xF]) << 32))
	res += rd

	c.R[(opcode>>12)&0xF] = uint32(res)
	c.R[(opcode>>16)&0xF] = uint32(res >> 32)
}

func (c *CPU) handleCLZ(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	val := c.R[opcode&0xF]

	res := uint32(0)
	for (val & 0xFF000000) == 0 {
		res += 8
		val <<= 8
		val |= 0xFF
	}
	for (val & 0x80000000) == 0 {
		res++
		val <<= 1
		val |= 0x1
	}

	c.R[(opcode>>12)&0xF] = res
}

func (c *CPU) handleQADD(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rn := c.R[(opcode>>16)&0xF]

	res := rm + rn
	if checkOverflowAdd(rm, rn) {
		if (res & 0x80000000) != 0 {
			res = 0x7FFFFFFF
		} else {
			res = 0x80000000
		}
		c.CPSR |= 0x08000000
	}

	c.R[(opcode>>12)&0xF] = res
}

func (c *CPU) handleQSUB(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rn := c.R[(opcode>>16)&0xF]

	res := rm - rn
	if checkOverflowSub(rm, rn) {
		if (res & 0x80000000) != 0 {
			res = 0x7FFFFFFF
		} else {
			res = 0x80000000
		}
		c.CPSR |= 0x08000000
	}

	c.R[(opcode>>12)&0xF] = res
}

func (c *CPU) handleQDADD(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rn := c.R[(opcode>>16)&0xF]

	if checkOverflowAdd(rn, rn) {
		if (rn & 0x80000000) != 0 {
			rn = 0x80000000
		} else {
			rn = 0x7FFFFFFF
		}
		c.CPSR |= 0x08000000
	} else {
		rn <<= 1
	}

	res := rm + rn
	if checkOverflowAdd(rm, rn) {
		if (res & 0x80000000) != 0 {
			res = 0x7FFFFFFF
		} else {
			res = 0x80000000
		}
		c.CPSR |= 0x08000000
	}

	c.R[(opcode>>12)&0xF] = res
}

func (c *CPU) handleQDSUB(opcode uint32) {
	if !c.IsARM9 {
		return
	}
	rm := c.R[opcode&0xF]
	rn := c.R[(opcode>>16)&0xF]

	if checkOverflowAdd(rn, rn) {
		if (rn & 0x80000000) != 0 {
			rn = 0x80000000
		} else {
			rn = 0x7FFFFFFF
		}
		c.CPSR |= 0x08000000
	} else {
		rn <<= 1
	}

	res := rm - rn
	if checkOverflowSub(rm, rn) {
		if (res & 0x80000000) != 0 {
			res = 0x7FFFFFFF
		} else {
			res = 0x80000000
		}
		c.CPSR |= 0x08000000
	}

	c.R[(opcode>>12)&0xF] = res
}
