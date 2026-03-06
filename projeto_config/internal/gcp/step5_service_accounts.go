package gcp

import (
	"fmt"
	"projeto_config/internal/models"
	"strings"
	"time"
)

// Step5CreateServiceAccounts implementa o passo 5: criar service accounts e roles
func Step5CreateServiceAccounts(project *models.GCPProject) error {
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

	for _, env := range environments {
		if env.ProjectID == "" {
			fmt.Printf("⚠️  Pulando ambiente %s - Project ID nao disponivel\n", env.Name)
			continue
		}

		fmt.Printf("🔧 Configurando service accounts no projeto: %s (%s)\n", env.ProjectID, env.Name)

		gitlabAccountID := fmt.Sprintf("sa-%s-git", project.Name)
		gsaAccountID := fmt.Sprintf("sa-%s-%s", project.Name, env.Name)

		// Service account da pipeline (GitLab)
		if err := CreateServiceAccount(env.ProjectID, gitlabAccountID, "GitLab Pipeline Service Account"); err != nil {
			return err
		}
		if err := WaitForServiceAccount(env.ProjectID, gitlabAccountID, 5, 2*time.Second); err != nil {
			return err
		}

		gitlabEmail := ServiceAccountEmail(gitlabAccountID, env.ProjectID)
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

		gsaEmail := ServiceAccountEmail(gsaAccountID, env.ProjectID)
		customRoleName := fmt.Sprintf("projects/%s/roles/%s", env.ProjectID, customRoleID)
		if err := AddProjectIamBinding(env.ProjectID, fmt.Sprintf("serviceAccount:%s", gsaEmail), customRoleName); err != nil {
			return err
		}

		// Secret Manager Viewer conforme requisicao
		if err := AddProjectIamBinding(env.ProjectID, fmt.Sprintf("serviceAccount:%s", gsaEmail), "roles/secretmanager.viewer"); err != nil {
			return err
		}

		fmt.Printf("   ✓ Service accounts e roles configuradas\n\n")
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Passo 5 concluido com sucesso!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return nil
}
