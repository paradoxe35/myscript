// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"myscript/internal/database"
	"myscript/internal/repository"
	"myscript/internal/synchronizer"
	"myscript/internal/utils"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) IsGoogleAuthEnabled() bool {
	return a.synchronizer.googleClient.HasCredentials()
}

func (a *App) GetGoogleAuthToken() *repository.GoogleAuthToken {
	return repository.
		NewGoogleAuthTokenRepository(a.unSyncedDB).
		GetGoogleAuthToken()
}

// DeleteGoogleAuthToken disconnects here and, best effort, on the account too,
// so a revoked build does not leave a live grant behind.
func (a *App) DeleteGoogleAuthToken() {
	a.synchronizer.sync.StopScheduler()

	if err := a.synchronizer.googleClient.Revoke(); err != nil {
		slog.Warn("Could not revoke the Google token", "error", err)
	}

	repository.
		NewGoogleAuthTokenRepository(a.unSyncedDB).
		DeleteGoogleAuthToken()
}

func (a *App) RefreshGoogleAuthToken() (*repository.GoogleAuthToken, error) {
	_, err := a.synchronizer.googleClient.GetClientFromSavedToken()
	if err != nil {
		return nil, err
	}

	return a.GetGoogleAuthToken(), nil
}

func (a *App) StopSynchronizer() {
	a.synchronizer.sync.StopScheduler()
}

// This function is being called from the frontend
func (a *App) StartSynchronizer() error {
	var httpClient *http.Client
	var err error

	if utils.HasInternet() {
		httpClient, err = a.synchronizer.googleClient.GetClientFromSavedToken()
		if err != nil {
			slog.Error("[StartSynchronizer] Error getting Google client", "error", err)
			return err
		}
	} else {
		token := a.synchronizer.googleClient.GetSavedToken()
		if token == nil {
			return errors.New("no token found")
		}

		httpClient, err = a.synchronizer.googleClient.GetClient(token.AuthToken.Data())
		if err != nil {
			return err
		}
	}

	googleDriveService, err := synchronizer.NewGoogleDriveService(httpClient)
	if err != nil {
		return err
	}

	a.synchronizer.sync.SetDriveService(googleDriveService)

	// Set on sync success callback
	a.synchronizer.sync.SetOnSyncSuccess(func(affectedTables database.AffectedTables) {
		// A restored backup can carry credentials written by an older build.
		repository.AdoptLegacyKeys(a.mainDB, a.unSyncedDB)
		runtime.EventsEmit(a.ctx, "on-sync-success", affectedTables)
	})

	// Set on sync failure callback
	a.synchronizer.sync.SetOnSyncFailure(func(err error) {
		runtime.EventsEmit(a.ctx, "on-sync-failure", err.Error())
	})

	// A grant that keeps being refused is cleared, so the UI can ask for a new
	// sign-in instead of failing every ten seconds.
	a.synchronizer.sync.SetOnAuthLost(func(err error) {
		slog.Error("Google authorization lost", "error", err)
		a.DeleteGoogleAuthToken()
		runtime.EventsEmit(a.ctx, "on-google-authorization-lost", err.Error())
	})

	return a.synchronizer.sync.StartScheduler()
}

// Just to get the affected tables binding generated
func (a *App) AffectedTablesPlaceholder() database.AffectedTables {
	return nil
}

const googleAuthPort = 43056
const googleAuthTimeout = 2 * time.Minute

// StartGoogleAuthorization opens the consent screen and returns once the
// browser has come back, or once the wait times out. Both endings close the
// same channel exactly once: two senders on an unbuffered channel would have
// parked whichever arrived second for the life of the process.
func (a *App) StartGoogleAuthorization() error {
	addr := fmt.Sprintf("http://localhost:%d", googleAuthPort)

	done := make(chan struct{})
	var finish sync.Once
	settle := func() { finish.Do(func() { close(done) }) }

	googleClient := a.synchronizer.googleClient
	tmpServer := synchronizer.NewAuthServerRedirection()

	go tmpServer.Start(googleAuthPort)
	defer tmpServer.Stop()

	timeout := time.NewTimer(googleAuthTimeout)
	defer timeout.Stop()

	go func() {
		select {
		case <-timeout.C:
			slog.Debug("Google authorization timeout")
			runtime.EventsEmit(a.ctx, "on-google-authorization-timeout")
			settle()
		case <-done:
		}
	}()

	tmpServer.Handler(func(authorizationCode string) {
		// The browser still has to render the landing page before the server
		// goes away, so the wait ends a moment after the code arrives.
		defer func() {
			time.Sleep(3 * time.Second)
			settle()
		}()

		slog.Debug("Google authorization received")

		if strings.TrimSpace(authorizationCode) == "" {
			return
		}

		if _, err := googleClient.SaveAuthToken(authorizationCode, addr); err != nil {
			slog.Error("Error saving Google authorization token", "error", err)
			runtime.EventsEmit(a.ctx, "on-google-authorization-error", err.Error())
			return
		}
	})

	authURL, err := googleClient.GenerateAuthURL(addr)
	if err != nil {
		return err
	}
	runtime.BrowserOpenURL(a.ctx, authURL)

	<-done

	return nil
}
