package cmd

import (
	"fmt"

	"projeto_config/internal/gcp"
	"projeto_config/internal/models"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCreateCommand() *cobra.Command {
	var runAll bool
	var allOptionalAPIs bool
	var interactiveAPIs bool
	var optionalAPIs []string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Cria recursos no GCP",
	}

	createProjectCmd := &cobra.Command{
		Use:     "projeto <nome-do-projeto>",
		Aliases: []string{"project"},
		Short:   "Cria a estrutura base (passo 1)",
		Long: `Exemplos:
  projeto_config create projeto benner-cloud
  projeto_config create projeto benner-cloud --all
  projeto_config create projeto benner-cloud --all --optional-api secretmanager,firestore
  projeto_config create projeto benner-cloud --all --all-optional-apis`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			printExecutionHeader(projectName)

			if !runAll && (allOptionalAPIs || interactiveAPIs || len(optionalAPIs) > 0) {
				return fmt.Errorf("as flags de API opcional so podem ser usadas com --all no create")
			}

			config := &models.ProjectConfig{
				ProjectName:      projectName,
				OrgID:            viper.GetString("org-id"),
				ParentFolderID:   viper.GetString("parent-folder"),
				BillingAccountID: viper.GetString("billing-account"),
			}

			gcpProject, err := gcp.Step1CreateFolderStructure(config)
			if err != nil {
				return fmt.Errorf("erro no passo 1 (create projeto): %w", err)
			}

			if runAll {
				if err := gcp.Step2AddLabels(gcpProject); err != nil {
					return fmt.Errorf("erro no passo 2 (labels): %w", err)
				}

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

				if err := gcp.Step3EnableAPIsWithOptions(gcpProject, apiOptions); err != nil {
					return fmt.Errorf("erro no passo 3 (apis): %w", err)
				}

				if err := gcp.Step4AttachToNetworks(gcpProject); err != nil {
					return fmt.Errorf("erro no passo 4 (networks): %w", err)
				}
			}

			gcp.PrintProjectStructure(gcpProject)
			if runAll {
				fmt.Println("✅ Create concluido com sucesso (passos 1, 2, 3 e 4).")
			} else {
				fmt.Println("✅ Create concluido com sucesso (passo 1).")
			}
			return nil
		},
	}

	createProjectCmd.Flags().BoolVar(&runAll, "all", false, "Executa create completo (passos 1, 2, 3 e 4)")
	createProjectCmd.Flags().StringSliceVar(&optionalAPIs, "optional-api", nil, "API opcional para o passo 3 (artifactregistry|secretmanager|firestore). Aceita virgula")
	createProjectCmd.Flags().BoolVar(&allOptionalAPIs, "all-optional-apis", false, "Inclui todas as APIs opcionais no passo 3")
	createProjectCmd.Flags().BoolVar(&interactiveAPIs, "interactive-apis", false, "Pergunta interativamente APIs opcionais no passo 3")

	createCmd.AddCommand(createProjectCmd)
	return createCmd
}
