package input

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
