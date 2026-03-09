# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto segue [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [1.0.0] - 2026-03-09

### Adicionado
- Flag `--version` / `-v` para exibir a versão da aplicação
- Comando `completion` para autocompletar no shell (bash, zsh, fish, powershell)
- Documentação sobre autocompletar no README.md
- Flag `--show-gcloud-commands` para controlar exibição de comandos gcloud executados
- Sistema de rastreamento de comandos gcloud executados por passo
- Comandos `get` e `describe` para consultar informações de projetos
- Comando `delete` para remover projetos e estrutura de pastas
- Testes unitários para validação de flags (`cmd/create_test.go`, `cmd/update_test.go`)

### Funcionalidades principais
- Passo 1: Criação automática de pastas e projetos GCP (dev, qld, prd)
- Passo 2: Aplicação de labels padronizados
- Passo 3: Habilitação de APIs obrigatórias e opcionais
- Passo 4: Associação com Shared VPCs (vpc-spoke-dev/qld/prd)
- Passo 5: Criação de service accounts, aplicação de roles e geração de chaves JSON
- Flag `--all` em `create` executa passos 1-4 em sequência
- Flag `--all` em `update` executa passos 2-5 em sequência
- Suporte a APIs opcionais: `artifactregistry`, `secretmanager`, `firestore`
- Modo interativo para seleção de APIs opcionais (`--interactive-apis`)
- Tratamento de erros com retry automático para propagação de IAM
- Indicadores de progresso inline para operações de espera

### Otimizações
- Retry automático para bindings IAM (eventual consistency do GCP)
- Espera inteligente para criação de service accounts antes de usar
- Espera de estabilização de políticas IAM antes de criar chaves
- Progress indicators visuais consistentes em todas as operações de retry/wait

## Formato de Versionamento

Este projeto usa **[Versionamento Semântico](https://semver.org/lang/pt-BR/)**: `MAJOR.MINOR.PATCH`

- **MAJOR**: Mudanças incompatíveis na API/CLI (breaking changes)
- **MINOR**: Novas funcionalidades mantendo compatibilidade
- **PATCH**: Correções de bugs e melhorias mantendo compatibilidade

### Quando incrementar cada número:

**MAJOR (1.0.0 → 2.0.0):**
- Remoção de flags existentes
- Mudança no comportamento padrão que quebra scripts existentes
- Alteração na estrutura de pastas/projetos criados
- Mudança incompatível em comandos

**MINOR (1.0.0 → 1.1.0):**
- Adição de novos comandos
- Adição de novas flags opcionais
- Novos passos de automação
- Novas funcionalidades que não afetam uso existente

**PATCH (1.0.0 → 1.0.1):**
- Correção de bugs
- Melhorias de performance
- Ajustes em mensagens de erro
- Refatoração interna sem impacto externo
- Atualização de documentação
