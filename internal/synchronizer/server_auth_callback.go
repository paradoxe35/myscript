package synchronizer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
			errorParam := r.URL.Query().Get("error")

			if a.authCodeReceived != nil {
				go a.authCodeReceived(authorizationCode)
			}

			containerHTML := `
			<div class="container">
				<div class="checkmark">✓</div>
				<h1>Authorization <span>Successful</span></h1>
				<p class="pulse">You may now close this window</p>
			</div>
			`

			if strings.TrimSpace(errorParam) != "" {
				containerHTML = `
			<div class="container">
				<div class="error-icon">×</div>
				<h1>Authorization <span class="error">Failed</span></h1>
				<p class="pulse">You may now close this window</p>
			</div>
			`
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<!DOCTYPE html>
			<html>
			<head>
				<title>Authorization Success</title>
				<style>
					body {
						font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
						background: #09090b;
						height: 100vh;
						display: flex;
						justify-content: center;
						align-items: center;
						margin: 0;
						color: #fafafa;
					}
					.container {
						background: #18181b;
						padding: 2.5rem;
						border-radius: 0.75rem;
						border: 1px solid #27272a;
						box-shadow: 0 20px 25px -5px rgba(0,0,0,0.4);
						max-width: 420px;
						text-align: center;
					}
					.checkmark {
						font-size: 3.5rem;
						color: #3b82f6;
						margin-bottom: 1.5rem;
						display: inline-block;
					}
					.error-icon {
						font-size: 3.5rem;
						color: #ef4444;
						margin-bottom: 1.5rem;
					}
					h1 {
						font-size: 1.5rem;
						font-weight: 600;
						margin: 0 0 1rem 0;
						letter-spacing: -0.025em;
					}
					h1 span {
						background: #3b82f6;
						-webkit-background-clip: text;
						background-clip: text;
						color: transparent;
					}
					h1 span.error {
						color: #ef4444;
					}
					p {
						color: #a1a1aa;
						margin: 0;
						line-height: 1.5;
						font-size: 0.95rem;
					}
					.pulse {
						animation: pulse 2s infinite;
					}
					@keyframes pulse {
						0%% { opacity: 1; }
						50%% { opacity: 0.6; }
						100%% { opacity: 1; }
					}
				</style>
				<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap">
			</head>
			<body>
				%s
			</body>
			</html>`, containerHTML)

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
