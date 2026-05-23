package system

import (
	"errors"
	"log"

	"goe-mu/cartridge"
	"goe-mu/cpu"
	"goe-mu/gpu"
	"goe-mu/input"
	"goe-mu/memory"
)

type NDS struct {
	CPU        *cpu.Processor
	Memory     *memory.MemoryBus
	GPU        *gpu.GPU
	Input      *input.Controller
	Cartridge  *cartridge.Cartridge
	Running    bool
	FrameCount uint64
}

func NewNDS() (*NDS, error) {
	nds := &NDS{
		Memory:     memory.NewMemoryBus(),
		GPU:        gpu.NewGPU(),
		Input:      input.NewController(),
		Running:    false,
		FrameCount: 0,
	}

	var err error
	nds.CPU, err = cpu.NewProcessor(nds.Memory)
	if err != nil {
		return nil, err
	}

	return nds, nil
}

func (n *NDS) LoadCartridge(path string) error {
	var err error
	n.Cartridge, err = cartridge.NewCartridge(path)
	if err != nil {
		return err
	}
	return nil
}

func (n *NDS) Run() error {
	if n.CPU == nil || n.Memory == nil {
		return errors.New("system: emulator is not initialized")
	}

	n.Running = true

	const demoOpcode uint32 = 0xE1A00000
	n.Memory.Write32(0x02000000, demoOpcode)
	n.CPU.Reset()

	opcode, err := n.CPU.Step()
	if err != nil {
		return err
	}

	if n.GPU != nil {
		n.GPU.UpdateFrame()
	}
	n.FrameCount++

	log.Printf("demo step: pc=%08x opcode=%08x", n.CPU.PC-4, opcode)

	n.Stop()

	return nil
}

func (n *NDS) Stop() {
	n.Running = false
}
