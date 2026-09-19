// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package repository

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UNSYNCED MODEL

type ChangeLog struct {
	gorm.Model
	ChangeID  string `gorm:"uniqueIndex"`
	TableName string
	RowID     string
	Operation string
	NewData   datatypes.JSON
	Synced    bool `gorm:"default:false"`
}

type ChangeLogRepository struct {
	BaseRepository
}

func NewChangeLogRepository(unSyncedDB *gorm.DB) *ChangeLogRepository {
	return &ChangeLogRepository{
		BaseRepository: BaseRepository{db: unSyncedDB},
	}
}

var unSyncedDB *gorm.DB

func SetUnSyncedDB(db *gorm.DB) {
	unSyncedDB = db
}

func logChange(tx *gorm.DB, model interface{}, operation string) error {
	if unSyncedDB == nil {
		return nil
	}

	rowId := GetModelID(model)
	if rowId == "" {
		return nil
	}

	newData, _ := json.Marshal(model)

	change := ChangeLog{
		ChangeID:  fmt.Sprintf("%s-%s-%s", tx.Statement.Table, rowId, operation),
		TableName: tx.Statement.Table,
		RowID:     rowId,
		Operation: operation,
		NewData:   datatypes.JSON(newData),
		Synced:    false,
	}

	// One statement: reading first and saving after races another writer into
	// the unique index, and that error would surface on the user's own save.
	return unSyncedDB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "change_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"new_data":   change.NewData,
			"synced":     false,
			"updated_at": time.Now(),
			"deleted_at": nil,
		}),
	}).Create(&change).Error
}

// InvalidateStaleChangeLogs drops local changes for rows a pull just rewrote.
// Without it a local delete made before the pull would be pushed afterwards and
// erase the record everywhere, having already been overruled here.
func (r *ChangeLogRepository) InvalidateStaleChangeLogs(affected map[string][]string) int64 {
	var invalidated int64

	for table, rows := range affected {
		for _, batch := range batches(unique(rows), 400) {
			result := r.db.Model(&ChangeLog{}).
				Where("table_name = ? AND synced = ? AND row_id IN ?", table, false, batch).
				Update("synced", true)
			invalidated += result.RowsAffected
		}
	}

	if invalidated > 0 {
		slog.Info("Dropped local changes overruled by the pull", "changes", invalidated)
	}
	return invalidated
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))

	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// SQLite caps how many values one statement can bind.
func batches(values []string, size int) [][]string {
	var out [][]string
	for start := 0; start < len(values); start += size {
		end := min(start+size, len(values))
		out = append(out, values[start:end])
	}
	return out
}

func (r *ChangeLogRepository) GetUnSyncedChanges() []ChangeLog {
	var changes []ChangeLog

	r.db.Where("synced = ?", false).Order("updated_at asc, id asc").Find(&changes)

	return changes
}

func (r *ChangeLogRepository) GetChangeLogByChangeID(ID uint) *ChangeLog {
	var change ChangeLog
	r.db.Where("id = ?", ID).Find(&change)
	return &change
}

func (r *ChangeLogRepository) MarkChangeLogAsSyncedIfNotChanged(item ChangeLog) {
	changeLog := r.GetChangeLogByChangeID(item.ID)

	if changeLog.ChangeID == item.ChangeID && changeLog.UpdatedAt.After(item.UpdatedAt) {
		return
	}

	r.db.Model(&ChangeLog{}).Where("id = ?", item.ID).Update("synced", true)
}

func (r *ChangeLogRepository) MarkChangeLogAsSynced(item ChangeLog) {
	r.db.Model(&ChangeLog{}).Where("id = ?", item.ID).Update("synced", true)
}

func (r *ChangeLogRepository) DeleteOldChangeLogs(timeOffset time.Time) error {
	return r.db.Unscoped().Where("created_at < ? and synced = ?", timeOffset, true).Delete(&ChangeLog{}).Error
}
