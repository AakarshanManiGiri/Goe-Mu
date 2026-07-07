package system

type DMAChannel struct {
	Source    uint32
	Dest      uint32
	WordCount uint32
	Control   uint32
}

type DMA struct {
	Channels [4]DMAChannel
}

func NewDMA() *DMA {
	return &DMA{}
}

func (d *DMA) Check() {
	// Execute transfers immediately based on Control register
}
