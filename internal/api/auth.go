package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

const (
	authorizeURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	tokenURL     = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	clientID     = "1950a258-227b-4e31-a9cf-717495945fc2"
	scope        = "https://analysis.windows.net/powerbi/api/.default offline_access"
)

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func buildAuthorizeURL(redirectURI, state string) string {
	params := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURI},
		"scope":         {scope},
		"state":         {state},
		"prompt":        {"select_account"},
	}
	return authorizeURL + "?" + params.Encode()
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func exchangeCode(code, redirectURI string) (string, error) {
	data := url.Values{
		"client_id":    {clientID},
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
		"scope":        {scope},
	}
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}
	if tr.Error != "" {
		return "", fmt.Errorf("auth error: %s — %s", tr.Error, tr.ErrorDesc)
	}
	return tr.AccessToken, nil
}

// GetAccessToken opens a browser for Entra ID login and returns an access token.
func GetAccessToken() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)
	state := randomState()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errCh <- fmt.Errorf("auth error: %s — %s", errMsg, r.URL.Query().Get("error_description"))
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>%s</p><p>You can close this tab.</p></body></html>", errMsg)
			return
		}
		code := r.URL.Query().Get("code")
		codeCh <- code
		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab and return to the terminal.</p></body></html>")
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	authURL := buildAuthorizeURL(redirectURI, state)
	if err := openBrowser(authURL); err != nil {
		return "", fmt.Errorf("failed to open browser: %w\nOpen manually: %s", err, authURL)
	}

	select {
	case code := <-codeCh:
		return exchangeCode(code, redirectURI)
	case err := <-errCh:
		return "", err
	case <-time.After(2 * time.Minute):
		return "", fmt.Errorf("authentication timed out (2 minutes)")
	}
}
