package memory

const (
	mainMemoryBase uint32 = 0x02000000
	vramBase       uint32 = 0x06000000
)

// MemoryBus manages all memory operations in the NDS
type MemoryBus struct {
	ARM9ITCM   [32 * 1024]byte
	ARM9DTCM   [16 * 1024]byte
	SharedWRAM [96 * 1024]byte
	MainMemory [4 * 1024 * 1024]byte
	VRAM       []byte
	OAM        []byte
	IORegs     [0x1000]byte
	ReadJoypad func() uint16
}

func NewMemoryBus(vram []byte, oam []byte) *MemoryBus {
	return &MemoryBus{
		VRAM: vram,
		OAM:  oam,
	}
}

func (m *MemoryBus) Read8(address uint32) byte {
	if address == 0x04000130 {
		if m.ReadJoypad != nil {
			return byte(m.ReadJoypad())
		}
		return 0xFF
	}
	if address == 0x04000131 {
		if m.ReadJoypad != nil {
			return byte(m.ReadJoypad() >> 8)
		}
		return 0x0F
	}
	if address >= vramBase && address < vramBase+0x000A4000 {
		offset := address - vramBase
		return m.VRAM[offset]
	}

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
	if address >= vramBase && address < vramBase+0x000A4000 {
		offset := address - vramBase
		m.VRAM[offset] = value
		return
	}
	
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
