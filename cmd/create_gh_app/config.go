package main

import (
	"fmt"
	"os"
	"strings"
)

const HOST = "127.0.0.1"
const DEFAULT_PORT = "3784"
const CREATE_APP_PATH = "/create-app"
const CALLBACK_PATH = "/callback"

type Config struct {
	Port           string
	AppName        string
	AppUrl         string
	AppPath        string
	PrivateKeyPath string
}

func getEnv(key string, fallback string) string {
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

func setupConfig() (*Config, error) {
	port := getEnv("CREATE_GH_APP_PORT", DEFAULT_PORT)
	appName, err := requireEnv("GITHUB_APP_NAME")
	if err != nil {
		return nil, fmt.Errorf("failed to get GITHUB_APP_NAME: %w", err)
	}
	appUrl, err := requireEnv("GITHUB_APP_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to get GITHUB_APP_URL: %w", err)
	}
	githubOrg := strings.TrimSpace(os.Getenv("GITHUB_APP_ORGANIZATION"))

	appPath := "/settings/apps/new"
	if githubOrg != "" {
		appPath = fmt.Sprintf("/organizations/%s%s", githubOrg, appPath)
	}
	privateKeyPath, err := requireEnv("GITHUB_APP_PRIVATE_KEY")
	if err != nil {
		return nil, fmt.Errorf("failed to get GITHUB_APP_PRIVATE_KEY: %w", err)
	}

	return &Config{
		Port:           port,
		AppName:        appName,
		AppUrl:         appUrl,
		AppPath:        appPath,
		PrivateKeyPath: privateKeyPath,
	}, nil
}
