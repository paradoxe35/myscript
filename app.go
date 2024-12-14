package main

import (
	"context"
	"myscript/internal/repository"

	"gorm.io/gorm"
)

// App struct
type App struct {
	ctx context.Context
	db  *gorm.DB
}

// NewApp creates a new App application struct
func NewApp(db *gorm.DB) *App {
	return &App{db: db}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Get Config
func (a *App) GetConfig() *repository.Config {
	return repository.GetConfig(a.db)
}

// Save Config
func (a *App) SaveConfig(config *repository.Config) {
	repository.SaveConfig(a.db, config)
}
