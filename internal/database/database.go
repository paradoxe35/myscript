// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package database

import (
	"fmt"
	"log/slog"
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

	// Older builds recorded one row per file per sync cycle. The unique index
	// below cannot be created over those, and AutoMigrate would fail silently.
	dropDuplicateFileIDs(db, "processed_changes")
	dropDuplicateFileIDs(db, "apply_failures")
	dropDuplicateFileIDs(db, "remote_apply_failures")

	migrate(db,
		&repository.ChangeLog{},
		&repository.ProcessedChange{},
		&repository.RemoteApplyFailure{},
		&repository.GoogleAuthToken{},
		&repository.SyncState{},
		&repository.Secret{},
	)

	return db
}

func migrate(db *gorm.DB, models ...any) {
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			slog.Error("Could not migrate a table", "model", fmt.Sprintf("%T", model), "error", err)
		}
	}
}

// dropDuplicateFileIDs keeps the most recent row per file, which is the one a
// unique index would have kept anyway.
func dropDuplicateFileIDs(db *gorm.DB, table string) {
	if !db.Migrator().HasTable(table) {
		return
	}

	statement := fmt.Sprintf(
		"DELETE FROM %s WHERE id NOT IN (SELECT MAX(id) FROM %s GROUP BY file_id)",
		table, table,
	)
	if err := db.Exec(statement).Error; err != nil {
		slog.Warn("Could not remove duplicate rows before migrating", "table", table, "error", err)
	}
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
