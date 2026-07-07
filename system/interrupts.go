package system
type Interrupts struct {
	IME uint16
	IE  uint32
	IF  uint32
}
func NewInterrupts() *Interrupts { return &Interrupts{} }
func (irq *Interrupts) WriteIME(val uint16) { irq.IME = val & 1 }
func (irq *Interrupts) ReadIME() uint16 { return irq.IME }
func (irq *Interrupts) WriteIE(val uint32) { irq.IE = val }
func (irq *Interrupts) ReadIE() uint32 { return irq.IE }
func (irq *Interrupts) WriteIF(val uint32) { irq.IF &= ^val } // writing 1 acknowledges
func (irq *Interrupts) ReadIF() uint32 { return irq.IF }
func (irq *Interrupts) Request(bit int) { irq.IF |= (1 << bit) }
func (irq *Interrupts) Pending() bool { return (irq.IME & 1) != 0 && (irq.IE & irq.IF) != 0 }
