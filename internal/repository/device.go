package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UNSYNCED MODEL

type Device struct {
	gorm.Model
	DeviceID string
}

type DeviceRepository struct {
	BaseRepository
}

func NewDeviceRepository(unsyncedDb *gorm.DB) *DeviceRepository {
	return &DeviceRepository{
		BaseRepository: BaseRepository{db: unsyncedDb},
	}
}

func (r *DeviceRepository) GetDeviceID() string {
	device := &Device{}

	if r.db.First(device).Error != nil {
		device.DeviceID = uuid.NewString()
		r.db.Create(device)
	}

	return device.DeviceID
}
