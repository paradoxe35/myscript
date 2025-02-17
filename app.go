// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"context"
	"myscript/internal/google"
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

	mainDB         *gorm.DB
	unSyncedDB     *gorm.DB
	audioSequencer *microphone.AudioSequencer
	lwt            *local_whisper.LocalWhisperTranscriber
	updater        *updater.Updater

	// synchronizer
	synchronizer *synchronizer.Synchronizer
	googleClient *google.GoogleClient
}

type AppOption func(app *App)

func WithMainDB(db *gorm.DB) AppOption {
	return func(app *App) {
		app.mainDB = db
	}
}

func WithUnSyncedDB(db *gorm.DB) AppOption {
	return func(app *App) {
		app.unSyncedDB = db
	}
}

func WithLocalWhisper(lwt *local_whisper.LocalWhisperTranscriber) AppOption {
	return func(app *App) {
		app.lwt = lwt
	}
}

func WithAudioSequencer(sequencer *microphone.AudioSequencer) AppOption {
	return func(app *App) {
		app.audioSequencer = sequencer
	}
}

func WithUpdater(updater *updater.Updater) AppOption {
	return func(app *App) {
		app.updater = updater
	}
}

func WithGoogleClient(client *google.GoogleClient) AppOption {
	return func(app *App) {
		app.googleClient = client
	}
}

func WithSynchronizer(synchronizer *synchronizer.Synchronizer) AppOption {
	return func(app *App) {
		app.synchronizer = synchronizer
	}
}

// NewApp creates a new App application struct
func NewApp(options ...AppOption) *App {
	app := &App{}

	for _, option := range options {
		option(app)
	}

	return app
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
