package synchronizer

import (
	"myscript/internal/repository"
	"strings"
)

type Synchronizer struct {
	driveService DriveService

	// Repository
	syncStateRepository       *repository.SyncStateRepository
	changeLogRepository       *repository.ChangeLogRepository
	processedChangeRepository *repository.ProcessedChangeRepository
}

// Option
type Option func(s *Synchronizer)

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
			if err := s.applyFromSnapshot(file); err != nil {
				return err
			}
		} else {
			if err := s.applyFromChangeLogs(file); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Synchronizer) applyFromSnapshot(file File) error {
	return nil
}

func (s *Synchronizer) applyFromChangeLogs(file File) error {
	return nil
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
