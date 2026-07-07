#include "../logic/src/types.h"
#include "../logic/src/NDS.h"
#include "../logic/src/GPU.h"
#include "../logic/src/Args.h"
#include "../logic/src/NDSCart.h"
#include <string.h>
#include <stdio.h>
#include <stdlib.h>

static melonDS::NDS* emu = nullptr;
static melonDS::u8* romBuffer = nullptr;
static size_t romSize = 0;

extern "C" {
    bool InitMelon(const char* romPath) {
        if (emu != nullptr) {
            delete emu;
            emu = nullptr;
        }
        if (romBuffer != nullptr) {
            delete[] romBuffer;
            romBuffer = nullptr;
        }

        // Load ROM into memory
        FILE* f = fopen(romPath, "rb");
        if (!f) return false;
        fseek(f, 0, SEEK_END);
        romSize = ftell(f);
        fseek(f, 0, SEEK_SET);
        romBuffer = new melonDS::u8[romSize];
        fread(romBuffer, 1, romSize, f);
        fclose(f);

        melonDS::NDSArgs args;
        // The default NDSArgs constructor populates FreeBIOS for ARM7/9 and Firmware.
        args.JIT = std::nullopt;
        
        emu = new melonDS::NDS(std::move(args), nullptr);
        
        // Let's parse the ROM
        auto cart = melonDS::NDSCart::ParseROM(romBuffer, romSize);
        if (!cart) {
            return false;
        }
        
        emu->SetNDSCart(std::move(cart));
        emu->SetupDirectBoot(romPath);
        emu->Start();
        
        return true;
    }

    void RunFrameMelon() {
        if (emu) {
            emu->RunFrame();
        }
    }

    void GetFramesMelon(uint32_t* topScreen, uint32_t* bottomScreen) {
        if (!emu) return;
        
        // In Software Renderer mode, GetFramebuffers gives us the pointers.
        void* topPtr = nullptr;
        void* bottomPtr = nullptr;
        if (emu->GPU.GetFramebuffers(&topPtr, &bottomPtr)) {
            // It is an array of 256*192 uint32_t
            if (topPtr) {
                memcpy(topScreen, topPtr, 256 * 192 * 4);
            }
            if (bottomPtr) {
                memcpy(bottomScreen, bottomPtr, 256 * 192 * 4);
            }
        }
    }

    void SetInputMelon(uint32_t keys, bool touch, uint8_t touchX, uint8_t touchY) {
        if (!emu) return;
        emu->SetKeyMask(keys);
        if (touch) {
            emu->TouchScreen(touchX, touchY);
        } else {
            emu->ReleaseScreen();
        }
    }
}
