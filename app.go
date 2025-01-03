package main

import (
	"context"
	local_whisper "myscript/internal/transcribe/whisper/local"
	"myscript/internal/utils/microphone"

	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx            context.Context
	audioSequencer *microphone.AudioSequencer
	lwt            *local_whisper.LocalWhisperTranscriber
	db             *gorm.DB
}

// NewApp creates a new App application struct
func NewApp(db *gorm.DB, audioSequencer *microphone.AudioSequencer, lwt *local_whisper.LocalWhisperTranscriber) *App {
	return &App{db: db, audioSequencer: audioSequencer, lwt: lwt}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetAppVersion() string {
	return AppVersion
}
