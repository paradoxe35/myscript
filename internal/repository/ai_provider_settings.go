// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import "gorm.io/gorm"

// The model this machine asks a provider for, and how. The provider itself
// syncs; the choice of model does not, like the API key.
type AIProviderSettings struct {
	Name         string `gorm:"primaryKey"`
	Model        string
	Temperature  float64
	LowReasoning bool
}

type AIProviderSettingsRepository struct {
	BaseRepository
}

func NewAIProviderSettingsRepository(unSyncedDB *gorm.DB) *AIProviderSettingsRepository {
	return &AIProviderSettingsRepository{BaseRepository: BaseRepository{db: unSyncedDB}}
}

func (r *AIProviderSettingsRepository) Get(name string) AIProviderSettings {
	settings := AIProviderSettings{Name: name}
	r.db.First(&settings, "name = ?", name)
	return settings
}

func (r *AIProviderSettingsRepository) Save(settings AIProviderSettings) error {
	return r.db.Save(&settings).Error
}

func (r *AIProviderSettingsRepository) Delete(name string) error {
	return r.db.Delete(&AIProviderSettings{}, "name = ?", name).Error
}
