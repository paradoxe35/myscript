package main

import (
	"fmt"
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

func (a *App) GoogleStartAuthorization() error {
	port := 43056
	wait := make(chan bool)

	tmpServer := synchronizer.NewAuthServerRedirection()

	go tmpServer.Start(port)
	defer tmpServer.Stop()

	tmpServer.Handler(func(authorizationCode string) {
		a.googleClient.SaveAuthToken(authorizationCode)

		// Stop the server
		tmpServer.Stop()
		close(wait)
	})

	// Timeout after 30 seconds
	go func() {
		time.Sleep(30 * time.Second)
		// Emit timeout event
		runtime.EventsEmit(a.ctx, "on-google-authorization-timeout")

		// Stop the server
		tmpServer.Stop()
		close(wait)
	}()

	if authURL, err := a.googleClient.GenerateAuthURL(fmt.Sprintf("http://localhost:%d", port)); err != nil {
		return err
	} else {
		runtime.BrowserOpenURL(a.ctx, authURL)
	}

	<-wait

	return nil
}
