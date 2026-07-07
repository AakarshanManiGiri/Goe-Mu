package memory

import "testing"

func TestMemoryReadWrite32(t *testing.T) {
	bus := NewMemoryBus(nil, nil, nil)
	bus.Write32(0x02000000, 0xDEADBEEF) // Write to Main RAM
	val := bus.Read32(0x02000000)
	if val != 0xDEADBEEF {
		t.Errorf("Expected 0xDEADBEEF, got 0x%X", val)
	}
}

func TestCartridgeMemoryMapping(t *testing.T) {
	bus := NewMemoryBus(nil, nil, nil)
	cartData := make([]byte, 1024)
	cartData[0] = 0xAA
	cartData[1] = 0xBB
	bus.CartridgeData = cartData

	if val := bus.Read8(0x08000000); val != 0xAA {
		t.Errorf("Expected 0xAA, got 0x%X", val)
	}
	if val := bus.Read16(0x08000000); val != 0xBBAA {
		t.Errorf("Expected 0xBBAA, got 0x%X", val)
	}
}
