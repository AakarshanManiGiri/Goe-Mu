package cartridge
import "encoding/binary"
type SPI struct {
	ROM         []byte
	Command     [8]byte
	ROMCTRL     uint32
	DataBuffer  []byte
	BufferIndex int
}
func NewSPI(rom []byte) *SPI { return &SPI{ROM: rom} }
func (s *SPI) WriteCommand(cmd []byte) { copy(s.Command[:], cmd) }
func (s *SPI) WriteROMCTRL(val uint32) {
	s.ROMCTRL = val
	if (val & 0x80000000) != 0 {
		if s.Command[0] == 0xB7 { // Read Data
			addr := binary.BigEndian.Uint32(s.Command[1:5]) & 0x0FFFFFFF
			length := (val >> 24) & 7
			transferSize := uint32(0x100)
			if length == 7 { transferSize = 4 }
			end := addr + transferSize
			if int(end) > len(s.ROM) {
				end = uint32(len(s.ROM))
			}
			if int(addr) > len(s.ROM) {
			    addr = uint32(len(s.ROM))
			}
			s.DataBuffer = s.ROM[addr : end]
			s.BufferIndex = 0
		} else if s.Command[0] == 0xB8 { // Chip ID
			s.DataBuffer = []byte{0xC2, 0x1F, 0x00, 0x00}
			s.BufferIndex = 0
		}
	}
}
func (s *SPI) ReadData32() uint32 {
	if s.BufferIndex+4 <= len(s.DataBuffer) {
		val := binary.LittleEndian.Uint32(s.DataBuffer[s.BufferIndex:])
		s.BufferIndex += 4
		return val
	}
	return 0
}
