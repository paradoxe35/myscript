// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package synchronizer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"myscript/internal/database"
	"myscript/internal/google"
	"myscript/internal/repository"
	"myscript/internal/utils"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const MAX_CHANGE_LOGS_APPLY_FAILURES = 5

// Drive rate limits a burst; a grant refused this many times in a row will
// not recover on the next tick.
const PUSH_CONCURRENCY = 8
const MAX_CONSECUTIVE_AUTH_FAILURES = 3
const MAX_SNAPSHOT_APPLY_FAILURES = 10
const SCHEDULER_INTERVAL = time.Second * 10

type Synchronizer struct {
	mainDB *gorm.DB

	syncStateRepository          *repository.SyncStateRepository
	changeLogRepository          *repository.ChangeLogRepository
	remoteApplyFailureRepository *repository.RemoteApplyFailureRepository
	processedChangeRepository    *repository.ProcessedChangeRepository

	// mu guards everything the frontend and the scheduler goroutine both touch.
	mu            sync.Mutex
	driveService  DriveService
	stop          chan struct{}
	isSyncing     bool
	authFailures  int
	onSyncSuccess func(affectedTables database.AffectedTables)
	onSyncFailure func(err error)
	onAuthLost    func(err error)

	// Only the sync worker reads or writes these, and two workers never overlap.
	affectedTables          database.AffectedTables
	lastSnapshotCreatedTime *time.Time
}

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

// Called once the grant has been refused often enough that only signing in
// again will help.
func (s *Synchronizer) SetOnAuthLost(onAuthLost func(err error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onAuthLost = onAuthLost
}

// The frontend can swap the service in while the scheduler is mid-cycle.
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

	// A replaced scheduler sees a stop it can select on; a stopped ticker never
	// closes its channel.
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
				s.schedulerWorker()
			}
			s.endSync()
		}
	}
}

// Fails when a cycle is running or this goroutine was replaced by a newer
// scheduler, so two workers never write the same database.
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
	s.resetAffectedTables()

	var failure error
	if err := s.applyRemoteChanges(); err != nil {
		failure = err
	}

	// A local change to a row the pull rewrote is already overruled; pushing it
	// would undo the pull elsewhere.
	if len(s.affectedTables) > 0 {
		s.changeLogRepository.InvalidateStaleChangeLogs(s.affectedTables)
	}
	// After the pull, so the snapshot captures what was just applied.
	if err := s.createDBSnapshot(); err != nil {
		failure = err
	}

	// Pushing on top of a failed pull could send a stale delete back, so the
	// push waits for a clean cycle.
	if failure == nil {
		if err := s.syncChangesLogsToDrive(); err != nil {
			failure = err
		}
	}

	s.report(failure)
}

// Gives up on a grant that keeps being refused rather than retrying forever.
func (s *Synchronizer) report(failure error) {
	onSuccess, onFailure := s.callbacks()

	if failure == nil {
		s.mu.Lock()
		s.authFailures = 0
		s.mu.Unlock()

		if onSuccess != nil {
			onSuccess(s.affectedTables)
		}
		return
	}

	slog.Error("Synchronizer: sync cycle failed", "error", failure)
	if onFailure != nil {
		onFailure(failure)
	}

	if !google.IsAuthError(failure) {
		return
	}

	s.mu.Lock()
	s.authFailures++
	lost := s.authFailures >= MAX_CONSECUTIVE_AUTH_FAILURES
	onAuthLost := s.onAuthLost
	s.mu.Unlock()

	if lost {
		slog.Error("Synchronizer: giving up after repeated authentication failures")
		s.StopScheduler()
		if onAuthLost != nil {
			onAuthLost(failure)
		}
	}
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

		s.remoteApplyFailureRepository.SaveRemoteApplyFailure(file.ID, 0)
		s.syncStateRepository.SaveSyncState(file.CreatedTime)
		s.processedChangeRepository.SaveProcessedChange(file.ID)
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

	tmpDir, err := os.MkdirTemp("", "snapshot")
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to create temporary directory",
			"filename", file.Name, "error", err,
		)
		return err
	}
	defer os.RemoveAll(tmpDir)

	dbPath, err := DecompressSnapshotFile(fileContent, tmpDir)
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to decompress snapshot file",
			"filename", file.Name, "error", err,
		)
		return err
	}

	sourceDB, err := database.MountDatabase(dbPath)
	if err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to mount database",
			"filename", file.Name, "error", err,
		)
		return err
	}

	dbSynchronizer := database.NewDatabaseSynchronizer(sourceDB, s.mainDB)
	if err := dbSynchronizer.SynchronizeAll(); err != nil {
		slog.Error("Synchronizer[applyRemoteSnapshot] Failed to synchronize databases",
			"filename", file.Name, "error", err,
		)
		return err
	}

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

		s.mergeAffectedTables(dbSynchronizer.GetAffectedTables())

		slog.Debug("Synchronizer[applyRemoteChangeLog] Change logs applied successfully", "file", file.Name)

		return nil
	})
}

// Changes to one row go up in order; different rows go up together.
func (s *Synchronizer) syncChangesLogsToDrive() error {
	drive := s.drive()
	if drive == nil {
		return errors.New("drive service is not initialized")
	}

	changes := s.changeLogRepository.GetUnSyncedChanges()
	if len(changes) == 0 {
		return nil
	}

	groups := groupChangesByRow(changes)
	slog.Debug("Synchronizer: pushing local changes", "changes", len(changes), "rows", len(groups))

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		failures int
		slots    = make(chan struct{}, PUSH_CONCURRENCY)
	)

	for _, group := range groups {
		wg.Add(1)
		slots <- struct{}{}

		go func(group []repository.ChangeLog) {
			defer wg.Done()
			defer func() { <-slots }()

			for _, change := range group {
				if err := s.pushChangeLog(drive, change); err != nil {
					slog.Error("Synchronizer: could not upload a change",
						"change", change.ChangeID, "error", err)
					mu.Lock()
					failures++
					mu.Unlock()
					// The rest of this row would land out of order.
					return
				}
			}
		}(group)
	}

	wg.Wait()

	if failures > 0 {
		return fmt.Errorf("could not upload %d of %d local changes", failures, len(groups))
	}
	return nil
}

func (s *Synchronizer) pushChangeLog(drive DriveService, change repository.ChangeLog) error {
	// Best effort: a stale copy left behind is applied and then superseded,
	// while failing here would block the upload.
	if deleted, err := drive.DeleteChangeLog(change); err == nil {
		for _, file := range deleted {
			s.processedChangeRepository.DeleteProcessedChange(file.ID)
		}
	} else {
		slog.Warn("Synchronizer: could not remove the previous copy of a change",
			"change", change.ChangeID, "error", err)
	}

	file, err := drive.UploadChangeLog(change)
	if err != nil {
		return err
	}

	s.processedChangeRepository.SaveProcessedChange(file.ID)
	s.changeLogRepository.MarkChangeLogAsSyncedIfNotChanged(change)
	return nil
}

// One row's history stays together and in order, so an update never overtakes
// the insert it depends on.
func groupChangesByRow(changes []repository.ChangeLog) [][]repository.ChangeLog {
	ordered := append([]repository.ChangeLog(nil), changes...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := changeOrder(ordered[i]), changeOrder(ordered[j])
		if left.Equal(right) {
			return ordered[i].ID < ordered[j].ID
		}
		return left.Before(right)
	})

	groups := make([][]repository.ChangeLog, 0, len(ordered))
	index := make(map[string]int, len(ordered))

	for _, change := range ordered {
		key := change.TableName + ":" + change.RowID
		if at, ok := index[key]; ok {
			groups[at] = append(groups[at], change)
			continue
		}
		index[key] = len(groups)
		groups = append(groups, []repository.ChangeLog{change})
	}

	return groups
}

func changeOrder(change repository.ChangeLog) time.Time {
	if !change.UpdatedAt.IsZero() {
		return change.UpdatedAt
	}
	return change.CreatedAt
}

func (s *Synchronizer) createDBSnapshot() error {
	// Cached to avoid calling GetLatestDBSnapshot every cycle.
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
		if !utils.IsAOlderThanBByOneWeek(dbSnapshot.CreatedTime, time.Now()) {
			return nil
		}
	}

	timeOffset := s.syncStateRepository.GetSyncState().SyncTimeOffset
	changesFiles, err := s.drive().GetChangeFilesAfterTimeOffset(timeOffset)
	if err != nil {
		return err
	}

	// Never snapshot over pending changes.
	if len(changesFiles) > 0 {
		return nil
	}

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

	pruneTimeOffset := newSnapshot.CreatedTime.Add(-time.Second)
	s.clearupAfterSnapshot(pruneTimeOffset)

	return nil
}

func (s *Synchronizer) clearupAfterSnapshot(pruneTimeOffset time.Time) error {
	if err := s.drive().PruneOldChanges(pruneTimeOffset); err != nil {
		slog.Error("Synchronizer[clearupAfterSnapshot]: failed to prune old changes (retry again)", "error", err)

		time.Sleep(time.Second * 5)
		s.drive().PruneOldChanges(pruneTimeOffset)
	}

	s.remoteApplyFailureRepository.DeleteOldRemoteApplyFailures(pruneTimeOffset)

	s.processedChangeRepository.DeleteOldProcessedChanges(pruneTimeOffset)

	s.changeLogRepository.DeleteOldChangeLogs(pruneTimeOffset)

	return nil
}
