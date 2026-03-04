# Projeto Config - Automação de Projetos GCP

Aplicação Go para automatizar a criação de projetos no GCP seguindo a estrutura definida pela Eletrobras.

## 📋 Pré-requisitos

- **Go** 1.16+ (para compilação)
- **gcloud CLI** instalado e configurado
- Autenticação no GCP: `gcloud auth login`
- Permissões na organização/pastas (Folder Creator, Project Creator)

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
./projeto_config -project <nome> [-step <1-4>]
```

### Exemplos

**Passo 1 - Criar estrutura de pastas:**
```bash
./projeto_config -project benner-cloud -step 1
```

**Com ID da pasta pai específico:**
```bash
./projeto_config -project benner-cloud -parent-folder 196427624856 -step 1
```

**Com nome de pasta pai (será resolvido automaticamente):**
```bash
./projeto_config -project benner-cloud -parent-folder fldr-scge -step 1
```

**Todos os passos sequencialmente:**
```bash
./projeto_config -project benner-cloud -step 1
./projeto_config -project benner-cloud -step 2
./projeto_config -project benner-cloud -step 3
./projeto_config -project benner-cloud -step 4
```

### Flags Disponíveis

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `-project` | Obrigatório | Nome do projeto (ex: benner-cloud) |
| `-parent-folder` | `fldr-scge` | ID ou nome da pasta pai |
| `-org-id` | `727440331682` | ID da organização Eletrobras |
| `-step` | `1` | Qual passo executar (1-4) |
| `-help` | - | Mostra ajuda |

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
- ✓ Tratamento robusto de erros com mensagens detalhadas

### 📋 Passo 2: Adicionar Labels (Em desenvolvimento)

Adiciona os seguintes labels a cada projeto:
- `ambiente`: dev | qld | prd
- `companhia`: elet
- `projeto`: <nome do projeto>

### 🔌 Passo 3: Habilitar APIs (Em desenvolvimento)

Habilita automaticamente as seguintes APIs em cada projeto:
- Compute Engine (`compute.googleapis.com`)
- Service Networking (`servicenetworking.googleapis.com`)

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
    │   └── step1_folders.go        # Orquestrador do Passo 1
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

### Erro: "gcloud not found"
Instale o Google Cloud SDK: https://cloud.google.com/sdk/docs/install

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

- [ ] ✅ Passo 1 - Criar pastas
- [ ] Passo 2 - Adicionar labels
- [ ] Passo 3 - Habilitar APIs
- [ ] Passo 4 - Atachar às redes spokes
- [ ] Adicionar testes unitários
- [ ] Adicionar arquivo de configuração YAML/JSON
- [ ] Sistema de logging mais robusto
- [ ] Validação de nomes de projeto

## 📄 Licença

Projeto Eletrobras - Interno

## 👨‍💻 Suporte

Para dúvidas ou problemas, contate a equipe de infraestrutura.

