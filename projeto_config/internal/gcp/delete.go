package gcp

import (
	"fmt"
	"projeto_config/internal/models"
	"strings"
	"time"
)

// DeleteProject executa shutdown de um projeto GCP
func DeleteProject(projectID string) error {
	cmd := newGCloudCommand("projects", "delete", projectID, "--quiet", "--format=json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao deletar projeto %s: %v\nOutput: %s", projectID, err, string(output))
	}
	return nil
}

// DeleteFolder deleta uma pasta no Resource Manager
func DeleteFolder(folderID string) error {
	cmd := newGCloudCommand("resource-manager", "folders", "delete", folderID, "--quiet", "--format=json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao deletar pasta %s: %v\nOutput: %s", folderID, err, string(output))
	}
	return nil
}

// GetProjectState retorna o estado atual de um projeto (ACTIVE, DELETE_REQUESTED, etc)
func GetProjectState(projectID string) (string, error) {
	projectInfo, err := GetProjectByID(projectID)
	if err != nil {
		return "", err
	}

	if state, ok := projectInfo["lifecycleState"].(string); ok {
		return state, nil
	}

	return "UNKNOWN", nil
}

// WaitForProjectDeletion aguarda até que o projeto seja efetivamente deletado
func WaitForProjectDeletion(projectID string, maxAttempts int, delay time.Duration) error {
	progressShown := false

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		state, err := GetProjectState(projectID)
		if err != nil {
			if progressShown {
				clearInlineProgress()
			}
			// Se GetProjectByID falhar, projeto pode já ter sido deletado
			if strings.Contains(err.Error(), "NOT_FOUND") || strings.Contains(err.Error(), "does not exist") {
				return nil
			}
			return fmt.Errorf("erro ao verificar estado do projeto: %w", err)
		}

		if state == "DELETE_IN_PROGRESS" || state == "DELETE_REQUESTED" {
			if attempt < maxAttempts {
				printInlineProgress("      ⏳ Aguardando conclusao da delecao (%d/%d)...", attempt, maxAttempts)
				progressShown = true
				time.Sleep(delay)
				continue
			}
		}

		if progressShown {
			clearInlineProgress()
		}

		// Se chegou aqui e ainda está ACTIVE, algo deu errado
		if state == "ACTIVE" {
			return fmt.Errorf("projeto ainda está ACTIVE após comando de delete")
		}

		return nil
	}

	return fmt.Errorf("timeout aguardando deleção do projeto")
}

// StepDeleteProject implementa a deleção completa da estrutura criada no passo 1
func StepDeleteProject(project *models.GCPProject) error {
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("OPERAÇÃO DE DELEÇÃO: %s\n", project.Name)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Validar autenticação
	if err := ValidateAuthentication(); err != nil {
		return err
	}

	account, _ := GetCurrentAccount()
	fmt.Printf("✓ Autenticado como: %s\n\n", account)

	// Avisos de segurança
	fmt.Printf("⚠️  ATENÇÃO: Esta operação irá:\n")
	fmt.Printf("   1. Fazer shutdown dos projetos GCP\n")
	fmt.Printf("   2. Deletar as pastas de ambiente (fldr-dev, fldr-qld, fldr-prd)\n")
	fmt.Printf("   3. Deletar a pasta principal (fldr-%s)\n\n", project.Name)

	environments := []*models.GCPEnvironment{project.Dev, project.Qld, project.Prd}

	// Passo 1: Shutdown dos projetos
	fmt.Printf("📋 Fase 1: Shutdown dos projetos GCP\n\n")
	for _, env := range environments {
		if env.ProjectID == "" {
			fmt.Printf("⚠️  Pulando ambiente %s - Project ID não disponível\n", env.Name)
			continue
		}

		fmt.Printf("🗑️  Deletando projeto: %s (%s)\n", env.ProjectID, env.Name)

		// Verificar se projeto existe
		state, err := GetProjectState(env.ProjectID)
		if err != nil {
			if strings.Contains(err.Error(), "NOT_FOUND") {
				fmt.Printf("   ℹ️  Projeto não encontrado, já foi deletado\n\n")
				continue
			}
			return fmt.Errorf("erro ao verificar estado do projeto %s: %w", env.ProjectID, err)
		}

		if state == "DELETE_REQUESTED" || state == "DELETE_IN_PROGRESS" {
			fmt.Printf("   ℹ️  Projeto já está em processo de deleção\n\n")
			continue
		}

		// Executar delete
		if err := DeleteProject(env.ProjectID); err != nil {
			return err
		}

		fmt.Printf("   ✓ Comando de delete enviado\n")

		// Aguardar propagação (GCP pode levar alguns segundos)
		if err := WaitForProjectDeletion(env.ProjectID, 10, 3*time.Second); err != nil {
			fmt.Printf("   ⚠️  Aviso: %v\n", err)
			fmt.Printf("   ℹ️  Projeto pode ainda estar sendo deletado em background\n")
		} else {
			fmt.Printf("   ✓ Projeto deletado\n")
		}

		fmt.Printf("\n")
	}

	// Passo 2: Deletar pastas de ambiente
	fmt.Printf("📋 Fase 2: Deletando pastas de ambiente\n\n")
	for _, env := range environments {
		if env.FolderID == "" {
			fmt.Printf("⚠️  Pulando ambiente %s - Folder ID não disponível\n", env.Name)
			continue
		}

		fmt.Printf("📁 Deletando pasta: fldr-%s (ID: %s)\n", env.Name, env.FolderID)

		if err := DeleteFolder(env.FolderID); err != nil {
			// Se falhar, pode ser porque a pasta já foi deletada ou ainda tem recursos
			if strings.Contains(err.Error(), "NOT_FOUND") {
				fmt.Printf("   ℹ️  Pasta não encontrada, já foi deletada\n\n")
				continue
			}
			fmt.Printf("   ⚠️  Erro: %v\n", err)
			fmt.Printf("   ℹ️  Pasta pode conter recursos pendentes ou já estar em processo de deleção\n\n")
			continue
		}

		fmt.Printf("   ✓ Pasta deletada\n\n")
	}

	// Passo 3: Deletar pasta principal
	fmt.Printf("📋 Fase 3: Deletando pasta principal\n\n")

	// Tentar encontrar a pasta principal
	mainFolderName := fmt.Sprintf("fldr-%s", project.Name)
	fmt.Printf("🔍 Buscando pasta principal: %s\n", mainFolderName)

	// Buscar folder ID da pasta principal (pode estar em qualquer ambiente que tenha FolderID)
	var mainFolderID string
	for _, env := range environments {
		if env.FolderID != "" {
			// Buscar parent folder do ambiente
			cmd := newGCloudCommand("resource-manager", "folders", "describe", env.FolderID, "--format=json")
			output, err := cmd.Output()
			if err == nil {
				// Parse JSON simplificado para pegar parent
				outputStr := string(output)
				if strings.Contains(outputStr, `"parent":`) {
					// Extrair parent: "folders/123456"
					parts := strings.Split(outputStr, `"parent":`)
					if len(parts) > 1 {
						parentPart := strings.Split(parts[1], `"`)[1]
						if strings.HasPrefix(parentPart, "folders/") {
							mainFolderID = strings.TrimPrefix(parentPart, "folders/")
							break
						}
					}
				}
			}
		}
	}

	if mainFolderID == "" {
		fmt.Printf("⚠️  Não foi possível determinar o ID da pasta principal\n")
		fmt.Printf("ℹ️  Você pode deletá-la manualmente se necessário:\n")
		fmt.Printf("   gcloud resource-manager folders list | grep %s\n\n", mainFolderName)
	} else {
		fmt.Printf("📁 Deletando pasta principal: %s (ID: %s)\n", mainFolderName, mainFolderID)

		if err := DeleteFolder(mainFolderID); err != nil {
			if strings.Contains(err.Error(), "NOT_FOUND") {
				fmt.Printf("   ℹ️  Pasta não encontrada, já foi deletada\n\n")
			} else {
				fmt.Printf("   ⚠️  Erro: %v\n", err)
				fmt.Printf("   ℹ️  Pasta pode conter recursos pendentes\n\n")
			}
		} else {
			fmt.Printf("   ✓ Pasta principal deletada\n\n")
		}
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Operação de deleção concluída\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return nil
}
