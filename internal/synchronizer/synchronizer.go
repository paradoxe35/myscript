package synchronizer

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"myscript/internal/database"
	"myscript/internal/repository"
	"myscript/internal/utils"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

const MAX_CHANGE_LOGS_APPLY_FAILURES = 5
const MAX_SNAPSHOT_APPLY_FAILURES = 10

type Synchronizer struct {
	driveService DriveService

	// Synced
	mainDB *gorm.DB

	// Repositories
	syncStateRepository          *repository.SyncStateRepository
	changeLogRepository          *repository.ChangeLogRepository
	remoteApplyFailureRepository *repository.RemoteApplyFailureRepository
	processedChangeRepository    *repository.ProcessedChangeRepository
}

// Option
type Option func(s *Synchronizer)

func WithMainDatabase(mainDB *gorm.DB) Option {
	return func(s *Synchronizer) {
		s.mainDB = mainDB
	}
}

func WithChangeLogRepository(repository *repository.ChangeLogRepository) Option {
	return func(s *Synchronizer) {
		s.changeLogRepository = repository
	}
}

func WithProcessedChangeRepository(repository *repository.ProcessedChangeRepository) Option {
	return func(s *Synchronizer) {
		s.processedChangeRepository = repository
	}
}

func WithDriveService(driveService DriveService) Option {
	return func(s *Synchronizer) {
		s.driveService = driveService
	}
}

func WithSyncStateRepository(repository *repository.SyncStateRepository) Option {
	return func(s *Synchronizer) {
		s.syncStateRepository = repository
	}
}

func WithRemoteApplyFailureRepository(repository *repository.RemoteApplyFailureRepository) Option {
	return func(s *Synchronizer) {
		s.remoteApplyFailureRepository = repository
	}
}

// Init
func NewSynchronizer(options ...Option) *Synchronizer {
	s := &Synchronizer{}

	for _, option := range options {
		option(s)
	}
	return s
}

func (s *Synchronizer) StartScheduler() error {
	return nil
}

func (s *Synchronizer) StopScheduler() error {
	return nil
}

func (s *Synchronizer) ApplyRemoteChanges() error {
	timeOffset := s.syncStateRepository.GetSyncState().SyncTimeOffset
	changesFiles, err := s.driveService.GetChangeFilesAfterTimeOffset(timeOffset)
	if err != nil {
		return err
	}

	for _, file := range changesFiles {
		if !s.canBeApplied(file) {
			continue
		}

		// Ignore if already applied
		if isApplied := s.processedChangeRepository.ChangeProcessed(file.ID); isApplied {
			return nil
		}

		var err error
		if file.IsSnapshot {
			err = s.applyRemoteSnapshot(file)
		} else {
			err = s.applyRemoteChangeLogs(file)
		}

		if !s.canIgnoreRemoteApplyFailure(file, err) {
			return err
		}

		s.remoteApplyFailureRepository.SaveRemoteApplyFailure(file.ID, 0) // Reset failure count
		s.syncStateRepository.SaveSyncState(file.CreatedTime)             // Update the sync state
		s.processedChangeRepository.SaveProcessedChange(file.ID)          // Update the processed change
	}

	return nil
}

func (s *Synchronizer) canBeApplied(file File) bool {
	return strings.HasPrefix(file.Name, DB_SNAPSHOT_PREFIX) || strings.HasSuffix(file.Name, CHANGES_FILE_PREFIX)
}

func (s *Synchronizer) canIgnoreRemoteApplyFailure(file File, err error) bool {
	maxFailures := MAX_CHANGE_LOGS_APPLY_FAILURES
	if file.IsSnapshot {
		maxFailures = MAX_SNAPSHOT_APPLY_FAILURES
	}

	if err == nil {
		return true
	}

	failure := s.remoteApplyFailureRepository.GetRemoteApplyFailure(file.ID)
	if failure.Count > maxFailures {
		slog.Error("Synchronizer[applyRemoteChanges] Failed to apply remote changes (Ignored, too many failures)", "error", err)
		return true
	}

	slog.Error("Synchronizer[applyRemoteChanges] Failed to apply remote changes", "error", err)
	s.remoteApplyFailureRepository.SaveRemoteApplyFailure(file.ID, failure.Count+1)
	return false
}

func (s *Synchronizer) applyRemoteSnapshot(file File) error {
	fileContent, err := s.driveService.GetFileContent(file.ID)
	if err != nil {
		return err
	}

	// Create a temporary directory for decompression
	tmpDir, err := os.MkdirTemp("", "snapshot")
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to create temporary directory", "error", err)
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Decompress the file and get the path to the SQLite database
	dbPath, err := DecompressSnapshotFile(fileContent, tmpDir)
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to decompress snapshot file", "error", err)
		return err
	}

	// Mount the database
	sourceDB, err := database.MountDatabase(dbPath)
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to mount database", "error", err)
		return err
	}

	// Synchronize the databases
	dbSynchronizer := database.NewDatabaseSynchronizer(sourceDB, s.mainDB)
	if err := dbSynchronizer.SynchronizeAll(); err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to synchronize databases", "error", err)
		return err
	}

	slog.Info("Synchronizer[applyRemoteSnapshot] Snapshot applied successfully", "file", file.Name)

	return nil
}

func (s *Synchronizer) applyRemoteChangeLogs(file File) error {
	dbSynchronizer := database.NewDatabaseSynchronizer(nil, s.mainDB)

	return s.mainDB.Transaction(func(tx *gorm.DB) error {
		fileContent, err := s.driveService.GetFileContent(file.ID)
		if err != nil {
			return err
		}

		var remoteChanges []repository.ChangeLog
		if err := json.Unmarshal(fileContent, &remoteChanges); err != nil {
			return err
		}

		err = dbSynchronizer.SynchronizeChangeLogs(remoteChanges)
		if err != nil {
			slog.Error("Synchronizer[applyRemoteChangeLogs] Failed to synchronize change logs", "error", err)
			return err
		}

		slog.Info("Synchronizer[applyRemoteChangeLogs] Change logs applied successfully", "file", file.Name)

		return nil
	})
}

func (s *Synchronizer) pruneOldChangeLogs() error {
	if latestSnapshot, err := s.driveService.GetLatestDBSnapshot(); err != nil {
		return err
	} else {
		return s.driveService.PruneOldChanges(latestSnapshot.CreatedTime)
	}
}

func (s *Synchronizer) syncChangesLogsToDrive() error {
	changes := s.changeLogRepository.GetUnSyncedChanges()

	if len(changes) == 0 {
		return nil
	}

	if file, err := s.driveService.UploadChangeLogs(changes); err != nil {
		return err
	} else {
		s.processedChangeRepository.SaveProcessedChange(file.ID)
		s.changeLogRepository.MarkChangeLogsAsSynced(changes)
	}

	return nil
}

func (s *Synchronizer) createDBSnapshot() error {
	dbSnapshot, err := s.driveService.GetLatestDBSnapshot()
	if err != nil && !errors.Is(err, ErrSnapshotNotFound) {
		return err
	}

	// The latest snapshot should older than 1 week
	if dbSnapshot != nil && !utils.IsAOlderThanBByOneWeek(dbSnapshot.CreatedTime, time.Now()) {
		return nil
	}

	// Check if there is any pending changes to be applied
	timeOffset := s.syncStateRepository.GetSyncState().SyncTimeOffset
	changesFiles, err := s.driveService.GetChangeFilesAfterTimeOffset(timeOffset)
	if err != nil {
		return err
	}

	// If there are pending changes, we should not create a snapshot
	if len(changesFiles) > 0 {
		return nil
	}

	// Get main database path
	dbPath, err := database.GetSQLitePath(s.mainDB)
	if err != nil {
		slog.Error("Synchronizer[createSnapshot]: failed to get SQLite path", "error", err)
		return err
	}

	buf, err := utils.NewFileArchiver(dbPath + "*").Archive()
	if err != nil {
		slog.Error("Synchronizer[createSnapshot]: failed to archive database", "error", err)
		return err
	}

	newSnapshot, err := s.driveService.SaveDBSnapshot(bytes.NewReader(buf.Bytes()))
	if err != nil {
		slog.Error("Synchronizer[createSnapshot]: failed to save snapshot", "error", err)
		return err
	}

	s.syncStateRepository.SaveSyncState(newSnapshot.CreatedTime)
	s.processedChangeRepository.SaveProcessedChange(newSnapshot.ID)

	// Prune old changes
	if err := s.driveService.PruneOldChanges(newSnapshot.CreatedTime); err != nil {
		slog.Error("Synchronizer[createSnapshot]: failed to prune old changes (retry again)", "error", err)

		// Retry again after 5 seconds
		time.Sleep(time.Second * 5)
		return s.driveService.PruneOldChanges(newSnapshot.CreatedTime)
	}

	return nil
}
