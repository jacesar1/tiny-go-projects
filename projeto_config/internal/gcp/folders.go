package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// CreateFolder cria uma pasta no Resource Manager do GCP
// Retorna o folder ID da pasta criada
func CreateFolder(parentFolderID, displayName string) (string, error) {
	// Comando: gcloud resource-manager folders create --display-name=<name> --folder=<folder-id>
	// Referência: gcloud help resource-manager folders create

	cmd := newGCloudCommand(
		"resource-manager", "folders", "create",
		fmt.Sprintf("--display-name=%s", displayName),
		fmt.Sprintf("--folder=%s", parentFolderID),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("erro ao criar pasta %s: %w\nStdout: %s\nStderr: %s", displayName, err, stdout.String(), stderr.String())
	}

	output := stdout.Bytes()

	// Parse do JSON retornado
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("erro ao fazer parse do resultado: %w", err)
	}

	folderID, ok := result["name"].(string)
	if !ok {
		return "", fmt.Errorf("não foi possível extrair folder ID da resposta")
	}

	// O resultado vem como "folders/123456", precisamos extrair só o número
	parts := strings.Split(folderID, "/")
	if len(parts) == 2 {
		folderID = parts[1]
	}

	return folderID, nil
}

// ListFolders lista as pastas dentro de uma pasta pai
func ListFolders(parentFolderID string) ([]map[string]interface{}, error) {
	cmd := newGCloudCommand(
		"resource-manager", "folders", "list",
		fmt.Sprintf("--folder=%s", parentFolderID),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("erro ao listar pastas: %w\nStderr: %s", err, stderr.String())
	}

	var folders []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &folders); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse das pastas: %w", err)
	}

	return folders, nil
}

// FindFolderIDByName procura uma pasta pelo nome dentro de uma pasta pai
func FindFolderIDByName(parentFolderID, folderName string) (string, error) {
	folders, err := ListFolders(parentFolderID)
	if err != nil {
		return "", err
	}

	for _, folder := range folders {
		if name, ok := folder["displayName"].(string); ok && name == folderName {
			if folderPath, ok := folder["name"].(string); ok {
				// O resultado vem como "folders/123456", precisamos extrair só o número
				parts := strings.Split(folderPath, "/")
				if len(parts) == 2 {
					return parts[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("pasta '%s' não encontrada", folderName)
}

// FindFolderIDByNameInOrg procura uma pasta pelo nome em uma organização
func FindFolderIDByNameInOrg(orgID, folderName string) (string, error) {
	cmd := newGCloudCommand(
		"resource-manager", "folders", "list",
		fmt.Sprintf("--organization=%s", orgID),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("erro ao listar pastas da organização: %w\nStderr: %s", err, stderr.String())
	}

	var folders []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &folders); err != nil {
		return "", fmt.Errorf("erro ao fazer parse das pastas: %w", err)
	}

	for _, folder := range folders {
		if name, ok := folder["displayName"].(string); ok && name == folderName {
			if folderPath, ok := folder["name"].(string); ok {
				parts := strings.Split(folderPath, "/")
				if len(parts) == 2 {
					return parts[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("pasta '%s' não encontrada na organização", folderName)
}
