package gcp

import (
	"bufio"
	"fmt"
	"os"
	"projeto_config/internal/models"
	"strings"
)

// Step3EnableAPIs implementa o passo 3: habilitar APIs nos projetos
// Habilita APIs obrigatórias e pergunta sobre APIs opcionais
func Step3EnableAPIs(project *models.GCPProject) error {
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("PASSO 3: Habilitando APIs nos projetos\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Validar autenticação
	if err := ValidateAuthentication(); err != nil {
		return err
	}

	account, _ := GetCurrentAccount()
	fmt.Printf("✓ Autenticado como: %s\n\n", account)

	// APIs obrigatórias
	requiredAPIs := []string{
		"compute.googleapis.com",           // Compute Engine
		"servicenetworking.googleapis.com", // Service Networking
	}

	// APIs opcionais
	optionalAPIs := map[string]string{
		"artifactregistry.googleapis.com": "Artifact Registry",
		"secretmanager.googleapis.com":    "Secret Manager",
		"firestore.googleapis.com":        "Firestore",
	}

	// Perguntar sobre APIs opcionais
	selectedOptionalAPIs := askForOptionalAPIs(optionalAPIs)

	// Combinar todas as APIs a serem habilitadas
	allAPIs := append(requiredAPIs, selectedOptionalAPIs...)

	// Habilitar APIs em cada projeto
	environments := []*models.GCPEnvironment{project.Dev, project.Qld, project.Prd}

	for _, env := range environments {
		if env.ProjectID == "" {
			fmt.Printf("⚠️  Pulando ambiente %s - Project ID não disponível\n", env.Name)
			continue
		}

		fmt.Printf("🔌 Habilitando APIs no projeto: %s (%s)\n", env.ProjectID, env.Name)

		for _, apiName := range allAPIs {
			fmt.Printf("   ⏳ Habilitando %s...\n", apiName)
			if err := EnableAPI(env.ProjectID, apiName); err != nil {
				fmt.Printf("      ❌ Erro ao habilitar API: %v\n", err)
				return err
			}
			fmt.Printf("      ✓ Habilitada\n")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Passo 3 concluído com sucesso!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return nil
}

// askForOptionalAPIs pergunta ao usuário quais APIs opcionais habilitar
func askForOptionalAPIs(optionalAPIs map[string]string) []string {
	fmt.Printf("📋 APIs Opcionais (Digite 's' para sim, 'n' para não):\n\n")

	var selectedAPIs []string
	reader := bufio.NewReader(os.Stdin)

	for apiID, apiName := range optionalAPIs {
		fmt.Printf("   ❓ Habilitar %s? (s/n): ", apiName)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Erro ao ler entrada: %v\n", err)
			continue
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response == "s" || response == "sim" || response == "yes" || response == "y" {
			selectedAPIs = append(selectedAPIs, apiID)
			fmt.Printf("      ✓ Selecionado: %s\n", apiName)
		} else {
			fmt.Printf("      ✗ Pulado: %s\n", apiName)
		}
	}

	fmt.Printf("\n")
	return selectedAPIs
}
