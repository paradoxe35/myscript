package synchronizer

import "time"

const (
	DEVICES_SYNC_STATE_FILE = "devices-sync-state.json"
)

type DriveService interface {
	InitDevice(deviceId string) error
	UpdateDeviceSyncTimeOffset(deviceId string, syncTimeOffset time.Time) error
}
