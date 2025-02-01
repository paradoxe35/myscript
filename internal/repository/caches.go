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

// Hooks
func (n *Cache) AfterCreate(tx *gorm.DB) error {
	return logChange(tx, n, "CREATE")
}

func (n *Cache) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, "UPDATE")
}

func (n *Cache) AfterDelete(tx *gorm.DB) error {
	return logChange(tx, n, "DELETE")
}

type CacheRepository struct {
	BaseRepository
}

func NewCacheRepository() *CacheRepository {
	return &CacheRepository{}
}

func (*CacheRepository) GetCache(db *gorm.DB, key string) *CacheValue {
	var cache Cache

	db.Where("key = ?", key).First(&cache)

	return cache.Value.Data()
}

func (*CacheRepository) SaveCache(db *gorm.DB, key string, value interface{}) *Cache {
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

func (*CacheRepository) DeleteCache(db *gorm.DB, key string) {
	var cache Cache

	db.Where("key = ?", key).First(&cache)
	db.Delete(&cache)
}
