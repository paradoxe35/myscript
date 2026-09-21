// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package database

import (
	"encoding/json"
	"myscript/internal/repository"
	"path/filepath"
	"testing"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open(filepath.Join(t.TempDir(), "t.sqlite")),
		&gorm.Config{Logger: logger.Discard},
	)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// A change pulled from Drive must not look like a local edit, or it would be
// uploaded straight back with a fresh timestamp.
func TestApplyingARemoteChangeDoesNotLogALocalOne(t *testing.T) {
	mainDB := openDB(t, &repository.Page{}, &repository.Config{}, &repository.Cache{})
	unsynced := openDB(t, &repository.ChangeLog{})

	repository.SetUnSyncedDB(unsynced)
	t.Cleanup(func() { repository.SetUnSyncedDB(nil) })

	page := repository.Page{Title: "From another machine"}
	page.ID = "page-1"
	payload, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}

	err = NewDatabaseSynchronizer(nil, mainDB).SynchronizeChangeLog(repository.ChangeLog{
		ChangeID:  "pages-page-1-SAVE",
		TableName: "pages",
		RowID:     "page-1",
		Operation: repository.OPERATION_SAVE,
		NewData:   datatypes.JSON(payload),
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	var applied repository.Page
	if err := mainDB.First(&applied, "id = ?", "page-1").Error; err != nil {
		t.Fatalf("the change should have been applied: %v", err)
	}

	var logged int64
	unsynced.Model(&repository.ChangeLog{}).Count(&logged)
	if logged != 0 {
		var echo repository.ChangeLog
		unsynced.First(&echo)
		t.Errorf("applying a remote change logged %d local change(s) (%s) — it would be pushed back",
			logged, echo.ChangeID)
	}
}

// Control for the test above: an ordinary local edit must log a change.
func TestALocalEditDoesLogAChange(t *testing.T) {
	mainDB := openDB(t, &repository.Page{}, &repository.Config{}, &repository.Cache{})
	unsynced := openDB(t, &repository.ChangeLog{})

	repository.SetUnSyncedDB(unsynced)
	t.Cleanup(func() { repository.SetUnSyncedDB(nil) })

	page := repository.Page{Title: "Typed here"}
	page.ID = "page-1"
	if err := mainDB.Create(&page).Error; err != nil {
		t.Fatal(err)
	}

	var logged int64
	unsynced.Model(&repository.ChangeLog{}).Count(&logged)
	if logged != 1 {
		t.Fatalf("a local edit logged %d changes, want 1 — the hooks are not wired in this test", logged)
	}
}

// The pull reports touched rows so the cycle can drop overruled local changes;
// a delete must not wipe the list collected for the table.
func TestADeleteKeepsTheRowsAlreadyCollected(t *testing.T) {
	mainDB := openDB(t, &repository.Page{}, &repository.Config{}, &repository.Cache{})

	repository.SetUnSyncedDB(nil)
	syncer := NewDatabaseSynchronizer(nil, mainDB)

	saved := repository.Page{Title: "kept"}
	saved.ID = "page-1"
	payload, _ := json.Marshal(saved)

	if err := syncer.SynchronizeChangeLog(repository.ChangeLog{
		TableName: "pages", RowID: "page-1",
		Operation: repository.OPERATION_SAVE,
		NewData:   datatypes.JSON(payload),
	}); err != nil {
		t.Fatal(err)
	}

	if err := syncer.SynchronizeChangeLog(repository.ChangeLog{
		TableName: "pages", RowID: "page-2",
		Operation: repository.OPERATION_DELETE,
	}); err != nil {
		t.Fatal(err)
	}

	rows := syncer.GetAffectedTables()["pages"]
	if len(rows) != 2 {
		t.Fatalf("got %v, both the saved and the deleted row should be reported", rows)
	}

	seen := map[string]bool{}
	for _, row := range rows {
		seen[row] = true
	}
	if !seen["page-1"] || !seen["page-2"] {
		t.Errorf("got %v, want page-1 and page-2", rows)
	}
}
