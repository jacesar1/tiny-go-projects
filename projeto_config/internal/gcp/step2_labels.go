package gcp

import (
	"fmt"
	"projeto_config/internal/models"
)

// Step2AddLabels implementa o passo 2: adicionar labels aos projetos
// Adiciona os labels:
//   - ambiente: dev | qld | prd
//   - companhia: elet
//   - projeto: <nome do projeto>
func Step2AddLabels(project *models.GCPProject) error {
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("PASSO 2: Adicionando labels aos projetos\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Validar autenticação
	if err := ValidateAuthentication(); err != nil {
		return err
	}

	account, _ := GetCurrentAccount()
	fmt.Printf("✓ Autenticado como: %s\n\n", account)

	// Adicionar labels para cada ambiente
	environments := []*models.GCPEnvironment{project.Dev, project.Qld, project.Prd}

	for _, env := range environments {
		if env.ProjectID == "" {
			fmt.Printf("⚠️  Pulando ambiente %s - Project ID não disponível\n", env.Name)
			continue
		}

		fmt.Printf("🏷️  Adicionando labels ao projeto: %s (%s)\n", env.ProjectID, env.Name)

		// Preparar labels para este ambiente
		labels := map[string]string{
			"ambiente":  env.Name,
			"companhia": "elet",
			"projeto":   project.Name,
		}

		// Adicionar os labels
		err := SetProjectLabels(env.ProjectID, labels)
		if err != nil {
			fmt.Printf("   ❌ Erro: %v\n\n", err)
			return err
		}

		fmt.Printf("   ✓ Labels adicionados com sucesso:\n")
		fmt.Printf("      - ambiente: %s\n", labels["ambiente"])
		fmt.Printf("      - companhia: %s\n", labels["companhia"])
		fmt.Printf("      - projeto: %s\n\n", labels["projeto"])
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Passo 2 concluído com sucesso!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return nil
}
