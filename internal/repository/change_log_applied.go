package repository

import (
	"gorm.io/gorm"
)

// UNSYNCED MODEL

type ChangeLogApplied struct {
	gorm.Model
	UpdateID string
}

type ChangeLogAppliedRepository struct {
	BaseRepository
}

func NewChangeLogAppliedRepository(unsyncedDb *gorm.DB) *ChangeLogAppliedRepository {
	return &ChangeLogAppliedRepository{
		BaseRepository: BaseRepository{db: unsyncedDb},
	}
}

func (r *ChangeLogAppliedRepository) GetChangeLogApplied(updateID string) *ChangeLogApplied {
	var changeLogApplied ChangeLogApplied

	err := r.db.Where("update_id = ?", updateID).
		First(&changeLogApplied).Error

	if err != nil {
		return nil
	}

	return &changeLogApplied
}

func (r *ChangeLogAppliedRepository) ChangeLogApplied(updateID string) bool {
	return r.GetChangeLogApplied(updateID) == nil
}

func (r *ChangeLogAppliedRepository) SaveChangeLogApplied(updateID string) error {
	item := &ChangeLogApplied{
		UpdateID: updateID,
	}

	return r.db.Save(item).Error
}
