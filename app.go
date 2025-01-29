package main

import (
	"context"
	"myscript/internal/synchronizer"
	local_whisper "myscript/internal/transcribe/whisper/local"
	"myscript/internal/updater"
	"myscript/internal/utils"
	"myscript/internal/utils/microphone"

	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx context.Context

	syncedDb   *gorm.DB
	UnSyncedDb *gorm.DB

	audioSequencer *microphone.AudioSequencer
	lwt            *local_whisper.LocalWhisperTranscriber
	updater        *updater.Updater

	googleClient *synchronizer.GoogleClient
}

type AppOptions struct {
	SyncedDb   *gorm.DB
	UnSyncedDb *gorm.DB

	Lwt            *local_whisper.LocalWhisperTranscriber
	AudioSequencer *microphone.AudioSequencer
	Updater        *updater.Updater

	// Synchronizer
	GoogleClient *synchronizer.GoogleClient
}

// NewApp creates a new App application struct
func NewApp(options AppOptions) *App {
	return &App{
		syncedDb:       options.SyncedDb,
		UnSyncedDb:     options.UnSyncedDb,
		googleClient:   options.GoogleClient,
		audioSequencer: options.AudioSequencer,
		lwt:            options.Lwt,
		updater:        options.Updater,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetAppVersion() string {
	return AppVersion
}

func (a *App) IsDevMode() bool {
	return utils.IsDevMode()
}
