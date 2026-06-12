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
	ARM9       *cpu.CPU
	ARM7       *cpu.CPU
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

	nds.ARM9 = cpu.NewARM9(nds.Memory)
	nds.ARM7 = cpu.NewARM7(nds.Memory)

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
	if n.ARM9 == nil || n.Memory == nil {
		return errors.New("system: emulator is not initialized")
	}

	n.Running = true

	const demoOpcode uint32 = 0xE1A00000 // NOP
	n.Memory.Write32(0x02000000, demoOpcode)
	n.ARM9.Reset()
	n.ARM7.Reset()

	// Step the ARM9
	n.ARM9.Step()

	if n.GPU != nil {
		n.GPU.UpdateFrame()
	}
	n.FrameCount++

	log.Printf("demo step complete")

	n.Stop()

	return nil
}

func (n *NDS) Stop() {
	n.Running = false
}
