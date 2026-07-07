package cpu

type CP15 struct {
	ControlRegister uint32
	DTCMRegion      uint32
	ITCMRegion      uint32
}
