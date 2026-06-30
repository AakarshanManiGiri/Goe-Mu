#ifndef GPU_WRAPPER_H
#define GPU_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

void StartGPUThread(void* vram, void* oam, void* output_fb);

#ifdef __cplusplus
}
#endif

#endif // GPU_WRAPPER_H
