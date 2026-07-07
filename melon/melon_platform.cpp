#include "../../logic/src/Platform.h"
#include <stdio.h>
#include <string>
#include <thread>
#include <mutex>
#include <chrono>
#include <condition_variable>
#include <stdarg.h>

namespace melonDS::Platform {

void SignalStop(StopReason reason, void* userdata) {}

std::string GetLocalFilePath(const std::string& filename) { return filename; }

struct FileHandle {
    FILE* f;
};

FileHandle* OpenFile(const std::string& path, FileMode mode) {
    const char* m = "rb";
    if (mode & Write) m = "wb";
    if (mode & ReadWrite) m = "r+b";
    if (mode & Append) m = "ab";
    FILE* f = fopen(path.c_str(), m);
    if (!f) return nullptr;
    return new FileHandle{f};
}

FileHandle* OpenLocalFile(const std::string& path, FileMode mode) {
    return OpenFile(path, mode);
}

bool FileExists(const std::string& name) {
    FILE* f = fopen(name.c_str(), "rb");
    if (f) { fclose(f); return true; }
    return false;
}

bool LocalFileExists(const std::string& name) { return FileExists(name); }

bool CheckFileWritable(const std::string& filepath) { return true; }
bool CheckLocalFileWritable(const std::string& filepath) { return true; }

bool CloseFile(FileHandle* file) {
    if (file) { fclose(file->f); delete file; return true; }
    return false;
}

bool IsEndOfFile(FileHandle* file) { return feof(file->f); }
bool FileReadLine(char* str, int count, FileHandle* file) { return fgets(str, count, file->f) != nullptr; }
u64 FilePosition(FileHandle* file) { return ftell(file->f); }
bool FileSeek(FileHandle* file, s64 offset, FileSeekOrigin origin) {
    int o = SEEK_SET;
    if (origin == FileSeekOrigin::Current) o = SEEK_CUR;
    if (origin == FileSeekOrigin::End) o = SEEK_END;
    return fseek(file->f, offset, o) == 0;
}
void FileRewind(FileHandle* file) { rewind(file->f); }
u64 FileRead(void* data, u64 size, u64 count, FileHandle* file) { return fread(data, size, count, file->f); }
bool FileFlush(FileHandle* file) { return fflush(file->f) == 0; }
u64 FileWrite(const void* data, u64 size, u64 count, FileHandle* file) { return fwrite(data, size, count, file->f); }
u64 FileWriteFormatted(FileHandle* file, const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    int res = vfprintf(file->f, fmt, args);
    va_end(args);
    return res;
}
u64 FileLength(FileHandle* file) {
    long pos = ftell(file->f);
    fseek(file->f, 0, SEEK_END);
    long len = ftell(file->f);
    fseek(file->f, pos, SEEK_SET);
    return len;
}

void Log(LogLevel level, const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    vprintf(fmt, args);
    va_end(args);
}

struct Thread { std::thread t; };
Thread* Thread_Create(std::function<void()> func) { return new Thread{std::thread(func)}; }
void Thread_Free(Thread* thread) { delete thread; }
void Thread_Wait(Thread* thread) { if (thread->t.joinable()) thread->t.join(); }

struct Semaphore {
    std::mutex m;
    std::condition_variable cv;
    int count = 0;
};
Semaphore* Semaphore_Create() { return new Semaphore(); }
void Semaphore_Free(Semaphore* sema) { delete sema; }
void Semaphore_Reset(Semaphore* sema) { sema->count = 0; }
void Semaphore_Wait(Semaphore* sema) {
    std::unique_lock<std::mutex> lock(sema->m);
    sema->cv.wait(lock, [sema]() { return sema->count > 0; });
    sema->count--;
}
bool Semaphore_TryWait(Semaphore* sema, int timeout_ms) {
    std::unique_lock<std::mutex> lock(sema->m);
    if (timeout_ms == 0) {
        if (sema->count > 0) { sema->count--; return true; }
        return false;
    }
    if (sema->cv.wait_for(lock, std::chrono::milliseconds(timeout_ms), [sema]() { return sema->count > 0; })) {
        sema->count--;
        return true;
    }
    return false;
}
void Semaphore_Post(Semaphore* sema, int count) {
    std::lock_guard<std::mutex> lock(sema->m);
    sema->count += count;
    for(int i=0; i<count; i++) sema->cv.notify_one();
}

struct Mutex { std::mutex m; };
Mutex* Mutex_Create() { return new Mutex(); }
void Mutex_Free(Mutex* mutex) { delete mutex; }
void Mutex_Lock(Mutex* mutex) { mutex->m.lock(); }
void Mutex_Unlock(Mutex* mutex) { mutex->m.unlock(); }
bool Mutex_TryLock(Mutex* mutex) { return mutex->m.try_lock(); }

void Sleep(u64 usecs) { std::this_thread::sleep_for(std::chrono::microseconds(usecs)); }

u64 GetMSCount() {
    auto now = std::chrono::steady_clock::now();
    return std::chrono::duration_cast<std::chrono::milliseconds>(now.time_since_epoch()).count();
}

u64 GetUSCount() {
    auto now = std::chrono::steady_clock::now();
    return std::chrono::duration_cast<std::chrono::microseconds>(now.time_since_epoch()).count();
}

void WriteNDSSave(const u8* savedata, u32 savelen, u32 writeoffset, u32 writelen, void* userdata) {}
void WriteGBASave(const u8* savedata, u32 savelen, u32 writeoffset, u32 writelen, void* userdata) {}
void WriteFirmware(const Firmware& firmware, u32 writeoffset, u32 writelen, void* userdata) {}
void WriteDateTime(int year, int month, int day, int hour, int minute, int second, void* userdata) {}
void MP_Begin(void* userdata) {}
void MP_End(void* userdata) {}
int MP_SendPacket(u8* data, int len, u64 timestamp, void* userdata) { return 0; }
int MP_RecvPacket(u8* data, u64* timestamp, void* userdata) { return 0; }
int MP_SendCmd(u8* data, int len, u64 timestamp, void* userdata) { return 0; }
int MP_SendReply(u8* data, int len, u64 timestamp, u16 aid, void* userdata) { return 0; }
int MP_SendAck(u8* data, int len, u64 timestamp, void* userdata) { return 0; }
int MP_RecvHostPacket(u8* data, u64* timestamp, void* userdata) { return 0; }
u16 MP_RecvReplies(u8* data, u64 timestamp, u16 aidmask, void* userdata) { return 0; }
int Net_SendPacket(u8* data, int len, void* userdata) { return 0; }
int Net_RecvPacket(u8* data, void* userdata) { return 0; }

void Camera_Start(int num, void* userdata) {}
void Camera_Stop(int num, void* userdata) {}
void Camera_CaptureFrame(int num, u32* frame, int width, int height, bool yuv, void* userdata) {}
void Mic_Start(void* userdata) {}
void Mic_Stop(void* userdata) {}
int Mic_ReadInput(s16* data, int maxlength, void* userdata) { return 0; }

struct AACDecoder {};
AACDecoder* AAC_Init() { return nullptr; }
void AAC_DeInit(AACDecoder* dec) {}
bool AAC_Configure(AACDecoder* dec, int frequency, int channels) { return false; }
bool AAC_DecodeFrame(AACDecoder* dec, const void* input, int inputlen, void* output, int outputlen) { return false; }

bool Addon_KeyDown(KeyType type, void* userdata) { return false; }
void Addon_RumbleStart(u32 len, void* userdata) {}
void Addon_RumbleStop(void* userdata) {}
float Addon_MotionQuery(MotionQueryType type, void* userdata) { return 0.0f; }

struct DynamicLibrary {};
DynamicLibrary* DynamicLibrary_Load(const char* lib) { return nullptr; }
void DynamicLibrary_Unload(DynamicLibrary* lib) {}
void* DynamicLibrary_LoadFunction(DynamicLibrary* lib, const char* name) { return nullptr; }

} // namespace melonDS::Platform
