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

type Synchronizer struct {
	driveService DriveService

	// Synced
	syncedDatabase *gorm.DB

	// Repository
	syncStateRepository       *repository.SyncStateRepository
	changeLogRepository       *repository.ChangeLogRepository
	processedChangeRepository *repository.ProcessedChangeRepository
}

// Option
type Option func(s *Synchronizer)

func WithSyncedDatabase(syncedDatabase *gorm.DB) Option {
	return func(s *Synchronizer) {
		s.syncedDatabase = syncedDatabase
	}
}

func WithChangeLogRepository(changeLogRepository *repository.ChangeLogRepository) Option {
	return func(s *Synchronizer) {
		s.changeLogRepository = changeLogRepository
	}
}

func WithProcessedChangeRepository(processedChangeRepository *repository.ProcessedChangeRepository) Option {
	return func(s *Synchronizer) {
		s.processedChangeRepository = processedChangeRepository
	}
}

func WithDriveService(driveService DriveService) Option {
	return func(s *Synchronizer) {
		s.driveService = driveService
	}
}

func WithSyncStateRepository(syncStateRepository *repository.SyncStateRepository) Option {
	return func(s *Synchronizer) {
		s.syncStateRepository = syncStateRepository
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
	if isApplied := s.processedChangeRepository.ChangeProcessed("empty", file.ID); isApplied {
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

	slog.Info("Synchronizer[applyRemoteSnapshot] Snapshot applied successfully", "file", file.Name)

	return nil
}

func (s *Synchronizer) applyRemoteChangeLogs(file File) error {
	// Synchronize change logs to target database
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

		if err := dbSynchronizer.SynchronizeChangeLogs(remoteChanges); err != nil {
			slog.Error("Synchronizer[applyRemoteChangeLogs] Failed to synchronize change logs", "error", err)
			return err
		}

		// Update the sync state
		s.syncStateRepository.SaveSyncState(file.CreatedTime)

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
