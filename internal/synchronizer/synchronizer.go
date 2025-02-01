package synchronizer

import "myscript/internal/repository"

type Synchronizer struct {
	deviceRepository *repository.DeviceRepository
}

// Option
type Option func(s *Synchronizer)

func WithDeviceRepository(deviceRepository *repository.DeviceRepository) Option {
	return func(s *Synchronizer) {
		s.deviceRepository = deviceRepository
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

func (s *Synchronizer) StartScheduler() {}
