package gcp

import (
	"encoding/json"
	"fmt"
	"projeto_config/internal/models"
	"strings"
)

// ProjectDetails contém informações detalhadas de um projeto
type ProjectDetails struct {
	ProjectID      string
	Name           string
	Number         string
	State          string
	Labels         map[string]string
	Parent         string
	CreateTime     string
	BillingAccount string
	EnabledAPIs    []string
	SharedVPCHost  string
}

// GetProjectDetails obtém informações detalhadas de um projeto
func GetProjectDetails(projectID string) (*ProjectDetails, error) {
	details := &ProjectDetails{
		ProjectID: projectID,
		Labels:    make(map[string]string),
	}

	// Obter informações básicas do projeto
	projectInfo, err := GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	// Extrair campos básicos
	if name, ok := projectInfo["name"].(string); ok {
		details.Name = name
	}
	if number, ok := projectInfo["projectNumber"].(string); ok {
		details.Number = number
	}
	if state, ok := projectInfo["lifecycleState"].(string); ok {
		details.State = state
	}
	if createTime, ok := projectInfo["createTime"].(string); ok {
		details.CreateTime = createTime
	}

	// Extrair labels
	if labelsRaw, ok := projectInfo["labels"].(map[string]interface{}); ok {
		for k, v := range labelsRaw {
			if vStr, ok := v.(string); ok {
				details.Labels[k] = vStr
			}
		}
	}

	// Extrair parent
	if parent, ok := projectInfo["parent"].(map[string]interface{}); ok {
		if parentType, ok := parent["type"].(string); ok {
			if parentID, ok := parent["id"].(string); ok {
				details.Parent = fmt.Sprintf("%s/%s", parentType, parentID)
			}
		}
	}

	// Obter billing account
	billingAccount, _ := GetProjectBillingAccount(projectID)
	details.BillingAccount = billingAccount

	// Obter APIs habilitadas
	enabledAPIs, _ := GetEnabledAPIs(projectID)
	details.EnabledAPIs = enabledAPIs

	// Obter status de Shared VPC
	sharedVPCHost, _ := GetSharedVPCStatus(projectID)
	details.SharedVPCHost = sharedVPCHost

	return details, nil
}

// GetEnabledAPIs retorna lista de APIs habilitadas em um projeto
func GetEnabledAPIs(projectID string) ([]string, error) {
	cmd := newGCloudCommand("services", "list", "--enabled", "--project="+projectID, "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erro ao listar APIs: %v", err)
	}

	var services []map[string]interface{}
	if err := json.Unmarshal(output, &services); err != nil {
		return nil, fmt.Errorf("erro ao parsear APIs: %v", err)
	}

	var apis []string
	for _, svc := range services {
		if name, ok := svc["config"].(map[string]interface{})["name"].(string); ok {
			apis = append(apis, name)
		}
	}

	return apis, nil
}

// DescribeProject exibe informações detalhadas de todos os ambientes
func DescribeProject(project *models.GCPProject) error {
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("DESCRIÇÃO DETALHADA: %s\n", project.Name)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Validar autenticação
	if err := ValidateAuthentication(); err != nil {
		return err
	}

	account, _ := GetCurrentAccount()
	fmt.Printf("✓ Autenticado como: %s\n\n", account)

	environments := []*models.GCPEnvironment{project.Dev, project.Qld, project.Prd}

	for _, env := range environments {
		if env.ProjectID == "" {
			fmt.Printf("⚠️  Ambiente %s: Project ID não disponível\n\n", env.Name)
			continue
		}

		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("📊 AMBIENTE: %s\n", strings.ToUpper(env.Name))
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		details, err := GetProjectDetails(env.ProjectID)
		if err != nil {
			fmt.Printf("❌ Erro ao obter detalhes: %v\n\n", err)
			continue
		}

		// Informações básicas
		fmt.Printf("🏷️  IDENTIFICAÇÃO\n")
		fmt.Printf("   Project ID:     %s\n", details.ProjectID)
		fmt.Printf("   Display Name:   %s\n", details.Name)
		fmt.Printf("   Project Number: %s\n", details.Number)
		fmt.Printf("   Estado:         %s\n", details.State)
		fmt.Printf("   Criado em:      %s\n", details.CreateTime)
		fmt.Printf("\n")

		// Hierarquia
		fmt.Printf("📁 HIERARQUIA\n")
		fmt.Printf("   Parent:    %s\n", details.Parent)
		fmt.Printf("   Folder ID: %s\n", env.FolderID)
		fmt.Printf("\n")

		// Billing
		fmt.Printf("💳 BILLING\n")
		if details.BillingAccount != "" {
			fmt.Printf("   Billing Account: %s\n", details.BillingAccount)
		} else {
			fmt.Printf("   Billing Account: Não vinculada\n")
		}
		fmt.Printf("\n")

		// Labels
		fmt.Printf("🏷️  LABELS\n")
		if len(details.Labels) > 0 {
			for k, v := range details.Labels {
				fmt.Printf("   %s: %s\n", k, v)
			}
		} else {
			fmt.Printf("   (nenhum label configurado)\n")
		}
		fmt.Printf("\n")

		// APIs habilitadas
		fmt.Printf("🔌 APIs HABILITADAS (%d)\n", len(details.EnabledAPIs))
		if len(details.EnabledAPIs) > 0 {
			// Mostrar apenas as principais
			importantAPIs := []string{
				"compute.googleapis.com",
				"servicenetworking.googleapis.com",
				"artifactregistry.googleapis.com",
				"secretmanager.googleapis.com",
				"firestore.googleapis.com",
			}

			for _, api := range importantAPIs {
				found := false
				for _, enabled := range details.EnabledAPIs {
					if enabled == api {
						found = true
						break
					}
				}
				status := "❌"
				if found {
					status = "✓"
				}
				fmt.Printf("   %s %s\n", status, api)
			}
		} else {
			fmt.Printf("   (nenhuma API habilitada)\n")
		}
		fmt.Printf("\n")

		// Shared VPC
		fmt.Printf("🔗 SHARED VPC\n")
		if details.SharedVPCHost != "" {
			fmt.Printf("   Host Project: %s\n", details.SharedVPCHost)
		} else {
			fmt.Printf("   (não atachado a Shared VPC)\n")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("✓ Descrição concluída\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	return nil
}
