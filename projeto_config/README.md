# Projeto Config - Automação de Projetos GCP

Aplicação Go para automatizar a criação de projetos no GCP seguindo a estrutura definida pela Eletrobras.

## 📋 Pré-requisitos

- **Go** 1.16+ (para compilação)
- **gcloud CLI** instalado e configurado
- Autenticação no GCP: `gcloud auth login`
- Permissões na organização/pastas (Folder Creator, Project Creator)
- A conta de Billing **01F7C9-60D131-20DC44** deve estar válida e ativa na organização

## 🚀 Instalação

### 1. Compilar o projeto

**No Windows (compila para Linux e Windows):**
```powershell
cd d:\workspace\tiny-go-projects\projeto_config
powershell -ExecutionPolicy Bypass -File build.ps1
```

**Resultado:** Dois binários serão criados
- `projeto_config` - Para Linux/WSL/Cloud Shell
- `projeto_config.exe` - Para Windows

### 2. Copiar para WSL ou Cloud Shell

```bash
# No WSL ou Cloud Shell
cp /mnt/d/workspace/tiny-go-projects/projeto_config/projeto_config .
chmod +x projeto_config
```

## 📌 Como Usar

### Sintaxe Básica

```bash
./projeto_config -project <nome> [optional: -step <1-5>]
```

### Exemplos

**Executar TODOS os passos (1-4):**
```bash
./projeto_config -project benner-cloud
```

**Passo 1 - Criar estrutura de pastas:**
```bash
./projeto_config -project benner-cloud -step 1
```

**Passo 2 - Adicionar labels (requer passo 1 já executado):**
```bash
./projeto_config -project benner-cloud -step 2
```

**Passo 3 - Habilitar APIs (requer passo 1 já executado):**
```bash
./projeto_config -project benner-cloud -step 3
```

**Passo 4 - Atachar nas redes (requer passo 1 já executado):**
```bash
./projeto_config -project benner-cloud -step 4
```

**Passo 5 - Criar Service Accounts (requer passo 1 já executado):**
```bash
./projeto_config -project benner-cloud -step 5
```

### Flags Disponíveis

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `-project` | Obrigatório | Nome do projeto (ex: benner-cloud) |
| `-parent-folder` | `fldr-scge` | ID ou nome da pasta pai (usado apenas no passo 1) |
| `-org-id` | `727440331682` | ID da organização Eletrobras |
| `-step` | Vazio (executa 1-4) | Qual passo executar (1-5). Se omitido, executa passos 1-4 |
| `-help` | - | Mostra ajuda |

**Importante:** 
- **Se nenhuma flag `-step` for especificada**, executa **todos os 4 passos sequencialmente**
- Se `-step` for especificado (1-4), executa **apenas aquele passo específico**
- Se `-step 5` for especificado, executa **apenas o passo 5**
- Os passos 2, 3 e 4 carregam automaticamente os dados dos projetos já criados no passo 1
- **Billing Account é vinculada automaticamente** (01F7C9-60D131-20DC44) ao criar projetos no passo 1
- Execute o passo 1 primeiro para criar a estrutura base

## 📚 Passos de Automação

### ✅ Passo 1: Criar Pastas no Resource Manager

Cria automaticamente a seguinte estrutura:

```
fldr-<nome do projeto>/
├── fldr-dev/
│   └── elet-<nome do projeto>-dev (projeto GCP)
├── fldr-qld/
│   └── elet-<nome do projeto>-qld (projeto GCP)
└── fldr-prd/
    └── elet-<nome do projeto>-prd (projeto GCP)
```

**Características:**
- ✓ Resolve automaticamente nome da pasta para ID
- ✓ Cria estrutura em 3 ambientes (dev, qld, prd)
- ✓ Fornece IDs de pasta e projeto na saída
- ✓ **Vincula Billing Account automaticamente (01F7C9-60D131-20DC44)**
- ✓ Tratamento robusto de erros com mensagens detalhadas

### 📋 Passo 2: Adicionar Labels ✅

Adiciona automaticamente os seguintes labels a cada projeto:
- `ambiente`: dev | qld | prd (variante para cada ambiente)
- `companhia`: elet
- `projeto`: <nome do projeto>

### 🔌 Passo 3: Habilitar APIs ✅

Habilita automaticamente as seguintes APIs obrigatórias em cada projeto:
- Compute Engine (`compute.googleapis.com`)
- Service Networking (`servicenetworking.googleapis.com`)

Além disso, oferece opção de habilitar APIs adicionais:
- Artifact Registry (`artifactregistry.googleapis.com`)
- Secret Manager (`secretmanager.googleapis.com`)
- Firestore (`firestore.googleapis.com`)

**Nota:** Requer que o projeto já tenha sido criado no passo 1 (com Billing Account vinculado automaticamente).

### 🔗 Passo 4: Atachar nas Redes Spokes ✅

Associa automaticamente cada projeto à sua VPC spoke correspondente (Shared VPC):

**Mapeamento de ambientes:**
- dev: `elet-<nome>-dev` → Shared VPC `vpc-spoke-dev` (host project ID: `redes-spoke-dev-002b`)
- qld: `elet-<nome>-qld` → Shared VPC `vpc-spoke-qld` (host project ID: `redes-spoke-qld-7e83`)
- prd: `elet-<nome>-prd` → Shared VPC `vpc-spoke-prd` (host project ID: `redes-spoke-prd-bd15`)

**Nota:** Requer que os projetos já tenham sido criados no passo 1.

### 👤 Passo 5: Criar Service Accounts ✅

Cria duas service accounts por ambiente, aplica as roles e gera chaves JSON:

**SA da Pipeline (GitLab):**
- Nome: `sa-<nome do projeto>-git` (uma única SA compartilhada entre ambientes)
- Role: `roles/artifactregistry.createOnPushWriter`
- Chave JSON: `sa-<nome do projeto>-git-<env>.json` (salva no diretório atual)

**SA GSA:**
- Nome: `sa-<nome do projeto>-dev|qld|prd`
- Role: `projects/<project_id>/roles/customRole_SA_<nome_do_projeto>` (hífens substituídos por underscores)
- Role adicional: `roles/secretmanager.viewer`
- Chave JSON: `sa-<nome do projeto>-<env>.json` (salva no diretório atual)

**Permissões da custom role:**
- `artifactregistry.repositories.downloadArtifacts`
- `autoscaling.sites.writeMetrics`
- `datastore.entities.get`
- `datastore.entities.list`
- `datastore.entities.update`
- `datastore.entities.create`
- `logging.logEntries.create`
- `monitoring.dashboards.get`
- `monitoring.timeSeries.create`
- `pubsub.subscriptions.consume`
- `pubsub.topics.publish`

**Criação de chaves JSON:**
Para permitir a criação das chaves, o passo 5 executa em 4 fases:
1. **Fase 1:** Cria service accounts e aplica roles em todos os projetos
2. **Fase 2:** Desabilita as org policies em todos os projetos:
  - `constraints/iam.disableServiceAccountKeyCreation`
  - `constraints/iam.disableServiceAccountKeyUpload`
  - Em seguida, valida a policy efetiva (`enforced=false`) antes de continuar
3. **Fase 3:** Cria as chaves JSON com retry robusto (até 8 tentativas por chave, com 15s entre tentativas)
4. **Fase 4:** Reseta as policies em todos os projetos (volta a herdar do parent) e valida a remoção do override

Essa abordagem em lote aumenta a confiabilidade e evita seguir para criação de chave antes da propagação real da policy.

**Nota:** Este passo só é executado com `-step 5`.

## 📂 Estrutura do Projeto

```
projeto_config/
├── main.go                          # Entrada principal
├── go.mod                           # Dependências do Go
├── go.sum                           # Hash de dependências
├── Makefile                         # Compilação multi-plataforma
├── build.ps1                        # Script PowerShell para compilar
├── build.sh                         # Script Bash para compilar
├── projeto_config                   # Binário Linux compilado
├── projeto_config.exe               # Binário Windows compilado
├── README.md                        # Esta documentação
├── cmd/                             # Comandos CLI (expandível)
└── internal/
    ├── gcp/                        # Operações GCP
    │   ├── client.go               # Cliente GCP (autenticação)
    │   ├── folders.go              # Gerenciamento de pastas
    │   ├── projects.go             # Gerenciamento de projetos
    │   ├── apis.go                 # Gerenciamento de APIs
    │   ├── billing.go              # Gerenciamento de Billing
    │   ├── networks.go             # Gerenciamento de Redes Shared VPC
    │   ├── loader.go               # Carregamento de projetos existentes
    │   ├── step1_folders.go        # Orquestrador do Passo 1
    │   ├── step2_labels.go         # Orquestrador do Passo 2
    │   ├── step3_apis.go           # Orquestrador do Passo 3
    │   └── step4_networks.go       # Orquestrador do Passo 4
    ├── models/                     # Estruturas de dados
    │   └── types.go                # Tipos do projeto
    └── config/                     # Configurações (expandível)
```

## 🔧 Troubleshooting

### Erro: "pasta não encontrada"
**Solução:** Use o ID numérico da pasta em vez do nome:
```bash
# Listar pastas e encontrar o ID
gcloud resource-manager folders list --organization=727440331682

# Usar o ID encontrado
./projeto_config -project test -parent-folder 196427624856
```

### Erro: "permissão negada"
Você precisa de permissões na organização. Peça admin para adicionar:
```bash
gcloud organizations add-iam-policy-binding 727440331682 \
  --member=user:seu-email@eletrobras.com \
  --role=roles/resourcemanager.folderCreator \
  --role=roles/resourcemanager.projectCreator
```

### Erro: "Billing account for project is not found"

Isso significa que a conta de billing **01F7C9-60D131-20DC44** não está vinculada corretamente. Para resolver:

**1. Verificar se a conta de billing está ativa:**
```bash
gcloud billing accounts list
```

**2. Se necessário, vincular manualmente a conta de billing ao projeto:**
```bash
gcloud billing projects link <PROJECT_ID> \
  --billing-account=01F7C9-60D131-20DC44
```

**Exemplo:**
```bash
gcloud billing projects link elet-axiaauth-dev \
  --billing-account=01F7C9-60D131-20DC44
```

Ou configure via Console GCP: https://console.cloud.google.com/billing

### Erro: "Shared VPC attachment failed"

Isso significa que o projeto host (como `redes-spoke-dev`) não está configurado corretamente ou o projeto de serviço não tem permissões. Para resolver:

**1. Verificar se o host project existe e está habilitado:**
```bash
gcloud compute projects describe redes-spoke-dev
```

**2. Verificar se a Shared VPC está habilitada no host project:**
```bash
gcloud compute shared-vpc enable redes-spoke-dev-002b
gcloud compute shared-vpc enable redes-spoke-qld-7e83
gcloud compute shared-vpc enable redes-spoke-prd-bd15
```

**3. Garantir que o seu usuário tem role `roles/compute.admin` no host project:**
```bash
gcloud projects add-iam-policy-binding redes-spoke-dev-002b \
  --member=user:seu-email@eletrobras.com \
  --role=roles/compute.admin
  
gcloud projects add-iam-policy-binding redes-spoke-qld-7e83 \
  --member=user:seu-email@eletrobras.com \
  --role=roles/compute.admin
  
gcloud projects add-iam-policy-binding redes-spoke-prd-bd15 \
  --member=user:seu-email@eletrobras.com \
  --role=roles/compute.admin
```

**4. Para vincular manualmente um projeto à Shared VPC:**
```bash
gcloud compute shared-vpc associated-projects add \
  elet-benner-cloud-dev \
  --host-project=redes-spoke-dev-002b
```

## 🛠️ Desenvolvimento

### Compilar localmente

```bash
cd projeto_config

# Para o SO atual
go build -o projeto_config main.go

# Para plataformas específicas
GOOS=linux GOARCH=amd64 go build -o projeto_config main.go
GOOS=windows GOARCH=amd64 go build -o projeto_config.exe main.go
```

### Adicionar novos passos

Crie novos arquivos em `internal/gcp/stepX_*.go` e implemente a função correspondente. Depois atualize `main.go` para integrá-la.

## 📝 Próximos Passos

- [x] ✅ Passo 1 - Criar pastas
- [x] ✅ Passo 2 - Adicionar labels
- [x] ✅ Passo 3 - Habilitar APIs
- [x] ✅ Passo 4 - Atachar às redes spokes
- [x] ✅ Passo 5 - Criar service accounts e roles
- [ ] Adicionar testes unitários
- [ ] Adicionar arquivo de configuração YAML/JSON
- [ ] Sistema de logging mais robusto
- [ ] Validação de nomes de projeto

## 📄 Licença

Projeto Eletrobras - Interno

## 👨‍💻 Suporte

Para dúvidas ou problemas, contate a equipe de infraestrutura.

