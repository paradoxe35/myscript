package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseUUIDModel struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *BaseUUIDModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == "" {
		base.ID = uuid.New().String()
	}
	return nil
}
