package gpu

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Engine3D struct {
	CommandFIFO []uint32
	Vertices    []ebiten.Vertex
	Indices     []uint16
}

func NewEngine3D() *Engine3D {
	return &Engine3D{
		CommandFIFO: make([]uint32, 0),
		Vertices:    make([]ebiten.Vertex, 0),
		Indices:     make([]uint16, 0),
	}
}

func (e *Engine3D) WriteCommand(cmd uint32) {
	e.CommandFIFO = append(e.CommandFIFO, cmd)
}

func (e *Engine3D) ProcessCommands() {
	// 3D Geometry command processing will go here
	// For now, clear FIFO
	e.CommandFIFO = e.CommandFIFO[:0]
}
