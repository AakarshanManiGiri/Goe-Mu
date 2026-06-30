package gpu

/*
#cgo CXXFLAGS: -std=c++11
#include <stdlib.h>
#include "wrapper.h"
*/
import "C"
import (
	"unsafe"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 384
)

type GPU struct {
	VRAM         []byte
	OAM          []byte
	OutputBuffer []byte
	FrameCount   uint64
	vramPtr      unsafe.Pointer
	oamPtr       unsafe.Pointer
	outPtr       unsafe.Pointer
}

func NewGPU() *GPU {
	vramPtr := C.malloc(C.size_t(656 * 1024))
	oamPtr := C.malloc(C.size_t(2 * 1024))
	outPtr := C.malloc(C.size_t(256 * 384 * 4))

	return &GPU{
		VRAM:         unsafe.Slice((*byte)(vramPtr), 656*1024),
		OAM:          unsafe.Slice((*byte)(oamPtr), 2*1024),
		OutputBuffer: unsafe.Slice((*byte)(outPtr), 256*384*4),
		vramPtr:      vramPtr,
		oamPtr:       oamPtr,
		outPtr:       outPtr,
	}
}

// Start begins the headless C++ rendering thread
func (g *GPU) Start() {
	// Pass pointers to the C++ thread (allocated via C.malloc, 100% safe)
	C.StartGPUThread(g.vramPtr, g.oamPtr, g.outPtr)
}

func (g *GPU) UpdateFrame() {
	g.FrameCount++
}
