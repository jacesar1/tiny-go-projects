package gcp

import (
	"fmt"
	"projeto_config/internal/models"
	"strings"
)

// LoadExistingProject busca os dados de um projeto já criado no GCP
// Útil quando precisamos executar apenas um passo específico (ex: step 2)
func LoadExistingProject(projectName string) (*models.GCPProject, error) {
	fmt.Printf("🔍 Buscando projetos existentes para: %s\n", projectName)

	gcpProject := &models.GCPProject{
		Name: projectName,
		Dev:  &models.GCPEnvironment{Name: "dev"},
		Qld:  &models.GCPEnvironment{Name: "qld"},
		Prd:  &models.GCPEnvironment{Name: "prd"},
	}

	environments := []*models.GCPEnvironment{gcpProject.Dev, gcpProject.Qld, gcpProject.Prd}
	foundAny := false

	for _, env := range environments {
		// Montar o ID do projeto esperado
		projectID := fmt.Sprintf("elet-%s-%s", projectName, env.Name)

		// Tentar buscar o projeto
		projectInfo, err := GetProjectByID(projectID)
		if err != nil {
			errMsg := err.Error()
			// Se for crash do gcloud, retornar erro fatal em vez de continuar
			if strings.Contains(errMsg, "gcloud CLI crashou") {
				return nil, fmt.Errorf("❌ ERRO CRÍTICO: %v", err)
			}
			fmt.Printf("   ⚠️  Projeto %s não encontrado: %v\n", projectID, err)
			continue
		}

		// Extrair o Project ID
		if pid, ok := projectInfo["projectId"].(string); ok {
			env.ProjectID = pid
			foundAny = true
			fmt.Printf("   ✓ Encontrado: %s\n", pid)
		}

		// Tentar extrair o Folder ID do parent
		if parent, ok := projectInfo["parent"].(map[string]interface{}); ok {
			if folderID, ok := parent["id"].(string); ok {
				env.FolderID = folderID
			}
		}
	}

	if !foundAny {
		return nil, fmt.Errorf("nenhum projeto encontrado para '%s'. Execute primeiro o passo 1", projectName)
	}

	fmt.Printf("\n")
	return gcpProject, nil
}
