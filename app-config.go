// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import "myscript/internal/repository"

func (a *App) GetConfig() *repository.Config {
	return repository.NewConfigRepository(a.mainDB).
		GetConfig()
}

func (a *App) SaveConfig(config *repository.Config) *repository.Config {
	repository.NewConfigRepository(a.mainDB).
		SaveConfig(config)

	return a.GetConfig()
}

func (a *App) deviceSettings() *repository.DeviceSettingsRepository {
	return repository.NewDeviceSettingsRepository(a.unSyncedDB)
}

func (a *App) GetDeviceSettings() repository.DeviceSettings {
	return a.deviceSettings().Get()
}

func (a *App) SaveDeviceSettings(settings repository.DeviceSettings) repository.DeviceSettings {
	return a.deviceSettings().Save(settings)
}
