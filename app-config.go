package main

import "myscript/internal/repository"

// --- Config ---

func (a *App) GetConfig() *repository.Config {
	return repository.NewConfigRepository(a.syncedDb).
		GetConfig()
}

func (a *App) SaveConfig(config *repository.Config) *repository.Config {
	repository.NewConfigRepository(a.syncedDb).
		SaveConfig(config)

	return a.GetConfig()
}
