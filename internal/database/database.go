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

	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}

	pragmas := []string{
		"PRAGMA busy_timeout=5000;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA cache_size=-2000;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA temp_store=MEMORY;",
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

	synced := []any{&repository.Config{}, &repository.Page{}, &repository.Cache{}}
	migrate(db, synced...)
	pruneColumns(db, synced...)
	// Only the page language still lives here; other keys are per-device now.
	db.Unscoped().Where("key NOT LIKE ?", "page-%-language").Delete(&repository.Cache{})

	return db
}

func NewUnSyncedDatabase(homeDir string) *gorm.DB {
	db, err := MountDatabase(filepath.Join(homeDir, "unsynced-"+DB_BASE_NAME))
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Duplicate rows would make the unique index below fail silently in AutoMigrate.
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
		&repository.DeviceSettings{},
		&repository.AIProviderSettings{},
		&repository.PageState{},
		&repository.LocalCache{},
	)

	return db
}

// Drops columns the models no longer declare. A snapshot only applies where
// every column of it exists, so the synced schema has to match a fresh install.
func pruneColumns(db *gorm.DB, models ...any) {
	for _, model := range models {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(model); err != nil {
			continue
		}
		declared := make(map[string]bool, len(stmt.Schema.DBNames))
		for _, name := range stmt.Schema.DBNames {
			declared[name] = true
		}

		columns, err := db.Migrator().ColumnTypes(model)
		if err != nil {
			continue
		}
		for _, column := range columns {
			if declared[column.Name()] {
				continue
			}
			// GORM's SQLite DropColumn ignores columns added by ALTER TABLE.
			drop := fmt.Sprintf("ALTER TABLE %q DROP COLUMN %q", stmt.Schema.Table, column.Name())
			if err := db.Exec(drop).Error; err != nil {
				slog.Warn("Could not drop a stale column", "table", stmt.Schema.Table, "column", column.Name(), "error", err)
			}
		}
	}
}

func migrate(db *gorm.DB, models ...any) {
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			slog.Error("Could not migrate a table", "model", fmt.Sprintf("%T", model), "error", err)
		}
	}
}

// Keeps the most recent row per file, as the unique index would.
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

	result := db.Raw("PRAGMA database_list;").Scan(&databases)
	if result.Error != nil {
		return "", result.Error
	}

	for _, dbInfo := range databases {
		if dbInfo.Name == "main" {
			return dbInfo.File, nil
		}
	}

	return "", fmt.Errorf("main database not found")
}
