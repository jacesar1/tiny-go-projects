package gcp

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ServiceAccountEmail monta o email completo de uma service account.
func ServiceAccountEmail(accountID, projectID string) string {
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountID, projectID)
}

// CreateServiceAccount cria uma service account se ela ainda nao existir.
func CreateServiceAccount(projectID, accountID, displayName string) error {
	exists, err := serviceAccountExists(projectID, accountID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	cmd := exec.Command(
		"gcloud",
		"iam", "service-accounts", "create", accountID,
		fmt.Sprintf("--project=%s", projectID),
		fmt.Sprintf("--display-name=%s", displayName),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao criar service account %s no projeto %s: %w\nStderr: %s", accountID, projectID, err, stderr.String())
	}

	return nil
}

// WaitForServiceAccount aguarda a propagacao da service account no IAM.
func WaitForServiceAccount(projectID, accountID string, maxAttempts int, delay time.Duration) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		exists, err := serviceAccountExists(projectID, accountID)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		time.Sleep(delay)
	}

	return fmt.Errorf("service account %s ainda nao disponivel apos %d tentativas", accountID, maxAttempts)
}

// AddProjectIamBinding adiciona um binding IAM em um projeto.
func AddProjectIamBinding(projectID, member, role string) error {
	cmd := exec.Command(
		"gcloud",
		"projects", "add-iam-policy-binding", projectID,
		fmt.Sprintf("--member=%s", member),
		fmt.Sprintf("--role=%s", role),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao adicionar role %s para %s no projeto %s: %w\nStderr: %s", role, member, projectID, err, stderr.String())
	}

	return nil
}

// CreateCustomRole cria uma custom role no projeto, se ainda nao existir.
func CreateCustomRole(projectID, roleID, title, description string, permissions []string) error {
	cmd := exec.Command(
		"gcloud",
		"iam", "roles", "create", roleID,
		fmt.Sprintf("--project=%s", projectID),
		fmt.Sprintf("--title=%s", title),
		fmt.Sprintf("--description=%s", description),
		fmt.Sprintf("--permissions=%s", strings.Join(permissions, ",")),
		"--stage=GA",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		// Se a role já existe, ignorar o erro
		if strings.Contains(strings.ToLower(stderrStr), "already exists") ||
			strings.Contains(stderrStr, "ALREADY_EXISTS") {
			return nil
		}
		return fmt.Errorf("erro ao criar custom role %s no projeto %s: %w\nStderr: %s", roleID, projectID, err, stderrStr)
	}

	return nil
}

func serviceAccountExists(projectID, accountID string) (bool, error) {
	email := ServiceAccountEmail(accountID, projectID)
	cmd := exec.Command(
		"gcloud",
		"iam", "service-accounts", "describe", email,
		fmt.Sprintf("--project=%s", projectID),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if strings.Contains(strings.ToLower(stderrStr), "not found") || strings.Contains(stderrStr, "NOT_FOUND") {
			return false, nil
		}
		return false, fmt.Errorf("erro ao verificar service account %s: %w\nStderr: %s", email, err, stderrStr)
	}

	return true, nil
}
