package main

import "myscript/internal/repository"

// --- Cache ---

func (a *App) GetCache(key string) *repository.CacheValue {
	return repository.NewCacheRepository().
		GetCache(a.syncedDb, key)
}

func (a *App) SaveCache(key string, value interface{}) *repository.Cache {
	return repository.NewCacheRepository().
		SaveCache(a.syncedDb, key, value)
}

func (a *App) DeleteCache(key string) {
	repository.NewCacheRepository().
		DeleteCache(a.syncedDb, key)
}
