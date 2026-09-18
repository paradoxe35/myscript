// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.sqlite")), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newStores(t *testing.T) (main *gorm.DB, unsynced *gorm.DB) {
	t.Helper()
	return newDB(t, &Config{}), newDB(t, &Secret{})
}
