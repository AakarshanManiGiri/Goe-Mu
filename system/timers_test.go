package system

import "testing"

func TestTimers(t *testing.T) {
	timers := NewTimers()
	timers.WriteControl(0, 0x0080) // Enable Timer 0, prescaler 1
	timers.Tick(100)
	
	val := timers.ReadCounter(0)
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}
}
