package cpu

func NewARM7(bus MemoryInterface) *CPU {
	cpu := NewCPU(false, bus)
	cpu.Reset()
	return cpu
}
