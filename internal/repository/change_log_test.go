// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

// Main and unsynced are separate files in the app; sharing one here would make
// a hook write into the transaction that triggered it.
func newChangeLogs(t *testing.T) (repo *ChangeLogRepository, mainDB, unsynced *gorm.DB) {
	t.Helper()

	mainDB = newDB(t, &Page{})
	unsynced = newDB(t, &ChangeLog{})

	SetUnSyncedDB(unsynced)
	t.Cleanup(func() { SetUnSyncedDB(nil) })

	return NewChangeLogRepository(unsynced), mainDB, unsynced
}

func addChange(t *testing.T, db *gorm.DB, table, row string, synced bool) ChangeLog {
	t.Helper()

	change := ChangeLog{
		ChangeID:  table + "-" + row + "-SAVE",
		TableName: table,
		RowID:     row,
		Operation: OPERATION_SAVE,
		Synced:    synced,
	}
	if err := db.Create(&change).Error; err != nil {
		t.Fatalf("seed change: %v", err)
	}
	return change
}

func TestInvalidateDropsLocalChangesThePullOverruled(t *testing.T) {
	repo, _, db := newChangeLogs(t)

	overruled := addChange(t, db, "pages", "a", false)
	untouched := addChange(t, db, "pages", "b", false)
	otherTable := addChange(t, db, "configs", "a", false)

	invalidated := repo.InvalidateStaleChangeLogs(map[string][]string{"pages": {"a"}})

	if invalidated != 1 {
		t.Fatalf("invalidated %d changes, want 1", invalidated)
	}

	pending := repo.GetUnSyncedChanges()
	if len(pending) != 2 {
		t.Fatalf("got %d pending changes, want 2", len(pending))
	}
	for _, change := range pending {
		if change.ID == overruled.ID {
			t.Error("the overruled change should no longer be pending")
		}
	}

	if ids := []uint{untouched.ID, otherTable.ID}; len(ids) != 2 {
		t.Fatal("unreachable")
	}
}

func TestInvalidateLeavesAlreadySyncedChangesAlone(t *testing.T) {
	repo, _, db := newChangeLogs(t)
	addChange(t, db, "pages", "a", true)

	if invalidated := repo.InvalidateStaleChangeLogs(map[string][]string{"pages": {"a"}}); invalidated != 0 {
		t.Errorf("invalidated %d, a synced change is already out of the way", invalidated)
	}
}

func TestInvalidateHandlesMoreRowsThanOneStatementCanBind(t *testing.T) {
	repo, _, db := newChangeLogs(t)

	rows := make([]string, 0, 1200)
	for i := range 1200 {
		row := string(rune('a'+i%26)) + string(rune('0'+i/26%10)) + string(rune('A'+i/260))
		rows = append(rows, row)
		addChange(t, db, "pages", row, false)
	}

	if invalidated := repo.InvalidateStaleChangeLogs(map[string][]string{"pages": rows}); invalidated != 1200 {
		t.Errorf("invalidated %d of 1200", invalidated)
	}
	if pending := repo.GetUnSyncedChanges(); len(pending) != 0 {
		t.Errorf("%d changes still pending", len(pending))
	}
}

func TestInvalidateIgnoresEmptyInput(t *testing.T) {
	repo, _, db := newChangeLogs(t)
	addChange(t, db, "pages", "a", false)

	if invalidated := repo.InvalidateStaleChangeLogs(map[string][]string{"pages": {}}); invalidated != 0 {
		t.Errorf("invalidated %d with no rows named", invalidated)
	}
	if invalidated := repo.InvalidateStaleChangeLogs(nil); invalidated != 0 {
		t.Errorf("invalidated %d with no tables named", invalidated)
	}
}

// Repeated edits to one row must collapse into a single pending change, keeping
// the original creation time so ordering against other rows stays stable.
func TestRepeatedEditsCollapseIntoOnePendingChange(t *testing.T) {
	_, mainDB, db := newChangeLogs(t)

	page := Page{Title: "first"}
	page.ID = "page-1"
	if err := mainDB.Create(&page).Error; err != nil {
		t.Fatal(err)
	}

	var first ChangeLog
	db.First(&first)

	time.Sleep(10 * time.Millisecond)

	page.Title = "second"
	if err := mainDB.Save(&page).Error; err != nil {
		t.Fatal(err)
	}

	var count int64
	db.Model(&ChangeLog{}).Count(&count)
	if count != 1 {
		t.Fatalf("got %d change rows for one row's edits, want 1", count)
	}

	var latest ChangeLog
	db.First(&latest)
	if !latest.CreatedAt.Equal(first.CreatedAt) {
		t.Error("the creation time should survive so ordering stays stable")
	}
	if !latest.UpdatedAt.After(first.UpdatedAt) {
		t.Error("the update time should move so the push order reflects the latest edit")
	}
	if latest.Synced {
		t.Error("a fresh edit is not synced")
	}
}

func TestAnEditAfterAPushMakesTheChangePendingAgain(t *testing.T) {
	repo, mainDB, db := newChangeLogs(t)

	page := Page{Title: "first"}
	page.ID = "page-1"
	mainDB.Create(&page)

	var change ChangeLog
	db.First(&change)
	repo.MarkChangeLogAsSynced(change)

	page.Title = "edited after the upload"
	mainDB.Save(&page)

	if pending := repo.GetUnSyncedChanges(); len(pending) != 1 {
		t.Fatalf("got %d pending changes, an edit after a push must be pushed again", len(pending))
	}
}

func TestPendingChangesComeBackOldestFirst(t *testing.T) {
	repo, _, db := newChangeLogs(t)

	for _, row := range []string{"a", "b", "c"} {
		addChange(t, db, "pages", row, false)
		time.Sleep(2 * time.Millisecond)
	}

	pending := repo.GetUnSyncedChanges()
	for i := 1; i < len(pending); i++ {
		if pending[i].UpdatedAt.Before(pending[i-1].UpdatedAt) {
			t.Fatal("pending changes should come back in the order they were made")
		}
	}
}
