package main

import (
	"fmt"
	"image/color"
	"log"

	"goe-mu/melon"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sqweek/dialog"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 192 * 2 // 2 screens
)

type Game struct {
	running      bool
	topFrame     []uint32
	bottomFrame  []uint32
	screenBuffer []byte
}

func (g *Game) Update() error {
	// Check for UI click on the Load ROM button
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if mx >= 10 && mx <= 110 && my >= 10 && my <= 40 {
			// Trigger file selection dialog
			file, err := dialog.File().Filter("NDS ROMs", "nds").Load()
			if err == nil {
				// User picked a file, attempt to load it
				if melon.Init(file) {
					g.running = true
				} else {
					log.Println("Failed to init melonDS with ROM:", file)
				}
			}
		}
	}

	if g.running {
		// Capture input
		var keys uint32 = 0xFFF // default all released (melonDS usually uses active low for keys, but let's check input handling if needed)
		// Wait, NDS::SetKeyMask usually takes active-low keys for NDS? Actually in melonDS keys might be active low.
		// Let's pass 0xFFF for now.
		
		touchX, touchY := 0, 0
		touch := false
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			if my >= 192 {
				touchX = mx
				touchY = my - 192
				touch = true
			}
		}
		
		melon.SetInput(keys, touch, uint8(touchX), uint8(touchY))
		
		// Run a frame
		melon.RunFrame()
		
		// Grab framebuffers
		melon.GetFrames(g.topFrame, g.bottomFrame)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.running {
		// Convert []uint32 frames to []byte RGBA for Ebiten
		// Top screen
		for i := 0; i < 256*192; i++ {
			c := g.topFrame[i]
			// format might be RGBA or BGRA depending on melonDS GPU
			g.screenBuffer[i*4+0] = byte(c & 0xFF)         // R
			g.screenBuffer[i*4+1] = byte((c >> 8) & 0xFF)  // G
			g.screenBuffer[i*4+2] = byte((c >> 16) & 0xFF) // B
			g.screenBuffer[i*4+3] = 255                    // A
		}
		
		// Bottom screen
		for i := 0; i < 256*192; i++ {
			c := g.bottomFrame[i]
			g.screenBuffer[(i+256*192)*4+0] = byte(c & 0xFF)
			g.screenBuffer[(i+256*192)*4+1] = byte((c >> 8) & 0xFF)
			g.screenBuffer[(i+256*192)*4+2] = byte((c >> 16) & 0xFF)
			g.screenBuffer[(i+256*192)*4+3] = 255
		}
		
		screen.WritePixels(g.screenBuffer)
	}

	// Draw UI over the framebuffer
	if !g.running {
		// Red button rectangle for Load ROM
		vector.DrawFilledRect(screen, 10, 10, 100, 30, color.RGBA{200, 50, 50, 255}, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	game := &Game{
		running:      false,
		topFrame:     make([]uint32, 256*192),
		bottomFrame:  make([]uint32, 256*192),
		screenBuffer: make([]byte, 256*192*2*4),
	}

	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowTitle("Goe-Mu (Hybrid NDS Emulator)")

	if err := ebiten.RunGame(game); err != nil {
		fmt.Println("Error running game:", err)
	}
}
