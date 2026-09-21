// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"myscript/internal/stt"
	"myscript/internal/transcribe/languages"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	eventModelDownloadProgress = "on-model-download-progress"
	eventModelDownloadDone     = "on-model-download-done"
	eventModelDownloadError    = "on-model-download-error"
	eventModelDownloadCancel   = "on-model-download-cancelled"
)

// SpeechModel is a catalogue entry as the settings list shows it.
type SpeechModel struct {
	ID             string
	Name           string
	Description    string
	SizeMB         float64
	Languages      []languages.Language
	LanguageDetect bool
	Streaming      bool
	Custom         bool
	// Accuracy is the catalogue score as a percentage; Speed is the estimated
	// multiple of realtime on this machine, 0 when unmeasured.
	Accuracy    int
	Speed       float64
	Fit         string
	FitLabel    string
	Downloaded  bool
	Downloading bool
	Recommended bool
	Suggested   bool
}

type MachineInfo struct {
	Cores    int
	MemoryMB int
}

func (a *App) GetMachine() MachineInfo {
	host := stt.Host()
	return MachineInfo{Cores: host.Cores, MemoryMB: host.MemoryMB}
}

type ModelDownloadEvent struct {
	ID         string
	Name       string
	Downloaded int64
	Total      int64
	Stage      string
	Error      string
}

// GetSpeechModels lists the catalogue ranked for this machine: downloaded
// first, then what runs comfortably here.
func (a *App) GetSpeechModels() []SpeechModel {
	store := a.speech.Store()
	host := stt.Host()

	models := stt.Catalogue()
	stt.RankForMachine(models, host, store.Downloaded)
	suggested, _ := stt.Suggested(models, host)

	summaries := make([]SpeechModel, 0, len(models))
	for _, model := range models {
		summaries = append(summaries, SpeechModel{
			ID:             model.ID,
			Name:           model.Name,
			Description:    model.Description,
			SizeMB:         model.SizeMB(),
			Languages:      languages.Named(model.Languages),
			LanguageDetect: model.LanguageDetect,
			Streaming:      model.Streaming,
			Custom:         strings.HasPrefix(model.ID, "custom/"),
			Accuracy:       int(math.Round(model.AccuracyScore * 100)),
			Speed:          math.Round(model.EstimatedRealtime(host)*10) / 10,
			Fit:            model.Fit(host).String(),
			FitLabel:       model.FitLabel(host),
			Downloaded:     store.Downloaded(model),
			Downloading:    store.Downloading(model),
			Recommended:    model.Recommended,
			Suggested:      model.ID == suggested.ID,
		})
	}
	return summaries
}

// DownloadSpeechModel returns at once; progress and the outcome arrive as events.
func (a *App) DownloadSpeechModel(id string) error {
	model, ok := stt.FindModel(id)
	if !ok {
		return fmt.Errorf("unknown model %q", id)
	}
	store := a.speech.Store()
	if store.Downloading(model) {
		return fmt.Errorf("%s is already downloading", model.Name)
	}

	go func() {
		err := store.Download(context.Background(), model, func(p stt.Progress) {
			runtime.EventsEmit(a.ctx, eventModelDownloadProgress, ModelDownloadEvent{
				ID: model.ID, Name: model.Name, Downloaded: p.Downloaded, Total: p.Total, Stage: string(p.Stage),
			})
		})
		switch {
		case errors.Is(err, context.Canceled):
			runtime.EventsEmit(a.ctx, eventModelDownloadCancel, ModelDownloadEvent{ID: model.ID, Name: model.Name})
			return
		case err != nil:
			slog.Error("Model download failed", "model", model.Name, "error", err)
			runtime.EventsEmit(a.ctx, eventModelDownloadError, ModelDownloadEvent{ID: model.ID, Name: model.Name, Error: err.Error()})
			return
		}
		runtime.EventsEmit(a.ctx, eventModelDownloadDone, ModelDownloadEvent{ID: model.ID, Name: model.Name})
	}()

	return nil
}

func (a *App) CancelSpeechModelDownload(id string) {
	if model, ok := stt.FindModel(id); ok {
		a.speech.Store().CancelDownload(model)
	}
}

func (a *App) DeleteSpeechModel(id string) error {
	model, ok := stt.FindModel(id)
	if !ok {
		return fmt.Errorf("unknown model %q", id)
	}
	if a.speech.Recording() {
		return fmt.Errorf("stop reading before deleting a model")
	}
	return a.speech.Store().Delete(model)
}

// RefreshSpeechModels rebuilds the catalogue now, regardless of the cache's
// age, so a user who heard about a new model need not wait for the scheduler.
func (a *App) RefreshSpeechModels() error {
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, stt.RefreshTimeout)
	defer cancel()
	return stt.Refresh(ctx)
}

// HasLegacyWhisperFiles reports ggml ".bin" models left by earlier releases.
func (a *App) HasLegacyWhisperFiles() bool {
	return len(a.speech.Store().LegacyFiles()) > 0
}

func (a *App) RemoveLegacyWhisperFiles() (int, error) {
	return a.speech.Store().RemoveLegacyFiles()
}
