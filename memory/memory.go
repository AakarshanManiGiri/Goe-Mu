package memory

const mainMemoryBase uint32 = 0x02000000

// MemoryBus manages all memory operations in the NDS
type MemoryBus struct {
	ARM9ITCM   [32 * 1024]byte
	ARM9DTCM   [16 * 1024]byte
	SharedWRAM [96 * 1024]byte
	MainMemory [4 * 1024 * 1024]byte
	VRAM       [656 * 1024]byte
	OAM        [2 * 1024]byte
	IORegs     [0x1000]byte
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{}
}

func (m *MemoryBus) Read8(address uint32) byte {
	if address < mainMemoryBase {
		return 0
	}

	offset := address - mainMemoryBase
	if offset >= uint32(len(m.MainMemory)) {
		return 0
	}

	return m.MainMemory[offset]
}

func (m *MemoryBus) Read16(address uint32) uint16 {
	first := m.Read8(address)
	second := m.Read8(address + 1)
	return uint16(first) | uint16(second)<<8
}

func (m *MemoryBus) Read32(address uint32) uint32 {
	byte0 := uint32(m.Read8(address))
	byte1 := uint32(m.Read8(address + 1))
	byte2 := uint32(m.Read8(address + 2))
	byte3 := uint32(m.Read8(address + 3))
	return byte0 | byte1<<8 | byte2<<16 | byte3<<24
}

func (m *MemoryBus) Write8(address uint32, value byte) {
	if address < mainMemoryBase {
		return
	}

	offset := address - mainMemoryBase
	if offset >= uint32(len(m.MainMemory)) {
		return
	}

	m.MainMemory[offset] = value
}

func (m *MemoryBus) Write16(address uint32, value uint16) {
	m.Write8(address, byte(value))
	m.Write8(address+1, byte(value>>8))
}

func (m *MemoryBus) Write32(address uint32, value uint32) {
	m.Write8(address, byte(value))
	m.Write8(address+1, byte(value>>8))
	m.Write8(address+2, byte(value>>16))
	m.Write8(address+3, byte(value>>24))
}
