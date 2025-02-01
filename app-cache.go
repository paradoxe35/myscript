package main

import "myscript/internal/repository"

// --- Cache ---

func (a *App) GetCache(key string) *repository.CacheValue {
	return repository.NewCacheRepository(a.syncedDb).
		GetCache(key)
}

func (a *App) SaveCache(key string, value interface{}) *repository.Cache {
	return repository.NewCacheRepository(a.syncedDb).
		SaveCache(key, value)
}

func (a *App) DeleteCache(key string) {
	repository.NewCacheRepository(a.syncedDb).
		DeleteCache(key)
}
