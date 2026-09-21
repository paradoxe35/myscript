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
