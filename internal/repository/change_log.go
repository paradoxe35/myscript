package repository

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
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

	// Check if the change already exists
	var existing ChangeLog
	if unSyncedDB.Where("change_id = ?", change.ChangeID).
		First(&existing).Error == nil {
		change.ID = existing.ID
		change.CreatedAt = existing.CreatedAt

		return unSyncedDB.Save(&change).Error
	}

	return unSyncedDB.Create(&change).Error
}

func (r *ChangeLogRepository) GetUnSyncedChanges() []ChangeLog {
	var changes []ChangeLog

	r.db.Where("synced = ?", false).Find(&changes)

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
