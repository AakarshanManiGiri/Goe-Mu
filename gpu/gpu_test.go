package gpu

import (
	"testing"
	"time"
)

func TestGPU_StartThread(t *testing.T) {
	vram := make([]byte, 656*1024)
	oam := make([]byte, 2*1024)
	gpu := NewGPU(vram, oam)
	gpu.OutputBuffer[0] = 0x00 // Initialize memory

	// Start the background C++ thread
	gpu.Start()

	// Wait a moment for the C++ thread to run and modify OutputBuffer
	time.Sleep(100 * time.Millisecond)

	// The C++ thread should have written something to the output buffer
	if gpu.OutputBuffer[0] == 0x00 && gpu.OutputBuffer[1] == 0x00 {
		t.Errorf("Expected C++ thread to modify OutputBuffer, got %v", gpu.OutputBuffer[:4])
	}
}
