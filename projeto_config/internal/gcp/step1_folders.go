package gcp

import (
	"fmt"
	"projeto_config/internal/models"
	"strconv"
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

	// Garantir pasta principal: fldr-<nome do projeto>
	mainFolderName := fmt.Sprintf("fldr-%s", config.ProjectName)
	fmt.Printf("📁 Garantindo pasta principal: %s\n", mainFolderName)
	mainFolderID, err := ensureFolder(parentFolderID, mainFolderName)
	if err != nil {
		return nil, err
	}
	fmt.Printf("   ✓ Folder ID: %s\n\n", mainFolderID)

	// Criar as subpastas para cada ambiente selecionado (padrao: dev, qld, prd)
	environments := selectedCreateEnvironments(config.TargetEnvironments)
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
		// Garantir pasta de ambiente
		envFolderName := fmt.Sprintf("fldr-%s", env)
		fmt.Printf("📁 Garantindo pasta de ambiente: fldr-%s/%s\n", config.ProjectName, envFolderName)
		envFolderID, err := ensureFolder(mainFolderID, envFolderName)
		if err != nil {
			return nil, err
		}
		fmt.Printf("   ✓ Folder ID: %s\n", envFolderID)

		// Armazenar Folder ID
		envMap[env].FolderID = envFolderID

		// Garantir projeto GCP dentro da pasta de ambiente
		projectName := fmt.Sprintf("elet-%s-%s", config.ProjectName, env)
		projectID := strings.ReplaceAll(projectName, "_", "-")
		projectID = strings.ReplaceAll(projectID, ".", "-")

		fmt.Printf("   📊 Garantindo projeto GCP: %s (ID: %s)\n", projectName, projectID)
		createdProjectID, err := ensureProject(projectID, projectName, envFolderID)
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

func selectedCreateEnvironments(selected []string) []string {
	if len(selected) == 0 {
		return []string{"dev", "qld", "prd"}
	}

	ordered := []string{"dev", "qld", "prd"}
	set := map[string]struct{}{}
	for _, env := range selected {
		set[env] = struct{}{}
	}

	result := make([]string, 0, len(ordered))
	for _, env := range ordered {
		if _, ok := set[env]; ok {
			result = append(result, env)
		}
	}

	return result
}

func ensureFolder(parentFolderID, displayName string) (string, error) {
	folderID, err := FindFolderIDByName(parentFolderID, displayName)
	if err == nil {
		fmt.Printf("   ↺ Pasta já existe, reutilizando\n")
		return folderID, nil
	}

	if !strings.Contains(strings.ToLower(err.Error()), "não encontrada") {
		return "", err
	}

	return CreateFolder(parentFolderID, displayName)
}

func ensureProject(projectID, displayName, expectedFolderID string) (string, error) {
	projectInfo, err := GetProjectByID(projectID)
	if err == nil {
		currentFolderID := extractParentFolderID(projectInfo)
		if currentFolderID != "" && currentFolderID != expectedFolderID {
			return "", fmt.Errorf("projeto %s já existe, mas está na pasta %s (esperado: %s)", projectID, currentFolderID, expectedFolderID)
		}

		fmt.Printf("      ↺ Projeto já existe, reutilizando\n")
		return projectID, nil
	}

	errMsg := strings.ToLower(err.Error())
	isNotFound := strings.Contains(errMsg, "não encontrado")
	isAmbiguousPermission := strings.Contains(errMsg, "does not have permission") || strings.Contains(errMsg, "caller does not have permission")

	if !isNotFound && !isAmbiguousPermission {
		return "", err
	}

	if isAmbiguousPermission {
		fmt.Printf("      ⚠️  Não foi possível confirmar existência via describe (permissão). Tentando criar projeto...\n")
	}

	return CreateProject(projectID, displayName, expectedFolderID)
}

func extractParentFolderID(projectInfo map[string]interface{}) string {
	parent, ok := projectInfo["parent"].(map[string]interface{})
	if !ok {
		return ""
	}

	if id, ok := parent["id"].(string); ok && id != "" {
		return id
	}

	if idFloat, ok := parent["id"].(float64); ok {
		return strconv.FormatInt(int64(idFloat), 10)
	}

	return ""
}

// PrintProjectStructure exibe a estrutura criada
func PrintProjectStructure(project *models.GCPProject) {
	fmt.Printf("📊 Estrutura criada:\n\n")
	fmt.Printf("fldr-%s/\n", project.Name)

	for _, env := range []*models.GCPEnvironment{project.Dev, project.Qld, project.Prd} {
		if env == nil || (env.FolderID == "" && env.ProjectID == "") {
			continue
		}
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
