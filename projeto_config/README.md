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
./projeto_config -project <nome> [optional: -step <1-4>]
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

### Flags Disponíveis

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `-project` | Obrigatório | Nome do projeto (ex: benner-cloud) |
| `-parent-folder` | `fldr-scge` | ID ou nome da pasta pai (usado apenas no passo 1) |
| `-org-id` | `727440331682` | ID da organização Eletrobras |
| `-step` | Vazio (executa todos) | Qual passo executar (1-4). Se omitido, executa todos os passos |
| `-help` | - | Mostra ajuda |

**Importante:** 
- **Se nenhuma flag `-step` for especificada**, executa **todos os 4 passos sequencialmente**
- Se `-step` for especificado (1-4), executa **apenas aquele passo específico**
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

### 🔗 Passo 4: Atachar nas Redes Spokes (Em desenvolvimento)

Associa cada projeto à VPC spoke do seu ambiente:
- dev → spoke-dev
- qld → spoke-qld  
- prd → spoke-prd

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
    │   ├── loader.go               # Carregamento de projetos existentes
    │   ├── step1_folders.go        # Orquestrador do Passo 1
    │   ├── step2_labels.go         # Orquestrador do Passo 2
    │   └── step3_apis.go           # Orquestrador do Passo 3
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
- [ ] Passo 4 - Atachar às redes spokes
- [ ] Adicionar testes unitários
- [ ] Adicionar arquivo de configuração YAML/JSON
- [ ] Sistema de logging mais robusto
- [ ] Validação de nomes de projeto

## 📄 Licença

Projeto Eletrobras - Interno

## 👨‍💻 Suporte

Para dúvidas ou problemas, contate a equipe de infraestrutura.

