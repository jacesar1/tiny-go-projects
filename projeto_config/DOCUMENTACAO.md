# 📋 Documentação Completa - Automação de Criação de Projetos GCP

**Automação de Criação de Projetos GCP - Axia Energia**

## 📑 Índice

1. [Visão Geral](#-visão-geral)
2. [Fluxo Geral do Programa](#-fluxo-geral-do-programa)
3. [Estrutura de Dados](#-estrutura-de-dados)
4. [Execução dos Passos](#-execução-dos-passos)
5. [Funções e Comandos GCloud](#-funções-e-comandos-gcloud)

---

## 🎯 Visão Geral

O programa `projeto_config` automatiza todo o ciclo de criação de projetos GCP para a Axia Energia, incluindo:

- ✅ Criação de estrutura de pastas (Resource Manager)
- ✅ Adição de labels de identificação
- ✅ Habilitação de APIs necessárias
- ✅ Vínculo a redes Shared VPC
- ✅ Criação de service accounts e roles
- ✅ Armazenamento de chaves em Secret Manager

**Localização:** `main.go` → Orquestra chamadas para `internal/gcp/*.go`

---

## 🔄 Fluxo Geral do Programa

```
┌─────────────────────────────────────────────────────────────┐
│ main.go                                                     │
│ ┌───────────────────────────────────────────────────────┐   │
│ │ 1. Parsear flags: -project, -org-id, -parent-folder  │   │
│ │ 2. Validar argumentos obrigatórios                    │   │
│ │ 3. Criar objeto ProjectConfig                         │   │
│ │ 4. Executar passos (Step 1-5)                         │   │
│ └───────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────────────────┐
│ Lógica dos Passos (internal/gcp)                            │
│                                                             │
│ [Step 1] → [Step 2] → [Step 3] → [Step 4] → [Step 5]       │
│ (folders) (labels)   (APIs)    (Shared VPC) (Service Accs) │
└─────────────────────────────────────────────────────────────┘
```

**Execução:**
- `-step 0` (padrão): Executa passos 1, 2, 3, 4
- `-step 1-5`: Executa apenas o passo especificado
- Cada passo pode ser rodado independentemente (carrega dados existentes)

---

## 📊 Estrutura de Dados

### ProjectConfig

```go
type ProjectConfig struct {
    ProjectName      string  // Nome do projeto (ex: "benner-cloud")
    OrgID           string  // ID da organização GCP (ex: "727440331682")
    ParentFolderID  string  // ID ou nome da pasta pai (ex: "fldr-scge")
    BillingAccountID string // ID da conta de billing (ex: "01F7C9-60D131-20DC44")
}
```

### GCPProject

```go
type GCPProject struct {
    Name string                    // Nome do projeto
    Dev  *GCPEnvironment          // Ambiente de desenvolvimento
    Qld  *GCPEnvironment          // Ambiente de qualidade
    Prd  *GCPEnvironment          // Ambiente de produção
}
```

### GCPEnvironment

```go
type GCPEnvironment struct {
    Name      string  // "dev", "qld" ou "prd"
    ProjectID string  // ID do projeto (ex: "elet-benner-cloud-dev")
    FolderID  string  // ID da pasta no Resource Manager
}
```

---

## 🚀 Execução dos Passos

### Comando de Uso

```bash
# Executar todos os passos (1-4)
./projeto_config -project benner-cloud

# Executar passo específico
./projeto_config -project benner-cloud -step 1

# Com argumentos adicionais
./projeto_config -project benner-cloud -org-id 727440331682 -parent-folder fldr-scge -step 2
```

---

## 🔧 Funções e Comandos GCloud

### 📌 PASSO 1: Criar Estrutura de Pastas

**Função:** `Step1CreateFolderStructure(config *ProjectConfig)` 
**Localização:** `internal/gcp/step1_folders.go`
**Retorno:** `*GCPProject, error`

#### Estrutura Criada

```
Organização (ID: 727440331682)
│
└── fldr-<projeto> (pasta principal)
    │
    ├── fldr-dev (pasta ambiente)
    │   └── elet-<projeto>-dev (projeto GCP)
    │
    ├── fldr-qld (pasta ambiente)
    │   └── elet-<projeto>-qld (projeto GCP)
    │
    └── fldr-prd (pasta ambiente)
        └── elet-<projeto>-prd (projeto GCP)
```

#### Passos Executados

1. **Validar autenticação GCP**
   - Função: `ValidateAuthentication()` (client.go)
   - Verifica se usuário está autenticado

2. **Resolver ID da pasta pai**
   - Se input é numérico → usa diretamente
   - Se input é nome → resolve para ID
   - Função: `FindFolderIDByNameInOrg()` (folders.go)
   - **Comando GCloud:**
     ```bash
     gcloud resource-manager folders list --organization=<org-id> --format=json
     ```

3. **Criar pasta principal**
   - Função: `CreateFolder(parentFolderID, displayName)` (folders.go)
   - **Comando GCloud:**
     ```bash
     gcloud resource-manager folders create \
       --display-name=fldr-<projeto> \
       --folder=<parent-folder-id> \
       --format=json
     ```

4. **Para cada ambiente (dev, qld, prd):**

   **a) Criar pasta de ambiente**
   - Função: `CreateFolder()` (folders.go)
   - **Comando GCloud:**
     ```bash
     gcloud resource-manager folders create \
       --display-name=fldr-<ambiente> \
       --folder=<main-folder-id> \
       --format=json
     ```

   **b) Criar projeto GCP**
   - Função: `CreateProject(projectID, displayName, folderID)` (projects.go)
   - **Comando GCloud:**
     ```bash
     gcloud projects create elet-<projeto>-<ambiente> \
       --name="<projeto> - <ambiente>" \
       --folder=<env-folder-id> \
       --format=json
     ```

   **c) Vincular Billing Account**
   - Função: `LinkBillingAccount(projectID, billingAccountID)` (billing.go)
   - **Comando GCloud:**
     ```bash
     gcloud billing projects link elet-<projeto>-<ambiente> \
       --billing-account=01F7C9-60D131-20DC44
     ```

5. **Exibir estrutura criada**
   - Função: `PrintProjectStructure(project)` (step1_folders.go)
   - Mostra resumo visual da estrutura

---

### 📌 PASSO 2: Adicionar Labels

**Função:** `Step2AddLabels(project *GCPProject)` 
**Localização:** `internal/gcp/step2_labels.go`
**Retorno:** `error`

#### Labels Adicionados

| Label | Valor |
|-------|-------|
| `ambiente` | dev \| qld \| prd |
| `companhia` | elet |
| `projeto` | \<nome-do-projeto> |

#### Passos Executados

1. **Validar autenticação**

2. **Para cada projeto (dev, qld, prd):**
   - Função: `SetProjectLabels(projectID, labels)` (projects.go)
   - **Comando GCloud:**
     ```bash
     gcloud alpha projects update elet-<projeto>-<ambiente> \
       --update-labels=ambiente=<ambiente>,companhia=elet,projeto=<projeto>
     ```

---

### 📌 PASSO 3: Habilitar APIs

**Função:** `Step3EnableAPIs(project *GCPProject)` 
**Localização:** `internal/gcp/step3_apis.go`
**Retorno:** `error`

#### APIs Habilitadas

**Obrigatórias:**
- `compute.googleapis.com` → Compute Engine
- `servicenetworking.googleapis.com` → Service Networking

**Opcionais (perguntar ao usuário):**
- `artifactregistry.googleapis.com` → Artifact Registry
- `secretmanager.googleapis.com` → Secret Manager
- `firestore.googleapis.com` → Firestore

#### Passos Executados

1. **Validar autenticação**

2. **Perguntar sobre APIs opcionais**
   - Função: `askForOptionalAPIs()` (step3_apis.go)
   - Ler entrada do usuário para cada API

3. **Para cada projeto (dev, qld, prd):**
   - Função: `EnableAPI(projectID, apiName)` (apis.go)
   - **Comando GCloud:**
     ```bash
     gcloud services enable <api-name> \
       --project=elet-<projeto>-<ambiente>
     ```

---

### 📌 PASSO 4: Atachar a Shared VPC

**Função:** `Step4AttachToNetworks(project *GCPProject)` 
**Localização:** `internal/gcp/step4_networks.go`
**Retorno:** `error`

#### Mapeamento de Ambientes para Host Projects

| Ambiente | Host Project | VPC |
|----------|--------------|-----|
| dev | redes-spoke-dev-002b | vpc-spoke-dev |
| qld | redes-spoke-qld-7e83 | vpc-spoke-qld |
| prd | redes-spoke-prd-bd15 | vpc-spoke-prd |

#### Passos Executados

1. **Validar autenticação**

2. **Para cada projeto (dev, qld, prd):**
   - Função: `AttachToSharedVPC(serviceProject, hostProject)` (networks.go)
   - **Comando GCloud:**
     ```bash
     gcloud compute shared-vpc associated-projects add \
       elet-<projeto>-<ambiente> \
       --host-project=redes-spoke-<ambiente>-<id>
     ```

---

### 📌 PASSO 5: Criar Service Accounts e Roles

**Função:** `Step5CreateServiceAccounts(project *GCPProject)` 
**Localização:** `internal/gcp/step5_service_accounts.go`
**Retorno:** `error` (com defer para garantir reset de policies)

#### Service Accounts Criadas

**Por Projeto (3 ambientes × 2 contas = 6 contas):**

1. **GitLab Pipeline SA**
   - ID: `sa-<projeto>-git`
   - Email: `sa-<projeto>-git@elet-<projeto>-<ambiente>.iam.gserviceaccount.com`
   - Role: `roles/artifactregistry.createOnPushWriter`
   - Arquivo JSON: `sa-<projeto>-git-<ambiente>.json`

2. **GSA (Google Service Account)**
   - ID: `sa-<projeto>-<ambiente>`
   - Email: `sa-<projeto>-<ambiente>@elet-<projeto>-<ambiente>.iam.gserviceaccount.com`
   - Roles: `customRole_SA_<projeto>` + `roles/secretmanager.viewer`
   - Arquivo JSON: `sa-<projeto>-<ambiente>.json`

#### Custom Role

- **ID:** `customRole_SA_<projeto>` (hífens do nome → underscores)
- **Permissões (11 total):**
  ```
  artifactregistry.repositories.downloadArtifacts
  autoscaling.sites.writeMetrics
  datastore.entities.get
  datastore.entities.list
  datastore.entities.update
  datastore.entities.create
  logging.logEntries.create
  monitoring.dashboards.get
  monitoring.timeSeries.create
  pubsub.subscriptions.consume
  pubsub.topics.publish
  ```

#### Fase 1: Criar Service Accounts e Roles

**Para cada ambiente (dev, qld, prd):**

1. **Criar Service Account GitLab**
   - Função: `CreateServiceAccount(projectID, accountID, displayName)` (service_accounts.go)
   - Verifica existência antes de criar (idempotente)
   - **Comando GCloud:**
     ```bash
     gcloud iam service-accounts create sa-<projeto>-git \
       --display-name="GitLab Pipeline Service Account" \
       --project=elet-<projeto>-<ambiente>
     ```

2. **Aguardar propagação da SA GitLab**
   - Função: `WaitForServiceAccount(projectID, accountID, maxAttempts, delay)` (service_accounts.go)
   - 5 tentativas com 2s de intervalo
   - **Comando GCloud (check):**
     ```bash
     gcloud iam service-accounts describe \
       sa-<projeto>-git@elet-<projeto>-<ambiente>.iam.gserviceaccount.com \
       --project=elet-<projeto>-<ambiente>
     ```

3. **Vincular Role ao GitLab SA**
   - Função: `AddProjectIamBinding(projectID, member, role)` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud projects add-iam-policy-binding elet-<projeto>-<ambiente> \
       --member=serviceAccount:sa-<projeto>-git@elet-<projeto>-<ambiente>.iam.gserviceaccount.com \
       --role=roles/artifactregistry.createOnPushWriter
     ```

4. **Criar Custom Role**
   - Função: `CreateCustomRole(projectID, roleID, title, description, permissions)` (service_accounts.go)
   - Ignora erro se role já existe
   - **Comando GCloud:**
     ```bash
     gcloud iam roles create customRole_SA_<projeto> \
       --project=elet-<projeto>-<ambiente> \
       --title="Custom Role SA <projeto>" \
       --description="Custom role para service account GSA" \
       --permissions=...permissões separadas por vírgula... \
       --stage=GA
     ```

5. **Criar Service Account GSA**
   - Função: `CreateServiceAccount()` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud iam service-accounts create sa-<projeto>-<ambiente> \
       --display-name="GSA Service Account" \
       --project=elet-<projeto>-<ambiente>
     ```

6. **Aguardar propagação da SA GSA**
   - Função: `WaitForServiceAccount()` (service_accounts.go)

7. **Vincular Custom Role ao GSA**
   - Função: `AddProjectIamBinding()` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud projects add-iam-policy-binding elet-<projeto>-<ambiente> \
       --member=serviceAccount:sa-<projeto>-<ambiente>@elet-<projeto>-<ambiente>.iam.gserviceaccount.com \
       --role=projects/elet-<projeto>-<ambiente>/roles/customRole_SA_<projeto>
     ```

8. **Vincular Secret Manager Viewer ao GSA**
   - Função: `AddProjectIamBinding()` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud projects add-iam-policy-binding elet-<projeto>-<ambiente> \
       --member=serviceAccount:sa-<projeto>-<ambiente>@elet-<projeto>-<ambiente>.iam.gserviceaccount.com \
       --role=roles/secretmanager.viewer
     ```

#### Fase 2: Desabilitar Org Policies

**Constraints a desabilitar:**
- `constraints/iam.disableServiceAccountKeyCreation`
- `constraints/iam.disableServiceAccountKeyUpload`

**Para cada projeto e constraint:**

1. **Desabilitar enforce**
   - Função: `DisableProjectOrgPolicyEnforce(projectID, constraint)` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud resource-manager org-policies disable-enforce \
       <constraint> \
       --project=elet-<projeto>-<ambiente>
     ```

2. **Aguardar propagação do desabilitar (enforced=false)**
   - Função: `WaitForPolicyEnforcementState(projectID, constraint, false, 18, 10s)` (service_accounts.go)
   - 18 tentativas com 10s de intervalo
   - Função helper: `IsProjectOrgPolicyEnforced()` → **Comando GCloud:**
     ```bash
     gcloud resource-manager org-policies describe \
       <constraint> \
       --project=elet-<projeto>-<ambiente> \
       --effective \
       --format=json
     ```
     Parse: `booleanPolicy.enforced` === `false`

#### Fase 3: Criar Chaves JSON

**Para cada projeto e conta (GitLab + GSA):**

1. **Verificar se chave já existe**
   - Função: `ServiceAccountHasUserManagedKeys(projectID, accountID)` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud iam service-accounts keys list \
       --iam-account=sa-<conta>@elet-<projeto>-<ambiente>.iam.gserviceaccount.com \
       --project=elet-<projeto>-<ambiente> \
       --format=json
     ```
     Verifica: existe `keyType === "USER_MANAGED"`

2. **Se chave NÃO existe:**
   - Criar com retry: `createKeyWithRetry()` (step5_service_accounts.go)
   - Função: `CreateServiceAccountKey(projectID, accountID, outputPath)` (service_accounts.go)
   - Máx 8 tentativas com 15s de intervalo
   - **Comando GCloud:**
     ```bash
     gcloud iam service-accounts keys create <pa-<projeto>-git-<ambiente>.json \
       --iam-account=sa-<projeto>-git@elet-<projeto>-<ambiente>.iam.gserviceaccount.com \
       --project=elet-<projeto>-<ambiente>
     ```

3. **Se chave JÁ existe E arquivo JSON local encontrado:**
   - Apenas atualizar secret (não criar nova chave)

4. **Armazenar JSON na Secret Manager**
   - Função: `StoreSecretFromFile(projectID, secretID, filePath)` (service_accounts.go)

   **4a) Garantir que secret existe**
   - Função: `EnsureSecretExists(projectID, secretID)` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud secrets describe sa-<conta> \
       --project=elet-<projeto>-<ambiente> \
       --format=json
     ```
     Se não existe:
     ```bash
     gcloud secrets create sa-<conta> \
       --project=elet-<projeto>-<ambiente> \
       --replication-policy=automatic
     ```

   **4b) Adicionar versão com o JSON criado**
   - Função: `AddSecretVersionFromFile(projectID, secretID, filePath)` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud secrets versions add sa-<conta> \
       --project=elet-<projeto>-<ambiente> \
       --data-file=sa-<projeto>-<ambiente>.json \
       --format=json
     ```
     Retorna: versão criada (ex: "3")

#### Fase 4: Restaurar Org Policies (sempre executa via defer)

**Para cada projeto e constraint:**

1. **Remover override da policy (voltar a herdar do parent)**
   - Função: `ResetProjectOrgPolicy(projectID, constraint)` (service_accounts.go)
   - **Comando GCloud:**
     ```bash
     gcloud resource-manager org-policies delete \
       <constraint> \
       --project=elet-<projeto>-<ambiente> \
       --quiet
     ```
     Se não encontrar override (NOT_FOUND) → considera sucesso

2. **Aguardar remoção do override (validar que está herdando)**
   - Função: `WaitForPolicyReset(projectID, constraint, 12, 5s)` (service_accounts.go)
   - 12 tentativas com 5s de intervalo
   - Função helper: `HasProjectOrgPolicyOverride()` → **Comando GCloud:**
     ```bash
     gcloud resource-manager org-policies describe \
       <constraint> \
       --project=elet-<projeto>-<ambiente> \
       --format=json
     ```
     Verifica: campo `spec` ausente ou vazio = não há override (herdando)

---

## 🔄 Funções de Carregamento e Exibição

### LoadExistingProject

**Função:** `LoadExistingProject(projectName string)` 
**Localização:** `internal/gcp/loader.go`
**Retorno:** `*GCPProject, error`

**Propósito:** Carregar dados de projetos já existentes no GCP

**Passos:**
1. Monta ID esperado: `elet-<projeto>-<ambiente>`
2. **Comando GCloud:**
   ```bash
   gcloud projects describe elet-<projeto>-<ambiente> \
     --format=json
   ```
3. Extrai `projectId` e `parent.id` (Folder ID)
4. Retorna `GCPProject` com dados preenchidos

### PrintProjectStructure

**Função:** `PrintProjectStructure(project *GCPProject)` 
**Localização:** `internal/gcp/step1_folders.go`
**Retorno:** `void`

**Propósito:** Exibir visualmente a estrutura criada

**Exemplo de saída:**
```
fldr-benner-cloud/
├── fldr-dev/
│   └── elet-benner-cloud-dev
│       ├── Folder ID: 123456789
│       └── Project ID: elet-benner-cloud-dev
...
```

---

## 🔐 Funções de Autenticação

### ValidateAuthentication

**Função:** `ValidateAuthentication()` 
**Localização:** `internal/gcp/client.go`
**Retorno:** `error`

**Propósito:** Validar autenticação do gcloud

**Passos:**
- **Comando GCloud:**
  ```bash
  gcloud config get-value account
  ```
- Se retornar erro → usuário não está autenticado

### GetCurrentAccount

**Função:** `GetCurrentAccount()` 
**Localização:** `internal/gcp/client.go`
**Retorno:** `(string, error)`

**Propósito:** Obter email da conta autenticada

**Passos:**
- **Comando GCloud:**
  ```bash
  gcloud config get-value account
  ```
- Retorna email (ex: "usuario@example.com")

---

## 📝 Resumo de Todos os Comandos GCloud

### Resource Manager (Pastas e Projetos)

```bash
# Criar pasta
gcloud resource-manager folders create \
  --display-name=<name> \
  --folder=<parent-folder-id> \
  --format=json

# Listar pastas
gcloud resource-manager folders list \
  --folder=<folder-id> \
  --format=json

gcloud resource-manager folders list \
  --organization=<org-id> \
  --format=json

# Criar projeto
gcloud projects create <project-id> \
  --name=<display-name> \
  --folder=<folder-id> \
  --format=json

# Descrever projeto
gcloud projects describe <project-id> \
  --format=json
```

### Projects (Labels)

```bash
# Adicionar labels
gcloud alpha projects update <project-id> \
  --update-labels=key1=value1,key2=value2
```

### Services (APIs)

```bash
# Habilitar API
gcloud services enable <api-name> \
  --project=<project-id>
```

### Billing

```bash
# Listar contas de billing
gcloud billing accounts list --format=json

# Vincular billing
gcloud billing projects link <project-id> \
  --billing-account=<billing-account-id>

# Descrever billing de projeto
gcloud billing projects describe <project-id> \
  --format=value(billingAccountName)
```

### Compute (Shared VPC)

```bash
# Atachar projeto a Shared VPC
gcloud compute shared-vpc associated-projects add \
  <service-project-id> \
  --host-project=<host-project-id>

# Listar projetos atachados
gcloud compute shared-vpc associated-projects list \
  --filter=name=<service-project-id> \
  --format=value(name)
```

### IAM (Service Accounts)

```bash
# Criar service account
gcloud iam service-accounts create <account-id> \
  --display-name=<display-name> \
  --project=<project-id>

# Descrever service account
gcloud iam service-accounts describe \
  <account-id>@<project-id>.iam.gserviceaccount.com \
  --project=<project-id>

# Listar service accounts
gcloud iam service-accounts list \
  --project=<project-id> \
  --format=json

# Criar custom role
gcloud iam roles create customRole_SA_<nome> \
  --project=<project-id> \
  --title=<title> \
  --description=<description> \
  --permissions=perm1,perm2,... \
  --stage=GA

# Adicionar IAM binding
gcloud projects add-iam-policy-binding <project-id> \
  --member=serviceAccount:<email> \
  --role=<role-name>

# Criar chave service account
gcloud iam service-accounts keys create <file.json> \
  --iam-account=<email> \
  --project=<project-id>

# Listar chaves
gcloud iam service-accounts keys list \
  --iam-account=<email> \
  --project=<project-id> \
  --format=json
```

### Organization Policies

```bash
# Desabilitar enforce de policy
gcloud resource-manager org-policies disable-enforce \
  <constraint> \
  --project=<project-id>

# Descrever policy efetiva
gcloud resource-manager org-policies describe \
  <constraint> \
  --project=<project-id> \
  --effective \
  --format=json

# Descrever policy local
gcloud resource-manager org-policies describe \
  <constraint> \
  --project=<project-id> \
  --format=json

# Deletar override (voltar a herdar)
gcloud resource-manager org-policies delete \
  <constraint> \
  --project=<project-id> \
  --quiet
```

### Secret Manager

```bash
# Descrever secret
gcloud secrets describe <secret-id> \
  --project=<project-id> \
  --format=json

# Criar secret
gcloud secrets create <secret-id> \
  --project=<project-id> \
  --replication-policy=automatic

# Adicionar versão
gcloud secrets versions add <secret-id> \
  --project=<project-id> \
  --data-file=<file> \
  --format=json
```

### Config (Autenticação)

```bash
# Obter projeto atual
gcloud config get-value project

# Obter conta atual
gcloud config get-value account
```

---

## 🎯 Fluxo de Exemplo Completo

Executando: `./projeto_config -project benner-cloud`

### Passo 1: Criar Pastas

```
✓ Autenticado como: usuario@example.com

🔍 Resolvendo pasta pai: fldr-scge
   ✓ ID encontrado: 123456789

📁 Criando pasta principal: fldr-benner-cloud
   ✓ Folder ID: 111111111

📁 Criando pasta de ambiente: fldr-dev
   ✓ Folder ID: 222222222
   ✓ Criando projeto GCP: elet-benner-cloud-dev
      ✓ Project ID: elet-benner-cloud-dev
   💳 Vinculando billing account: 01F7C9-60D131-20DC44
      ✓ Billing account vinculado

[Repete para qld e prd]

📊 Estrutura criada:
fldr-benner-cloud/
├── fldr-dev/
│   └── elet-benner-cloud-dev
│       ├── Folder ID: 222222222
│       └── Project ID: elet-benner-cloud-dev
...
```

### Passo 2: Adicionar Labels

```
🏷️  Adicionando labels ao projeto: elet-benner-cloud-dev (dev)
   ✓ Labels adicionados com sucesso:
      - ambiente: dev
      - companhia: elet
      - projeto: benner-cloud

[Repete para qld e prd]
```

### Passo 3: Habilitar APIs

```
📋 APIs Opcionais (Digite 's' para sim, 'n' para não):

   ❓ Habilitar Artifact Registry? (s/n): s
      ✓ Selecionado: Artifact Registry
   ❓ Habilitar Secret Manager? (s/n): s
      ✓ Selecionado: Secret Manager
   ❓ Habilitar Firestore? (s/n): n
      ✗ Pulado: Firestore

🔌 Habilitando APIs no projeto: elet-benner-cloud-dev (dev)
   ⏳ Habilitando compute.googleapis.com...
      ✓ Habilitada
   ⏳ Habilitando servicenetworking.googleapis.com...
      ✓ Habilitada
   ⏳ Habilitando artifactregistry.googleapis.com...
      ✓ Habilitada
   ⏳ Habilitando secretmanager.googleapis.com...
      ✓ Habilitada

[Repete para qld e prd]
```

### Passo 4: Atachar a Shared VPC

```
🔗 Atachando projeto à rede Spoke-dev
   Projeto de Serviço: elet-benner-cloud-dev
   Host Project: redes-spoke-dev-002b
   VPC: vpc-spoke-dev
   ✓ Projeto atachado com sucesso

[Repete para qld e prd]
```

### Passo 5: Criar Service Accounts (exemplo)

```
📋 Fase 1: Service accounts e roles por projeto

🔧 Configurando service accounts no projeto: elet-benner-cloud-dev (dev)
   - GitLab: sa-benner-cloud-git (sa-benner-cloud-git@elet-benner-cloud-dev.iam.gserviceaccount.com)
   - GSA: sa-benner-cloud-dev (sa-benner-cloud-dev@elet-benner-cloud-dev.iam.gserviceaccount.com)
   ✓ Service accounts e roles configuradas

[Repete para qld e prd]

📋 Fase 2: Desabilitando policies de org em todos os projetos

🔓 Desabilitando policies no projeto: elet-benner-cloud-dev
   ✓ constraints/iam.disableServiceAccountKeyCreation desabilitada e efetiva (enforced=false)
   ✓ constraints/iam.disableServiceAccountKeyUpload desabilitada e efetiva (enforced=false)

[Repete para qld e prd]

📋 Fase 3: Criando chaves JSON

🔑 Criando chaves para projeto: elet-benner-cloud-dev (dev)
   ✓ Chave GitLab criada: sa-benner-cloud-git-dev.json
   ✓ Secret GitLab atualizada: sa-benner-cloud-git (versao 1)
   ✓ Chave GSA criada: sa-benner-cloud-dev.json
   ✓ Secret GSA atualizada: sa-benner-cloud-dev (versao 1)

[Repete para qld e prd]

📋 Fase 4: Resetando policies de org em todos os projetos

🔒 Resetando policies no projeto: elet-benner-cloud-dev
   ✓ constraints/iam.disableServiceAccountKeyCreation voltou a herdar do parent
   ✓ constraints/iam.disableServiceAccountKeyUpload voltou a herdar do parent

[Repete para qld e prd]

✓ Passo 5 concluido com sucesso!
```

---

## 📚 Referências GCloud

- **Resource Manager:** `gcloud resource-manager --help`
- **Projects:** `gcloud projects --help`
- **IAM:** `gcloud iam --help`
- **Compute:** `gcloud compute --help`
- **Secret Manager:** `gcloud secrets --help`
- **Billing:** `gcloud billing --help`
- **Organization Policies:** `gcloud resource-manager org-policies --help`

---

**Última atualização:** 06/03/2026
