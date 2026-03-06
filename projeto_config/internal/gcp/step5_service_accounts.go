package gcp

import (
	"fmt"
	"os"
	"projeto_config/internal/models"
	"strings"
	"time"
)

type envConfig struct {
	env             *models.GCPEnvironment
	gitlabAccountID string
	gsaAccountID    string
}

// Step5CreateServiceAccounts implementa o passo 5: criar service accounts e roles
func Step5CreateServiceAccounts(project *models.GCPProject) (retErr error) {
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("PASSO 5: Criando Service Accounts e Roles\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Validar autenticacao
	if err := ValidateAuthentication(); err != nil {
		return err
	}

	account, _ := GetCurrentAccount()
	fmt.Printf("✓ Autenticado como: %s\n\n", account)

	// Nome da custom role: substituir hifens por underscores (padrao: [a-zA-Z0-9_\.]{3,64})
	projectNameSafe := strings.ReplaceAll(project.Name, "-", "_")
	customRoleID := fmt.Sprintf("customRole_SA_%s", projectNameSafe)
	customRoleTitle := fmt.Sprintf("Custom Role SA %s", project.Name)
	customRoleDescription := "Custom role para service account GSA"

	customRolePermissions := []string{
		"artifactregistry.repositories.downloadArtifacts",
		"autoscaling.sites.writeMetrics",
		"datastore.entities.get",
		"datastore.entities.list",
		"datastore.entities.update",
		"datastore.entities.create",
		"logging.logEntries.create",
		"monitoring.dashboards.get",
		"monitoring.timeSeries.create",
		"pubsub.subscriptions.consume",
		"pubsub.topics.publish",
	}

	// Processar cada ambiente
	environments := []*models.GCPEnvironment{project.Dev, project.Qld, project.Prd}

	constraints := []string{
		"constraints/iam.disableServiceAccountKeyCreation",
		"constraints/iam.disableServiceAccountKeyUpload",
	}

	// Fase 1: Criar service accounts e aplicar roles em todos os projetos
	fmt.Printf("📋 Fase 1: Service accounts e roles por projeto\n\n")

	var envConfigs []envConfig

	for _, env := range environments {
		if env.ProjectID == "" {
			fmt.Printf("⚠️  Pulando ambiente %s - Project ID nao disponivel\n", env.Name)
			continue
		}

		fmt.Printf("🔧 Configurando service accounts no projeto: %s (%s)\n", env.ProjectID, env.Name)

		gitlabAccountID := fmt.Sprintf("sa-%s-git", project.Name)
		gsaAccountID := fmt.Sprintf("sa-%s-%s", project.Name, env.Name)
		gitlabEmail := ServiceAccountEmail(gitlabAccountID, env.ProjectID)
		gsaEmail := ServiceAccountEmail(gsaAccountID, env.ProjectID)

		fmt.Printf("   - GitLab: %s (%s)\n", gitlabAccountID, gitlabEmail)
		fmt.Printf("   - GSA: %s (%s)\n", gsaAccountID, gsaEmail)

		// Service account da pipeline (GitLab)
		if err := CreateServiceAccount(env.ProjectID, gitlabAccountID, "GitLab Pipeline Service Account"); err != nil {
			return err
		}
		if err := WaitForServiceAccount(env.ProjectID, gitlabAccountID, 5, 2*time.Second); err != nil {
			return err
		}

		if err := AddProjectIamBinding(env.ProjectID, fmt.Sprintf("serviceAccount:%s", gitlabEmail), "roles/artifactregistry.createOnPushWriter"); err != nil {
			return err
		}

		// Custom role e service account GSA
		if err := CreateCustomRole(env.ProjectID, customRoleID, customRoleTitle, customRoleDescription, customRolePermissions); err != nil {
			return err
		}

		if err := CreateServiceAccount(env.ProjectID, gsaAccountID, "GSA Service Account"); err != nil {
			return err
		}
		if err := WaitForServiceAccount(env.ProjectID, gsaAccountID, 5, 2*time.Second); err != nil {
			return err
		}

		customRoleName := fmt.Sprintf("projects/%s/roles/%s", env.ProjectID, customRoleID)
		if err := AddProjectIamBinding(env.ProjectID, fmt.Sprintf("serviceAccount:%s", gsaEmail), customRoleName); err != nil {
			return err
		}

		// Secret Manager Viewer conforme requisicao
		if err := AddProjectIamBinding(env.ProjectID, fmt.Sprintf("serviceAccount:%s", gsaEmail), "roles/secretmanager.viewer"); err != nil {
			return err
		}

		fmt.Printf("   ✓ Service accounts e roles configuradas\n\n")

		envConfigs = append(envConfigs, envConfig{
			env:             env,
			gitlabAccountID: gitlabAccountID,
			gsaAccountID:    gsaAccountID,
		})
	}

	// Fase 2: Desabilitar policies em todos os projetos
	fmt.Printf("📋 Fase 2: Desabilitando policies de org em todos os projetos\n\n")

	for _, cfg := range envConfigs {
		fmt.Printf("🔓 Desabilitando policies no projeto: %s\n", cfg.env.ProjectID)
		for _, constraint := range constraints {
			if err := DisableProjectOrgPolicyEnforce(cfg.env.ProjectID, constraint); err != nil {
				return err
			}
			if err := WaitForPolicyEnforcementState(cfg.env.ProjectID, constraint, false, 18, 10*time.Second); err != nil {
				return fmt.Errorf("falha ao aguardar propagacao da policy %s no projeto %s: %w", constraint, cfg.env.ProjectID, err)
			}
			fmt.Printf("   ✓ %s desabilitada e efetiva (enforced=false)\n", constraint)
		}
	}

	// Fase 3: Criar chaves JSON com retry
	fmt.Printf("\n📋 Fase 3: Criando chaves JSON\n\n")

	defer func() {
		// Fase 4: Resetar policies em todos os projetos (sempre executa)
		fmt.Printf("\n📋 Fase 4: Resetando policies de org em todos os projetos\n\n")
		if err := resetPoliciesForAllProjects(envConfigs, constraints); err != nil {
			if retErr != nil {
				retErr = fmt.Errorf("%v | erro ao resetar policies: %w", retErr, err)
				return
			}
			retErr = err
		}
	}()

	for _, cfg := range envConfigs {
		fmt.Printf("🔑 Criando chaves para projeto: %s (%s)\n", cfg.env.ProjectID, cfg.env.Name)

		// Chave GitLab com retry
		gitlabKeyPath := fmt.Sprintf("%s-%s.json", cfg.gitlabAccountID, cfg.env.Name)
		gitlabHasKey, err := ServiceAccountHasUserManagedKeys(cfg.env.ProjectID, cfg.gitlabAccountID)
		if err != nil {
			return err
		}
		if gitlabHasKey {
			if _, err := os.Stat(gitlabKeyPath); err == nil {
				gitlabVersion, err := StoreSecretFromFile(cfg.env.ProjectID, cfg.gitlabAccountID, gitlabKeyPath)
				if err != nil {
					return err
				}
				fmt.Printf("   ✓ Chave GitLab ja existe, secret atualizada: %s (versao %s)\n", cfg.gitlabAccountID, gitlabVersion)
			} else {
				fmt.Printf("   ℹ️  Chave GitLab ja existe, mas o JSON local nao foi encontrado: %s\n", gitlabKeyPath)
			}
		} else {
			if err := createKeyWithRetry(cfg.env.ProjectID, cfg.gitlabAccountID, gitlabKeyPath, 8, 15*time.Second); err != nil {
				return err
			}
			fmt.Printf("   ✓ Chave GitLab criada: %s\n", gitlabKeyPath)
			gitlabVersion, err := StoreSecretFromFile(cfg.env.ProjectID, cfg.gitlabAccountID, gitlabKeyPath)
			if err != nil {
				return err
			}
			fmt.Printf("   ✓ Secret GitLab atualizada: %s (versao %s)\n", cfg.gitlabAccountID, gitlabVersion)
		}

		// Chave GSA com retry
		gsaKeyPath := fmt.Sprintf("%s.json", cfg.gsaAccountID)
		gsaHasKey, err := ServiceAccountHasUserManagedKeys(cfg.env.ProjectID, cfg.gsaAccountID)
		if err != nil {
			return err
		}
		if gsaHasKey {
			if _, err := os.Stat(gsaKeyPath); err == nil {
				gsaVersion, err := StoreSecretFromFile(cfg.env.ProjectID, cfg.gsaAccountID, gsaKeyPath)
				if err != nil {
					return err
				}
				fmt.Printf("   ✓ Chave GSA ja existe, secret atualizada: %s (versao %s)\n\n", cfg.gsaAccountID, gsaVersion)
			} else {
				fmt.Printf("   ℹ️  Chave GSA ja existe, mas o JSON local nao foi encontrado: %s\n\n", gsaKeyPath)
			}
		} else {
			if err := createKeyWithRetry(cfg.env.ProjectID, cfg.gsaAccountID, gsaKeyPath, 8, 15*time.Second); err != nil {
				return err
			}
			fmt.Printf("   ✓ Chave GSA criada: %s\n", gsaKeyPath)
			gsaVersion, err := StoreSecretFromFile(cfg.env.ProjectID, cfg.gsaAccountID, gsaKeyPath)
			if err != nil {
				return err
			}
			fmt.Printf("   ✓ Secret GSA atualizada: %s (versao %s)\n\n", cfg.gsaAccountID, gsaVersion)
		}
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Passo 5 concluido com sucesso!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return nil
}

func resetPoliciesForAllProjects(envConfigs []envConfig, constraints []string) error {
	var resetErrors []string

	for _, cfg := range envConfigs {
		fmt.Printf("🔒 Resetando policies no projeto: %s\n", cfg.env.ProjectID)
		for _, constraint := range constraints {
			if err := ResetProjectOrgPolicy(cfg.env.ProjectID, constraint); err != nil {
				resetErrors = append(resetErrors, err.Error())
				continue
			}

			if err := WaitForPolicyReset(cfg.env.ProjectID, constraint, 12, 5*time.Second); err != nil {
				resetErrors = append(resetErrors, err.Error())
				continue
			}

			fmt.Printf("   ✓ %s voltou a herdar do parent\n", constraint)
		}
	}

	if len(resetErrors) > 0 {
		return fmt.Errorf(strings.Join(resetErrors, " | "))
	}

	return nil
}

// createKeyWithRetry tenta criar uma chave com retry em caso de falha
func createKeyWithRetry(projectID, accountID, outputPath string, maxAttempts int, delay time.Duration) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := CreateServiceAccountKey(projectID, accountID, outputPath)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxAttempts {
			fmt.Printf("      ⚠️  Tentativa %d falhou, aguardando %v antes de tentar novamente...\n", attempt, delay)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("falhou apos %d tentativas: %w", maxAttempts, lastErr)
}
