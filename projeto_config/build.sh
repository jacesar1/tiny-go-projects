#!/bin/bash

# Script para compilar projeto_config para diferentes plataformas

set -e

BINARY_NAME="projeto_config"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "╔════════════════════════════════════════════╗"
echo "║  Compilando Projeto Config                ║"
echo "╚════════════════════════════════════════════╝"
echo ""

# Navegar para o diretório do projeto
cd "$SCRIPT_DIR"

# Verificar se Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✓ Go versão: $GO_VERSION"

# Baixar dependências
echo ""
echo "📦 Baixando dependências..."
go mod download
go mod tidy

# Compilar para Linux (padrão)
echo ""
echo "🐧 Compilando para Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "${BINARY_NAME}" main.go
echo "✓ Binário Linux criado: $BINARY_NAME"
echo "  Tamanho: $(ls -lh $BINARY_NAME | awk '{print $5}')"

# Compilar para Windows (opcional)
if [ "$1" != "--linux-only" ]; then
    echo ""
    echo "🪟 Compilando para Windows (amd64)..."
    GOOS=windows GOARCH=amd64 go build -o "${BINARY_NAME}.exe" main.go
    echo "✓ Binário Windows criado: ${BINARY_NAME}.exe"
fi

echo ""
echo "✅ Compilação concluída com sucesso!"
echo ""
echo "📝 Próximos passos:"
echo "   1. Copiar o binário para WSL ou Cloud Shell"
echo "   2. Executar: ./projeto_config -project <nome>"
echo ""
