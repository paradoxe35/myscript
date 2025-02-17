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
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) IsGoogleAuthEnabled() bool {
	return a.googleClient.HasCredentials()
}

func (a *App) GetGoogleAuthToken() *repository.GoogleAuthToken {
	return repository.
		NewGoogleAuthTokenRepository(a.unSyncedDB).
		GetGoogleAuthToken()
}

func (a *App) DeleteGoogleAuthToken() {
	repository.
		NewGoogleAuthTokenRepository(a.unSyncedDB).
		DeleteGoogleAuthToken()

	a.synchronizer.StopScheduler()
}

func (a *App) RefreshGoogleAuthToken() (*repository.GoogleAuthToken, error) {
	_, err := a.googleClient.GetClientFromSavedToken()
	if err != nil {
		return nil, err
	}

	return a.GetGoogleAuthToken(), nil
}

func (a *App) StopSynchronizer() {
	a.synchronizer.StopScheduler()
}

// This function is being called from the frontend
func (a *App) StartSynchronizer() error {
	var httpClient *http.Client
	var err error

	if utils.HasInternet() {
		httpClient, err = a.googleClient.GetClientFromSavedToken()
		if err != nil {
			slog.Error("[StartSynchronizer] Error getting Google client", "error", err)
			return err
		}
	} else {
		token := a.googleClient.GetSavedToken()
		if token == nil {
			return errors.New("no token found")
		}

		httpClient, err = a.googleClient.GetClient(token.AuthToken.Data())
		if err != nil {
			return err
		}
	}

	googleDriveService, err := synchronizer.NewGoogleDriveService(httpClient)
	if err != nil {
		return err
	}

	a.synchronizer.SetDriveService(googleDriveService)

	// Set on sync success callback
	a.synchronizer.SetOnSyncSuccess(func(affectedTables database.AffectedTables) {
		runtime.EventsEmit(a.ctx, "on-sync-success", affectedTables)
	})

	// Set on sync failure callback
	a.synchronizer.SetOnSyncFailure(func(err error) {
		runtime.EventsEmit(a.ctx, "on-sync-failure", err.Error())
	})

	return a.synchronizer.StartScheduler()
}

// Just to get the affected tables binding generated
func (a *App) AffectedTablesPlaceholder() database.AffectedTables {
	return nil
}

func (a *App) StartGoogleAuthorization() error {
	port := 43056
	addr := fmt.Sprintf("http://localhost:%d", port)

	done := make(chan bool)
	ticker := time.NewTicker(2 * time.Minute)

	tmpServer := synchronizer.NewAuthServerRedirection()

	go tmpServer.Start(port)
	defer func() {
		ticker.Stop()
		tmpServer.Stop()
	}()

	go func() {
		for range ticker.C {
			done <- true
			slog.Debug("Google authorization timeout")
			runtime.EventsEmit(a.ctx, "on-google-authorization-timeout")
			break
		}
	}()

	tmpServer.Handler(func(authorizationCode string) {
		defer func() {
			time.Sleep(3 * time.Second)
			done <- true
		}()

		slog.Debug("Google authorization received", "code", authorizationCode)

		if strings.TrimSpace(authorizationCode) == "" {
			return
		}

		_, err := a.googleClient.SaveAuthToken(authorizationCode, addr)
		if err != nil {
			slog.Error("Error saving Google authorization token", "error", err)
			runtime.EventsEmit(a.ctx, "on-google-authorization-error", err.Error())
			return
		}
	})

	if authURL, err := a.googleClient.GenerateAuthURL(addr); err != nil {
		return err
	} else {
		runtime.BrowserOpenURL(a.ctx, authURL)
	}

	<-done

	return nil
}
