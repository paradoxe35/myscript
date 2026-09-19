package database

import (
	"myscript/internal/repository"
	"testing"
)

func TestUnsyncedMigrationSurvivesOldDuplicates(t *testing.T) {
	db := openDB(t)

	db.Exec(`CREATE TABLE processed_changes (
		id integer PRIMARY KEY AUTOINCREMENT, created_at datetime, updated_at datetime,
		deleted_at datetime, file_id text)`)
	for range 3 {
		db.Exec(`INSERT INTO processed_changes (file_id) VALUES ('file-1')`)
	}
	db.Exec(`INSERT INTO processed_changes (file_id) VALUES ('file-2')`)

	dropDuplicateFileIDs(db, "processed_changes")
	if err := db.AutoMigrate(&repository.ProcessedChange{}); err != nil {
		t.Fatalf("migration should succeed over an old database: %v", err)
	}

	var rows int64
	db.Model(&repository.ProcessedChange{}).Count(&rows)
	if rows != 2 {
		t.Errorf("kept %d rows, want one per file", rows)
	}

	repo := repository.NewProcessedChangeRepository(db)
	if err := repo.SaveProcessedChange("file-1"); err != nil {
		t.Fatal(err)
	}
	db.Model(&repository.ProcessedChange{}).Count(&rows)
	if rows != 2 {
		t.Errorf("got %d rows, re-recording a file must not add one", rows)
	}
	if !repo.ChangeProcessed("file-1") || !repo.ChangeProcessed("file-2") {
		t.Error("both files should still be recognised")
	}
}
