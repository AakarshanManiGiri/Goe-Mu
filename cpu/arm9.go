package cpu

import (
	"errors"

	"goe-mu/memory"
)

const demoProgramAddress uint32 = 0x02000000

type Processor struct {
	Memory *memory.MemoryBus
	PC     uint32
	Clock  uint64
}

func NewProcessor(mem *memory.MemoryBus) (*Processor, error) {
	if mem == nil {
		return nil, errors.New("cpu: memory bus is required")
	}

	processor := &Processor{Memory: mem}
	processor.Reset()
	return processor, nil
}

func (p *Processor) Reset() {
	p.PC = demoProgramAddress
	p.Clock = 0
}

func (p *Processor) Step() (uint32, error) {
	if p.Memory == nil {
		return 0, errors.New("cpu: memory bus is nil")
	}

	opcode := p.Memory.Read32(p.PC)
	p.PC += 4
	p.Clock++
	return opcode, nil
}
