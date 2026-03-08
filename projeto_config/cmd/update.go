package cmd

import (
	"fmt"
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
	var optionalAPIs []string

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza recursos de um projeto existente",
	}

	updateProjectCmd := &cobra.Command{
		Use:     "projeto <nome-do-projeto>",
		Aliases: []string{"project"},
		Short:   "Executa passos de atualizacao (labels/apis/networks/service-accounts)",
		Long: `Exemplos:
  projeto_config update projeto benner-cloud --labels
  projeto_config update projeto benner-cloud --networks
  projeto_config update projeto benner-cloud --service-accounts
  projeto_config update projeto benner-cloud --apis --optional-api secretmanager --optional-api firestore
  projeto_config update projeto benner-cloud --apis --optional-api secretmanager,firestore
  projeto_config update projeto benner-cloud --all`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			printExecutionHeader(projectName)

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

			if runLabels {
				if err := gcp.Step2AddLabels(project); err != nil {
					return fmt.Errorf("erro no passo labels: %w", err)
				}
			}

			if runAPIs {
				apiOptions := gcp.Step3Options{
					Interactive: interactiveAPIs,
				}

				if allOptionalAPIs {
					apiOptions.OptionalAPIs = gcp.AvailableOptionalAPIs()
				} else {
					normalized, err := normalizeOptionalAPIs(optionalAPIs)
					if err != nil {
						return err
					}
					apiOptions.OptionalAPIs = normalized
				}

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

	updateCmd.AddCommand(updateProjectCmd)
	return updateCmd
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
