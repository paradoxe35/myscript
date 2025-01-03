package main

import "myscript/internal/repository"

// --- Cache ---

func (a *App) GetCache(key string) *repository.CacheValue {
	return repository.GetCache(a.db, key)
}

func (a *App) SaveCache(key string, value interface{}) *repository.Cache {
	return repository.SaveCache(a.db, key, value)
}

func (a *App) DeleteCache(key string) {
	repository.DeleteCache(a.db, key)
}
