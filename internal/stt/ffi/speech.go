// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

// Package ffi binds the Rust capture and transcription library. Callbacks
// arrive on library threads and are relayed by one goroutine, so Wails is
// never touched from a foreign thread.
package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../../rust-ffi
#cgo LDFLAGS: ${SRCDIR}/../../../lib/libmyscript_stt.a
#cgo linux LDFLAGS: -lstdc++ -lasound -lpthread -ldl -lm
// Accelerate: ggml-cpu calls vDSP directly under GGML_USE_ACCELERATE, but its
// own link manifest only requests the framework when BLAS is on.
#cgo darwin LDFLAGS: -lc++ -framework Accelerate -framework AudioToolbox -framework CoreAudio -framework AudioUnit -framework CoreFoundation
#cgo windows LDFLAGS: -lstdc++ -lole32 -lavrt -lws2_32 -luserenv -lbcrypt -lntdll -static

#include <stdlib.h>
#include "bindings.h"

// cgo exports pointers without const; the casts below match the header's types.
extern void sttTextGateway(char *text);
extern void sttAudioGateway(int16_t *samples, uintptr_t len);
extern void sttLevelGateway(float rms);
extern void sttStoppedGateway(bool autoStopped);
extern void sttErrorGateway(char *message);
*/
import "C"

import (
	"fmt"
	"myscript/internal/stt"
	"sync"
	"unsafe"
)

const deliveryQueue = 256

var (
	callbacksMu sync.RWMutex
	callbacks   stt.Callbacks
	deliveries  = make(chan func(), deliveryQueue)
	deliverOnce sync.Once
)

// The Rust side orders start, stop and cancel; calls only need the handle kept alive.
type Speech struct {
	mu     sync.RWMutex
	handle C.myscript_stt_SttHandle
}

func NewSpeech() (*Speech, error) {
	deliverOnce.Do(func() { go deliver() })

	handle := C.myscript_stt_new(C.myscript_stt_Callbacks{
		text:    C.myscript_stt_TextCallback(C.sttTextGateway),
		audio:   C.myscript_stt_AudioCallback(C.sttAudioGateway),
		level:   C.myscript_stt_LevelCallback(C.sttLevelGateway),
		stopped: C.myscript_stt_StoppedCallback(C.sttStoppedGateway),
		error:   C.myscript_stt_ErrorCallback(C.sttErrorGateway),
	})
	if handle == nil {
		return nil, fmt.Errorf("failed to create the speech recogniser: %s", lastError())
	}
	return &Speech{handle: handle}, nil
}

func (s *Speech) SetCallbacks(c stt.Callbacks) {
	callbacksMu.Lock()
	defer callbacksMu.Unlock()
	callbacks = c
}

func current() stt.Callbacks {
	callbacksMu.RLock()
	defer callbacksMu.RUnlock()
	return callbacks
}

func deliver() {
	for call := range deliveries {
		call()
	}
}

//export sttTextGateway
func sttTextGateway(text *C.char) {
	if handler := current().Text; handler != nil {
		copied := C.GoString(text)
		deliveries <- func() { handler(copied) }
	}
}

//export sttAudioGateway
func sttAudioGateway(samples *C.int16_t, length C.uintptr_t) {
	if handler := current().Audio; handler != nil {
		copied := C.GoBytes(unsafe.Pointer(samples), C.int(length)*2)
		deliveries <- func() { handler(copied) }
	}
}

//export sttLevelGateway
func sttLevelGateway(rms C.float) {
	if handler := current().Level; handler != nil {
		// Dropped when the host is behind: a stale meter reading is worthless,
		// and the audio thread must never wait.
		select {
		case deliveries <- func() { handler(float32(rms)) }:
		default:
		}
	}
}

//export sttStoppedGateway
func sttStoppedGateway(autoStopped C.bool) {
	if handler := current().Stopped; handler != nil {
		auto := bool(autoStopped)
		deliveries <- func() { handler(auto) }
	}
}

//export sttErrorGateway
func sttErrorGateway(message *C.char) {
	if handler := current().Error; handler != nil {
		copied := C.GoString(message)
		deliveries <- func() { handler(copied) }
	}
}

// Blocks until the model is resident or the load failed.
func (s *Speech) Load(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	s.mu.RLock()
	result := C.myscript_stt_load(s.handle, cPath)
	s.mu.RUnlock()
	return check(result)
}

func (s *Speech) Unload() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	C.myscript_stt_unload(s.handle)
}

func (s *Speech) Start() error {
	s.mu.RLock()
	result := C.myscript_stt_start(s.handle)
	s.mu.RUnlock()
	return check(result)
}

// Blocks until the last utterance has been transcribed and delivered.
func (s *Speech) Stop() error {
	s.mu.RLock()
	result := C.myscript_stt_stop(s.handle)
	s.mu.RUnlock()
	return check(result)
}

func (s *Speech) Cancel() error {
	s.mu.RLock()
	result := C.myscript_stt_cancel(s.handle)
	s.mu.RUnlock()
	return check(result)
}

// Empty means the system default; applies to the next take.
func (s *Speech) SetDevice(name string) error {
	return s.setString(name, func(value *C.char) C.int {
		return C.myscript_stt_set_device(s.handle, value)
	})
}

// ISO code, empty to detect; applies to the next take.
func (s *Speech) SetLanguage(code string) error {
	return s.setString(code, func(value *C.char) C.int {
		return C.myscript_stt_set_language(s.handle, value)
	})
}

func (s *Speech) setString(value string, set func(*C.char) C.int) error {
	var cValue *C.char
	if value != "" {
		cValue = C.CString(value)
		defer C.free(unsafe.Pointer(cValue))
	}

	s.mu.RLock()
	result := set(cValue)
	s.mu.RUnlock()
	return check(result)
}

// Routes utterances to the audio callback instead of the model; applies to the next take.
func (s *Speech) SetCaptureOnly(enabled bool) error {
	s.mu.RLock()
	result := C.myscript_stt_set_capture_only(s.handle, C.bool(enabled))
	s.mu.RUnlock()
	return check(result)
}

func (s *Speech) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handle != nil {
		C.myscript_stt_free(s.handle)
		s.handle = nil
	}
}

func InputDevices() []stt.Device {
	listed := C.myscript_stt_devices()
	if listed == nil {
		return nil
	}
	defer C.myscript_stt_free_string(listed)
	return stt.ParseDevices(C.GoString(listed))
}

func check(result C.int) error {
	if result != 0 {
		return fmt.Errorf("%s", lastError())
	}
	return nil
}

func lastError() string {
	cErr := C.myscript_stt_get_last_error()
	if cErr == nil {
		return "unknown error"
	}
	defer C.myscript_stt_free_string((*C.char)(unsafe.Pointer(cErr)))
	return C.GoString(cErr)
}
