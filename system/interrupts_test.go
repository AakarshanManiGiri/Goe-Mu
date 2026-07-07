package system
import "testing"
func TestInterrupts(t *testing.T) {
	irq := NewInterrupts()
	irq.WriteIME(1)
	irq.WriteIE(1) // Enable VBlank
	irq.Request(0) // Request VBlank
	if irq.ReadIF() != 1 { t.Errorf("Expected IF=1") }
}
