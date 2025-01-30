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

func GetDeviceID(db *gorm.DB) string {
	device := &Device{}

	if db.First(device).Error != nil {
		device.DeviceID = uuid.NewString()
		db.Create(device)
	}

	return device.DeviceID
}
