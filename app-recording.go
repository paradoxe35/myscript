// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"myscript/internal/permissions"
	"myscript/internal/repository"
	"myscript/internal/stt"
	"myscript/internal/stt/ffi"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	eventTranscribedText  = "on-transcribed-text"
	eventTranscribeError  = "on-transcribe-error"
	eventRecordingStopped = "on-recording-stopped"
	eventTranscriberState = "on-transcriber-state"
	eventMicLevel         = "on-mic-level"
)

type TranscriberState struct {
	State     stt.State
	ModelName string
}

func (a *App) speechListener() stt.Listener {
	return stt.Listener{
		State: func(state stt.State, model stt.Model) {
			runtime.EventsEmit(a.ctx, eventTranscriberState, TranscriberState{State: state, ModelName: model.Name})
		},
		Text: func(text string) {
			runtime.EventsEmit(a.ctx, eventTranscribedText, text)
		},
		Error: func(message string) {
			slog.Error("Transcription error", "error", message)
			runtime.EventsEmit(a.ctx, eventTranscribeError, message)
		},
		Level: func(rms float32) {
			runtime.EventsEmit(a.ctx, eventMicLevel, rms)
		},
		Stopped: func(auto bool) {
			runtime.EventsEmit(a.ctx, eventRecordingStopped, auto)
		},
	}
}

// Returns once accepted; state changes and failures arrive as events.
func (a *App) StartRecording(language string, micInputDevice string) error {
	if status := permissions.RequestMicrophone(); status != permissions.MicrophoneGranted {
		return errMicrophone(status)
	}

	config := a.GetConfig()
	if config.TranscriberSource == "" {
		return fmt.Errorf("No transcription source has been configured.")
	}

	opts := stt.Options{Language: language, Device: micInputDevice}
	if config.TranscriberSource == "local" {
		model, err := a.selectedSpeechModel(config)
		if err != nil {
			return err
		}
		opts.ModelID = model.ID
	} else {
		remote, err := a.remoteTranscriber(config.TranscriberSource)
		if err != nil {
			return err
		}
		opts.Remote = remote
	}

	slog.Debug("Starting recording", "language", language, "source", config.TranscriberSource)

	go func() {
		err := a.speech.Start(opts)
		if err != nil && !errors.Is(err, stt.ErrCancelled) {
			slog.Error("Could not start recording", "error", err)
			runtime.EventsEmit(a.ctx, eventTranscribeError, err.Error())
		}
	}()

	return nil
}

// Falls back to the best downloaded model and remembers it, so a first read
// works without a trip to Settings.
func (a *App) selectedSpeechModel(config *repository.Config) (stt.Model, error) {
	store := a.speech.Store()
	if config.SpeechModelID != nil {
		if model, ok := stt.FindModel(*config.SpeechModelID); ok && store.Downloaded(model) {
			return model, nil
		}
	}

	best, ok := stt.Recommended(stt.Catalogue(), stt.Host(), store.Downloaded)
	if !ok || !store.Downloaded(best) {
		return stt.Model{}, stt.ErrNoModel
	}
	config.SpeechModelID = &best.ID
	a.SaveConfig(config)
	return best, nil
}

func (a *App) StopRecording() error {
	slog.Debug("Stopping recording")
	return a.speech.Stop()
}

// Ends the take and drops undelivered text.
func (a *App) CancelRecording() error {
	slog.Debug("Cancelling recording")
	return a.speech.Cancel()
}

func (a *App) IsRecording() bool {
	return a.speech.Recording()
}

func (a *App) GetMicInputDevices() []stt.Device {
	return ffi.InputDevices()
}

func errMicrophone(status permissions.MicrophoneStatus) error {
	if status == permissions.MicrophoneRestricted {
		return errors.New("Microphone access is restricted on this Mac.")
	}
	return errors.New("Microphone access was denied. Allow MyScript in System Settings, Privacy & Security, Microphone.")
}

// RequestMicrophoneAccess shows the system prompt when undecided and reports
// the outcome, so the window can offer the settings pane on a refusal.
func (a *App) RequestMicrophoneAccess() string {
	return string(permissions.RequestMicrophone())
}

func (a *App) OpenMicrophoneSettings() {
	permissions.OpenMicrophoneSettings()
}
