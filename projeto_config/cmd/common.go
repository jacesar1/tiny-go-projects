package cmd

import (
	"fmt"

	"github.com/spf13/viper"
)

func printExecutionHeader(projectName string) {
	fmt.Printf("╔═══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║     Automacao de Criacao de Projetos GCP - Axia Energia  ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Configuracoes:\n")
	fmt.Printf("  Nome do projeto: %s\n", projectName)
	fmt.Printf("  ID da Org: %s\n", viper.GetString("org-id"))
	fmt.Printf("  Pasta pai: %s\n", viper.GetString("parent-folder"))
	fmt.Printf("  Billing Account: %s\n\n", viper.GetString("billing-account"))
}
