package cartridge
import "testing"
func TestSPIRead(t *testing.T) {
	spi := NewSPI([]byte{0x11, 0x22, 0x33, 0x44})
	spi.WriteCommand([]byte{0xB7, 0, 0, 0, 0, 0, 0, 0}) // Read Data
	spi.WriteROMCTRL(0x80000000) // Start transfer
	if val := spi.ReadData32(); val != 0x44332211 { t.Errorf("Got %X", val) }
}
