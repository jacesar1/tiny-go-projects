package gcp

import (
	"bufio"
	"fmt"
	"os"
	"projeto_config/internal/models"
	"sort"
	"strings"
)

var optionalAPICatalog = map[string]string{
	"artifactregistry.googleapis.com": "Artifact Registry",
	"secretmanager.googleapis.com":    "Secret Manager",
	"firestore.googleapis.com":        "Firestore",
}

// Step3Options define como o passo 3 seleciona APIs opcionais.
type Step3Options struct {
	Interactive  bool
	OptionalAPIs []string
}

// AvailableOptionalAPIs retorna os IDs das APIs opcionais suportadas.
func AvailableOptionalAPIs() []string {
	keys := make([]string, 0, len(optionalAPICatalog))
	for apiID := range optionalAPICatalog {
		keys = append(keys, apiID)
	}
	sort.Strings(keys)
	return keys
}

// Step3EnableAPIs implementa o passo 3: habilitar APIs nos projetos
// Habilita APIs obrigatórias e pergunta sobre APIs opcionais
func Step3EnableAPIs(project *models.GCPProject) error {
	return Step3EnableAPIsWithOptions(project, Step3Options{Interactive: true})
}

// Step3EnableAPIsWithOptions implementa o passo 3 com selecao de APIs opcionais por flags.
func Step3EnableAPIsWithOptions(project *models.GCPProject, options Step3Options) error {
	BeginStepCommandTrace("passo 3")
	defer EndStepCommandTrace()

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
	selectedOptionalAPIs, err := resolveOptionalAPIs(options)
	if err != nil {
		return err
	}

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

func resolveOptionalAPIs(options Step3Options) ([]string, error) {
	if options.Interactive {
		return askForOptionalAPIs(optionalAPICatalog), nil
	}

	if len(options.OptionalAPIs) == 0 {
		return nil, nil
	}

	allowed := map[string]struct{}{}
	for _, apiID := range AvailableOptionalAPIs() {
		allowed[apiID] = struct{}{}
	}

	unique := map[string]struct{}{}
	selected := make([]string, 0, len(options.OptionalAPIs))
	for _, apiID := range options.OptionalAPIs {
		apiID = strings.TrimSpace(apiID)
		if apiID == "" {
			continue
		}
		if _, ok := allowed[apiID]; !ok {
			return nil, fmt.Errorf("api opcional nao suportada: %s", apiID)
		}
		if _, seen := unique[apiID]; seen {
			continue
		}
		unique[apiID] = struct{}{}
		selected = append(selected, apiID)
	}

	return selected, nil
}

// askForOptionalAPIs pergunta ao usuário quais APIs opcionais habilitar
func askForOptionalAPIs(optionalAPIs map[string]string) []string {
	fmt.Printf("📋 APIs Opcionais (Digite 's' para sim, 'n' para não):\n\n")

	var selectedAPIs []string
	reader := bufio.NewReader(os.Stdin)

	keys := make([]string, 0, len(optionalAPIs))
	for apiID := range optionalAPIs {
		keys = append(keys, apiID)
	}
	sort.Strings(keys)

	for _, apiID := range keys {
		apiName := optionalAPIs[apiID]
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
