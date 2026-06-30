package cartridge

import (
	"encoding/binary"
	"errors"
	"os"
)

type NDSHeader struct {
	ARM9Offset  uint32
	ARM9Entry   uint32
	ARM9RAMAddr uint32
	ARM9Size    uint32
}

type Cartridge struct {
	Data   []byte
	Header NDSHeader
}

func NewCartridge(path string) (*Cartridge, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("failed to read ROM: " + err.Error())
	}
	return &Cartridge{Data: data}, nil
}

func (c *Cartridge) ParseHeader() error {
	if len(c.Data) < 0x200 {
		return errors.New("ROM too small for header")
	}

	c.Header.ARM9Offset = binary.LittleEndian.Uint32(c.Data[0x020:0x024])
	c.Header.ARM9Entry = binary.LittleEndian.Uint32(c.Data[0x024:0x028])
	c.Header.ARM9RAMAddr = binary.LittleEndian.Uint32(c.Data[0x028:0x02C])
	c.Header.ARM9Size = binary.LittleEndian.Uint32(c.Data[0x02C:0x030])

	return nil
}
