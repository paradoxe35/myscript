package synchronizer

import (
	"errors"
	"fmt"
	"io"
	"myscript/internal/repository"
	"time"
)

const (
	DB_SNAPSHOT_PREFIX  = "snapshot_"
	CHANGES_FILE_PREFIX = "changes_"
)

type DeviceSyncStateValue struct {
	SyncTimeOffset string
}

var ErrSnapshotNotFound = errors.New("no DB snapshot found")

type DeviceSyncState map[string]*DeviceSyncStateValue

type File struct {
	ID          string
	IsSnapshot  bool
	Name        string
	CreatedTime time.Time
}

type UploadFile struct {
	Name    string
	Content interface{}
}

func snapshotFileName(timestamp time.Time) string {
	return fmt.Sprintf("%s%d.db.gz", DB_SNAPSHOT_PREFIX, timestamp.Unix())
}

func changeLogsFileName(changeLog repository.ChangeLog) string {
	return fmt.Sprintf("%s%s.json", CHANGES_FILE_PREFIX, changeLog.ChangeID)
}

type DriveService interface {
	// DB Snapshot
	GetLatestDBSnapshot() (*File, error)
	SaveDBSnapshot(content io.ReadSeeker) (*File, error)
	// Get Changes from Drive
	DeleteChangeLog(changeLog repository.ChangeLog) error
	UploadChangeLog(changeLog repository.ChangeLog) (*File, error)
	GetChangeFilesAfterTimeOffset(timeOffset time.Time) ([]*File, error)
	// Drive
	GetFileContent(fileId string) ([]byte, error)
	// Prune
	PruneOldChanges(timestamp time.Time) error
}
