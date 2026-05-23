package main

import (
	"log"

	"goe-mu/system"
)

func main() {
	log.Println("Initializing NDS Emulator (Goe-Mu)...")

	// Initialize the NDS system
	nds, err := system.NewNDS()
	if err != nil {
		log.Fatalf("Failed to initialize NDS: %v", err)
	}

	// Start the emulator
	if err := nds.Run(); err != nil {
		log.Fatalf("Emulator error: %v", err)
	}
}
