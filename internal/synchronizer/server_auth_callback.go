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

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<!DOCTYPE html>
	<html>
	<head>
		<title>Authorization Success</title>
		<style>
			body {
				font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
				background: #f5f5f5;
				height: 100vh;
				display: flex;
				justify-content: center;
				align-items: center;
				margin: 0;
				text-align: center;
			}
			.container {
				background: white;
				padding: 3rem;
				border-radius: 1rem;
				box-shadow: 0 2px 10px rgba(0,0,0,0.1);
				max-width: 90%;
			}
			.checkmark {
				font-size: 4rem;
				color: #34D399;
				margin-bottom: 1rem;
			}
			h1 {
				color: #1F2937;
				margin: 0 0 1rem 0;
			}
			p {
				color: #6B7280;
				margin: 0;
				font-size: 1.1rem;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="checkmark">✔</div>
			<h1>Authorization Successful!</h1>
			<p>You can now safely close this window.</p>
		</div>
	</body>
	</html>`)
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
