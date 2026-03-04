.PHONY: help build build-linux build-windows build-mac run clean test

BINARY_NAME=projeto_config
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LD_FLAGS=-ldflags "-X main.Version=$(VERSION)"

help:
	@echo "Targets disponíveis:"
	@echo "  make build        - Compilar para o SO atual"
	@echo "  make build-linux  - Compilar para Linux (amd64)"
	@echo "  make build-windows- Compilar para Windows (amd64)"
	@echo "  make build-mac    - Compilar para macOS (amd64)"
	@echo "  make run          - Compilar e executar localmente"
	@echo "  make clean        - Limpar binários"
	@echo "  make test         - Executar testes"

build:
	@echo "Compilando para o SO atual..."
	go build -o $(BINARY_NAME) main.go

build-linux:
	@echo "Compilando para Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) main.go
	@echo "✓ Binário Linux criado: $(BINARY_NAME)"

build-windows:
	@echo "Compilando para Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe main.go
	@echo "✓ Binário Windows criado: $(BINARY_NAME).exe"

build-mac:
	@echo "Compilando para macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-mac main.go
	@echo "✓ Binário macOS criado: $(BINARY_NAME)-mac"

run: build
	@./$(BINARY_NAME) -help

clean:
	@echo "Limpando binários..."
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME)-mac

test:
	@echo "Executando testes..."
	go test -v ./...

fmt:
	@echo "Formatando código..."
	go fmt ./...

lint:
	@echo "Verificando código..."
	go vet ./...

deps:
	@echo "Baixando dependências..."
	go mod download
	go mod tidy
