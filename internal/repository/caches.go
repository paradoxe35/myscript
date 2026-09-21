// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Small per-page values that follow the page to every machine; Value is JSON.
type Cache struct {
	gorm.Model
	Key   string `json:"key" gorm:"uniqueIndex"`
	Value string `json:"value"`
}

func (n *Cache) AfterCreate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Cache) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_SAVE)
}

func (n *Cache) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_DELETE)
}

type CacheRepository struct {
	BaseRepository
}

func NewCacheRepository(db *gorm.DB) *CacheRepository {
	return &CacheRepository{
		BaseRepository: BaseRepository{db: db},
	}
}

func pageLanguageKey(pageID string) string {
	return "page-" + pageID + "-language"
}

func (r *CacheRepository) PageLanguage(pageID string) string {
	var language string
	r.get(pageLanguageKey(pageID), &language)
	return language
}

func (r *CacheRepository) SetPageLanguage(pageID, code string) error {
	return r.set(pageLanguageKey(pageID), code)
}

func (r *CacheRepository) get(key string, out any) {
	var cache Cache
	if err := r.db.First(&cache, "key = ?", key).Error; err != nil {
		return
	}
	json.Unmarshal([]byte(cache.Value), out)
}

// The row is reused so the change log keeps one entry per key.
func (r *CacheRepository) set(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var cache Cache
	r.db.First(&cache, "key = ?", key)

	cache.Key = key
	cache.Value = string(raw)
	return r.db.Save(&cache).Error
}
