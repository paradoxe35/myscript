// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package database

import (
	"myscript/internal/repository"
	"path/filepath"
	"testing"
)

func TestPruneColumnsDropsWhatTheModelNoLongerDeclares(t *testing.T) {
	db, err := MountDatabase(filepath.Join(t.TempDir(), "main.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	migrate(db, &repository.Config{}, &repository.Page{})
	db.Exec("ALTER TABLE configs ADD COLUMN transcriber_source text")
	db.Exec("ALTER TABLE pages ADD COLUMN expanded numeric")

	pruneColumns(db, &repository.Config{}, &repository.Page{})

	if db.Migrator().HasColumn(&repository.Config{}, "transcriber_source") {
		t.Error("configs.transcriber_source should be gone")
	}
	if db.Migrator().HasColumn(&repository.Page{}, "expanded") {
		t.Error("pages.expanded should be gone")
	}
	for _, column := range []string{"remote_provider", "ai_providers"} {
		if !db.Migrator().HasColumn(&repository.Config{}, column) {
			t.Errorf("configs.%s should survive", column)
		}
	}
}

func TestSnapshotsWithColumnsTheModelDroppedStillApply(t *testing.T) {
	dir := t.TempDir()
	snapshot, err := MountDatabase(filepath.Join(dir, "snapshot.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	local, err := MountDatabase(filepath.Join(dir, "local.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	synced := []any{&repository.Config{}, &repository.Page{}, &repository.Cache{}}
	migrate(snapshot, synced...)
	migrate(local, synced...)
	snapshot.Exec("ALTER TABLE pages ADD COLUMN expanded numeric")
	snapshot.Create(&repository.Page{Title: "kept", HtmlContent: "<p>body</p>"})

	if err := NewDatabaseSynchronizer(snapshot, local).SynchronizeAll(); err != nil {
		t.Fatalf("snapshot with an extra column should apply: %v", err)
	}

	var page repository.Page
	local.First(&page, "title = ?", "kept")
	if page.HtmlContent != "<p>body</p>" {
		t.Errorf("page was not restored, got %+v", page)
	}

	snapshot.Exec("ALTER TABLE pages ADD COLUMN clash text")
	local.Exec("ALTER TABLE pages ADD COLUMN clash integer")
	if err := NewDatabaseSynchronizer(snapshot, local).SynchronizeAll(); err == nil {
		t.Error("a shared column with a different type should still be rejected")
	}
}
