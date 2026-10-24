package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type GithubApp struct {
	Id  int64  `json:"id"`
	Pem string `json:"pem"`
}

func createGitHubAppFromManifest(code string, httpClient *http.Client) (*GithubApp, error) {
	request, err := http.NewRequest(
		http.MethodPost,
		"https://api.github.com/app-manifests/"+url.PathEscape(code)+"/conversions",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create manifest conversion request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "g5a-github-app-creator")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("convert GitHub App manifest: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("convert GitHub App manifest: GitHub returned %s", response.Status)
	}

	var app GithubApp
	if err := json.NewDecoder(response.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("decode manifest conversion response: %w", err)
	}
	if app.Id == 0 || app.Pem == "" {
		return nil, fmt.Errorf("manifest conversion response did not include an App ID and private key")
	}
	return &app, nil
}
