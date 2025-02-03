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
	return logChange(tx, n, OPERATION_CREATE)
}

func (n *Cache) AfterUpdate(tx *gorm.DB) error {
	return logChange(tx, n, OPERATION_UPDATE)
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

func (r *CacheRepository) GetCache(key string) *CacheValue {
	var cache Cache

	r.db.Where("key = ?", key).First(&cache)

	return cache.Value.Data()
}

func (r *CacheRepository) SaveCache(key string, value interface{}) *Cache {
	var cache Cache
	item := &Cache{
		Key: key,
		Value: datatypes.NewJSONType(&CacheValue{
			Value: value,
		}),
	}

	if item.Key != "" {
		r.db.Where("key = ?", item.Key).First(&cache)
	} else if item.ID != 0 {
		r.db.First(&cache, item.ID)
	}

	cache.Key = key
	cache.Value = item.Value
	r.db.Save(&cache)

	return &cache
}

func (r *CacheRepository) DeleteCache(key string) {
	var cache Cache

	r.db.Where("key = ?", key).First(&cache)
	r.db.Delete(&cache)
}
