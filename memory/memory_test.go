package memory

import "testing"

func TestMemoryReadWrite32(t *testing.T) {
	bus := NewMemoryBus()
	bus.Write32(0x02000000, 0xDEADBEEF) // Write to Main RAM
	val := bus.Read32(0x02000000)
	if val != 0xDEADBEEF {
		t.Errorf("Expected 0xDEADBEEF, got 0x%X", val)
	}
}
