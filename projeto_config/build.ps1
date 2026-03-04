# Script para compilar projeto_config para diferentes plataformas

$BinaryName = "projeto_config"
$ProjectDir = Split-Path -Parent $PSCommandPath

Write-Host "`n╔════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Compilando Projeto Config                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# Verificar se Go está instalado
$goVersion = go version
if ($?) {
    Write-Host "✓ Go versão: $goVersion" -ForegroundColor Green
} else {
    Write-Host "❌ Go não está instalado" -ForegroundColor Red
    exit 1
}

# Navegar para o diretório do projeto
Set-Location $ProjectDir

# Baixar dependências
Write-Host "`n📦 Baixando dependências..." -ForegroundColor Yellow
go mod download
go mod tidy

# Compilar para Linux
Write-Host "`n🐧 Compilando para Linux (amd64)..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o $BinaryName main.go

if ($?) {
    $size = (Get-Item $BinaryName).Length / 1MB
    Write-Host "✓ Binário Linux criado: $BinaryName (${size:F2} MB)" -ForegroundColor Green
} else {
    Write-Host "❌ Erro ao compilar para Linux" -ForegroundColor Red
    exit 1
}

# Compilar para Windows
Write-Host "`n🪟 Compilando para Windows (amd64)..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "${BinaryName}.exe" main.go

if ($?) {
    $size = (Get-Item "${BinaryName}.exe").Length / 1MB
    Write-Host "✓ Binário Windows criado: ${BinaryName}.exe (${size:F2} MB)" -ForegroundColor Green
} else {
    Write-Host "❌ Erro ao compilar para Windows" -ForegroundColor Red
    exit 1
}

# Limpar variáveis de ambiente
$env:GOOS = ""
$env:GOARCH = ""

Write-Host "`n✅ Compilação concluída com sucesso!" -ForegroundColor Green
Write-Host "`n📝 Próximos passos:" -ForegroundColor Cyan
Write-Host "   1. Copiar o binário 'projeto_config' para WSL ou Cloud Shell" -ForegroundColor White
Write-Host "   2. Executar: ./projeto_config -project <nome>" -ForegroundColor White
Write-Host "`n"
