package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"projeto_config/internal/gcp"

	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	var skipConfirm bool

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Deleta recursos do GCP",
	}

	deleteProjectCmd := &cobra.Command{
		Use:     "projeto <nome-do-projeto>",
		Aliases: []string{"project"},
		Short:   "Deleta completamente um projeto (shutdown + pastas)",
		Long: `Deleta completamente a estrutura de um projeto criada no passo 1.

Esta operação irá:
  1. Fazer shutdown dos projetos GCP (dev, qld, prd)
  2. Deletar as pastas de ambiente (fldr-dev, fldr-qld, fldr-prd)
  3. Deletar a pasta principal (fldr-<nome-do-projeto>)

Exemplos:
  projeto_config delete projeto benner-cloud
  projeto_config delete projeto benner-cloud --yes  # sem confirmacao`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// Carregar projeto existente
			project, err := gcp.LoadExistingProject(projectName)
			if err != nil {
				return fmt.Errorf("erro ao carregar projeto: %w", err)
			}

			// Confirmação interativa (a menos que --yes seja usado)
			if !skipConfirm {
				fmt.Printf("⚠️  ATENÇÃO: Você está prestes a deletar o projeto '%s'\n\n", projectName)
				fmt.Printf("Esta ação irá:\n")
				fmt.Printf("  • Deletar os projetos GCP:\n")
				if project.Dev.ProjectID != "" {
					fmt.Printf("    - %s\n", project.Dev.ProjectID)
				}
				if project.Qld.ProjectID != "" {
					fmt.Printf("    - %s\n", project.Qld.ProjectID)
				}
				if project.Prd.ProjectID != "" {
					fmt.Printf("    - %s\n", project.Prd.ProjectID)
				}
				fmt.Printf("  • Deletar as pastas e subpastas no Resource Manager\n\n")
				fmt.Printf("Esta operação é IRREVERSÍVEL.\n\n")
				fmt.Printf("Digite o nome do projeto para confirmar: ")

				reader := bufio.NewReader(os.Stdin)
				confirmation, _ := reader.ReadString('\n')
				confirmation = strings.TrimSpace(confirmation)

				if confirmation != projectName {
					return fmt.Errorf("confirmação não coincide. Operação cancelada")
				}

				fmt.Println()
			}

			// Executar deleção
			if err := gcp.StepDeleteProject(project); err != nil {
				return fmt.Errorf("erro durante deleção: %w", err)
			}

			fmt.Printf("✅ Projeto '%s' deletado com sucesso.\n", projectName)
			return nil
		},
	}

	deleteProjectCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Pular confirmação interativa")

	deleteCmd.AddCommand(deleteProjectCmd)
	return deleteCmd
}
