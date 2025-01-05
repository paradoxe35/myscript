package main

import (
	"fmt"
	"log/slog"
	"myscript/internal/utils"
	"myscript/internal/utils/microphone"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) StartRecording(language string, micInputDeviceID string) error {
	micDeviceID, err := utils.B64toBytes(micInputDeviceID)
	if err != nil {
		return fmt.Errorf("Invalid microphone input device")
	}

	if config := a.GetConfig(); config.TranscriberSource == "" {
		return fmt.Errorf("No transcription source has been configured.")
	}

	// If transcriber source is set to local, load the model
	if err := a.initLocalWhisperTranscriber(language); err != nil {
		return err
	}

	slog.Debug("Starting recording with language: %s", language)

	pq := utils.NewProcessQueue("transcriber-queue")

	a.audioSequencer.SetSequentializeCallback(func(buffer []byte) {
		bookId := pq.Book()

		waveBuffer, _ := a.audioSequencer.RawBytesToWAV(buffer)
		transcribed, err := a.Transcribe(waveBuffer, language)

		if err != nil {
			slog.Error("Transcription error: %s\n", err.Error())
			runtime.EventsEmit(a.ctx, "on-transcribe-error", err.Error())
			return
		}

		pq.Add(bookId, func() {
			runtime.EventsEmit(a.ctx, "on-transcribed-text", transcribed)
		})
	})

	a.audioSequencer.SetStopCallback(func(autoStopped bool) {
		runtime.EventsEmit(a.ctx, "on-recording-stopped", autoStopped)
		// Unload local whisper model, if it is loaded
		go a.lwt.Close()
	})

	return a.audioSequencer.Start(micDeviceID)
}

func (a *App) StopRecording() {
	slog.Debug("Stopping recording")

	a.audioSequencer.Stop(false)
	// Should be called after Stop()
	a.audioSequencer.SetSequentializeCallback(nil)
	a.audioSequencer.SetStopCallback(nil)
}

func (a *App) IsRecording() bool {
	return a.audioSequencer.Recording()
}

func (a *App) GetMicInputDevices() ([]microphone.MicInputDevice, error) {
	return a.audioSequencer.GetMicInputDevices()
}
