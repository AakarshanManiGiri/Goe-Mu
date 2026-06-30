#include "wrapper.h"
#include <thread>
#include <chrono>
#include <iostream>

static void GPUThreadLoop(void* vram, void* oam, void* output_fb) {
    std::cout << "[C++ GPU] Thread started." << std::endl;
    unsigned char* out = static_cast<unsigned char*>(output_fb);
    unsigned char* vram_bytes = static_cast<unsigned char*>(vram);
    
    // Run infinitely (for now, until program exit)
    while (true) {
        std::this_thread::sleep_for(std::chrono::milliseconds(16)); // ~60Hz
        
        if (out != nullptr && vram_bytes != nullptr) {
            // Render a 256x192 Bitmap from VRAM offset 0
            // Assuming 16-bit colors (RGB555)
            for (int y = 0; y < 192; y++) {
                for (int x = 0; x < 256; x++) {
                    int vram_index = (y * 256 + x) * 2; // 2 bytes per pixel
                    unsigned short color16 = vram_bytes[vram_index] | (vram_bytes[vram_index + 1] << 8);
                    
                    int r = (color16 & 0x1F) << 3;
                    int g = ((color16 >> 5) & 0x1F) << 3;
                    int b = ((color16 >> 10) & 0x1F) << 3;
                    
                    int out_index = (y * 256 + x) * 4;
                    out[out_index + 0] = r;
                    out[out_index + 1] = g;
                    out[out_index + 2] = b;
                    out[out_index + 3] = 255;
                }
            }
        }
    }
}

void StartGPUThread(void* vram, void* oam, void* output_fb) {
    std::thread gpuThread(GPUThreadLoop, vram, oam, output_fb);
    gpuThread.detach();
}
