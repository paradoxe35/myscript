package synchronizer

import "myscript/internal/repository"

type Synchronizer struct {
	driveService *DriveService

	// Repository
	deviceRepository          *repository.DeviceRepository
	changeLogRepository       *repository.ChangeLogRepository
	processedChangeRepository *repository.ProcessedChangeRepository
}

// Option
type Option func(s *Synchronizer)

func WithDeviceRepository(deviceRepository *repository.DeviceRepository) Option {
	return func(s *Synchronizer) {
		s.deviceRepository = deviceRepository
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

func WithDriveService(driveService *DriveService) Option {
	return func(s *Synchronizer) {
		s.driveService = driveService
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

func (s *Synchronizer) applyRemoteChanges() error {

	return nil
}

func (s *Synchronizer) pruneOldChangeLogs() error {
	return nil
}

func (s *Synchronizer) createDBSnapshot() error {
	return nil
}

func (s *Synchronizer) restoreFromSnapshot() error {
	return nil
}
