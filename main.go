package main

import (
	"embed"
	"myscript/internal/database"
	"myscript/internal/filesystem"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

const (
	APP_DIR = ".myscript"
	title   = "myscript"
	width   = 1024
	height  = 768
)

func main() {
	// Create the directory if it doesn't exist
	homeDir := filesystem.CreateDirectoryInHome(APP_DIR)

	// Create the database
	db := database.CreateDatabase(homeDir)

	// Create an instance of the app structure
	app := NewApp(db)

	// Create application with options
	err := wails.Run(&options.App{
		Title:     title,
		MinWidth:  width,
		MinHeight: height,
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
