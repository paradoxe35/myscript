package synchronizer

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type AuthServerRedirection struct {
	server           *http.Server
	authCodeReceived func(string)
}

func NewAuthServerRedirection() *AuthServerRedirection {
	return &AuthServerRedirection{}
}

func (a *AuthServerRedirection) Start(port int) error {
	a.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,

		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorizationCode := r.URL.Query().Get("code")

			if a.authCodeReceived != nil {
				go a.authCodeReceived(authorizationCode)
			}

			fmt.Fprintf(w, "You are now authorized, you can close this window.")
		}),
	}

	return a.server.ListenAndServe()
}

func (a *AuthServerRedirection) Handler(authCodeReceived func(string)) {
	a.authCodeReceived = authCodeReceived
}

func (a *AuthServerRedirection) Stop() error {
	if a.server != nil {
		return a.server.Shutdown(context.Background())
	}

	return nil
}
