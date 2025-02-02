package repository

import (
	"gorm.io/gorm"
)

// UNSYNCED MODEL

type ProcessedChange struct {
	gorm.Model
	ChangeID string
}

type ProcessedChangeRepository struct {
	BaseRepository
}

func NewProcessedChangeRepository(unsyncedDb *gorm.DB) *ProcessedChangeRepository {
	return &ProcessedChangeRepository{
		BaseRepository: BaseRepository{db: unsyncedDb},
	}
}

func (r *ProcessedChangeRepository) GetProcessedChange(updateID string) *ProcessedChange {
	var ProcessedChange ProcessedChange

	err := r.db.Where("update_id = ?", updateID).
		First(&ProcessedChange).Error

	if err != nil {
		return nil
	}

	return &ProcessedChange
}

func (r *ProcessedChangeRepository) ProcessedChange(changeID string) bool {
	return r.GetProcessedChange(changeID) == nil
}

func (r *ProcessedChangeRepository) SaveProcessedChange(changeID string) error {
	item := &ProcessedChange{
		ChangeID: changeID,
	}

	return r.db.Save(item).Error
}
