# Gestão de Liberação e Implantação

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir os processos de liberação (release) e implantação (deploy) do Sistema de Inventário de Ativos de TI, garantindo entregas controladas, testadas e reversíveis.

---

## 2. Escopo

Liberação e implantação de todos os componentes: Agent Windows, API, Dashboard e banco de dados.

---

## 3. Política de Versionamento

### 3.1 Versionamento Semântico (SemVer)

```
MAJOR.MINOR.PATCH

v1.0.0 — Release inicial de produção
v1.1.0 — Nova funcionalidade (backward-compatible)
v1.1.1 — Correção de bug
v2.0.0 — Mudança incompatível (breaking change)
```

### 3.2 Versionamento por Componente

Todos os componentes compartilham a **mesma versão** (monorepo release):

| Tag Git | Agent | API | Dashboard |
|---|---|---|---|
| v1.0.0 | agent v1.0.0 | api v1.0.0 | web v1.0.0 |
| v1.1.0 | agent v1.1.0 | api v1.1.0 | web v1.1.0 |

> **Justificativa:** Simplicidade para desenvolvedor solo. Versões independentes por componente são mais complexas de gerenciar.

---

## 4. Tipos de Release

| Tipo | Trigger | Branch | Testes | Deploy |
|---|---|---|---|---|
| **Release** | Tag `v*.*.*` em main | main | CI completo | Manual (docker compose) |
| **Hotfix** | Bug crítico em produção | hotfix/* → main | CI mínimo | Imediato |
| **Pre-release** | Teste de release candidate | develop | CI completo | Ambiente de teste |

---

## 5. Pipeline CI/CD

### 5.1 Visão Geral

```
Push / PR                    Tag v*.*.*
    │                            │
    ▼                            ▼
┌──────────┐              ┌──────────────┐
│  CI      │              │  Release     │
│          │              │              │
│ 1. Lint  │              │ 1. CI (full) │
│ 2. Test  │              │ 2. Build     │
│ 3. Build │              │ 3. Release   │
└──────────┘              │    - .exe    │
                          │    - Docker  │
                          └──────────────┘
```

### 5.2 CI — Em Cada Push/PR

```yaml
# .github/workflows/ci.yml (resumo)
on:
  push:
    branches: [develop, main]
  pull_request:
    branches: [develop, main]

jobs:
  lint-go:
    - golangci-lint run ./server/... ./agent/... ./shared/...

  test-go:
    - go test ./server/... -race -coverprofile=coverage.out
    - go test ./agent/... -race (tags: !windows se no Linux CI)

  test-integration:
    services:
      postgres: (testcontainers gerencia via código)
    - go test ./server/... -tags=integration

  lint-web:
    - npm run lint (web/)

  test-web:
    - npm run test (web/)

  build-agent:
    - GOOS=windows GOARCH=amd64 go build -o agent.exe ./agent/cmd/agent

  build-server:
    - docker build -t inventory-api ./server

  build-web:
    - docker build -t inventory-web ./web
```

### 5.3 Release — Em Tag

```yaml
# .github/workflows/release.yml (resumo)
on:
  push:
    tags: ['v*.*.*']

jobs:
  release:
    steps:
      - CI completo (lint + test + build)
      - Build agent.exe (GOOS=windows GOARCH=amd64)
      - Build Docker images (api, web)
      - Criar GitHub Release com:
        - agent.exe (binário)
        - agent-config.example.yaml
        - CHANGELOG.md
      - Push Docker images para GitHub Container Registry (ghcr.io)
```

---

## 6. Procedimento de Deploy — API e Dashboard

### 6.1 Pré-requisitos

- [ ] Docker e Docker Compose instalados no servidor
- [ ] Arquivo `.env` configurado com secrets de produção
- [ ] Backup do banco realizado (se houver dados)
- [ ] Changelog da versão revisado

### 6.2 Deploy Inicial (Primeira Vez)

```bash
# 1. Clonar o repositório (ou copiar arquivos de deploy)
git clone https://github.com/org/inventario.git /opt/inventory
cd /opt/inventory

# 2. Configurar variáveis de ambiente
cp .env.example .env
nano .env  # Configurar DATABASE_URL, JWT_SECRET, ENROLLMENT_KEY, etc.

# 3. Subir os containers
docker compose up -d

# 4. Verificar saúde
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz

# 5. Criar usuário admin inicial
# (via seed script ou endpoint de setup)
docker compose exec api ./server seed --username admin --password <senha>

# 6. Verificar dashboard
# Acessar http://<server-ip>:3000 no browser
```

### 6.3 Deploy de Atualização

```bash
# 1. Fazer backup do banco
docker exec inventory-postgres pg_dump -U inventory -Fc inventory > backup_$(date +%Y%m%d).dump

# 2. Atualizar imagens
cd /opt/inventory
git pull origin main  # ou: docker compose pull (se usando registry)

# 3. Aplicar atualização
docker compose up -d --build

# 4. Verificar saúde
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz

# 5. Verificar logs
docker compose logs --tail=50 api

# 6. Verificar dashboard
# Acessar http://<server-ip>:3000 e confirmar versão
```

### 6.4 Rollback

```bash
# Se algo deu errado após deploy:

# 1. Reverter para versão anterior
git checkout v1.0.0  # tag anterior
docker compose up -d --build

# 2. Se migration causou problema: restaurar backup
docker compose stop api web
docker exec -it inventory-postgres psql -U inventory -c "DROP DATABASE inventory;"
docker exec -it inventory-postgres psql -U inventory -c "CREATE DATABASE inventory;"
docker cp backup_YYYYMMDD.dump inventory-postgres:/tmp/restore.dump
docker exec -it inventory-postgres pg_restore -U inventory -d inventory /tmp/restore.dump
docker compose start api web

# 3. Verificar saúde
curl http://localhost:8080/healthz
```

---

## 7. Procedimento de Deploy — Agent Windows

### 7.1 Instalação Inicial

```powershell
# 1. Baixar agent.exe da GitHub Release (ou copiar de rede compartilhada)
# Destino: C:\ProgramData\InventoryAgent\

# 2. Criar arquivo de configuração
# C:\ProgramData\InventoryAgent\agent-config.yaml
@"
api_url: "http://<server-ip>:8080"
enrollment_key: "<enrollment-key>"
collection_interval: "4h"
log_level: "info"
log_file: "C:\ProgramData\InventoryAgent\agent.log"
"@ | Out-File -FilePath "C:\ProgramData\InventoryAgent\agent-config.yaml" -Encoding UTF8

# 3. Instalar como serviço
C:\ProgramData\InventoryAgent\agent.exe install

# 4. Iniciar o serviço
C:\ProgramData\InventoryAgent\agent.exe start

# 5. Verificar status
Get-Service InventoryAgent
```

### 7.2 Atualização do Agent

```powershell
# 1. Parar o serviço
C:\ProgramData\InventoryAgent\agent.exe stop

# 2. Substituir o binário
Copy-Item "\\server\share\agent-v1.1.0.exe" "C:\ProgramData\InventoryAgent\agent.exe" -Force

# 3. Iniciar o serviço
C:\ProgramData\InventoryAgent\agent.exe start

# 4. Verificar log
Get-Content "C:\ProgramData\InventoryAgent\agent.log" -Tail 20
```

### 7.3 Deploy em Massa (GPO/Script)

Para 100-500 máquinas, usar:
- **GPO (Group Policy Object)** para executar script de instalação
- **Script PowerShell** distribuído via rede compartilhada
- **PSEXEC** ou **PowerShell Remoting** para execução remota

```powershell
# Exemplo: script de deploy em massa via PowerShell Remoting
$computers = Get-Content "C:\deploy\computers.txt"
$credential = Get-Credential

foreach ($computer in $computers) {
    Invoke-Command -ComputerName $computer -Credential $credential -ScriptBlock {
        # Parar serviço existente (se houver)
        if (Get-Service InventoryAgent -ErrorAction SilentlyContinue) {
            & "C:\ProgramData\InventoryAgent\agent.exe" stop
        }

        # Copiar binário
        Copy-Item "\\fileserver\inventory\agent.exe" "C:\ProgramData\InventoryAgent\" -Force

        # Instalar e iniciar (se primeira vez)
        if (-not (Get-Service InventoryAgent -ErrorAction SilentlyContinue)) {
            & "C:\ProgramData\InventoryAgent\agent.exe" install
        }
        & "C:\ProgramData\InventoryAgent\agent.exe" start
    }
}
```

---

## 8. Checklist de Go/No-Go

Antes de cada release, validar:

### 8.1 Go (Prosseguir)

- [ ] CI pipeline passa 100% (lint, test, build)
- [ ] Testes de integração passam com banco real
- [ ] Changelog atualizado
- [ ] Documentação atualizada (se comportamento mudou)
- [ ] Backup do banco de produção realizado
- [ ] Rollback plan definido e testado
- [ ] Janela de deploy definida

### 8.2 No-Go (Adiar)

- [ ] Testes falhando no CI
- [ ] Vulnerabilidade crítica não resolvida
- [ ] Dependência de mudança em outro componente não pronta
- [ ] Sem backup disponível
- [ ] Fora da janela de manutenção (para mudanças de alto risco)

---

## 9. Ambientes

| Ambiente | Propósito | Configuração | Dados |
|---|---|---|---|
| **Local (dev)** | Desenvolvimento | `docker-compose.dev.yml` | Fixtures/seed |
| **CI** | Testes automatizados | GitHub Actions + testcontainers | Temporários |
| **Produção** | Serviço real | `docker-compose.yml` + `.env` | Dados reais |

> **Nota:** Na Fase 1, não há ambiente de staging separado. Testes completos no CI + ambiente local substituem staging.

---

## 10. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Release manager | Desenvolvedor | Criar tag, gerar release |
| Deploy da API/Dashboard | Administrador de TI ou Desenvolvedor | Executar docker compose |
| Deploy do Agent | Administrador de TI | Distribuir .exe nas estações |
| Validação pós-deploy | Administrador de TI | Health check, verificar dashboard |

---

## 11. Referências

- [Gestão de Mudanças](gestao-de-mudancas.md)
- [Validação e Testes](validacao-e-testes.md)
- [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md)
- [Diagrama — Fluxo de Deploy](../06-anexos/diagramas/fluxo-de-deploy.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
