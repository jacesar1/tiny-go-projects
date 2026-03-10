package models

// ProjectConfig contém as configurações para criação de um novo projeto GCP
type ProjectConfig struct {
	// Nome do projeto (ex: benner-cloud)
	ProjectName string
	// ID da organização (ex: 727440331682)
	OrgID string
	// ID da pasta pai onde criar as subpastas (ex: fldr-scge)
	ParentFolderID string
	// ID da conta de billing (ex: 01F7C9-60D131-20DC44) - opcional
	BillingAccountID string
	// Ambientes alvo para create/update parcial (ex: dev, qld, prd)
	TargetEnvironments []string
}

// GCPEnvironment representa um ambiente (dev, qld, prd)
type GCPEnvironment struct {
	Name      string // dev, qld, prd
	FolderID  string // ID da pasta criada
	ProjectID string // ID do projeto criado
}

// GCPProject contém todas as informações de um projeto GCP
type GCPProject struct {
	Name string
	Dev  *GCPEnvironment
	Qld  *GCPEnvironment
	Prd  *GCPEnvironment
}
