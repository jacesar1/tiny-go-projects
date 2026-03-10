package cmd

import (
	"bufio"
	"fmt"
	"os"
	"projeto_config/internal/models"
	"strings"

	"projeto_config/internal/gcp"

	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	var runLabels bool
	var runAPIs bool
	var runNetworks bool
	var runServiceAccounts bool
	var runAll bool
	var allOptionalAPIs bool
	var interactiveAPIs bool
	var interactiveEnvs bool
	var targetEnvs []string
	var optionalAPIs []string

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza recursos de um projeto existente",
		Long: `Atualiza recursos de um projeto existente.

Use a ajuda detalhada do recurso de projeto para ver todas as flags, incluindo selecao de ambientes:
  projeto_config update projeto -h`,
	}

	updateProjectCmd := &cobra.Command{
		Use:     "projeto <nome-do-projeto>",
		Aliases: []string{"project"},
		Short:   "Executa passos de atualizacao (labels/apis/networks/service-accounts)",
		Long: `Exemplos:
  projeto_config update projeto benner-cloud --labels
  projeto_config update projeto benner-cloud --networks
  projeto_config update projeto benner-cloud --service-accounts
	projeto_config update projeto benner-cloud --all --env qld --env prd
	projeto_config update projeto benner-cloud --all --env dev,qld
	projeto_config update projeto benner-cloud --all --interactive-envs
  projeto_config update projeto benner-cloud --apis --optional-api secretmanager --optional-api firestore
  projeto_config update projeto benner-cloud --apis --optional-api secretmanager,firestore
  projeto_config update projeto benner-cloud --all`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			printExecutionHeader(projectName)

			if interactiveEnvs && len(targetEnvs) > 0 {
				return fmt.Errorf("use apenas um modo de seleção de ambientes: --env ou --interactive-envs")
			}

			selectedAny := runLabels || runAPIs || runNetworks || runServiceAccounts || runAll
			if !selectedAny {
				// Comportamento padrao: equivalente aos passos 2, 3 e 4 do fluxo antigo.
				runLabels = true
				runAPIs = true
				runNetworks = true
			}

			if runAll {
				runLabels = true
				runAPIs = true
				runNetworks = true
				runServiceAccounts = true
			}

			project, err := gcp.LoadExistingProject(projectName)
			if err != nil {
				return fmt.Errorf("erro ao carregar projeto existente: %w", err)
			}

			selectedEnvs, err := resolveTargetEnvironments(targetEnvs, interactiveEnvs)
			if err != nil {
				return err
			}

			project = filterProjectEnvironments(project, selectedEnvs)
			if countSelectedProjects(project) == 0 {
				return fmt.Errorf("nenhum projeto encontrado para os ambientes selecionados (%s)", strings.Join(selectedEnvs, ", "))
			}

			if runLabels {
				if err := gcp.Step2AddLabels(project); err != nil {
					return fmt.Errorf("erro no passo labels: %w", err)
				}
			}

			if runAPIs {
				apiOptions := gcp.Step3Options{
					Interactive: interactiveAPIs,
				}

				resolvedOptionalAPIs, err := resolveOptionalAPIsForCommand(runAll, allOptionalAPIs, interactiveAPIs, optionalAPIs)
				if err != nil {
					return err
				}
				apiOptions.OptionalAPIs = resolvedOptionalAPIs

				if err := gcp.Step3EnableAPIsWithOptions(project, apiOptions); err != nil {
					return fmt.Errorf("erro no passo apis: %w", err)
				}
			}

			if runNetworks {
				if err := gcp.Step4AttachToNetworks(project); err != nil {
					return fmt.Errorf("erro no passo networks: %w", err)
				}
			}

			if runServiceAccounts {
				if err := gcp.Step5CreateServiceAccounts(project); err != nil {
					return fmt.Errorf("erro no passo service-accounts: %w", err)
				}
			}

			fmt.Println("✅ Update concluido com sucesso.")
			return nil
		},
	}

	updateProjectCmd.Flags().BoolVar(&runLabels, "labels", false, "Executa o passo 2 (labels)")
	updateProjectCmd.Flags().BoolVar(&runAPIs, "apis", false, "Executa o passo 3 (habilitar APIs)")
	updateProjectCmd.Flags().BoolVar(&runNetworks, "networks", false, "Executa o passo 4 (attach em Shared VPC)")
	updateProjectCmd.Flags().BoolVar(&runServiceAccounts, "service-accounts", false, "Executa o passo 5 (service accounts)")
	updateProjectCmd.Flags().BoolVar(&runAll, "all", false, "Executa todos os passos de update (2, 3, 4 e 5)")
	updateProjectCmd.Flags().StringSliceVar(&optionalAPIs, "optional-api", nil, "API opcional para incluir no passo 3 (artifactregistry|secretmanager|firestore). Aceita virgula")
	updateProjectCmd.Flags().BoolVar(&allOptionalAPIs, "all-optional-apis", false, "Inclui todas as APIs opcionais no passo 3")
	updateProjectCmd.Flags().BoolVar(&interactiveAPIs, "interactive-apis", false, "Pergunta interativamente APIs opcionais no passo 3")
	updateProjectCmd.Flags().StringSliceVar(&targetEnvs, "env", nil, "Ambientes alvo do update (dev|qld|prd). Aceita repeticao e virgula")
	updateProjectCmd.Flags().BoolVar(&interactiveEnvs, "interactive-envs", false, "Pergunta interativamente quais ambientes atualizar")

	updateCmd.AddCommand(updateProjectCmd)
	return updateCmd
}

func resolveTargetEnvironments(values []string, interactive bool) ([]string, error) {
	if interactive {
		return askForTargetEnvironments(), nil
	}

	normalized, err := normalizeTargetEnvironments(values)
	if err != nil {
		return nil, err
	}

	if len(normalized) == 0 {
		return []string{"dev", "qld", "prd"}, nil
	}

	return normalized, nil
}

func normalizeTargetEnvironments(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	allowed := map[string]string{
		"dev":        "dev",
		"qld":        "qld",
		"prd":        "prd",
		"prod":       "prd",
		"production": "prd",
	}

	unique := map[string]struct{}{}
	var normalized []string
	for _, raw := range values {
		key := strings.ToLower(strings.TrimSpace(raw))
		env, ok := allowed[key]
		if !ok {
			return nil, fmt.Errorf("ambiente invalido: %q (use dev, qld ou prd)", raw)
		}
		if _, seen := unique[env]; seen {
			continue
		}
		unique[env] = struct{}{}
		normalized = append(normalized, env)
	}

	return normalized, nil
}

func askForTargetEnvironments() []string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("🌍 Seleção de ambientes para update\n")
	fmt.Printf("   Atualizar todos os ambientes (dev, qld, prd)? (s/n): ")

	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "s" || answer == "sim" || answer == "y" || answer == "yes" {
		return []string{"dev", "qld", "prd"}
	}

	environments := []string{"dev", "qld", "prd"}
	var selected []string
	for _, env := range environments {
		fmt.Printf("   Atualizar %s? (s/n): ", env)
		ans, _ := reader.ReadString('\n')
		ans = strings.TrimSpace(strings.ToLower(ans))
		if ans == "s" || ans == "sim" || ans == "y" || ans == "yes" {
			selected = append(selected, env)
		}
	}

	if len(selected) == 0 {
		fmt.Printf("   ⚠️  Nenhum ambiente selecionado. Usando todos por padrão.\n\n")
		return []string{"dev", "qld", "prd"}
	}

	fmt.Printf("\n")
	return selected
}

func filterProjectEnvironments(project *models.GCPProject, selectedEnvs []string) *models.GCPProject {
	selected := map[string]struct{}{}
	for _, env := range selectedEnvs {
		selected[env] = struct{}{}
	}

	filtered := &models.GCPProject{
		Name: project.Name,
		Dev:  &models.GCPEnvironment{Name: "dev"},
		Qld:  &models.GCPEnvironment{Name: "qld"},
		Prd:  &models.GCPEnvironment{Name: "prd"},
	}

	if _, ok := selected["dev"]; ok {
		filtered.Dev = project.Dev
	}
	if _, ok := selected["qld"]; ok {
		filtered.Qld = project.Qld
	}
	if _, ok := selected["prd"]; ok {
		filtered.Prd = project.Prd
	}

	return filtered
}

func countSelectedProjects(project *models.GCPProject) int {
	count := 0
	if project.Dev != nil && project.Dev.ProjectID != "" {
		count++
	}
	if project.Qld != nil && project.Qld.ProjectID != "" {
		count++
	}
	if project.Prd != nil && project.Prd.ProjectID != "" {
		count++
	}
	return count
}

func normalizeOptionalAPIs(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	unique := map[string]struct{}{}
	allowed := map[string]string{
		"artifactregistry":                "artifactregistry.googleapis.com",
		"artifact-registry":               "artifactregistry.googleapis.com",
		"artifactregistry.googleapis.com": "artifactregistry.googleapis.com",
		"secretmanager":                   "secretmanager.googleapis.com",
		"secret-manager":                  "secretmanager.googleapis.com",
		"secretmanager.googleapis.com":    "secretmanager.googleapis.com",
		"firestore":                       "firestore.googleapis.com",
		"firestore.googleapis.com":        "firestore.googleapis.com",
	}

	var normalized []string
	for _, raw := range values {
		key := strings.ToLower(strings.TrimSpace(raw))
		apiID, ok := allowed[key]
		if !ok {
			return nil, fmt.Errorf("api opcional invalida: %q (use artifactregistry, secretmanager ou firestore)", raw)
		}
		if _, seen := unique[apiID]; seen {
			continue
		}
		unique[apiID] = struct{}{}
		normalized = append(normalized, apiID)
	}

	return normalized, nil
}

func resolveOptionalAPIsForCommand(runAll bool, allOptionalAPIs bool, interactiveAPIs bool, optionalAPIs []string) ([]string, error) {
	if allOptionalAPIs {
		return gcp.AvailableOptionalAPIs(), nil
	}

	if runAll && !interactiveAPIs && len(optionalAPIs) == 0 {
		return gcp.AvailableOptionalAPIs(), nil
	}

	return normalizeOptionalAPIs(optionalAPIs)
}
