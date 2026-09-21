// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"context"
	"myscript/internal/google"
	"myscript/internal/stt"
	"myscript/internal/synchronizer"
	"myscript/internal/updater"
	"myscript/internal/utils"

	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx context.Context

	mainDB        *gorm.DB
	unSyncedDB    *gorm.DB
	speech        *stt.Service
	updater       *updater.Updater
	synchronizer  *Synchronizer
	aiCompletions *aiCompletions
}

type Synchronizer struct {
	sync         *synchronizer.Synchronizer
	googleClient *google.GoogleClient
}

type AppOption func(app *App)

type SynchronizerOption func(synchronizer *Synchronizer)

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

// WithSpeech wires the speech service; its listener needs the app, so the
// service is built here rather than passed in.
func WithSpeech(newEngine func() (stt.Engine, error)) AppOption {
	return func(app *App) {
		app.speech = stt.NewService(stt.NewStore(stt.ModelsDir()), newEngine, app.speechListener())
	}
}

func WithUpdater(updater *updater.Updater) AppOption {
	return func(app *App) {
		app.updater = updater
	}
}

// Synchronizer Options

func WithSynchronizer(options ...SynchronizerOption) AppOption {
	sync := &Synchronizer{}

	for _, option := range options {
		option(sync)
	}

	return func(app *App) {
		app.synchronizer = sync
	}
}

func WithGoogleClient(client *google.GoogleClient) SynchronizerOption {
	return func(synch *Synchronizer) {
		synch.googleClient = client
	}
}

func WithSync(sync *synchronizer.Synchronizer) SynchronizerOption {
	return func(synch *Synchronizer) {
		synch.sync = sync
	}
}

// NewApp creates a new App application struct
func NewApp(options ...AppOption) *App {
	app := &App{aiCompletions: newAICompletions()}

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

func (a *App) shutdown(ctx context.Context) {
	a.aiCompletions.cancelAll()
	a.synchronizer.sync.StopScheduler()
	stt.StopRefreshing()
	a.speech.Close()
}

func (a *App) GetAppVersion() string {
	return AppVersion
}

func (a *App) IsDevMode() bool {
	return utils.IsDevMode()
}
