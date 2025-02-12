// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package database

import (
	"fmt"
	"myscript/internal/repository"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DB_BASE_NAME = "database.sqlite"

func MountDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Enable WAL mode
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}

	// Optional: Configure other PRAGMA settings for better performance
	pragmas := []string{
		"PRAGMA busy_timeout=5000;",  // Wait up to 5 seconds when database is locked
		"PRAGMA synchronous=NORMAL;", // Balance between safety and speed
		"PRAGMA cache_size=-2000;",   // Use 2MB of memory for page cache
		"PRAGMA foreign_keys=ON;",    // Enable foreign key constraints
		"PRAGMA temp_store=MEMORY;",  // Store temp tables in memory
	}

	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func NewMainDatabase(homeDir string) *gorm.DB {
	db, err := MountDatabase(filepath.Join(homeDir, DB_BASE_NAME))
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Migrate schemas
	db.AutoMigrate(&repository.Config{})
	db.AutoMigrate(&repository.Page{})
	db.AutoMigrate(&repository.Cache{})

	return db
}

func NewUnSyncedDatabase(homeDir string) *gorm.DB {
	db, err := MountDatabase(filepath.Join(homeDir, "unsynced-"+DB_BASE_NAME))
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Migrate schemas
	db.AutoMigrate(&repository.ChangeLog{})
	db.AutoMigrate(&repository.ProcessedChange{})
	db.AutoMigrate(&repository.RemoteApplyFailure{})
	db.AutoMigrate(&repository.GoogleAuthToken{})
	db.AutoMigrate(&repository.SyncState{})

	return db
}

type DatabaseInfo struct {
	Seq  int
	Name string
	File string
}

func GetSQLitePath(db *gorm.DB) (string, error) {
	var databases []DatabaseInfo

	// Execute PRAGMA query to get database list
	result := db.Raw("PRAGMA database_list;").Scan(&databases)
	if result.Error != nil {
		return "", result.Error
	}

	// Find the main database (where name is 'main')
	for _, dbInfo := range databases {
		if dbInfo.Name == "main" {
			return dbInfo.File, nil
		}
	}

	return "", fmt.Errorf("main database not found")
}
