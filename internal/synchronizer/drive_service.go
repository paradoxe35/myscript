package synchronizer

const (
	DEVICES_SYNC_STATE_FILE = "devices-sync-state.json"
)

type DriveService interface {
	GetFile(fileID string)
}
