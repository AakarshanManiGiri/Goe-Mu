package cpu

import (
	"testing"
)

type dummyMemory struct{}

func (d *dummyMemory) Read8(addr uint32) byte          { return 0 }
func (d *dummyMemory) Read16(addr uint32) uint16       { return 0 }
func (d *dummyMemory) Read32(addr uint32) uint32       { return 0xE1A00000 } // NOP (MOV R0, R0)
func (d *dummyMemory) Write8(addr uint32, val byte)    {}
func (d *dummyMemory) Write16(addr uint32, val uint16) {}
func (d *dummyMemory) Write32(addr uint32, val uint32) {}

func TestCPU_ExecutesNOP(t *testing.T) {
	// The CPU should not panic when stepping through a NOP
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CPU panicked on NOP: %v", r)
		}
	}()
	cpu := NewCPU(true, &dummyMemory{})
	cpu.Step()
}

func TestCPU_BranchInstruction(t *testing.T) {
	// B instruction: 0xEA000001 (Branch forward 1 instruction)
	cpu := NewCPU(true, &dummyMemory{})
	cpu.R[15] = 0 // Reset PC

	// We'll mock the memory to return a Branch opcode
	mem := &branchDummyMemory{opcode: 0xEA000001}
	cpu.Bus = mem

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CPU panicked: %v", r)
		}
	}()

	cpu.Step()

	if cpu.R[15] != 8 {
		t.Errorf("Expected PC to be 8, got %d", cpu.R[15])
	}
}

type branchDummyMemory struct{ opcode uint32 }

func (d *branchDummyMemory) Read8(addr uint32) byte          { return 0 }
func (d *branchDummyMemory) Read16(addr uint32) uint16       { return 0 }
func (d *branchDummyMemory) Read32(addr uint32) uint32       { return d.opcode }
func (d *branchDummyMemory) Write8(addr uint32, val byte)    {}
func (d *branchDummyMemory) Write16(addr uint32, val uint16) {}
func (d *branchDummyMemory) Write32(addr uint32, val uint32) {}

func TestCPU_DataProcessingADD(t *testing.T) {
	cpu := NewCPU(true, &branchDummyMemory{opcode: 0xE2812005}) // ADD R2, R1, #5
	cpu.R[1] = 10
	cpu.Step()

	if cpu.R[2] != 15 {
		t.Errorf("Expected R2 to be 15, got %d", cpu.R[2])
	}
}

func (c *CPU) StepOpcode(opcode uint32) {
	idx := (opcode >> 16) & 0xFFF
	c.ARMLUT[idx](opcode)
}

type singleDataDummyMemory struct {
	writtenAddr uint32
	writtenVal  uint32
}

func (d *singleDataDummyMemory) Read8(addr uint32) byte    { return 0 }
func (d *singleDataDummyMemory) Read16(addr uint32) uint16 { return 0 }
func (d *singleDataDummyMemory) Read32(addr uint32) uint32 {
	if addr == 0x02000004 {
		return 0xCAFEBABE
	}
	return 0
}
func (d *singleDataDummyMemory) Write8(addr uint32, val byte)    {}
func (d *singleDataDummyMemory) Write16(addr uint32, val uint16) {}
func (d *singleDataDummyMemory) Write32(addr uint32, val uint32) {
	d.writtenAddr = addr
	d.writtenVal = val
}

func TestCPU_SingleDataTransfer(t *testing.T) {
	cpu := NewCPU(true, nil)
	mem := &singleDataDummyMemory{}
	cpu.Bus = mem
	cpu.R[1] = 0x02000000

	// Test LDR
	cpu.StepOpcode(0xE5910004)
	if cpu.R[0] != 0xCAFEBABE {
		t.Errorf("Expected R0=0xCAFEBABE, got 0x%X", cpu.R[0])
	}

	// Test STR R0, [R1, #-4]
	cpu.StepOpcode(0xE5010004)
	if mem.writtenAddr != 0x01FFFFFC || mem.writtenVal != 0xCAFEBABE {
		t.Errorf("Store failed")
	}
}

func TestCPU_BlockDataTransfer(t *testing.T) {
	// STMIA R0!, {R1, R2}
	// AL(1110), 100, P=0, U=1, S=0, W=1, L=0 -> 0xE8A00006
	cpu := NewCPU(true, nil)
	mem := &blockDummyMemory{}
	cpu.Bus = mem

	cpu.R[0] = 0x1000
	cpu.R[1] = 0xAA
	cpu.R[2] = 0xBB

	cpu.StepOpcode(0xE8A00006)

	if mem.written == nil || mem.written[0x1000] != 0xAA || mem.written[0x1004] != 0xBB {
		t.Errorf("Block store failed")
	}
	if cpu.R[0] != 0x1008 {
		t.Errorf("Writeback failed, R0=%X", cpu.R[0])
	}
}

type blockDummyMemory struct {
	written map[uint32]uint32
}

func (d *blockDummyMemory) Read8(addr uint32) byte          { return 0 }
func (d *blockDummyMemory) Read16(addr uint32) uint16       { return 0 }
func (d *blockDummyMemory) Read32(addr uint32) uint32       { return 0 }
func (d *blockDummyMemory) Write8(addr uint32, val byte)    {}
func (d *blockDummyMemory) Write16(addr uint32, val uint16) {}
func (d *blockDummyMemory) Write32(addr uint32, val uint32) {
	if d.written == nil {
		d.written = make(map[uint32]uint32)
	}
	d.written[addr] = val
}
