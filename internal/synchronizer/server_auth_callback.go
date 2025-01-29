package synchronizer

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type AuthServerRedirection struct {
	server *http.Server
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
	}

	return a.server.ListenAndServe()
}

func (a *AuthServerRedirection) Handler(authCodeReceived func(string)) {
	if a.server != nil {
		a.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorizationCode := r.URL.Query().Get("code")

			fmt.Fprintf(w, "You are now authorized, you can close this window.")

			authCodeReceived(authorizationCode)
		})
	}
}

func (a *AuthServerRedirection) Stop() error {
	if a.server != nil {
		return a.server.Shutdown(context.Background())
	}

	return nil
}
