package synchronizer

import (
	"fmt"
	"io"
	"myscript/internal/repository"
	"time"
)

const (
	DEVICES_SYNC_STATE_FILE = "devices-sync-state.json"
	DB_SNAPSHOT_PREFIX      = "snapshot_"
	CHANGES_FILE_PREFIX     = "changes_"
)

type DeviceSyncStateValue struct {
	SyncTimeOffset string
}

type DeviceSyncState map[string]*DeviceSyncStateValue

type File struct {
	ID         string
	IsSnapshot bool
	Name       string
	ModifiedAt time.Time
}

type UploadFile struct {
	Name    string
	Content interface{}
}

func snapshotFileName(timestamp time.Time) string {
	return fmt.Sprintf("%s%d.db.gz", DB_SNAPSHOT_PREFIX, timestamp.Unix())
}

func changeLogsFileName(timestamp time.Time) string {
	return fmt.Sprintf("%s%d.json", CHANGES_FILE_PREFIX, timestamp.Unix())
}

type DriveService interface {
	InitDevice(deviceId string) error
	UpdateDeviceSyncTimeOffset(deviceId string, syncTimeOffset time.Time) error
	// DB Snapshot
	GetLatestDBSnapshot() (*File, error)
	SaveDBSnapshot(content io.ReadSeeker) (*File, error)
	// Get Changes from Drive
	UploadChangeLogs(changes []repository.ChangeLog) (*File, error)
	GetFilesAfterTimeOffset(timeOffset time.Time) ([]File, error)
	// Drive
	GetFileContent(fileId string) ([]byte, error)
	// Prune
	PruneOldChanges(timestamp time.Time) error
}
