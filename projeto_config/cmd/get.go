package cmd

import (
	"fmt"
	"strings"

	"projeto_config/internal/gcp"
	"projeto_config/internal/models"

	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Obtem informacoes de recursos existentes",
	}

	getProjectCmd := &cobra.Command{
		Use:     "projeto <nome-do-projeto>",
		Aliases: []string{"project"},
		Short:   "Lista informacoes basicas de um projeto",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			project, err := gcp.LoadExistingProject(projectName)
			if err != nil {
				return fmt.Errorf("erro ao carregar projeto: %w", err)
			}

			// Exibir resumo dos ambientes
			fmt.Printf("PROJETO: %s\n", project.Name)
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("%-8s %-30s %-20s\n", "ENV", "PROJECT ID", "FOLDER ID")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

			environments := []struct {
				name string
				env  *models.GCPEnvironment
			}{
				{"dev", project.Dev},
				{"qld", project.Qld},
				{"prd", project.Prd},
			}

			for _, item := range environments {
				projectID := item.env.ProjectID
				folderID := item.env.FolderID

				if projectID == "" {
					projectID = "(não encontrado)"
				}
				if folderID == "" {
					folderID = "(desconhecido)"
				}

				fmt.Printf("%-8s %-30s %-20s\n", strings.ToUpper(item.name), projectID, folderID)
			}

			fmt.Printf("\nUse 'projeto_config describe projeto %s' para ver detalhes completos.\n", projectName)
			return nil
		},
	}

	getCmd.AddCommand(getProjectCmd)
	return getCmd
}
