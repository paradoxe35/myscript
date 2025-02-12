// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

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
