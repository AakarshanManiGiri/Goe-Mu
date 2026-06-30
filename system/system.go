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
	gpuInst := gpu.NewGPU()
	
	nds := &NDS{
		GPU:        gpuInst,
		Memory:     memory.NewMemoryBus(gpuInst.VRAM, gpuInst.OAM),
		Input:      input.NewController(),
		Running:    false,
		FrameCount: 0,
	}
	
	nds.Memory.ReadJoypad = nds.Input.ReadJoypad
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

func (n *NDS) InjectBoot() error {
	if n.Cartridge == nil || len(n.Cartridge.Data) == 0 {
		return errors.New("no cartridge loaded")
	}

	if err := n.Cartridge.ParseHeader(); err != nil {
		return err
	}

	hdr := n.Cartridge.Header
	if uint32(len(n.Cartridge.Data)) < hdr.ARM9Offset+hdr.ARM9Size {
		return errors.New("ROM is smaller than expected ARM9 size")
	}

	arm9Payload := n.Cartridge.Data[hdr.ARM9Offset : hdr.ARM9Offset+hdr.ARM9Size]

	dest := hdr.ARM9RAMAddr
	log.Printf("Injecting ARM9 Boot Payload of size %d bytes to 0x%08X", len(arm9Payload), dest)
	for i := 0; i < len(arm9Payload); i++ {
		n.Memory.Write8(dest+uint32(i), arm9Payload[i])
	}

	// Reset CPUs to put them in a clean state (Supervisor mode, interrupts disabled, etc)
	n.ARM9.Reset()
	n.ARM7.Reset()

	// Override PC to the entry point
	n.ARM9.R[15] = hdr.ARM9Entry
	log.Printf("ARM9 PC Set to Entry Point: 0x%08X", hdr.ARM9Entry)

	n.Running = true
	return nil
}

func (n *NDS) Stop() {
	n.Running = false
}
