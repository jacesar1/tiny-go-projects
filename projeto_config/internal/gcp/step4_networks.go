package gcp

import (
	"fmt"
	"projeto_config/internal/models"
)

// Step4AttachToNetworks implementa o passo 4: atachar projetos às redes spoke
// Associa cada projeto à sua VPC spoke correspondente
func Step4AttachToNetworks(project *models.GCPProject) error {
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("PASSO 4: Atachando projetos às Redes Spokes VPC\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Validar autenticação
	if err := ValidateAuthentication(); err != nil {
		return err
	}

	account, _ := GetCurrentAccount()
	fmt.Printf("✓ Autenticado como: %s\n\n", account)

	// Mapeamento de ambientes para seus hosts projects
	// Nota: Usar o ID do projeto, não o display name
	environmentsConfig := map[string]struct {
		env         *models.GCPEnvironment
		hostProject string
		vpcName     string
	}{
		"dev": {
			env:         project.Dev,
			hostProject: "redes-spoke-dev-002b",
			vpcName:     "vpc-spoke-dev",
		},
		"qld": {
			env:         project.Qld,
			hostProject: "redes-spoke-qld-7e83",
			vpcName:     "vpc-spoke-qld",
		},
		"prd": {
			env:         project.Prd,
			hostProject: "redes-spoke-prd-bd15",
			vpcName:     "vpc-spoke-prd",
		},
	}

	// Atachar cada projeto ao seu host project
	for envName, config := range environmentsConfig {
		if config.env.ProjectID == "" {
			fmt.Printf("⚠️  Pulando ambiente %s - Project ID não disponível\n", envName)
			continue
		}

		fmt.Printf("🔗 Atachando projeto à rede Spoke-%s\n", envName)
		fmt.Printf("   Projeto de Serviço: %s\n", config.env.ProjectID)
		fmt.Printf("   Host Project: %s\n", config.hostProject)
		fmt.Printf("   VPC: %s\n", config.vpcName)

		// Atachar o projeto
		if err := AttachToSharedVPC(config.env.ProjectID, config.hostProject); err != nil {
			fmt.Printf("   ❌ Erro: %v\n\n", err)
			return err
		}

		fmt.Printf("   ✓ Projeto atachado com sucesso\n\n")
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Passo 4 concluído com sucesso!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return nil
}
