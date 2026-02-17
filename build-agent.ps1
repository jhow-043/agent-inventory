# Script para compilar o agent Windows

Write-Host "Compilando o agent Windows..." -ForegroundColor Cyan

# Criar diretório bin se não existir
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Usar Docker para compilar
docker run --rm `
    -v "${PWD}:/work" `
    -w /work `
    golang:1.24-alpine `
    sh -c "cd agent && go mod download && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' -o /work/bin/agent.exe ./cmd/agent"

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nAgent compilado com sucesso!" -ForegroundColor Green
    Write-Host "Executável: bin\agent.exe" -ForegroundColor Green
} else {
    Write-Host "`nErro ao compilar o agent!" -ForegroundColor Red
}
