package gcp

import (
	"fmt"
	"projeto_config/internal/models"
	"strings"
)

// Step1CreateFolderStructure implementa o passo 1: criar as pastas no Resource Manager
// Estrutura esperada:
// fldr-<nome do projeto>
//
//	└─ fldr-dev
//	     └─ elet-<nome do projeto>-dev (projeto)
//	└─ fldr-qld
//	     └─ elet-<nome do projeto>-qld (projeto)
//	└─ fldr-prd
//	     └─ elet-<nome do projeto>-prd (projeto)
func Step1CreateFolderStructure(config *models.ProjectConfig) (*models.GCPProject, error) {
	BeginStepCommandTrace("passo 1")
	defer EndStepCommandTrace()

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("PASSO 1: Criando estrutura de pastas\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Validar autenticação
	if err := ValidateAuthentication(); err != nil {
		return nil, err
	}

	account, _ := GetCurrentAccount()
	fmt.Printf("✓ Autenticado como: %s\n\n", account)

	// Resolver o ID da pasta pai (pode ser um nome ou um ID)
	parentFolderID := config.ParentFolderID

	// Se não for um número, tenta resolver como nome
	if !isNumeric(parentFolderID) {
		fmt.Printf("🔍 Resolvendo pasta pai: %s\n", parentFolderID)
		var err error
		parentFolderID, err = FindFolderIDByNameInOrg(config.OrgID, parentFolderID)
		if err != nil {
			return nil, fmt.Errorf("erro ao encontrar pasta '%s': %v", config.ParentFolderID, err)
		}
		fmt.Printf("   ✓ ID encontrado: %s\n\n", parentFolderID)
	}

	// Criar pasta principal: fldr-<nome do projeto>
	mainFolderName := fmt.Sprintf("fldr-%s", config.ProjectName)
	fmt.Printf("📁 Criando pasta principal: %s\n", mainFolderName)
	mainFolderID, err := CreateFolder(parentFolderID, mainFolderName)
	if err != nil {
		return nil, err
	}
	fmt.Printf("   ✓ Folder ID: %s\n\n", mainFolderID)

	// Criar as subpastas para cada ambiente (dev, qld, prd)
	environments := []string{"dev", "qld", "prd"}
	gcpProject := &models.GCPProject{
		Name: config.ProjectName,
		Dev:  &models.GCPEnvironment{Name: "dev"},
		Qld:  &models.GCPEnvironment{Name: "qld"},
		Prd:  &models.GCPEnvironment{Name: "prd"},
	}

	envMap := map[string]*models.GCPEnvironment{
		"dev": gcpProject.Dev,
		"qld": gcpProject.Qld,
		"prd": gcpProject.Prd,
	}

	for _, env := range environments {
		// Criar pasta de ambiente
		envFolderName := fmt.Sprintf("fldr-%s", env)
		fmt.Printf("📁 Criando pasta de ambiente: fldr-%s/%s\n", config.ProjectName, envFolderName)
		envFolderID, err := CreateFolder(mainFolderID, envFolderName)
		if err != nil {
			return nil, err
		}
		fmt.Printf("   ✓ Folder ID: %s\n", envFolderID)

		// Armazenar Folder ID
		envMap[env].FolderID = envFolderID

		// Criar projeto GCP dentro da pasta de ambiente
		projectName := fmt.Sprintf("elet-%s-%s", config.ProjectName, env)
		projectID := strings.ReplaceAll(projectName, "_", "-")
		projectID = strings.ReplaceAll(projectID, ".", "-")

		fmt.Printf("   📊 Criando projeto GCP: %s (ID: %s)\n", projectName, projectID)
		createdProjectID, err := CreateProject(projectID, projectName, envFolderID)
		if err != nil {
			return nil, err
		}
		fmt.Printf("      ✓ Project ID: %s\n", createdProjectID)

		// Vincular conta de billing se fornecida
		if config.BillingAccountID != "" {
			fmt.Printf("   💳 Vinculando billing account: %s\n", config.BillingAccountID)
			if err := LinkBillingAccount(createdProjectID, config.BillingAccountID); err != nil {
				return nil, err
			}
			fmt.Printf("      ✓ Billing account vinculado\n")
		}
		fmt.Printf("\n")

		// Armazenar Project ID
		envMap[env].ProjectID = createdProjectID
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Passo 1 concluído com sucesso!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return gcpProject, nil
}

// PrintProjectStructure exibe a estrutura criada
func PrintProjectStructure(project *models.GCPProject) {
	fmt.Printf("📊 Estrutura criada:\n\n")
	fmt.Printf("fldr-%s/\n", project.Name)

	for _, env := range []*models.GCPEnvironment{project.Dev, project.Qld, project.Prd} {
		fmt.Printf("├── fldr-%s/\n", env.Name)
		fmt.Printf("│   └── elet-%s-%s\n", project.Name, env.Name)
		fmt.Printf("│       ├── Folder ID: %s\n", env.FolderID)
		fmt.Printf("│       └── Project ID: %s\n", env.ProjectID)
	}
	fmt.Printf("\n")
}

// isNumeric verifica se uma string contém apenas dígitos
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
