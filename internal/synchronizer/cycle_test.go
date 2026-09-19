// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package synchronizer

import (
	"errors"
	"myscript/internal/database"
	"sync"
	"testing"
)

func TestAFailedCycleDoesNotReportSuccess(t *testing.T) {
	syncer := NewSynchronizer()

	var succeeded, failed bool
	syncer.SetOnSyncSuccess(func(database.AffectedTables) { succeeded = true })
	syncer.SetOnSyncFailure(func(error) { failed = true })

	syncer.report(errors.New("drive unavailable"))

	if succeeded {
		t.Error("a failed cycle must not be reported as success")
	}
	if !failed {
		t.Error("the failure should reach the host")
	}
}

func TestAuthFailuresGiveUpAfterRepeatedRefusals(t *testing.T) {
	syncer := NewSynchronizer()
	syncer.SetDriveService(&stubDrive{})

	var lost int
	syncer.SetOnAuthLost(func(error) { lost++ })
	syncer.SetOnSyncFailure(func(error) {})

	revoked := errors.New("oauth2: invalid_grant: Token has been expired or revoked.")

	for i := 1; i < MAX_CONSECUTIVE_AUTH_FAILURES; i++ {
		syncer.report(revoked)
		if lost != 0 {
			t.Fatalf("gave up after %d failures, the limit is %d", i, MAX_CONSECUTIVE_AUTH_FAILURES)
		}
	}

	syncer.report(revoked)
	if lost != 1 {
		t.Fatalf("onAuthLost called %d times, want once", lost)
	}
}

func TestAnOrdinaryFailureNeverGivesUpOnTheGrant(t *testing.T) {
	syncer := NewSynchronizer()
	syncer.SetOnSyncFailure(func(error) {})

	var lost bool
	syncer.SetOnAuthLost(func(error) { lost = true })

	for range MAX_CONSECUTIVE_AUTH_FAILURES * 3 {
		syncer.report(errors.New("dial tcp: no such host"))
	}

	if lost {
		t.Error("being offline is not a reason to make the user sign in again")
	}
}

func TestASuccessfulCycleForgivesEarlierAuthFailures(t *testing.T) {
	syncer := NewSynchronizer()
	syncer.SetOnSyncFailure(func(error) {})
	syncer.SetOnSyncSuccess(func(database.AffectedTables) {})

	var lost bool
	syncer.SetOnAuthLost(func(error) { lost = true })

	revoked := errors.New("invalid_grant")

	syncer.report(revoked)
	syncer.report(revoked)
	syncer.report(nil) // a refresh worked again
	syncer.report(revoked)

	if lost {
		t.Error("the count should restart after a cycle succeeds")
	}
}

func TestReportIsSafeFromSeveralGoroutines(t *testing.T) {
	syncer := NewSynchronizer()
	syncer.SetOnSyncFailure(func(error) {})
	syncer.SetOnAuthLost(func(error) {})

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			syncer.report(errors.New("invalid_grant"))
		}()
	}
	wg.Wait()
}
