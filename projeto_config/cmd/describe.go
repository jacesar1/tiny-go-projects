package cmd

import (
	"fmt"

	"projeto_config/internal/gcp"

	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Exibe informacoes detalhadas de recursos",
	}

	describeProjectCmd := &cobra.Command{
		Use:     "projeto <nome-do-projeto>",
		Aliases: []string{"project"},
		Short:   "Exibe informacoes detalhadas de todos os ambientes do projeto",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			project, err := gcp.LoadExistingProject(projectName)
			if err != nil {
				return fmt.Errorf("erro ao carregar projeto: %w", err)
			}

			if err := gcp.DescribeProject(project); err != nil {
				return fmt.Errorf("erro ao descrever projeto: %w", err)
			}

			return nil
		},
	}

	describeCmd.AddCommand(describeProjectCmd)
	return describeCmd
}
