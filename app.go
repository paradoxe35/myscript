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

	syncedDb       *gorm.DB
	UnSyncedDb     *gorm.DB
	audioSequencer *microphone.AudioSequencer
	lwt            *local_whisper.LocalWhisperTranscriber
	updater        *updater.Updater
	synchronizer   *Synchronizer
}

type Synchronizer struct {
	googleClient *synchronizer.GoogleClient
}

type AppOption func(app *App)

type SynchronizerOption func(synchronizer *Synchronizer)

func WithSyncedDB(db *gorm.DB) AppOption {
	return func(app *App) {
		app.syncedDb = db
	}
}

func WithUnSyncedDB(db *gorm.DB) AppOption {
	return func(app *App) {
		app.UnSyncedDb = db
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

func WithGoogleClient(client *synchronizer.GoogleClient) SynchronizerOption {
	return func(sync *Synchronizer) {
		sync.googleClient = client
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
