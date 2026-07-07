package cpu

import "testing"

type mockMem struct{}
func (mockMem) Read8(a uint32) byte { return 0 }
func (mockMem) Read16(a uint32) uint16 { return 0 }
func (mockMem) Read32(a uint32) uint32 { return 0 }
func (mockMem) Write8(a uint32, v byte) {}
func (mockMem) Write16(a uint32, v uint16) {}
func (mockMem) Write32(a uint32, v uint32) {}

func TestCP15(t *testing.T) {
	arm9 := NewARM9(mockMem{})
	
	// Write to CP15 Control Register (MCR p15, 0, R0, c1, c0, 0)
	arm9.R[0] = 0x12345678
	opcode := uint32(0xEE010F10) // MCR p15, 0, r0, cr1, cr0, {0}
	
	arm9.handleCoprocessorRegisterTransfer(opcode)
	
	if arm9.CP15.ControlRegister != 0x12345678 {
		t.Errorf("CP15 Control not set")
	}
}
