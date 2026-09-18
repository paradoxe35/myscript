// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package synchronizer

import (
	"io"
	"myscript/internal/repository"
	"runtime"
	"sync"
	"testing"
	"time"
)

type stubDrive struct {
	mu    sync.Mutex
	calls int
}

func (d *stubDrive) GetChangeFilesAfterTimeOffset(time.Time) ([]*File, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls++
	return nil, nil
}

func (d *stubDrive) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.calls
}

func (d *stubDrive) GetLatestDBSnapshot() (*File, error)                   { return nil, ErrSnapshotNotFound }
func (d *stubDrive) SaveDBSnapshot(io.ReadSeeker) (*File, error)           { return nil, nil }
func (d *stubDrive) DeleteChangeLog(repository.ChangeLog) ([]*File, error) { return nil, nil }
func (d *stubDrive) UploadChangeLog(repository.ChangeLog) (*File, error)   { return nil, nil }
func (d *stubDrive) GetFileContent(string) ([]byte, error)                 { return nil, nil }
func (d *stubDrive) PruneOldChanges(time.Time) error                       { return nil }

func TestStartSchedulerNeedsADriveService(t *testing.T) {
	if err := NewSynchronizer().StartScheduler(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestStopSchedulerEndsTheGoroutine(t *testing.T) {
	sync := NewSynchronizer()
	sync.SetDriveService(&stubDrive{})

	before := runtime.NumGoroutine()

	for range 5 {
		if err := sync.StartScheduler(); err != nil {
			t.Fatal(err)
		}
		if err := sync.StopScheduler(); err != nil {
			t.Fatal(err)
		}
	}

	if leaked := waitForGoroutines(before); leaked {
		t.Error("stopping the scheduler should not leave its goroutine parked")
	}
}

func TestRestartingReplacesTheRunningScheduler(t *testing.T) {
	sync := NewSynchronizer()
	sync.SetDriveService(&stubDrive{})

	before := runtime.NumGoroutine()

	for range 5 {
		if err := sync.StartScheduler(); err != nil {
			t.Fatal(err)
		}
	}
	sync.StopScheduler()

	if leaked := waitForGoroutines(before); leaked {
		t.Error("each restart should retire the previous scheduler")
	}
}

func TestIsSyncingIsFalseWhenIdle(t *testing.T) {
	sync := NewSynchronizer()

	if sync.IsSyncing() {
		t.Error("a synchronizer that never ran is not syncing")
	}
}

func waitForGoroutines(before int) bool {
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before {
			return false
		}
		time.Sleep(20 * time.Millisecond)
	}
	return true
}
