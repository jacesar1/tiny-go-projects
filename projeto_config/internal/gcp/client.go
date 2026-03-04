package gcp

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetCurrentProject retorna o projeto padrão configurado no gcloud
func GetCurrentProject() (string, error) {
	cmd := exec.Command(
		"gcloud",
		"config", "get-value", "project",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("erro ao obter projeto atual: %w\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// GetCurrentAccount retorna a conta autenticada no gcloud
func GetCurrentAccount() (string, error) {
	cmd := exec.Command(
		"gcloud",
		"config", "get-value", "account",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("erro ao obter conta atual: %w\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ValidateAuthentication verifica se o usuário está autenticado no GCP
func ValidateAuthentication() error {
	_, err := GetCurrentAccount()
	if err != nil {
		return fmt.Errorf("usuário não autenticado no GCP. Por favor, execute: gcloud auth login")
	}
	return nil
}
