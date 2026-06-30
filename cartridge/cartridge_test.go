package cartridge

import "testing"

func TestNewCartridge_InvalidPath(t *testing.T) {
	_, err := NewCartridge("non_existent_file.nds")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}

func TestCartridge_ParseHeader(t *testing.T) {
	// Create a dummy 1KB ROM in memory with a fake header
	dummyData := make([]byte, 1024)

	// Write ARM9 Rom Offset (0x020) = 0x200
	dummyData[0x020] = 0x00
	dummyData[0x021] = 0x02
	dummyData[0x022] = 0x00
	dummyData[0x023] = 0x00

	// Write ARM9 Entry Address (0x024) = 0x02000000
	dummyData[0x024] = 0x00
	dummyData[0x025] = 0x00
	dummyData[0x026] = 0x00
	dummyData[0x027] = 0x02

	// Write ARM9 RAM Address (0x028) = 0x02000000
	dummyData[0x028] = 0x00
	dummyData[0x029] = 0x00
	dummyData[0x02A] = 0x00
	dummyData[0x02B] = 0x02

	// Write ARM9 Size (0x02C) = 0x100
	dummyData[0x02C] = 0x00
	dummyData[0x02D] = 0x01
	dummyData[0x02E] = 0x00
	dummyData[0x02F] = 0x00

	cart := &Cartridge{Data: dummyData}
	err := cart.ParseHeader()

	if err != nil {
		t.Fatalf("ParseHeader failed: %v", err)
	}

	if cart.Header.ARM9Offset != 0x200 {
		t.Errorf("Expected ARM9Offset 0x200, got 0x%X", cart.Header.ARM9Offset)
	}
	if cart.Header.ARM9Entry != 0x02000000 {
		t.Errorf("Expected ARM9Entry 0x02000000, got 0x%X", cart.Header.ARM9Entry)
	}
}
