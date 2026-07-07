package system

import "goe-mu/cpu"

func HandleSWI(c *cpu.CPU, swiNum uint32) {
	switch swiNum {
	case 0x06: // Div
		num := int32(c.R[0])
		denom := int32(c.R[1])
		if denom == 0 {
			c.R[0] = 0
			c.R[1] = uint32(num)
		} else {
			c.R[0] = uint32(num / denom)
			c.R[1] = uint32(num % denom)
		}
	case 0x02: // Halt
		// Stub for now
	case 0x04: // IntrWait
		// Stub for now
	}
}
