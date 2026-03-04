package gcp

import (
	"bytes"
	"fmt"
	"os/exec"
)

// EnableAPI habilita uma API em um projeto GCP
func EnableAPI(projectID, apiName string) error {
	// Comando: gcloud services enable <api-name> --project=<project-id>
	cmd := exec.Command(
		"gcloud",
		"services", "enable", apiName,
		fmt.Sprintf("--project=%s", projectID),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("erro ao habilitar API %s no projeto %s: %w\nStderr: %s", apiName, projectID, err, stderr.String())
	}

	return nil
}

// EnableAPIs habilita múltiplas APIs em um projeto
func EnableAPIs(projectID string, apiNames []string) error {
	for _, apiName := range apiNames {
		if err := EnableAPI(projectID, apiName); err != nil {
			return err
		}
	}
	return nil
}

// GetRequiredAPIs retorna as APIs necessárias
func GetRequiredAPIs() []string {
	return []string{
		"compute.googleapis.com",           // Compute Engine
		"servicenetworking.googleapis.com", // Service Networking
	}
}
