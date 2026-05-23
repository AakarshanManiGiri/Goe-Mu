package gpu

const (
	ScreenWidth  = 256
	ScreenHeight = 384
)

type GPU struct {
	FrameBuffer [ScreenHeight * ScreenWidth * 4]byte
	FrameCount  uint64
}

func NewGPU() *GPU {
	return &GPU{}
}

func (g *GPU) UpdateFrame() {
	g.FrameCount++
}
