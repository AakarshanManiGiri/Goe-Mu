package system

type Timer struct {
	Counter uint16
	Reload  uint16
	Control uint16
}

type Timers struct {
	T [4]Timer
}

func NewTimers() *Timers {
	return &Timers{}
}

func (ts *Timers) WriteControl(id int, val uint16) {
	ts.T[id].Control = val
}

func (ts *Timers) ReadCounter(id int) uint16 {
	return ts.T[id].Counter
}

func (ts *Timers) Tick(cycles uint64) {
	for i := 0; i < 4; i++ {
		if (ts.T[i].Control & 0x0080) != 0 {
			ts.T[i].Counter += uint16(cycles)
		}
	}
}
