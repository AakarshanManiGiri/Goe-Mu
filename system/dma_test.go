package system
import "testing"

func TestDMA(t *testing.T) {
	dma := NewDMA()
	if dma.Channels[0].Control != 0 {
		t.Errorf("Expected channel 0 control to be 0")
	}
}
