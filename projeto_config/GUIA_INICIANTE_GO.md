# 🎓 Guia para Iniciantes em Go - Entendendo Estruturas e Ponteiros

**Projeto: Automação GCP Axia Energia**

---

## 📑 Índice

1. [Introdução às Estruturas (Structs)](#-introdução-às-estruturas-structs)
2. [Os Tipos do Projeto](#-os-tipos-do-projeto)
3. [Ponteiros vs Valores](#-ponteiros-vs-valores)
4. [Por que Usamos Ponteiros?](#-por-que-usamos-ponteiros)
5. [Passagem de Argumentos](#-passagem-de-argumentos)
6. [Exemplos Práticos Passo a Passo](#-exemplos-práticos-passo-a-passo)
7. [Boas Práticas](#-boas-práticas)

---

## 🏗️ Introdução às Estruturas (Structs)

Em Go, uma **struct** (estrutura) é um tipo de dado composto que agrupa diferentes campos relacionados. É similar a uma "classe" em outras linguagens, mas sem métodos (embora você possa adicionar métodos a structs).

### Sintaxe Básica

```go
type NomeDaStruct struct {
    Campo1 TipoDoCampo1
    Campo2 TipoDoCampo2
    // ... mais campos
}
```

### Exemplo Simples

```go
type Pessoa struct {
    Nome  string
    Idade int
}

// Criando uma instância
pessoa := Pessoa{
    Nome:  "João",
    Idade: 30,
}

// Acessando campos
fmt.Println(pessoa.Nome)  // "João"
fmt.Println(pessoa.Idade) // 30
```

---

## 📦 Os Tipos do Projeto

O arquivo `internal/models/types.go` define 3 estruturas principais:

### 1. ProjectConfig

```go
type ProjectConfig struct {
    ProjectName      string
    OrgID            string
    ParentFolderID   string
    BillingAccountID string
}
```

**Propósito:** Armazenar as configurações de entrada fornecidas pelo usuário via linha de comando.

**Campos:**
- `ProjectName`: Nome do projeto (ex: "benner-cloud")
- `OrgID`: ID da organização GCP (ex: "727440331682")
- `ParentFolderID`: ID da pasta pai onde criar as subpastas (ex: "fldr-scge")
- `BillingAccountID`: ID da conta de billing (ex: "01F7C9-60D131-20DC44")

**Uso:**
```go
config := &ProjectConfig{
    ProjectName:      "benner-cloud",
    OrgID:            "727440331682",
    ParentFolderID:   "fldr-scge",
    BillingAccountID: "01F7C9-60D131-20DC44",
}
```

**Por que todos são strings?** 
- IDs no GCP são geralmente strings (podem conter letras, números e hífens)
- Strings são mais flexíveis para entrada de dados

---

### 2. GCPEnvironment

```go
type GCPEnvironment struct {
    Name      string  // dev, qld, prd
    FolderID  string  // ID da pasta criada
    ProjectID string  // ID do projeto criado
}
```

**Propósito:** Representar um ambiente específico (desenvolvimento, qualidade ou produção).

**Campos:**
- `Name`: Nome do ambiente ("dev", "qld" ou "prd")
- `FolderID`: ID da pasta criada no Resource Manager (ex: "123456789")
- `ProjectID`: ID do projeto GCP criado (ex: "elet-benner-cloud-dev")

**Exemplo de uso:**
```go
dev := &GCPEnvironment{
    Name:      "dev",
    FolderID:  "123456789",
    ProjectID: "elet-benner-cloud-dev",
}
```

**Por que strings?**
- `Name`: Identificação textual
- `FolderID` e `ProjectID`: IDs do GCP são strings

---

### 3. GCPProject

```go
type GCPProject struct {
    Name string
    Dev  *GCPEnvironment
    Qld  *GCPEnvironment
    Prd  *GCPEnvironment
}
```

**Propósito:** Agrupar todos os 3 ambientes de um projeto.

**Campos:**
- `Name`: Nome do projeto (ex: "benner-cloud")
- `Dev`: **Ponteiro** para o ambiente de desenvolvimento
- `Qld`: **Ponteiro** para o ambiente de qualidade
- `Prd`: **Ponteiro** para o ambiente de produção

**⚠️ ATENÇÃO:** Observe que `Dev`, `Qld` e `Prd` são **ponteiros** (`*GCPEnvironment`), não valores diretos.

**Exemplo de uso:**
```go
project := &GCPProject{
    Name: "benner-cloud",
    Dev:  &GCPEnvironment{Name: "dev"},
    Qld:  &GCPEnvironment{Name: "qld"},
    Prd:  &GCPEnvironment{Name: "prd"},
}
```

---

## 🔍 Ponteiros vs Valores

### O que é um Ponteiro?

Um **ponteiro** é uma variável que armazena o **endereço de memória** de outra variável, não o valor diretamente.

**Analogia:** 
- **Valor:** É como ter uma cópia de um documento
- **Ponteiro:** É como ter o endereço onde o documento original está guardado

### Sintaxe de Ponteiros em Go

```go
// Declaração de um tipo ponteiro
var p *int  // p é um ponteiro para int

// Operador & (endereço de)
x := 42
p = &x      // p agora aponta para o endereço de x

// Operador * (desreferência)
fmt.Println(*p)  // Imprime o valor apontado por p (42)
*p = 100         // Altera o valor de x através do ponteiro
fmt.Println(x)   // Imprime 100
```

### Exemplo Completo: Valor vs Ponteiro

```go
package main

import "fmt"

type Pessoa struct {
    Nome  string
    Idade int
}

// Função que recebe VALOR (cópia)
func AniversarioValor(p Pessoa) {
    p.Idade++  // Altera a cópia, não o original
}

// Função que recebe PONTEIRO (referência)
func AniversarioPonteiro(p *Pessoa) {
    p.Idade++  // Altera o original
}

func main() {
    pessoa1 := Pessoa{Nome: "João", Idade: 30}
    
    AniversarioValor(pessoa1)
    fmt.Println(pessoa1.Idade)  // 30 (não mudou!)
    
    AniversarioPonteiro(&pessoa1)
    fmt.Println(pessoa1.Idade)  // 31 (mudou!)
}
```

**Resultado:**
```
30  // AniversarioValor não alterou o original
31  // AniversarioPonteiro alterou o original
```

---

## 🎯 Por que Usamos Ponteiros?

### Razão 1: Modificar o Original

Quando passamos um ponteiro, a função pode alterar o objeto original:

```go
func (env *GCPEnvironment) SetProjectID(id string) {
    env.ProjectID = id  // Altera o original
}

// Uso:
dev := &GCPEnvironment{Name: "dev"}
dev.SetProjectID("elet-benner-cloud-dev")
fmt.Println(dev.ProjectID)  // "elet-benner-cloud-dev"
```

### Razão 2: Evitar Cópias Desnecessárias

Structs grandes podem ser pesadas para copiar. Ponteiros são sempre do mesmo tamanho (8 bytes em sistemas 64-bit):

```go
// SEM ponteiro: copia todos os campos (pesado)
func ProcessarValor(project GCPProject) {
    // ...
}

// COM ponteiro: copia apenas o endereço (leve)
func ProcessarPonteiro(project *GCPProject) {
    // ...
}
```

### Razão 3: Compartilhar Estado

Múltiplas partes do código podem trabalhar com o mesmo objeto:

```go
project := &GCPProject{Name: "benner-cloud"}

// Passo 1 popula os ambientes
Step1CreateFolderStructure(config)  // Retorna *GCPProject

// Passo 2 usa os mesmos ambientes
Step2AddLabels(project)  // Recebe *GCPProject

// Todos trabalham com o MESMO objeto na memória
```

### Razão 4: Nil (Valor Nulo)

Ponteiros podem ser `nil` (nulo), indicando ausência de valor:

```go
var env *GCPEnvironment  // nil por padrão

if env == nil {
    fmt.Println("Ambiente não configurado")
}

// Verificação comum no código:
if project.Dev.ProjectID == "" {
    fmt.Println("Projeto dev ainda não criado")
}
```

---

## 🔄 Passagem de Argumentos

### Regra Geral em Go

1. **Por Valor (cópia):** Ao passar uma variável diretamente
2. **Por Referência (endereço):** Ao passar com `&` ou declarar parâmetro como ponteiro

### Exemplo do Projeto

#### Declaração da Função (step1_folders.go)

```go
func Step1CreateFolderStructure(config *models.ProjectConfig) (*models.GCPProject, error) {
    // config é um PONTEIRO para ProjectConfig
    // Retorna um PONTEIRO para GCPProject
}
```

**Assinatura explicada:**
- `config *models.ProjectConfig` → Recebe **ponteiro** para ProjectConfig
- `(*models.GCPProject, error)` → Retorna **ponteiro** para GCPProject ou erro

#### Chamada da Função (main.go)

```go
// Criando config (já é ponteiro por causa do &)
config := &models.ProjectConfig{
    ProjectName:      *projectName,
    OrgID:            *orgID,
    ParentFolderID:   *parentFolder,
    BillingAccountID: DefaultBillingAccountID,
}

// Chamando a função
gcpProject, err := gcp.Step1CreateFolderStructure(config)
// config já é ponteiro, passa diretamente
```

**Detalhe importante:**
- `&models.ProjectConfig{...}` cria E retorna um ponteiro
- Se tivéssemos `config := models.ProjectConfig{...}`, passaríamos `&config`

### Comparação: Com e Sem &

```go
// FORMA 1: Criando com & (já é ponteiro)
config1 := &ProjectConfig{ProjectName: "teste"}
Step1CreateFolderStructure(config1)  // Passa diretamente

// FORMA 2: Criando sem & (valor)
config2 := ProjectConfig{ProjectName: "teste"}
Step1CreateFolderStructure(&config2)  // Precisa do &
```

---

## 💡 Exemplos Práticos Passo a Passo

### Exemplo 1: Criação e Inicialização de GCPProject

```go
// Passo a passo do Step1CreateFolderStructure

// 1. Criar a estrutura de projeto vazia
gcpProject := &models.GCPProject{
    Name: config.ProjectName,          // Copia o nome da config
    Dev:  &models.GCPEnvironment{Name: "dev"},  // Cria ponteiro para ambiente dev
    Qld:  &models.GCPEnvironment{Name: "qld"},  // Cria ponteiro para ambiente qld
    Prd:  &models.GCPEnvironment{Name: "prd"},  // Cria ponteiro para ambiente prd
}
```

**O que acontece na memória:**

```
┌─────────────────────────────────────────┐
│ gcpProject (ponteiro)                   │
│ Aponta para → ┌────────────────────┐   │
│               │ GCPProject         │   │
│               │ Name: "benner"     │   │
│               │ Dev → ┌─────────┐  │   │
│               │       │ Name: "dev"│  │   │
│               │       │ FolderID: ""│  │ (será preenchido)
│               │       │ ProjectID: ""│ │
│               │       └─────────┘  │   │
│               │ Qld → [similar]    │   │
│               │ Prd → [similar]    │   │
│               └────────────────────┘   │
└─────────────────────────────────────────┘
```

### Exemplo 2: Populando os Ambientes

```go
// 2. Criar mapa para facilitar acesso aos ambientes
envMap := map[string]*models.GCPEnvironment{
    "dev": gcpProject.Dev,  // Copia o PONTEIRO (não o valor)
    "qld": gcpProject.Qld,
    "prd": gcpProject.Prd,
}

// 3. Loop pelos ambientes
environments := []string{"dev", "qld", "prd"}
for _, env := range environments {
    // 3a. Obter o ambiente do mapa
    envData := envMap[env]  // envData é um ponteiro
    
    // 3b. Criar pasta de ambiente
    envFolderID, err := CreateFolder(mainFolderID, "fldr-" + env)
    
    // 3c. Atribuir o FolderID ao ambiente
    envData.FolderID = envFolderID
    // ↑ Como envData é PONTEIRO, isso modifica o original em gcpProject
    
    // 3d. Criar projeto GCP
    projectID := "elet-" + config.ProjectName + "-" + env
    createdProjectID, err := CreateProject(projectID, projectName, envFolderID)
    
    // 3e. Atribuir o ProjectID ao ambiente
    envData.ProjectID = createdProjectID
    // ↑ Novamente, modifica o original
}

// 4. Retornar o projeto (ponteiro)
return gcpProject, nil
```

**Fluxo de dados:**

```
main.go
  ↓ passa config (ponteiro)
Step1CreateFolderStructure()
  ↓ cria gcpProject (ponteiro)
  ↓ preenche gcpProject.Dev.FolderID
  ↓ preenche gcpProject.Dev.ProjectID
  ↓ [mesmo para Qld e Prd]
  ↓ retorna gcpProject (ponteiro)
main.go
  ↓ recebe gcpProject (ponteiro)
  ↓ passa para Step2AddLabels(gcpProject)
Step2AddLabels()
  ↓ acessa project.Dev (ponteiro)
  ↓ usa project.Dev.ProjectID
  ✓ Tudo funciona pois é o MESMO objeto na memória
```

### Exemplo 3: Modificando Ambientes em Funções

```go
// step2_labels.go
func Step2AddLabels(project *models.GCPProject) error {
    // project é um ponteiro recebido
    
    // Criar lista de ambientes (também ponteiros)
    environments := []*models.GCPEnvironment{
        project.Dev,  // Ponteiro
        project.Qld,  // Ponteiro
        project.Prd,  // Ponteiro
    }
    
    // Loop pelos ambientes
    for _, env := range environments {
        // env é um ponteiro para o ambiente
        
        // Verificar se ProjectID existe
        if env.ProjectID == "" {
            fmt.Printf("⚠️  Pulando ambiente %s - Project ID não disponível\n", env.Name)
            continue
        }
        
        // Usar o ProjectID
        labels := map[string]string{
            "ambiente":  env.Name,        // Acessa campo Name
            "projeto":   project.Name,     // Acessa campo Name do projeto
        }
        
        // Chamar API do GCP
        err := SetProjectLabels(env.ProjectID, labels)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

**Por que `[]*models.GCPEnvironment`?**
- `[]` = slice (array dinâmico)
- `*models.GCPEnvironment` = slice de ponteiros para GCPEnvironment
- Cada elemento é um ponteiro (não uma cópia)

### Exemplo 4: Carregando Projeto Existente

```go
// loader.go
func LoadExistingProject(projectName string) (*models.GCPProject, error) {
    // 1. Criar estrutura de projeto nova
    gcpProject := &models.GCPProject{
        Name: projectName,
        Dev:  &models.GCPEnvironment{Name: "dev"},
        Qld:  &models.GCPEnvironment{Name: "qld"},
        Prd:  &models.GCPEnvironment{Name: "prd"},
    }
    
    // 2. Lista de ambientes (ponteiros)
    environments := []*models.GCPEnvironment{
        gcpProject.Dev,
        gcpProject.Qld,
        gcpProject.Prd,
    }
    
    // 3. Para cada ambiente, tentar carregar do GCP
    for _, env := range environments {
        // Montar ID esperado
        projectID := fmt.Sprintf("elet-%s-%s", projectName, env.Name)
        
        // Buscar projeto no GCP
        projectInfo, err := GetProjectByID(projectID)
        if err != nil {
            continue  // Pula se não encontrar
        }
        
        // Extrair informações do JSON retornado
        if pid, ok := projectInfo["projectId"].(string); ok {
            env.ProjectID = pid  // Atribui ao ambiente (modifica original)
        }
        
        if parent, ok := projectInfo["parent"].(map[string]interface{}); ok {
            if folderID, ok := parent["id"].(string); ok {
                env.FolderID = folderID  // Atribui ao ambiente
            }
        }
    }
    
    // 4. Retornar projeto (ponteiro)
    return gcpProject, nil
}
```

**Fluxo:**

```
LoadExistingProject("benner-cloud")
  ↓
Cria gcpProject com 3 ambientes vazios
  ↓
Para cada ambiente:
  ↓ Busca "elet-benner-cloud-dev" no GCP
  ↓ Se encontrou, preenche env.ProjectID e env.FolderID
  ↓ (modifica o original porque env é ponteiro)
  ↓
Retorna gcpProject (com dados preenchidos)
```

---

## 🧪 Exemplo Didático Completo

Vamos criar um exemplo simples para demonstrar tudo:

```go
package main

import "fmt"

// 1. Definição das structs (como em types.go)
type Environment struct {
    Name      string
    ProjectID string
}

type Project struct {
    Name string
    Dev  *Environment  // PONTEIRO
}

// 2. Função que cria o projeto
func CreateProject(name string) *Project {
    // Cria projeto e retorna ponteiro
    project := &Project{
        Name: name,
        Dev:  &Environment{Name: "dev"},  // Cria ambiente com ponteiro
    }
    return project
}

// 3. Função que preenche o projeto
func PopulateProject(project *Project, projectID string) {
    // Recebe ponteiro, modifica o original
    project.Dev.ProjectID = projectID
}

// 4. Função que exibe o projeto
func PrintProject(project *Project) {
    // Recebe ponteiro, apenas lê
    fmt.Printf("Projeto: %s\n", project.Name)
    fmt.Printf("Ambiente: %s\n", project.Dev.Name)
    fmt.Printf("Project ID: %s\n", project.Dev.ProjectID)
}

func main() {
    // Passo 1: Criar projeto
    myProject := CreateProject("meu-projeto")
    // myProject é um ponteiro para Project
    
    fmt.Println("=== Após criação ===")
    PrintProject(myProject)
    // Output:
    // Projeto: meu-projeto
    // Ambiente: dev
    // Project ID: 
    
    // Passo 2: Preencher dados
    PopulateProject(myProject, "elet-meu-projeto-dev")
    
    fmt.Println("\n=== Após popular ===")
    PrintProject(myProject)
    // Output:
    // Projeto: meu-projeto
    // Ambiente: dev
    // Project ID: elet-meu-projeto-dev
    
    // Passo 3: Verificar que é o mesmo objeto
    fmt.Printf("\nEndereço de myProject: %p\n", myProject)
    fmt.Printf("Endereço de myProject.Dev: %p\n", myProject.Dev)
    // Ambos mostram o mesmo endereço de memória
}
```

**Saída esperada:**

```
=== Após criação ===
Projeto: meu-projeto
Ambiente: dev
Project ID: 

=== Após popular ===
Projeto: meu-projeto
Ambiente: dev
Project ID: elet-meu-projeto-dev

Endereço de myProject: 0xc0000b4000
Endereço de myProject.Dev: 0xc0000b4020
```

---

## 🎯 Boas Práticas

### 1. Quando Usar Ponteiros

✅ **USE ponteiros quando:**
- A função precisa modificar o objeto
- O objeto é grande (evita cópias)
- Você quer compartilhar o mesmo objeto entre funções
- Você quer permitir `nil` como valor válido

❌ **NÃO use ponteiros quando:**
- O objeto é pequeno (ex: int, bool, structs com 1-2 campos)
- Você quer garantir que o objeto não será modificado
- Trabalhando com tipos primitivos simples

### 2. Criando Ponteiros

```go
// FORMA 1: Com &
config := &ProjectConfig{
    ProjectName: "teste",
}

// FORMA 2: Sem & (depois usar &)
config2 := ProjectConfig{
    ProjectName: "teste",
}
UsarConfig(&config2)

// FORMA 3: Com new() (menos comum)
config3 := new(ProjectConfig)
config3.ProjectName = "teste"
```

### 3. Verificando Nil

Sempre verifique ponteiros antes de usar:

```go
func ProcessarAmbiente(env *GCPEnvironment) error {
    // Verificar se ponteiro não é nil
    if env == nil {
        return fmt.Errorf("ambiente é nil")
    }
    
    // Verificar se campos estão preenchidos
    if env.ProjectID == "" {
        return fmt.Errorf("ProjectID vazio")
    }
    
    // Usar o ambiente
    fmt.Println(env.ProjectID)
    return nil
}
```

### 4. Retornando Ponteiros

```go
// ✅ BOM: Retornar ponteiro de estrutura criada na função
func NovoAmbiente(name string) *GCPEnvironment {
    return &GCPEnvironment{Name: name}
}

// ❌ RUIM: Retornar ponteiro de variável local primitiva
func GetNumber() *int {
    x := 42
    return &x  // Perigoso em algumas linguagens, mas OK em Go
}
// Em Go, o compilador move para heap automaticamente
```

### 5. Nomenclatura

Convenção comum em Go:

```go
// Funções que retornam ponteiros: New + Nome
func NewProjectConfig() *ProjectConfig { ... }

// Funções que recebem ponteiros: nome descritivo
func Step1CreateFolderStructure(config *ProjectConfig) { ... }
```

---

## 📚 Resumo da Jornada

### Estado Inicial (main.go)

```go
// 1. Usuário fornece flags
projectName := "benner-cloud"

// 2. Criar config (ponteiro)
config := &models.ProjectConfig{
    ProjectName: projectName,
    // ... outros campos
}

// 3. Chamar Step 1
gcpProject, err := gcp.Step1CreateFolderStructure(config)
// gcpProject agora é ponteiro com dados preenchidos
```

### Durante Step 1 (step1_folders.go)

```go
func Step1CreateFolderStructure(config *ProjectConfig) (*GCPProject, error) {
    // 1. Criar projeto (ponteiro)
    gcpProject := &GCPProject{
        Name: config.ProjectName,
        Dev:  &GCPEnvironment{Name: "dev"},
        Qld:  &GCPEnvironment{Name: "qld"},
        Prd:  &GCPEnvironment{Name: "prd"},
    }
    
    // 2. Preencher ambientes
    gcpProject.Dev.FolderID = "123"
    gcpProject.Dev.ProjectID = "elet-benner-cloud-dev"
    
    // 3. Retornar (ponteiro)
    return gcpProject, nil
}
```

### De Volta ao main.go

```go
// Recebe o ponteiro retornado
gcpProject, err := gcp.Step1CreateFolderStructure(config)

// Passa para próximo passo (mesmo ponteiro)
err = gcp.Step2AddLabels(gcpProject)

// Step2 acessa os mesmos dados na memória
```

---

## 🔬 Conceitos Avançados (Opcional)

### 1. Escape Analysis

Go automaticamente decide se uma variável fica na **stack** (rápido, local) ou **heap** (mais lento, compartilhado):

```go
func CreateObject() *MyStruct {
    obj := &MyStruct{Name: "teste"}
    return obj  // Go move para heap automaticamente
}
```

### 2. Garbage Collection

Go tem **garbage collector** automático. Você não precisa liberar memória manualmente:

```go
func Process() {
    project := &GCPProject{...}  // Aloca memória
    // ... usa project
}  // Quando função termina, Go libera memória automaticamente
```

### 3. Métodos em Ponteiros

Você pode definir métodos que operam em ponteiros:

```go
// Método que recebe *GCPProject
func (p *GCPProject) AddEnvironment(env *GCPEnvironment) {
    // Modifica p
}

// Uso:
project := &GCPProject{Name: "teste"}
project.AddEnvironment(newEnv)
```

---

## 🎓 Exercícios Práticos

### Exercício 1: Criar e Modificar

```go
// Crie uma função que:
// 1. Cria um GCPProject
// 2. Popula o ambiente dev
// 3. Retorna o ponteiro

func CreateAndPopulateProject(name string, devProjectID string) *GCPProject {
    // SEU CÓDIGO AQUI
}
```

<details>
<summary>Solução</summary>

```go
func CreateAndPopulateProject(name string, devProjectID string) *GCPProject {
    project := &GCPProject{
        Name: name,
        Dev:  &GCPEnvironment{Name: "dev"},
        Qld:  &GCPEnvironment{Name: "qld"},
        Prd:  &GCPEnvironment{Name: "prd"},
    }
    
    project.Dev.ProjectID = devProjectID
    
    return project
}
```

</details>

### Exercício 2: Modificar via Função

```go
// Crie uma função que recebe um ponteiro e modifica o FolderID

func SetFolderID(env *GCPEnvironment, folderID string) {
    // SEU CÓDIGO AQUI
}
```

<details>
<summary>Solução</summary>

```go
func SetFolderID(env *GCPEnvironment, folderID string) {
    if env != nil {
        env.FolderID = folderID
    }
}
```

</details>

---

## 📖 Recursos Adicionais

### Documentação Oficial

- [Tour of Go - Pointers](https://go.dev/tour/moretypes/1)
- [Effective Go - Pointers vs Values](https://go.dev/doc/effective_go#pointers_vs_values)
- [Go by Example - Pointers](https://gobyexample.com/pointers)

### Livros Recomendados

- "The Go Programming Language" (Donovan & Kernighan)
- "Learning Go" (Jon Bodner)
- "Go in Action" (William Kennedy)

---

**Última atualização:** 06/03/2026

---

## 🎉 Parabéns!

Você agora entende:
- ✅ Como structs funcionam em Go
- ✅ A diferença entre valores e ponteiros
- ✅ Por que usamos ponteiros no projeto
- ✅ Como passar argumentos corretamente
- ✅ Como as funções do projeto trabalham juntas

Continue praticando e explorando o código! 🚀
