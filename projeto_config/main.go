package main

import (
	"flag"
	"fmt"
	"log"
	"projeto_config/internal/gcp"
	"projeto_config/internal/models"
)

func main() {
	// Definir flags de linha de comando
	projectName := flag.String("project", "", "Nome do projeto (obrigatório)")
	orgID := flag.String("org-id", "727440331682", "ID da organização")
	parentFolder := flag.String("parent-folder", "fldr-scge", "ID da pasta pai")
	step := flag.Int("step", 1, "Qual passo executar (1-4)")

	flag.Parse()

	// Validar argumentos obrigatórios
	if *projectName == "" {
		fmt.Println("Uso: projeto_config -project <nome> [-org-id <id>] [-parent-folder <id>] [-step <1-4>]")
		fmt.Println("\nExemplos:")
		fmt.Println("  projeto_config -project benner-cloud")
		fmt.Println("  projeto_config -project benner-cloud -step 1")
		fmt.Println("  projeto_config -project benner-cloud -step 2")
		flag.PrintDefaults()
		return
	}

	fmt.Printf("╔═══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║     Automação de Criação de Projetos GCP - Eletrobras    ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Configurações:\n")
	fmt.Printf("  Nome do projeto: %s\n", *projectName)
	fmt.Printf("  ID da Org: %s\n", *orgID)
	fmt.Printf("  Pasta pai: %s\n", *parentFolder)
	fmt.Printf("  Passo a executar: %d\n\n", *step)

	// Validar passo
	if *step < 1 || *step > 4 {
		log.Fatal("Passo deve estar entre 1 e 4")
	}

	// Criar configuração
	config := &models.ProjectConfig{
		ProjectName:    *projectName,
		OrgID:          *orgID,
		ParentFolderID: *parentFolder,
	}

	// Executar os passos solicitados
	var gcpProject *models.GCPProject

	// Passo 1: Criar pastas
	if *step >= 1 {
		project, err := gcp.Step1CreateFolderStructure(config)
		if err != nil {
			log.Fatalf("❌ Erro no Passo 1: %v", err)
		}
		gcpProject = project
		gcp.PrintProjectStructure(gcpProject)
	}

	// Passo 2: Adicionar labels
	if *step >= 2 {
		// TODO: Implementar passo 2
		fmt.Println("Passo 2 ainda não implementado")
	}

	// Passo 3: Habilitar APIs
	if *step >= 3 {
		// TODO: Implementar passo 3
		fmt.Println("Passo 3 ainda não implementado")
	}

	// Passo 4: Atachar nas redes spokes
	if *step >= 4 {
		// TODO: Implementar passo 4
		fmt.Println("Passo 4 ainda não implementado")
	}

	fmt.Printf("✅ Todos os passos foram executados com sucesso!\n")
}
