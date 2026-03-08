package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	cmd := newGCloudCommand(
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
	cmd := newGCloudCommand(
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

// AddProjectIamBindingWithRetry repete o binding quando a SA ainda nao propagou no IAM.
func AddProjectIamBindingWithRetry(projectID, member, role string, maxAttempts int, delay time.Duration) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := AddProjectIamBinding(projectID, member, role)
		if err == nil {
			return nil
		}

		lastErr = err
		if !isServiceAccountNotFoundInBindingError(err) {
			return err
		}

		if attempt < maxAttempts {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("falha ao adicionar binding apos %d tentativas: %w", maxAttempts, lastErr)
}

func isServiceAccountNotFoundInBindingError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "service account") && strings.Contains(errMsg, "does not exist")
}

// CreateServiceAccountKey cria uma chave JSON para a service account.
func CreateServiceAccountKey(projectID, accountID, outputPath string) error {
	email := ServiceAccountEmail(accountID, projectID)
	cmd := newGCloudCommand(
		"iam", "service-accounts", "keys", "create", outputPath,
		fmt.Sprintf("--iam-account=%s", email),
		fmt.Sprintf("--project=%s", projectID),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao criar chave da service account %s no projeto %s: %w\nStderr: %s", email, projectID, err, stderr.String())
	}

	return nil
}

// ServiceAccountHasUserManagedKeys verifica se a service account possui chaves criadas pelo usuario.
func ServiceAccountHasUserManagedKeys(projectID, accountID string) (bool, error) {
	email := ServiceAccountEmail(accountID, projectID)
	cmd := newGCloudCommand(
		"iam", "service-accounts", "keys", "list",
		fmt.Sprintf("--iam-account=%s", email),
		fmt.Sprintf("--project=%s", projectID),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("erro ao listar chaves da service account %s no projeto %s: %w\nStderr: %s", email, projectID, err, stderr.String())
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return false, fmt.Errorf("erro ao parsear lista de chaves da service account %s no projeto %s: %w", email, projectID, err)
	}

	for _, item := range result {
		keyType, _ := item["keyType"].(string)
		if strings.EqualFold(keyType, "USER_MANAGED") {
			return true, nil
		}
	}

	return false, nil
}

// EnsureSecretExists cria a secret se ela ainda nao existir no projeto.
func EnsureSecretExists(projectID, secretID string) error {
	describeCmd := newGCloudCommand(
		"secrets", "describe", secretID,
		fmt.Sprintf("--project=%s", projectID),
		"--format=json",
	)

	var describeStdout, describeStderr bytes.Buffer
	describeCmd.Stdout = &describeStdout
	describeCmd.Stderr = &describeStderr

	if err := describeCmd.Run(); err == nil {
		return nil
	} else {
		stderrStr := describeStderr.String()
		if !strings.Contains(strings.ToLower(stderrStr), "not found") && !strings.Contains(stderrStr, "NOT_FOUND") {
			return fmt.Errorf("erro ao verificar secret %s no projeto %s: %w\nStderr: %s", secretID, projectID, err, stderrStr)
		}
	}

	createCmd := newGCloudCommand(
		"secrets", "create", secretID,
		fmt.Sprintf("--project=%s", projectID),
		"--replication-policy=automatic",
	)

	var createStdout, createStderr bytes.Buffer
	createCmd.Stdout = &createStdout
	createCmd.Stderr = &createStderr

	if err := createCmd.Run(); err != nil {
		stderrStr := createStderr.String()
		if strings.Contains(strings.ToLower(stderrStr), "already exists") || strings.Contains(stderrStr, "ALREADY_EXISTS") {
			return nil
		}
		return fmt.Errorf("erro ao criar secret %s no projeto %s: %w\nStderr: %s", secretID, projectID, err, stderrStr)
	}

	return nil
}

// AddSecretVersionFromFile adiciona uma nova versao a partir de um arquivo local e retorna o numero.
func AddSecretVersionFromFile(projectID, secretID, filePath string) (string, error) {
	cmd := newGCloudCommand(
		"secrets", "versions", "add", secretID,
		fmt.Sprintf("--project=%s", projectID),
		fmt.Sprintf("--data-file=%s", filePath),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("erro ao adicionar versao na secret %s no projeto %s: %w\nStderr: %s", secretID, projectID, err, stderr.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("erro ao parsear resposta da secret %s no projeto %s: %w", secretID, projectID, err)
	}

	nameValue, ok := result["name"].(string)
	if !ok || nameValue == "" {
		return "", fmt.Errorf("resposta da secret %s no projeto %s sem campo name", secretID, projectID)
	}

	parts := strings.Split(nameValue, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("formato inesperado do name da secret %s no projeto %s", secretID, projectID)
	}

	version := parts[len(parts)-1]
	if version == "" {
		return "", fmt.Errorf("nao foi possivel extrair versao da secret %s no projeto %s", secretID, projectID)
	}

	return version, nil
}

// StoreSecretFromFile garante a secret e adiciona uma nova versao.
func StoreSecretFromFile(projectID, secretID, filePath string) (string, error) {
	if err := EnsureSecretExists(projectID, secretID); err != nil {
		return "", err
	}

	version, err := AddSecretVersionFromFile(projectID, secretID, filePath)
	if err != nil {
		return "", err
	}

	return version, nil
}

// CreateCustomRole cria uma custom role no projeto, se ainda nao existir.
func CreateCustomRole(projectID, roleID, title, description string, permissions []string) error {
	cmd := newGCloudCommand(
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

// DisableProjectOrgPolicyEnforce desabilita o enforce de uma policy no projeto.
func DisableProjectOrgPolicyEnforce(projectID, constraint string) error {
	cmd := newGCloudCommand(
		"resource-manager", "org-policies", "disable-enforce", constraint,
		fmt.Sprintf("--project=%s", projectID),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao desabilitar policy %s no projeto %s: %w\nStdout: %s\nStderr: %s", constraint, projectID, err, stdout.String(), stderr.String())
	}

	return nil
}

// ResetProjectOrgPolicy remove o override da policy no projeto (volta a herdar do parent).
func ResetProjectOrgPolicy(projectID, constraint string) error {
	cmd := newGCloudCommand(
		"resource-manager", "org-policies", "delete", constraint,
		fmt.Sprintf("--project=%s", projectID),
		"--quiet",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		// Se nao houver override local, o comportamento desejado (heranca) ja esta ativo.
		if strings.Contains(strings.ToLower(stderrStr), "not found") || strings.Contains(stderrStr, "NOT_FOUND") {
			return nil
		}
		return fmt.Errorf("erro ao remover override da policy %s no projeto %s: %w\nStderr: %s", constraint, projectID, err, stderrStr)
	}

	return nil
}

// IsProjectOrgPolicyEnforced consulta o estado efetivo de enforce para uma constraint no projeto.
func IsProjectOrgPolicyEnforced(projectID, constraint string) (bool, error) {
	cmd := newGCloudCommand(
		"resource-manager", "org-policies", "describe", constraint,
		fmt.Sprintf("--project=%s", projectID),
		"--effective",
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("erro ao consultar policy efetiva %s no projeto %s: %w\nStderr: %s", constraint, projectID, err, stderr.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return false, fmt.Errorf("erro ao parsear policy efetiva %s no projeto %s: %w", constraint, projectID, err)
	}

	booleanPolicy, ok := result["booleanPolicy"].(map[string]interface{})
	if !ok {
		return false, nil
	}

	enforced, ok := booleanPolicy["enforced"].(bool)
	if !ok {
		return false, nil
	}

	return enforced, nil
}

// WaitForPolicyEnforcementState aguarda a policy atingir o estado efetivo esperado.
func WaitForPolicyEnforcementState(projectID, constraint string, wantEnforced bool, maxAttempts int, delay time.Duration) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		enforced, err := IsProjectOrgPolicyEnforced(projectID, constraint)
		if err != nil {
			return err
		}
		if enforced == wantEnforced {
			return nil
		}
		time.Sleep(delay)
	}

	return fmt.Errorf("policy %s no projeto %s nao atingiu enforced=%t apos %d tentativas", constraint, projectID, wantEnforced, maxAttempts)
}

// HasProjectOrgPolicyOverride verifica se ainda existe override local da policy no projeto.
func HasProjectOrgPolicyOverride(projectID, constraint string) (bool, error) {
	cmd := newGCloudCommand(
		"resource-manager", "org-policies", "describe", constraint,
		fmt.Sprintf("--project=%s", projectID),
		"--format=json",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		// Se não encontrou nenhuma policy (local ou herdada), não há override.
		if strings.Contains(strings.ToLower(stderrStr), "not found") || strings.Contains(stderrStr, "NOT_FOUND") {
			return false, nil
		}
		return false, fmt.Errorf("erro ao verificar override da policy %s no projeto %s: %w\nStderr: %s", constraint, projectID, err, stderrStr)
	}

	// Parsear o JSON para verificar se existe spec local (override).
	// Quando a policy é herdada, o campo "spec" não existe ou está vazio.
	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		// Se não conseguir parsear, assume que não há override (comportamento conservador).
		return false, nil
	}

	// Verifica se existe o campo "spec" que indica override local.
	spec, hasSpec := result["spec"]
	if !hasSpec {
		return false, nil // Não há spec = policy herdada
	}

	// Se spec existe mas está vazio/nil, também não há override.
	if spec == nil {
		return false, nil
	}

	// Se spec existe e não é nil, há override local.
	return true, nil
}

// WaitForPolicyReset aguarda remocao do override local da policy.
func WaitForPolicyReset(projectID, constraint string, maxAttempts int, delay time.Duration) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		hasOverride, err := HasProjectOrgPolicyOverride(projectID, constraint)
		if err != nil {
			return err
		}
		if !hasOverride {
			return nil
		}
		time.Sleep(delay)
	}

	return fmt.Errorf("policy %s no projeto %s ainda possui override apos %d tentativas", constraint, projectID, maxAttempts)
}

func serviceAccountExists(projectID, accountID string) (bool, error) {
	email := ServiceAccountEmail(accountID, projectID)
	cmd := newGCloudCommand(
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
