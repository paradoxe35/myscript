package repository

import (
	"gorm.io/gorm"
)

// UNSYNCED MODEL

type ProcessedChange struct {
	gorm.Model
	ChangeID string
	FileID   string
}

type ProcessedChangeRepository struct {
	BaseRepository
}

func NewProcessedChangeRepository(unsyncedDb *gorm.DB) *ProcessedChangeRepository {
	return &ProcessedChangeRepository{
		BaseRepository: BaseRepository{db: unsyncedDb},
	}
}

func (r *ProcessedChangeRepository) GetProcessedChange(changeID string, fileID string) *ProcessedChange {
	var ProcessedChange ProcessedChange

	err := r.db.
		Where("change_id = ?", changeID).
		Or("file_id = ?", fileID).
		First(&ProcessedChange).Error

	if err != nil {
		return nil
	}

	return &ProcessedChange
}

func (r *ProcessedChangeRepository) ChangeProcessed(changeID string, fileID string) bool {
	return r.GetProcessedChange(changeID, fileID) != nil
}

func (r *ProcessedChangeRepository) SaveProcessedChange(changeID string, fileID string) error {
	item := &ProcessedChange{
		ChangeID: changeID,
		FileID:   fileID,
	}

	return r.db.Save(item).Error
}
