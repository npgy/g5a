package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
)

const HOST = "127.0.0.1"
const DEFAULT_PORT = "3784"
const CREATE_APP_PATH = "/create-app"
const CALLBACK_PATH = "/callback"

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		return
	}
}

func run() error {
	port := getEnv("PORT", DEFAULT_PORT)
	appName, err := requireEnv("GITHUB_APP_NAME")
	if err != nil {
		return fmt.Errorf("failed to get GITHUB_APP_NAME: %w", err)
	}
	appUrl, err := requireEnv("GITHUB_APP_URL")
	if err != nil {
		return fmt.Errorf("failed to get GITHUB_APP_URL: %w", err)
	}
	githubOrg := strings.TrimSpace(os.Getenv("GITHUB_APP_ORGANIZATION"))

	appPath := "/settings/apps/new"
	if githubOrg != "" {
		appPath = fmt.Sprintf("/organizations/%s%s", githubOrg, appPath)
	}

	stateSecret, err := createStateSecret(32)
	if err != nil {
		return fmt.Errorf("failed to create state secret: %w", err)
	}

	httpMux := http.NewServeMux()

	httpMux.HandleFunc(CREATE_APP_PATH, func(w http.ResponseWriter, r *http.Request) {
		callbackUrl := fmt.Sprintf("http://%s:%s%s", HOST, port, CALLBACK_PATH)

		manifest := map[string]any{
			"name":         appName,
			"url":          appUrl,
			"redirect_url": callbackUrl,
			"public":       false,
			"default_permissions": map[string]string{
				"metadata": "read",
				"contents": "write",
			},
		}
		manifestJson, err := json.Marshal(manifest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		page := fmt.Sprintf(
			`<form id="create-app" action="https://github.com%s" method="post">`+
				`<input type="hidden" name="manifest" value="%s">`+
				`<input type="hidden" name="state" value="%s">`+
				`</form><script>document.getElementById('create-app').submit()</script>`,
			html.EscapeString(appPath),
			html.EscapeString(string(manifestJson)),
			html.EscapeString(stateSecret),
		)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(page))
	})

	httpMux.HandleFunc(CALLBACK_PATH, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		fmt.Printf("received data %s\n", query)
		w.Write([]byte("Hello, GitHub App!"))
	})

	http.ListenAndServe(fmt.Sprintf(":%s", port), httpMux)
	return nil
}

func createStateSecret(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Use NoPadding if you want to omit trailing '=' characters
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func requireEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value), nil
	}
	return "", fmt.Errorf("required environment variable %s not set", key)
}
