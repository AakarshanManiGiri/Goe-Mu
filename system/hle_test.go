package system

import "testing"

func TestHLEHalt(t *testing.T) {
	nds, _ := NewNDS()
	
	// Test the HLE dispatcher directly
	nds.ARM9.R[0] = 10 // dummy arg (num)
	nds.ARM9.R[1] = 3  // dummy arg (denom)
	
	// Call SWI 0x06 (Div)
	HandleSWI(nds.ARM9, 0x06)
	
	// R0 should be R0/R1 = 3, R1 = R0%R1 = 1
	if nds.ARM9.R[0] != 3 || nds.ARM9.R[1] != 1 {
		t.Errorf("Div SWI failed: got R0=%d, R1=%d", nds.ARM9.R[0], nds.ARM9.R[1])
	}
}
