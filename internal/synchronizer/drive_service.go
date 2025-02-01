package synchronizer

import (
	"fmt"
	"time"
)

const (
	DEVICES_SYNC_STATE_FILE = "devices-sync-state.json"
)

type File struct {
	Name       string
	ModifiedAt time.Time
}

type DBSnapshot struct {
	Name string
	Data []byte
}

func ComposeSnapshotName(timestamp time.Time) string {
	return fmt.Sprintf("database-snapshot-%s.zip", timestamp.Format("2006-01-02T15:04:05.000Z"))
}

type DriveService interface {
	InitDevice(deviceId string) error
	UpdateDeviceSyncTimeOffset(deviceId string, syncTimeOffset time.Time) error
	// DB Snapshot
	GetLatestDBSnapshot() (*File, error)
	SaveDBSnapshot(snapshot *DBSnapshot) error
	RemoveOldDBSnapshot() error
	// Get Changes from Drive
	GetChangesFilesAfterSyncTimeOffset(syncTimeOffset time.Time) ([]File, error)
	GetChangesFilesBeforeTimeStamp(timestamp time.Time) ([]File, error)
	UploadChangesFiles(changes []File) error
	DeleteChangesFile(changes []File) error
}
