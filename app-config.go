package main

import "myscript/internal/repository"

// --- Config ---

func (a *App) GetConfig() *repository.Config {
	return repository.NewConfigRepository(a.mainDB).
		GetConfig()
}

func (a *App) SaveConfig(config *repository.Config) *repository.Config {
	repository.NewConfigRepository(a.mainDB).
		SaveConfig(config)

	return a.GetConfig()
}
