package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"

	"github.com/paulakimenko/splitwise-cli/internal/config"
)

const (
	authURL  = "https://secure.splitwise.com/oauth/authorize"
	tokenURL = "https://secure.splitwise.com/oauth/token"
)

// StoredAuth persists OAuth credentials.
type StoredAuth struct {
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	Token        *oauth2.Token `json:"token"`
}

// oauthConfig builds an oauth2.Config from stored credentials.
func oauthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURL,
	}
}

// Login runs the full OAuth 2.0 authorization code flow.
func Login(clientID, clientSecret string) error {
	// Find an available port for the callback server.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)

	cfg := oauthConfig(clientID, clientSecret, redirectURL)
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	authCodeURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println("Opening browser for Splitwise authorization...")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", authCodeURL)
	openBrowser(authCodeURL)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch in OAuth callback")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errCh <- fmt.Errorf("no authorization code received")
			return
		}
		fmt.Fprint(w, "<html><body><h2>✓ Authorization successful!</h2><p>You can close this tab and return to the terminal.</p></body></html>")
		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("timed out waiting for authorization")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	stored := &StoredAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Token:        token,
	}
	return saveAuth(stored)
}

// LoadToken returns a valid token, refreshing if needed.
func LoadToken() (string, error) {
	stored, err := loadAuth()
	if err != nil {
		return "", fmt.Errorf("not logged in — run `splitwise auth` first")
	}

	if stored.Token.Valid() {
		return stored.Token.AccessToken, nil
	}

	// Attempt refresh.
	cfg := oauthConfig(stored.ClientID, stored.ClientSecret, "")
	src := cfg.TokenSource(context.Background(), stored.Token)
	newToken, err := src.Token()
	if err != nil {
		return "", fmt.Errorf("token expired and refresh failed — run `splitwise auth` again: %w", err)
	}

	stored.Token = newToken
	if err := saveAuth(stored); err != nil {
		// Non-fatal: we have a valid token, just can't persist.
		fmt.Fprintf(os.Stderr, "Warning: could not save refreshed token: %v\n", err)
	}
	return newToken.AccessToken, nil
}

// IsLoggedIn checks if credentials exist.
func IsLoggedIn() bool {
	_, err := loadAuth()
	return err == nil
}

// Logout removes stored credentials.
func Logout() error {
	path := config.AuthPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func loadAuth() (*StoredAuth, error) {
	data, err := os.ReadFile(config.AuthPath())
	if err != nil {
		return nil, err
	}
	var stored StoredAuth
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	if stored.Token == nil {
		return nil, fmt.Errorf("no token stored")
	}
	return &stored, nil
}

func saveAuth(stored *StoredAuth) error {
	if err := os.MkdirAll(config.Dir(), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.AuthPath(), data, 0o600)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
