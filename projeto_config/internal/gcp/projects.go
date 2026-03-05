package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// CreateProject cria um novo projeto GCP dentro de uma pasta
func CreateProject(projectID, displayName, parentFolderID string) (string, error) {
	// Comando: gcloud projects create <project-id> --name=<display-name> --folder=<folder-id>
	cmd := exec.Command(
		"gcloud",
		"projects", "create", projectID,
		fmt.Sprintf("--name=%s", displayName),
		fmt.Sprintf("--folder=%s", parentFolderID),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("erro ao criar projeto %s: %w\nStdout: %s\nStderr: %s", projectID, err, stdout.String(), stderr.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("erro ao fazer parse do resultado: %w", err)
	}

	createdProjectID, ok := result["projectId"].(string)
	if !ok {
		return "", fmt.Errorf("não foi possível extrair project ID da resposta")
	}

	return createdProjectID, nil
}

// SetProjectLabels adiciona labels a um projeto GCP
func SetProjectLabels(projectID string, labels map[string]string) error {
	if len(labels) == 0 {
		return nil
	}

	// Monta o parametro de labels
	labelsStr := ""
	for key, value := range labels {
		if labelsStr != "" {
			labelsStr += ","
		}
		labelsStr += fmt.Sprintf("%s=%s", key, value)
	}

	// Usar gcloud alpha projects update (suporta --update-labels)
	cmd := exec.Command(
		"gcloud",
		"alpha", "projects", "update", projectID,
		fmt.Sprintf("--update-labels=%s", labelsStr),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("erro ao adicionar labels ao projeto %s: %w\nStdout: %s\nStderr: %s", projectID, err, stdout.String(), stderr.String())
	}

	return nil
}

// GetProjectByID busca as informações de um projeto pelo ID
func GetProjectByID(projectID string) (map[string]interface{}, error) {
	cmd := exec.Command(
		"gcloud",
		"projects", "describe", projectID,
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar projeto %s: %w\nStderr: %s", projectID, err, stderr.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do resultado: %w", err)
	}

	return result, nil
}
