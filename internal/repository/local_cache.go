// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Fetched data kept for a quick first render; never synced. The value is JSON
// in a text column: a JSON column would turn a bare number into an integer.
type LocalCache struct {
	Key       string `gorm:"primaryKey"`
	Value     string
	UpdatedAt time.Time
}

type LocalCacheRepository struct {
	BaseRepository
}

func NewLocalCacheRepository(unSyncedDB *gorm.DB) *LocalCacheRepository {
	return &LocalCacheRepository{BaseRepository: BaseRepository{db: unSyncedDB}}
}

// Decodes the entry into out and reports whether one was found.
func (r *LocalCacheRepository) Get(key string, out any) bool {
	var entry LocalCache
	if err := r.db.First(&entry, "key = ?", key).Error; err != nil {
		return false
	}
	return json.Unmarshal([]byte(entry.Value), out) == nil
}

func (r *LocalCacheRepository) Set(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.db.Save(&LocalCache{Key: key, Value: string(raw)}).Error
}
