package main

var Config = struct {
	CyclesPerFrame uint64
	TargetFPS      float64
	DebugMode      bool
	VerboseLogging bool

	ARM9Clock    uint32 // Hz
	ARM7Clock    uint32 // Hz
	ScreenWidth  uint16
	ScreenHeight uint16

	ROMPath string
}{
	CyclesPerFrame: 560160,
	TargetFPS:      59.8261,
	DebugMode:      false,
	VerboseLogging: false,
	ARM9Clock:      33513982,
	ARM7Clock:      33513982,
	ScreenWidth:    256,
	ScreenHeight:   192,
	ROMPath:        "",
}
