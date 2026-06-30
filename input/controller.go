package input

import "github.com/hajimehoshi/ebiten/v2"

type ButtonMask uint16

const (
	ButtonA      ButtonMask = 1 << 0
	ButtonB      ButtonMask = 1 << 1
	ButtonSelect ButtonMask = 1 << 2
	ButtonStart  ButtonMask = 1 << 3
	ButtonRight  ButtonMask = 1 << 4
	ButtonLeft   ButtonMask = 1 << 5
	ButtonUp     ButtonMask = 1 << 6
	ButtonDown   ButtonMask = 1 << 7
	ButtonR      ButtonMask = 1 << 8
	ButtonL      ButtonMask = 1 << 9
)

type Controller struct {
	ButtonState ButtonMask
	TouchX      int16
	TouchY      int16
	IsTouching  bool
}

func NewController() *Controller {
	return &Controller{
		ButtonState: 0,
		IsTouching:  false,
	}
}

func (c *Controller) PressButton(btn ButtonMask) {
	c.ButtonState |= btn
}

func (c *Controller) ReleaseButton(btn ButtonMask) {
	c.ButtonState &= ^btn
}

func (c *Controller) IsButtonPressed(btn ButtonMask) bool {
	return (c.ButtonState & btn) != 0
}

func (c *Controller) SetTouchInput(x, y int16, touching bool) {
	c.TouchX = x
	c.TouchY = y
	c.IsTouching = touching
}

func (c *Controller) ReadJoypad() uint16 {
	// NDS active-low joypad: 0=pressed, 1=released
	var state uint16 = 0x0FFF // All released
	
	if ebiten.IsKeyPressed(ebiten.KeyZ) { state &^= (1 << 0) }
	if ebiten.IsKeyPressed(ebiten.KeyX) { state &^= (1 << 1) }
	if ebiten.IsKeyPressed(ebiten.KeyShift) { state &^= (1 << 2) }
	if ebiten.IsKeyPressed(ebiten.KeyEnter) { state &^= (1 << 3) }
	if ebiten.IsKeyPressed(ebiten.KeyRight) { state &^= (1 << 4) }
	if ebiten.IsKeyPressed(ebiten.KeyLeft) { state &^= (1 << 5) }
	if ebiten.IsKeyPressed(ebiten.KeyUp) { state &^= (1 << 6) }
	if ebiten.IsKeyPressed(ebiten.KeyDown) { state &^= (1 << 7) }
	if ebiten.IsKeyPressed(ebiten.KeyS) { state &^= (1 << 8) }
	if ebiten.IsKeyPressed(ebiten.KeyA) { state &^= (1 << 9) }
	if ebiten.IsKeyPressed(ebiten.KeyC) { state &^= (1 << 10) }
	if ebiten.IsKeyPressed(ebiten.KeyV) { state &^= (1 << 11) }
	
	return state
}
