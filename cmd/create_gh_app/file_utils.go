package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func savePrivateKey(privateKeyPath string, privateKey string) error {
	privateKeyDirectory := filepath.Dir(privateKeyPath)
	if err := os.MkdirAll(privateKeyDirectory, 0o700); err != nil {
		return fmt.Errorf("create private key directory: %w", err)
	}
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("create private key file: %w", err)
	}

	if err := privateKeyFile.Chmod(0o600); err != nil {
		privateKeyFile.Close()
		return fmt.Errorf("set private key permissions: %w", err)
	}
	if _, err := privateKeyFile.WriteString(privateKey); err != nil {
		privateKeyFile.Close()
		return fmt.Errorf("write private key: %w", err)
	}
	if err := privateKeyFile.Close(); err != nil {
		return fmt.Errorf("close private key file: %w", err)
	}
	return nil
}

func writeGitHubSecrets(appID int64, filePath string) error {
	githubSecretsFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create GitHub secrets file: %w", err)
	}
	if _, err := githubSecretsFile.WriteString(fmt.Sprintf("GITHUB_APP_ID=%d\n", appID)); err != nil {
		return fmt.Errorf("write GitHub secrets file: %w", err)
	}
	if err := githubSecretsFile.Close(); err != nil {
		return fmt.Errorf("close GitHub secrets file: %w", err)
	}
	return nil
}
