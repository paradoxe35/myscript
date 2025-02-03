package synchronizer

import (
	"encoding/json"
	"log/slog"
	"myscript/internal/database"
	"myscript/internal/repository"
	"os"
	"strings"

	"gorm.io/gorm"
)

const MAX_APPLY_FAILURES = 5

type Synchronizer struct {
	driveService DriveService

	// Synced
	syncedDatabase *gorm.DB

	// Repositories
	syncStateRepository          *repository.SyncStateRepository
	changeLogRepository          *repository.ChangeLogRepository
	remoteApplyFailureRepository *repository.RemoteApplyFailureRepository
	processedChangeRepository    *repository.ProcessedChangeRepository
}

// Option
type Option func(s *Synchronizer)

func WithSyncedDatabase(syncedDatabase *gorm.DB) Option {
	return func(s *Synchronizer) {
		s.syncedDatabase = syncedDatabase
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

func (s *Synchronizer) canBeApplied(file File) bool {
	return strings.HasPrefix(file.Name, DB_SNAPSHOT_PREFIX) || strings.HasSuffix(file.Name, CHANGES_FILE_PREFIX)
}

func (s *Synchronizer) applyRemoteChanges() error {
	timeOffset := s.syncStateRepository.GetSyncState().SyncTimeOffset
	changesFiles, err := s.driveService.GetChangeFilesAfterTimeOffset(timeOffset)
	if err != nil {
		return err
	}

	for _, file := range changesFiles {
		if !s.canBeApplied(file) {
			continue
		}

		if file.IsSnapshot {
			if err := s.applyRemoteSnapshot(file); err != nil {
				return err
			}
		} else {
			if err := s.applyRemoteChangeLogs(file); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Synchronizer) applyRemoteSnapshot(file File) error {
	// Ignore if already applied
	if isApplied := s.processedChangeRepository.ChangeProcessed(file.ID); isApplied {
		return nil
	}

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
	dbSynchronizer := database.NewDatabaseSynchronizer(sourceDB, s.syncedDatabase)
	if err := dbSynchronizer.SynchronizeAll(); err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to synchronize databases", "error", err)
		return err
	}

	// Update the sync state
	s.syncStateRepository.SaveSyncState(file.CreatedTime)
	s.processedChangeRepository.SaveProcessedChange(file.ID)

	slog.Info("Synchronizer[applyRemoteSnapshot] Snapshot applied successfully", "file", file.Name)

	return nil
}

func (s *Synchronizer) applyRemoteChangeLogs(file File) error {
	// Ignore if already applied
	if isApplied := s.processedChangeRepository.ChangeProcessed(file.ID); isApplied {
		return nil
	}

	dbSynchronizer := database.NewDatabaseSynchronizer(nil, s.syncedDatabase)

	return s.syncedDatabase.Transaction(func(tx *gorm.DB) error {
		fileContent, err := s.driveService.GetFileContent(file.ID)
		if err != nil {
			return err
		}

		var remoteChanges []repository.ChangeLog
		if err := json.Unmarshal(fileContent, &remoteChanges); err != nil {
			return err
		}

		failure := s.remoteApplyFailureRepository.GetRemoteApplyFailure(file.ID)

		err = dbSynchronizer.SynchronizeChangeLogs(remoteChanges)
		// If the failure count is greater than MAX_APPLY_FAILURES, ignore the error
		if err != nil && failure.Count <= MAX_APPLY_FAILURES {
			s.remoteApplyFailureRepository.SaveRemoteApplyFailure(file.ID, failure.Count+1)
			slog.Error("Synchronizer[applyRemoteChangeLogs] Failed to synchronize change logs", "error", err)
			return err
		} else if err != nil {
			slog.Error("Synchronizer[applyRemoteChangeLogs] Failed to synchronize change logs (ignoring)", "error", err)
		}

		// Reset the failure count
		s.remoteApplyFailureRepository.SaveRemoteApplyFailure(file.ID, 0)

		// Update the sync state
		s.syncStateRepository.SaveSyncState(file.CreatedTime)
		s.processedChangeRepository.SaveProcessedChange(file.ID)

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

func (s *Synchronizer) createDBSnapshot() error {
	return nil
}
