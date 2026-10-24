package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

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

func addHandler(httpMux *http.ServeMux, path string, handlerFunc HandlerFunc) {
	httpMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(w, r); err != nil {
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

	addHandler(httpMux, CREATE_APP_PATH, func(w http.ResponseWriter, r *http.Request) error {
		callbackUrl := fmt.Sprintf("http://%s:%s%s", HOST, config.Port, CALLBACK_PATH)

		manifest := Json{
			"name":         config.AppName,
			"url":          config.AppUrl,
			"redirect_url": callbackUrl,
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

	addHandler(httpMux, CALLBACK_PATH, func(w http.ResponseWriter, r *http.Request) error {
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

		if err := writeGitHubSecrets(app.Id, "github_app_secrets.env"); err != nil {
			return fmt.Errorf("write GitHub secrets: %w", err)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(successPage(config.PrivateKeyPath)); err != nil {
			fmt.Fprintf(os.Stderr, "write callback response: %v\n", err)
		}

		return nil
	})

	http.ListenAndServe(fmt.Sprintf(":%s", config.Port), httpMux)
	return nil
}
