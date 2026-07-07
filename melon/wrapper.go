package melon

/*
#cgo CXXFLAGS: -std=c++17 -I${SRCDIR}/../logic/src -I${SRCDIR}/../logic/src/teakra/include
#cgo LDFLAGS: -lstdc++

#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>

bool InitMelon(const char* romPath);
void RunFrameMelon();
void GetFramesMelon(uint32_t* topScreen, uint32_t* bottomScreen);
void SetInputMelon(uint32_t keys, bool touch, uint8_t touchX, uint8_t touchY);
*/
import "C"
import "unsafe"

func Init(romPath string) bool {
    cPath := C.CString(romPath)
    defer C.free(unsafe.Pointer(cPath))
    return bool(C.InitMelon(cPath))
}

func RunFrame() {
    C.RunFrameMelon()
}

func GetFrames(top []uint32, bottom []uint32) {
    if len(top) < 256*192 || len(bottom) < 256*192 {
        return
    }
    C.GetFramesMelon((*C.uint32_t)(&top[0]), (*C.uint32_t)(&bottom[0]))
}

func SetInput(keys uint32, touch bool, touchX, touchY uint8) {
    C.SetInputMelon(C.uint32_t(keys), C.bool(touch), C.uint8_t(touchX), C.uint8_t(touchY))
}
