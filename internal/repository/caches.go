package repository

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// !SYNCED MODEL

type CacheValue struct {
	Value interface{} `json:"value"`
}

type Cache struct {
	gorm.Model
	Key   string                          `json:"key" gorm:"uniqueIndex"`
	Value datatypes.JSONType[*CacheValue] `json:"value"`
}

func GetCache(db *gorm.DB, key string) *CacheValue {
	var cache Cache

	db.Where("key = ?", key).First(&cache)

	return cache.Value.Data()
}

func SaveCache(db *gorm.DB, key string, value interface{}) *Cache {
	var cache Cache
	item := &Cache{
		Key: key,
		Value: datatypes.NewJSONType(&CacheValue{
			Value: value,
		}),
	}

	if item.Key != "" {
		db.Where("key = ?", item.Key).First(&cache)
	} else if item.ID != 0 {
		db.First(&cache, item.ID)
	}

	cache.Key = key
	cache.Value = item.Value
	db.Save(&cache)

	return &cache
}

func DeleteCache(db *gorm.DB, key string) {
	var cache Cache

	db.Where("key = ?", key).First(&cache)
	db.Delete(&cache)
}
