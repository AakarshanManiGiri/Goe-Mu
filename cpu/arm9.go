package cpu

func NewARM9(bus MemoryInterface) *CPU {
	cpu := NewCPU(true, bus)
	cpu.Reset()
	return cpu
}

func (c *CPU) Reset() {
	// Initialize mode to System or Supervisor depending on standard NDS boot
	c.CPSR = 0x000000DF // System mode, ARM state, IRQ/FIQ disabled

	// Default PC for demo
	c.R[15] = 0x02000000
	c.Cycles = 0
}
