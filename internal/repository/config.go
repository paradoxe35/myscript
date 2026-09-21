// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Settings shared by every machine; per-device choices live in DeviceSettings.
type Config struct {
	gorm.Model

	// The hosted transcription service, used when the device transcribes remotely.
	RemoteProvider string `gorm:"column:remote_provider"`
	RemoteModel    string `gorm:"column:remote_model"`
	RemoteBaseURL  string `gorm:"column:remote_base_url"`

	AIProviders datatypes.JSON `gorm:"column:ai_providers"`
}

func (n *Config) AfterCreate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Config) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Config) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_DELETE)
}

type ConfigRepository struct {
	BaseRepository
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{
		BaseRepository: BaseRepository{db: db},
	}
}

func (r *ConfigRepository) GetConfig() *Config {
	var config Config
	r.db.First(&config)
	return &config
}

func (r *ConfigRepository) SaveConfig(config *Config) {
	if newConfig := r.GetConfig(); newConfig != nil {
		config.ID = newConfig.ID
	}

	r.db.Save(config)
}
