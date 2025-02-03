package main

import (
	"fmt"
	"log/slog"
	"myscript/internal/repository"
	"myscript/internal/synchronizer"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) IsGoogleAuthEnabled() bool {
	return a.synchronizer.googleClient.HasCredentials()
}

func (a *App) GetGoogleAuthToken() *repository.GoogleAuthToken {
	return repository.
		NewGoogleAuthTokenRepository(a.UnSyncedDB).
		GetGoogleAuthToken()
}

func (a *App) DeleteGoogleAuthToken() {
	repository.
		NewGoogleAuthTokenRepository(a.UnSyncedDB).
		DeleteGoogleAuthToken()
}

func (a *App) StartSynchronizer() error {
	gClient, err := a.synchronizer.googleClient.GetClientFromSavedToken()
	if err != nil {
		return err
	}

	googleDriveService, err := synchronizer.NewGoogleDriveService(gClient)
	if err != nil {
		return err
	}

	a.synchronizer.sync.SetDriveService(googleDriveService)

	// Set on sync success callback
	a.synchronizer.sync.SetOnSyncSuccess(func() {
		runtime.EventsEmit(a.ctx, "on-sync-success")
	})

	// Set on sync failure callback
	a.synchronizer.sync.SetOnSyncFailure(func(err error) {
		runtime.EventsEmit(a.ctx, "on-sync-failure", err.Error())
	})

	return a.synchronizer.sync.StartScheduler()
}

func (a *App) StartGoogleAuthorization() error {
	port := 43056
	addr := fmt.Sprintf("http://localhost:%d", port)

	done := make(chan bool)
	ticker := time.NewTicker(2 * time.Minute)
	googleClient := a.synchronizer.googleClient

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

		_, err := googleClient.SaveAuthToken(authorizationCode, addr)
		if err != nil {
			slog.Error("Error saving Google authorization token", "error", err)
			runtime.EventsEmit(a.ctx, "on-google-authorization-error", err.Error())
			return
		}

		// Start Synchronizer after authorization
		a.StartSynchronizer()
	})

	if authURL, err := googleClient.GenerateAuthURL(addr); err != nil {
		return err
	} else {
		runtime.BrowserOpenURL(a.ctx, authURL)
	}

	<-done

	return nil
}
