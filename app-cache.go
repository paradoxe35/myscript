// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import "myscript/internal/repository"

// --- Cache ---

func (a *App) GetCache(key string) *repository.CacheValue {
	return repository.NewCacheRepository(a.mainDB).
		GetCache(key)
}

func (a *App) SaveCache(key string, value interface{}) *repository.Cache {
	return repository.NewCacheRepository(a.mainDB).
		SaveCache(key, value)
}

func (a *App) DeleteCache(key string) {
	repository.NewCacheRepository(a.mainDB).
		DeleteCache(key)
}
