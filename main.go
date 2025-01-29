package main

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"myscript/internal/database"
	"myscript/internal/filesystem"
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

	appUpdater := updater.NewUpdater(REPO_OWNER, REPO_NAME, strings.TrimSpace(AppVersion))
	appUpdater.SetToken(readGitHubToken())

	// Create an instance of the app structure
	app := NewApp(
		AppOptions{
			SyncedDb:       database.NewSyncedDatabase(filesystem.HOME_DIR),
			AudioSequencer: microphone.NewAudioSequencer(),
			Lwt:            local_whisper.NewLocalWhisperTranscriber(),
			Updater:        appUpdater,
		},
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

func readGitHubToken() string {
	githubTokenFile, err := credentials.ReadFile("credentials/github-token.txt")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Error("Unexpected error opening GitHub token", "error", err)
		}
		return ""
	}

	// Remove UTF-8 BOM if present
	data := bytes.TrimPrefix(githubTokenFile, []byte{0xEF, 0xBB, 0xBF})

	// Remove UTF-16 BOM if present
	data = bytes.TrimPrefix(data, []byte{0xFF, 0xFE})

	// Handle Windows/Unix line endings
	token := strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n"))

	return token
}
