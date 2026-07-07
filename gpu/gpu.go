package gpu

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 192
)

type GPU struct {
	VRAM         []byte
	OAM          []byte
	PaletteRAM   []byte
	OutputBuffer []byte
	FrameCount   uint64
	Engine3D     *Engine3D
}

func NewGPU() *GPU {
	return &GPU{
		VRAM:         make([]byte, 656*1024),
		OAM:          make([]byte, 2*1024),
		PaletteRAM:   make([]byte, 2*1024),
		OutputBuffer: make([]byte, ScreenWidth*ScreenHeight*4),
		Engine3D:     NewEngine3D(),
	}
}

func (g *GPU) Start() {
	// No longer needs a C++ background thread.
	// Ebiten's Update/Draw loop will handle this gracefully.
}

func (g *GPU) UpdateFrame() {
	g.FrameCount++
	// Basic stub: Fill screen with magenta
	for i := 0; i < len(g.OutputBuffer); i += 4 {
		g.OutputBuffer[i] = 255   // R
		g.OutputBuffer[i+1] = 0   // G
		g.OutputBuffer[i+2] = 255 // B
		g.OutputBuffer[i+3] = 255 // A
	}
	
	g.Engine3D.ProcessCommands()
}

var emptyImage = ebiten.NewImage(3, 3)

func init() {
	emptyImage.Fill(color.White)
}

func (g *GPU) Draw(screen *ebiten.Image) {
	screen.WritePixels(g.OutputBuffer)

	if len(g.Engine3D.Indices) > 0 {
		screen.DrawTriangles(g.Engine3D.Vertices, g.Engine3D.Indices, emptyImage, &ebiten.DrawTrianglesOptions{})
	}
}

func (g *GPU) Write3DCommand(address, val uint32) {
	// NDS Geometry Engine commands are written to 0x04000400 (Command FIFO)
	// Or parameters to 0x04000440+
	g.Engine3D.WriteCommand(val)
}
