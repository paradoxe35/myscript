// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

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
	"sync"
	"time"

	"gorm.io/gorm"
)

const MAX_CHANGE_LOGS_APPLY_FAILURES = 5
const MAX_SNAPSHOT_APPLY_FAILURES = 10
const SCHEDULER_INTERVAL = time.Second * 10

type Synchronizer struct {
	mainDB *gorm.DB

	// Repositories
	syncStateRepository          *repository.SyncStateRepository
	changeLogRepository          *repository.ChangeLogRepository
	remoteApplyFailureRepository *repository.RemoteApplyFailureRepository
	processedChangeRepository    *repository.ProcessedChangeRepository

	// mu guards everything the frontend and the scheduler goroutine both touch.
	mu            sync.Mutex
	driveService  DriveService
	stop          chan struct{}
	isSyncing     bool
	onSyncSuccess func(affectedTables database.AffectedTables)
	onSyncFailure func(err error)

	// Only the sync worker reads or writes these, and two workers never overlap.
	affectedTables          database.AffectedTables
	lastSnapshotCreatedTime *time.Time
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

func (s *Synchronizer) SetDriveService(driveService DriveService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.driveService = driveService
}

func (s *Synchronizer) SetOnSyncSuccess(onSyncSuccess func(affectedTables database.AffectedTables)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onSyncSuccess = onSyncSuccess
}

func (s *Synchronizer) SetOnSyncFailure(onSyncFailure func(err error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onSyncFailure = onSyncFailure
}

// drive is read through the mutex because the frontend can swap the service in
// while the scheduler goroutine is mid-cycle.
func (s *Synchronizer) drive() DriveService {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.driveService
}

func (s *Synchronizer) callbacks() (func(database.AffectedTables), func(error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.onSyncSuccess, s.onSyncFailure
}

// IsSyncing reports whether a sync cycle is running right now.
func (s *Synchronizer) IsSyncing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isSyncing
}

func (s *Synchronizer) StartScheduler() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.driveService == nil {
		return errors.New("drive service is not initialized")
	}

	// Replacing a running scheduler: the old goroutine sees a stop it can
	// select on. Stopping a ticker never closes its channel, so a loop ranging
	// over one would have parked here forever.
	if s.stop != nil {
		close(s.stop)
	}
	stop := make(chan struct{})
	s.stop = stop

	go s.scheduler(stop)

	slog.Debug("Synchronizer[StartScheduler]: scheduler started")
	return nil
}

func (s *Synchronizer) StopScheduler() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stop != nil {
		close(s.stop)
		s.stop = nil
	}

	slog.Debug("Synchronizer[StopScheduler]: scheduler stopped")
	return nil
}

func (s *Synchronizer) scheduler(stop chan struct{}) {
	ticker := time.NewTicker(SCHEDULER_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if !s.beginSync(stop) {
				continue
			}
			if utils.HasInternet() {
				s.schedulerWorker() // run the sync
			}
			s.endSync()
		}
	}
}

// beginSync claims the right to run one cycle. It fails when a cycle is still
// running or when this goroutine has been replaced by a newer scheduler, so a
// restart never leaves two workers writing the same database.
func (s *Synchronizer) beginSync(stop chan struct{}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isSyncing || s.stop != stop {
		return false
	}
	s.isSyncing = true
	return true
}

func (s *Synchronizer) endSync() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isSyncing = false
}

func (s *Synchronizer) schedulerWorker() {
	// Reset affected tables
	s.resetAffectedTables()

	// Apply remote changes and create snapshots
	var failure error
	if err := s.applyRemoteChanges(); err != nil {
		failure = err
	}
	// This should come after ApplyRemoteChanges
	if err := s.createDBSnapshot(); err != nil {
		failure = err
	}

	onSuccess, onFailure := s.callbacks()
	if failure != nil {
		if onFailure != nil {
			onFailure(failure)
		}
	} else if onSuccess != nil {
		onSuccess(s.affectedTables)
	}

	// Synchronize changes (this should be the last step)
	s.syncChangesLogsToDrive()
}

func (s *Synchronizer) resetAffectedTables() {
	s.affectedTables = nil
}

func (s *Synchronizer) mergeAffectedTables(affectedTables database.AffectedTables) {
	if s.affectedTables == nil {
		s.affectedTables = affectedTables
		return
	}

	for table, columns := range affectedTables {
		if _, ok := s.affectedTables[table]; !ok {
			s.affectedTables[table] = columns
		} else {
			s.affectedTables[table] = append(s.affectedTables[table], columns...)
		}

		s.affectedTables[table] = utils.UniqueStrings(s.affectedTables[table])
	}
}

func (s *Synchronizer) applyRemoteChanges() error {
	drive := s.drive()
	if drive == nil {
		return errors.New("drive service is not initialized")
	}

	timeOffset := s.syncStateRepository.GetSyncState().SyncTimeOffset
	changesFiles, err := drive.GetChangeFilesAfterTimeOffset(timeOffset)
	if err != nil {
		return err
	}

	for _, file := range changesFiles {
		if !s.canBeApplied(file) ||
			file.CreatedTime.Equal(timeOffset) ||
			file.CreatedTime.Before(timeOffset) {
			continue
		}

		// Ignore if already applied
		if isApplied := s.processedChangeRepository.ChangeProcessed(file.ID); isApplied {
			s.syncStateRepository.SaveSyncState(file.CreatedTime)
			continue
		}

		var err error
		if file.IsSnapshot {
			err = s.applyRemoteSnapshot(file)
		} else {
			err = s.applyRemoteChangeLog(file)
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

func (s *Synchronizer) canBeApplied(file *File) bool {
	return strings.HasPrefix(file.Name, DB_SNAPSHOT_PREFIX) || strings.HasPrefix(file.Name, CHANGES_FILE_PREFIX)
}

func (s *Synchronizer) canIgnoreRemoteApplyFailure(file *File, err error) bool {
	maxFailures := MAX_CHANGE_LOGS_APPLY_FAILURES
	if file.IsSnapshot {
		maxFailures = MAX_SNAPSHOT_APPLY_FAILURES
	}

	if err == nil {
		return true
	}

	failure := s.remoteApplyFailureRepository.GetRemoteApplyFailure(file.ID)
	if failure.Count > maxFailures {
		slog.Error("Synchronizer[applyRemoteChanges] Failed to apply remote changes (Ignored, too many failures)",
			"filename", file.Name, "error", err,
		)
		return true
	}

	slog.Error("Synchronizer[applyRemoteChanges] Failed to apply remote changes",
		"filename", file.Name, "error", err,
	)
	s.remoteApplyFailureRepository.SaveRemoteApplyFailure(file.ID, failure.Count+1)
	return false
}

func (s *Synchronizer) applyRemoteSnapshot(file *File) error {
	fileContent, err := s.drive().GetFileContent(file.ID)
	if err != nil {
		return err
	}

	// Create a temporary directory for decompression
	tmpDir, err := os.MkdirTemp("", "snapshot")
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to create temporary directory",
			"filename", file.Name, "error", err,
		)
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Decompress the file and get the path to the SQLite database
	dbPath, err := DecompressSnapshotFile(fileContent, tmpDir)
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to decompress snapshot file",
			"filename", file.Name, "error", err,
		)
		return err
	}

	// Mount the database
	sourceDB, err := database.MountDatabase(dbPath)
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to mount database",
			"filename", file.Name, "error", err,
		)
		return err
	}

	// Synchronize the databases
	dbSynchronizer := database.NewDatabaseSynchronizer(sourceDB, s.mainDB)
	if err := dbSynchronizer.SynchronizeAll(); err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to synchronize databases",
			"filename", file.Name, "error", err,
		)
		return err
	}

	// Save affected tables
	s.mergeAffectedTables(dbSynchronizer.GetAffectedTables())

	slog.Debug("Synchronizer[applyRemoteSnapshot] Snapshot applied successfully", "file", file.Name)

	return nil
}

func (s *Synchronizer) applyRemoteChangeLog(file *File) error {
	dbSynchronizer := database.NewDatabaseSynchronizer(nil, s.mainDB)

	return s.mainDB.Transaction(func(tx *gorm.DB) error {
		fileContent, err := s.drive().GetFileContent(file.ID)
		if err != nil {
			return err
		}

		var remoteChangeLog repository.ChangeLog
		if err := json.Unmarshal(fileContent, &remoteChangeLog); err != nil {
			slog.Error("Synchronizer[applyRemoteChangeLog] Failed to unmarshal remote change log", "error", err)
			return err
		}

		err = dbSynchronizer.SynchronizeChangeLog(remoteChangeLog)
		if err != nil {
			slog.Error("Synchronizer[applyRemoteChangeLog] Failed to synchronize change logs", "error", err)
			return err
		}

		// Save affected tables
		s.mergeAffectedTables(dbSynchronizer.GetAffectedTables())

		slog.Debug("Synchronizer[applyRemoteChangeLog] Change logs applied successfully", "file", file.Name)

		return nil
	})
}

func (s *Synchronizer) syncChangesLogsToDrive() error {
	drive := s.drive()
	if drive == nil {
		return errors.New("drive service is not initialized")
	}

	changes := s.changeLogRepository.GetUnSyncedChanges()
	if len(changes) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(changes))

	for _, change := range changes {
		go func(change repository.ChangeLog) {
			defer wg.Done()

			// Delete change log from drive
			deletedFiles, err := drive.DeleteChangeLog(change)
			if err == nil && len(deletedFiles) > 0 {
				for _, file := range deletedFiles {
					s.processedChangeRepository.DeleteProcessedChange(file.ID)
				}
			}

			// Upload change log to drive
			if file, err := drive.UploadChangeLog(change); err != nil {
				slog.Error("Synchronizer[syncChangesLogsToDrive] Failed to upload change logs", "error", err, "change", change.ID)
				return
			} else {
				s.processedChangeRepository.SaveProcessedChange(file.ID)
				s.changeLogRepository.MarkChangeLogAsSyncedIfNotChanged(change)
			}
		}(change)
	}

	wg.Wait()

	return nil
}

func (s *Synchronizer) createDBSnapshot() error {
	// We cache the last snapshot created time
	// To avoid calling GetLatestDBSnapshot too often
	if s.lastSnapshotCreatedTime != nil {
		if !utils.IsAOlderThanBByOneWeek(*s.lastSnapshotCreatedTime, time.Now()) {
			return nil
		}
	}

	dbSnapshot, err := s.drive().GetLatestDBSnapshot()
	if err != nil && !errors.Is(err, ErrSnapshotNotFound) {
		return err
	}

	if dbSnapshot != nil {
		s.lastSnapshotCreatedTime = &dbSnapshot.CreatedTime
		// The latest snapshot should older than 1 week
		if !utils.IsAOlderThanBByOneWeek(dbSnapshot.CreatedTime, time.Now()) {
			return nil
		}
	}

	// Check if there is any pending changes to be applied
	timeOffset := s.syncStateRepository.GetSyncState().SyncTimeOffset
	changesFiles, err := s.drive().GetChangeFilesAfterTimeOffset(timeOffset)
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

	newSnapshot, err := s.drive().SaveDBSnapshot(bytes.NewReader(buf.Bytes()))
	if err != nil {
		slog.Error("Synchronizer[createSnapshot]: failed to save snapshot", "error", err)
		return err
	}

	s.syncStateRepository.SaveSyncState(newSnapshot.CreatedTime)
	s.processedChangeRepository.SaveProcessedChange(newSnapshot.ID)

	slog.Debug("Synchronizer[createSnapshot]: snapshot created successfully", "file", newSnapshot.Name)

	// Prune old changes
	pruneTimeOffset := newSnapshot.CreatedTime.Add(-time.Second) // Prune one second before the snapshot
	s.clearupAfterSnapshot(pruneTimeOffset)

	return nil
}

func (s *Synchronizer) clearupAfterSnapshot(pruneTimeOffset time.Time) error {
	if err := s.drive().PruneOldChanges(pruneTimeOffset); err != nil {
		slog.Error("Synchronizer[clearupAfterSnapshot]: failed to prune old changes (retry again)", "error", err)

		// Retry again after 5 seconds
		time.Sleep(time.Second * 5)
		s.drive().PruneOldChanges(pruneTimeOffset)
	}

	// Prune old apply failures
	s.remoteApplyFailureRepository.DeleteOldRemoteApplyFailures(pruneTimeOffset)

	// Prune old processed changes
	s.processedChangeRepository.DeleteOldProcessedChanges(pruneTimeOffset)

	// prune old change logs
	s.changeLogRepository.DeleteOldChangeLogs(pruneTimeOffset)

	return nil
}
