// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"embed"
	"log/slog"
	"myscript/internal/database"
	"myscript/internal/filesystem"
	"myscript/internal/google"
	"myscript/internal/repository"
	"myscript/internal/synchronizer"
	local_whisper "myscript/internal/transcribe/whisper/local"
	"myscript/internal/updater"
	"myscript/internal/utils"
	"myscript/internal/utils/microphone"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed version.txt
var AppVersion string

//go:embed all:credentials
var credentials embed.FS

const (
	REPO_OWNER = "paradoxe35"
	REPO_NAME  = "myscript"
)

const (
	title  = "myscript"
	width  = 1024
	height = 768
)

func main() {
	logger, err := utils.NewFileLogger(filesystem.HOME_DIR, utils.IsDevMode())
	if err != nil {
		panic(err)
	}

	// Set Slog as the default logger
	slog.SetDefault(logger.Slog)

	// Updater
	appUpdater := updater.NewUpdater(REPO_OWNER, REPO_NAME, strings.TrimSpace(AppVersion))
	appUpdater.SetToken(readGitHubToken())

	// Database
	mainDB := database.NewMainDatabase(filesystem.HOME_DIR)
	unSyncedDB := database.NewUnSyncedDatabase(filesystem.HOME_DIR)

	// Set global variable (used in change log repository)
	repository.SetUnSyncedDB(unSyncedDB)

	// Repositories
	googleAuthTokenRepository := repository.NewGoogleAuthTokenRepository(unSyncedDB)
	syncStateRepository := repository.NewSyncStateRepository(unSyncedDB)
	changeLogRepository := repository.NewChangeLogRepository(unSyncedDB)
	remoteApplyFailureRepository := repository.NewRemoteApplyFailureRepository(unSyncedDB)
	processedChangeRepository := repository.NewProcessedChangeRepository(unSyncedDB)

	// Google
	googleClient := google.NewGoogleClient(readGoogleCredentials(), googleAuthTokenRepository)

	// Synchronizer
	sync := synchronizer.NewSynchronizer(
		synchronizer.WithMainDatabase(mainDB),
		synchronizer.WithRemoteApplyFailureRepository(remoteApplyFailureRepository),
		synchronizer.WithSyncStateRepository(syncStateRepository),
		synchronizer.WithChangeLogRepository(changeLogRepository),
		synchronizer.WithProcessedChangeRepository(processedChangeRepository),
	)

	app := NewApp(
		WithMainDB(mainDB),
		WithUnSyncedDB(unSyncedDB),
		WithLocalWhisper(local_whisper.NewLocalWhisperTranscriber()),
		WithAudioSequencer(microphone.NewAudioSequencer()),
		WithUpdater(appUpdater),

		// Synchronizer
		WithSynchronizer(
			WithSync(sync),
			WithGoogleClient(googleClient),
		),
	)

	// Create application with options
	err = wails.Run(&options.App{
		Title:     title,
		MinWidth:  width,
		MinHeight: height,
		Logger:    logger.Logger,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		WindowStartState: options.Maximised,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
