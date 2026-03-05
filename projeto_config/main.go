package main

import (
	"flag"
	"fmt"
	"log"
	"projeto_config/internal/gcp"
	"projeto_config/internal/models"
)

const (
	// Billing Account da Eletrobras - vinculado automaticamente a todos os projetos
	DefaultBillingAccountID = "01F7C9-60D131-20DC44"
)

func main() {
	// Definir flags de linha de comando
	projectName := flag.String("project", "", "Nome do projeto (obrigatório)")
	orgID := flag.String("org-id", "727440331682", "ID da organização")
	parentFolder := flag.String("parent-folder", "fldr-scge", "ID da pasta pai")
	step := flag.Int("step", 0, "Qual passo executar (1-4). Se omitido (0), executa todos")

	flag.Parse()

	// Validar argumentos obrigatórios
	if *projectName == "" {
		fmt.Println("Uso: projeto_config -project <nome> [-org-id <id>] [-parent-folder <id>] [-step <1-4>]")
		fmt.Println("\nExemplos:")
		fmt.Println("  projeto_config -project benner-cloud          # Executa todos os passos (1-4)")
		fmt.Println("  projeto_config -project benner-cloud -step 1  # Criar estrutura apenas")
		fmt.Println("  projeto_config -project benner-cloud -step 2  # Adicionar labels apenas")
		fmt.Println("  projeto_config -project benner-cloud -step 3  # Habilitar APIs apenas")
		fmt.Println("  projeto_config -project benner-cloud -step 4  # Atachar nas redes apenas")
		fmt.Println("\nNota: Billing Account é vinculado automaticamente (01F7C9-60D131-20DC44)")
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
	fmt.Printf("  Billing Account: %s (automático)\n", DefaultBillingAccountID)
	if *step == 0 {
		fmt.Printf("  Passos a executar: 1, 2, 3, 4 (todos)\n\n")
	} else {
		fmt.Printf("  Passo a executar: %d\n\n", *step)
	}

	// Validar passo
	if *step < 0 || *step > 4 {
		log.Fatal("Passo deve estar entre 0 (todos) e 4")
	}

	// Criar configuração
	config := &models.ProjectConfig{
		ProjectName:      *projectName,
		OrgID:            *orgID,
		ParentFolderID:   *parentFolder,
		BillingAccountID: DefaultBillingAccountID,
	}

	// Executar os passos solicitados
	var gcpProject *models.GCPProject
	var err error

	// Passo 1: Criar pastas
	if *step == 0 || *step == 1 {
		gcpProject, err = gcp.Step1CreateFolderStructure(config)
		if err != nil {
			log.Fatalf("❌ Erro no Passo 1: %v", err)
		}
		gcp.PrintProjectStructure(gcpProject)
	}

	// Passo 2: Adicionar labels
	if *step == 0 || *step == 2 {
		// Carregar dados dos projetos existentes (se não foi executado passo 1)
		if gcpProject == nil {
			gcpProject, err = gcp.LoadExistingProject(*projectName)
			if err != nil {
				log.Fatalf("❌ Erro ao carregar projeto: %v", err)
			}
		}

		if err := gcp.Step2AddLabels(gcpProject); err != nil {
			log.Fatalf("❌ Erro no Passo 2: %v", err)
		}
	}

	// Passo 3: Habilitar APIs
	if *step == 0 || *step == 3 {
		// Carregar dados dos projetos existentes (se não foi executado passo 1)
		if gcpProject == nil {
			gcpProject, err = gcp.LoadExistingProject(*projectName)
			if err != nil {
				log.Fatalf("❌ Erro ao carregar projeto: %v", err)
			}
		}

		if err := gcp.Step3EnableAPIs(gcpProject); err != nil {
			log.Fatalf("❌ Erro no Passo 3: %v", err)
		}
	}

	// Passo 4: Atachar nas redes spokes
	if *step == 0 || *step == 4 {
		// Carregar dados dos projetos existentes (se não foi executado passo 1)
		if gcpProject == nil {
			gcpProject, err = gcp.LoadExistingProject(*projectName)
			if err != nil {
				log.Fatalf("❌ Erro ao carregar projeto: %v", err)
			}
		}

		if err := gcp.Step4AttachToNetworks(gcpProject); err != nil {
			log.Fatalf("❌ Erro no Passo 4: %v", err)
		}
	}

	if *step == 0 {
		fmt.Printf("✅ Todos os passos (1-4) foram executados com sucesso!\n")
	} else {
		fmt.Printf("✅ Passo %d executado com sucesso!\n", *step)
	}
}
