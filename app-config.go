package main

import "myscript/internal/repository"

// --- Config ---

func (a *App) GetConfig() *repository.Config {
	return repository.NewConfigRepository().
		GetConfig(a.syncedDb)
}

func (a *App) SaveConfig(config *repository.Config) *repository.Config {
	repository.NewConfigRepository().
		SaveConfig(a.syncedDb, config)

	return a.GetConfig()
}
