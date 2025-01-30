package main

import (
	"fmt"
	"log/slog"
	"myscript/internal/repository"
	"myscript/internal/synchronizer"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) IsSynchronizerEnabled() bool {
	return a.googleClient.HasCredentials()
}

func (a *App) GetGoogleAuthToken() *repository.GoogleAuthToken {
	return repository.GetGoogleAuthToken(a.UnSyncedDb)
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
