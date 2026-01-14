package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// CallbackServer handles OAuth callback
type CallbackServer struct {
	port   int
	server *http.Server
	code   chan string
	err    chan error
}

// NewCallbackServer creates a new callback server
func NewCallbackServer(port int) *CallbackServer {
	return &CallbackServer{
		port: port,
		code: make(chan string, 1),
		err:  make(chan error, 1),
	}
}

// Start begins listening for the OAuth callback
func (s *CallbackServer) Start(expectedState string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check for error
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			s.err <- fmt.Errorf("authorization denied: %s", errParam)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authorization Failed</title></head>
<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
<h1>Authorization Failed</h1>
<p>You denied access to the application.</p>
<p>You can close this window.</p>
</body>
</html>`)
			return
		}

		// Verify state
		state := r.URL.Query().Get("state")
		if state != expectedState {
			s.err <- fmt.Errorf("state mismatch: expected %s, got %s", expectedState, state)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authorization Failed</title></head>
<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
<h1>Authorization Failed</h1>
<p>Security validation failed. Please try again.</p>
<p>You can close this window.</p>
</body>
</html>`)
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			s.err <- fmt.Errorf("no authorization code received")
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authorization Failed</title></head>
<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
<h1>Authorization Failed</h1>
<p>No authorization code received.</p>
<p>You can close this window.</p>
</body>
</html>`)
			return
		}

		// Success!
		s.code <- code
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authorization Successful</title></head>
<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
<h1>Authorization Successful!</h1>
<p>You can close this window and return to the terminal.</p>
</body>
</html>`)
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.err <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// WaitForCode blocks until an authorization code is received or timeout
func (s *CallbackServer) WaitForCode(timeout time.Duration) (string, error) {
	select {
	case code := <-s.code:
		return code, nil
	case err := <-s.err:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for authorization")
	}
}

// Stop shuts down the callback server
func (s *CallbackServer) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

// GetRedirectURI returns the redirect URI for OAuth
func (s *CallbackServer) GetRedirectURI() string {
	return fmt.Sprintf("http://localhost:%d/callback", s.port)
}
