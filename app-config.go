package main

import "myscript/internal/repository"

// --- Config ---

func (a *App) GetConfig() *repository.Config {
	return repository.GetConfig(a.db)
}

func (a *App) SaveConfig(config *repository.Config) *repository.Config {
	repository.SaveConfig(a.db, config)
	return a.GetConfig()
}
