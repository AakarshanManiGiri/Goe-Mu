package gpu

import (
	"testing"
)

func TestGPU_UpdateFrame(t *testing.T) {
	gpu := NewGPU()
	gpu.OutputBuffer[0] = 0x00

	gpu.UpdateFrame()

	if gpu.OutputBuffer[0] == 0x00 && gpu.OutputBuffer[1] == 0x00 {
		t.Errorf("Expected GPU to modify OutputBuffer, got %v", gpu.OutputBuffer[:4])
	}
}
