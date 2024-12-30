package main

import (
	"embed"
	"log"
	"myscript/internal/database"
	"myscript/internal/filesystem"
	local_whisper "myscript/internal/transcribe/whisper/local"
	"myscript/internal/utils"
	"myscript/internal/utils/microphone"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed version.txt
var AppVersion string

const (
	title  = "myscript"
	width  = 1024
	height = 768
)

func main() {
	logger, err := utils.NewFileLogger(filesystem.HOME_DIR)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()

	// Create an instance of the app structure
	app := NewApp(
		database.NewDatabase(filesystem.HOME_DIR),
		microphone.NewAudioSequencer(),
		local_whisper.NewLocalWhisperTranscriber(),
	)

	// Create application with options
	iErr := wails.Run(&options.App{
		Title:     title,
		MinWidth:  width,
		MinHeight: height,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		WindowStartState: options.Maximised,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Logger:           logger,
		Bind: []interface{}{
			app,
		},
	})

	if iErr != nil {
		println("Error:", err.Error())
	}
}
