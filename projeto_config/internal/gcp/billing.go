package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// GetBillingAccounts lista as contas de billing disponíveis
func GetBillingAccounts() ([]map[string]interface{}, error) {
	cmd := newGCloudCommand(
		"billing", "accounts", "list",
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("erro ao listar contas de billing: %w\nStderr: %s", err, stderr.String())
	}

	var accounts []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &accounts); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse das contas de billing: %w", err)
	}

	return accounts, nil
}

// LinkBillingAccount vincula uma conta de billing a um projeto
func LinkBillingAccount(projectID, billingAccountID string) error {
	cmd := newGCloudCommand(
		"billing", "projects", "link", projectID,
		fmt.Sprintf("--billing-account=%s", billingAccountID),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("erro ao vincular conta de billing ao projeto %s: %w\nStderr: %s", projectID, err, stderr.String())
	}

	return nil
}

// GetProjectBillingAccount retorna a conta de billing vinculada a um projeto
func GetProjectBillingAccount(projectID string) (string, error) {
	cmd := newGCloudCommand(
		"billing", "projects", "describe", projectID,
		"--format=value(billingAccountName)",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Se não houver billing account, retorna string vazia
		return "", nil
	}

	result := stdout.String()
	if result == "" || result == "None\n" {
		return "", nil
	}

	return result, nil
}
