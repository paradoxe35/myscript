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
	driveService DriveService

	// Synced
	mainDB *gorm.DB

	// Repositories
	syncStateRepository          *repository.SyncStateRepository
	changeLogRepository          *repository.ChangeLogRepository
	remoteApplyFailureRepository *repository.RemoteApplyFailureRepository
	processedChangeRepository    *repository.ProcessedChangeRepository

	// Synced
	isSyncing               bool
	schedulerTicker         *time.Ticker
	affectedTables          database.AffectedTables
	onSyncSuccess           func(affectedTables database.AffectedTables)
	onSyncFailure           func(err error)
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
	s := &Synchronizer{isSyncing: false}

	for _, option := range options {
		option(s)
	}
	return s
}

func (s *Synchronizer) SetDriveService(driveService DriveService) {
	s.driveService = driveService
}

func (s *Synchronizer) SetOnSyncSuccess(onSyncSuccess func(affectedTables database.AffectedTables)) {
	s.onSyncSuccess = onSyncSuccess
}

func (s *Synchronizer) SetOnSyncFailure(onSyncFailure func(err error)) {
	s.onSyncFailure = onSyncFailure
}

func (s *Synchronizer) StartScheduler() error {
	if s.driveService == nil {
		return errors.New("drive service is not initialized")
	}

	if s.schedulerTicker != nil {
		s.schedulerTicker.Stop()
	}

	s.schedulerTicker = time.NewTicker(SCHEDULER_INTERVAL)

	go s.scheduler()

	slog.Debug("Synchronizer[StartScheduler]: scheduler started")
	return nil
}

func (s *Synchronizer) StopScheduler() error {
	if s.schedulerTicker != nil {
		s.schedulerTicker.Stop()
	}

	slog.Error("Synchronizer[StopScheduler]: scheduler stopped")
	return nil
}

func (s *Synchronizer) scheduler() {
	if s.schedulerTicker == nil {
		return
	}

	for range s.schedulerTicker.C {
		if !s.isSyncing {
			if utils.HasInternet() {
				s.schedulerWorker() // run the sync
			}
		}
	}
}

func (s *Synchronizer) schedulerWorker() {
	s.isSyncing = true

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

	if failure != nil {
		if s.onSyncFailure != nil {
			s.onSyncFailure(failure)
		}
	} else if s.onSyncSuccess != nil {
		s.onSyncSuccess(s.affectedTables)
	}

	// Synchronize changes (this should be the last step)
	s.syncChangesLogsToDrive()

	s.isSyncing = false
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
	if s.driveService == nil {
		return errors.New("drive service is not initialized")
	}

	timeOffset := s.syncStateRepository.GetSyncState().SyncTimeOffset
	changesFiles, err := s.driveService.GetChangeFilesAfterTimeOffset(timeOffset)
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
	fileContent, err := s.driveService.GetFileContent(file.ID)
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
		fileContent, err := s.driveService.GetFileContent(file.ID)
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
			deletedFiles, err := s.driveService.DeleteChangeLog(change)
			if err == nil && len(deletedFiles) > 0 {
				for _, file := range deletedFiles {
					s.processedChangeRepository.DeleteProcessedChange(file.ID)
				}
			}

			// Upload change log to drive
			if file, err := s.driveService.UploadChangeLog(change); err != nil {
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

	dbSnapshot, err := s.driveService.GetLatestDBSnapshot()
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

	slog.Debug("Synchronizer[createSnapshot]: snapshot created successfully", "file", newSnapshot.Name)

	// Prune old changes
	pruneTimeOffset := newSnapshot.CreatedTime.Add(-time.Second) // Prune one second before the snapshot
	s.clearupAfterSnapshot(pruneTimeOffset)

	return nil
}

func (s *Synchronizer) clearupAfterSnapshot(pruneTimeOffset time.Time) error {
	if err := s.driveService.PruneOldChanges(pruneTimeOffset); err != nil {
		slog.Error("Synchronizer[clearupAfterSnapshot]: failed to prune old changes (retry again)", "error", err)

		// Retry again after 5 seconds
		time.Sleep(time.Second * 5)
		s.driveService.PruneOldChanges(pruneTimeOffset)
	}

	// Prune old apply failures
	s.remoteApplyFailureRepository.DeleteOldRemoteApplyFailures(pruneTimeOffset)

	// Prune old processed changes
	s.processedChangeRepository.DeleteOldProcessedChanges(pruneTimeOffset)

	// prune old change logs
	s.changeLogRepository.DeleteOldChangeLogs(pruneTimeOffset)

	return nil
}
