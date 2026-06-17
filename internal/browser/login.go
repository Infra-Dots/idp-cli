// Package browser implements the loopback OAuth-style browser login flow used
// by `idp auth login`. It starts a local callback server on 127.0.0.1, opens
// the InfraDots web app to mint an API token in the user's session, and waits
// for the token to be handed back to the loopback server.
package browser

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// LoginTimeout bounds how long we wait for the user to finish in the browser.
const LoginTimeout = 3 * time.Minute

type callbackResult struct {
	token string
	err   error
}

// Login performs the browser login flow against the given web app origin
// (e.g. https://app.infradots.com) and returns the minted API token.
func Login(appURL string) (string, error) {
	appURL = strings.TrimRight(appURL, "/")

	state, err := randomState()
	if err != nil {
		return "", err
	}

	// Bind to an OS-assigned free port on the loopback interface only.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("starting local callback server: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	resCh := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			resCh <- callbackResult{err: fmt.Errorf("state mismatch — aborting (possible CSRF)")}
			return
		}
		token := q.Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			resCh <- callbackResult{err: fmt.Errorf("no token returned from browser")}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(successPage))
		resCh <- callbackResult{token: token}
	})

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	hostname, _ := os.Hostname()
	authURL := fmt.Sprintf("%s/cli/auth?port=%d&state=%s&host=%s",
		appURL, port, state, url.QueryEscape(hostname))

	fmt.Fprintf(os.Stderr, "Opening your browser to sign in…\nIf it doesn't open automatically, visit:\n\n  %s\n\n", authURL)
	if err := Open(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "(could not open a browser automatically: %v)\n", err)
	}

	select {
	case res := <-resCh:
		if res.err != nil {
			return "", res.err
		}
		return res.token, nil
	case <-time.After(LoginTimeout):
		return "", fmt.Errorf("timed out after %s waiting for browser sign-in", LoginTimeout)
	}
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

const successPage = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>InfraDots CLI</title>
<style>
  body{font-family:-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;
       background:#0b0d12;color:#e7eaf0;display:flex;align-items:center;
       justify-content:center;height:100vh;margin:0}
  .card{text-align:center;padding:2.5rem 3rem;border:1px solid #232838;
        border-radius:12px;background:#11141c}
  h1{font-size:1.25rem;margin:0 0 .5rem}
  p{color:#9aa3b2;margin:0}
</style></head>
<body><div class="card">
  <h1>✓ You're signed in</h1>
  <p>You can close this tab and return to your terminal.</p>
</div></body></html>`
