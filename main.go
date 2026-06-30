package main

import (
	"image/color"
	"log"

	"goe-mu/cartridge"
	"goe-mu/system"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sqweek/dialog"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 384
)

type Game struct {
	nds *system.NDS
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
				if loadErr := g.nds.LoadCartridge(file); loadErr == nil {
					_ = g.nds.InjectBoot() // Ignore errors for demo
				}
			}
		}
	}

	// Step the CPU if it's running
	if g.nds.Running {
		// Log PC once per frame to see where it is
		log.Printf("Frame %d: PC=%08X", g.nds.FrameCount, g.nds.ARM9.R[15])
		g.nds.FrameCount++
		
		for i := 0; i < 5000; i++ {
			// Recover panic in case we hit an unimplemented opcode
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("CPU Halted: %v", r)
						g.nds.Running = false
					}
				}()
				g.nds.ARM9.Step()
			}()
			if !g.nds.Running {
				break
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Write the C++ generated framebuffer to the screen
	screen.WritePixels(g.nds.GPU.OutputBuffer)

	// Draw UI over the framebuffer
	// Red button rectangle for Load ROM
	vector.DrawFilledRect(screen, 10, 10, 100, 30, color.RGBA{200, 50, 50, 255}, false)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	nds, err := system.NewNDS()
	if err != nil {
		log.Fatal(err)
	}

	// Create an empty mock cartridge so it doesn't crash without a file
	nds.Cartridge = &cartridge.Cartridge{
		Data: make([]byte, 1024),
	}
	_ = nds.InjectBoot()

	// Start the background C++ rendering thread
	nds.GPU.Start()

	game := &Game{nds: nds}

	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowTitle("Goe-Mu (Hybrid NDS Emulator)")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
