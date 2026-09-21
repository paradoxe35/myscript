// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import "gorm.io/gorm"

const (
	TranscriberLocal  = "local"
	TranscriberRemote = "remote"
	TranscriberWitAI  = "witai"
)

// Choices tied to this machine: which engine it transcribes with, which
// microphone it uses and which AI provider it talks to. Never synced.
type DeviceSettings struct {
	ID uint `gorm:"primaryKey" json:"-"`

	TranscriberSource string
	SpeechModelID     *string
	AIProvider        string
	MicInputDevice    string
}

// A single row, addressed by a fixed id so Save can upsert it.
const deviceSettingsID = 1

type DeviceSettingsRepository struct {
	BaseRepository
}

func NewDeviceSettingsRepository(unSyncedDB *gorm.DB) *DeviceSettingsRepository {
	return &DeviceSettingsRepository{BaseRepository: BaseRepository{db: unSyncedDB}}
}

func (r *DeviceSettingsRepository) Get() DeviceSettings {
	var settings DeviceSettings
	r.db.First(&settings, deviceSettingsID)

	if settings.TranscriberSource == "" {
		settings.TranscriberSource = TranscriberLocal
	}
	return settings
}

func (r *DeviceSettingsRepository) Save(settings DeviceSettings) DeviceSettings {
	settings.ID = deviceSettingsID
	r.db.Save(&settings)
	return r.Get()
}
