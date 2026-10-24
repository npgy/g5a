package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

var githubAppState GitHubAppState

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		return
	}
}

func createStateSecret(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

func addHandler(httpMux *http.ServeMux, server *http.Server, path string, handlerFunc HandlerFunc) {
	httpMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(server, w, r); err != nil {
			fmt.Printf("handler error for path %s: %v\n", path, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func run() error {
	config, err := setupConfig()
	if err != nil {
		return fmt.Errorf("setup config: %w", err)
	}

	stateSecret, err := createStateSecret(32)
	if err != nil {
		return fmt.Errorf("failed to create state secret: %w", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	httpMux := http.NewServeMux()
	server := &http.Server{
		Handler: httpMux,
	}

	addHandler(httpMux, server, CREATE_APP_PATH, func(server *http.Server, w http.ResponseWriter, r *http.Request) error {
		callbackUrl := fmt.Sprintf("http://%s:%s%s", HOST, config.Port, CALLBACK_PATH)
		setupUrl := fmt.Sprintf("http://%s:%s%s", HOST, config.Port, INSTALLATION_CALLBACK_PATH)

		manifest := Json{
			"name":         config.AppName,
			"url":          config.AppUrl,
			"redirect_url": callbackUrl,
			"setup_url":    setupUrl,
			"public":       false,
			"default_permissions": Json{
				"metadata": "read",
				"contents": "write",
			},
		}
		manifestJson, err := json.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("marshal manifest: %w", err)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(createFormPage(config.AppPath, manifestJson, stateSecret))

		return nil
	})

	addHandler(httpMux, server, CALLBACK_PATH, func(server *http.Server, w http.ResponseWriter, r *http.Request) error {
		code := r.URL.Query().Get("code")
		if code == "" {
			return fmt.Errorf("missing GitHub manifest code")
		}

		app, err := createGitHubAppFromManifest(code, httpClient)
		if err != nil {
			return fmt.Errorf("create GitHub App from manifest: %w", err)
		}

		if err := savePrivateKey(config.PrivateKeyPath, app.Pem); err != nil {
			return fmt.Errorf("save private key: %w", err)
		}
		githubAppState.setAppID(app.Id)

		installationUrl := fmt.Sprintf("https://github.com/apps/%s/installations/new", app.Slug)
		http.Redirect(w, r, installationUrl, http.StatusSeeOther)
		return nil
	})

	addHandler(httpMux, server, INSTALLATION_CALLBACK_PATH, func(server *http.Server, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Query().Get("setup_action") != "install" {
			return fmt.Errorf("expected GitHub setup action install")
		}

		installationID, err := strconv.ParseInt(r.URL.Query().Get("installation_id"), 10, 64)
		if err != nil || installationID <= 0 {
			return fmt.Errorf("invalid GitHub installation ID")
		}
		githubAppState.setInstallationID(installationID)

		appID, installationID, ok := githubAppState.ids()
		if !ok {
			return fmt.Errorf("GitHub App ID is unavailable")
		}
		if err := writeGitHubSecrets(appID, installationID, "github_app_secrets.env"); err != nil {
			return fmt.Errorf("write GitHub secrets: %w", err)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(successPage(config.PrivateKeyPath)); err != nil {
			fmt.Fprintf(os.Stderr, "write installation callback response: %v\n", err)
		}

		go func() {
			fmt.Printf("GitHub app created and installed successfully!\n")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "shut down server: %v\n", err)
			}
		}()

		return nil
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Port))
	if err != nil {
		return fmt.Errorf("listen on port %s: %w", config.Port, err)
	}

	url := fmt.Sprintf("http://%s:%s%s", HOST, config.Port, CREATE_APP_PATH)
	if err := openBrowser(url); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}
