# Fluxo de Deploy (CI/CD)

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  

---

## Pipeline CI/CD — Visão Geral

```mermaid
graph LR
    subgraph "Desenvolvedor"
        DEV["Push / PR"]
    end

    subgraph "GitHub Actions — CI"
        LINT["Lint<br/>golangci-lint<br/>ESLint"]
        TEST_GO["Test Go<br/>go test ./...<br/>-race -cover"]
        TEST_REACT["Test React<br/>vitest run<br/>--coverage"]
        BUILD_GO["Build Go<br/>go build agent<br/>go build server"]
        BUILD_REACT["Build React<br/>npm run build"]
    end

    subgraph "GitHub Actions — CD (on tag)"
        RELEASE["Create Release<br/>Upload binaries"]
    end

    subgraph "Servidor Produção"
        DEPLOY["Deploy<br/>git pull + docker compose up"]
    end

    DEV --> LINT
    LINT --> TEST_GO
    LINT --> TEST_REACT
    TEST_GO --> BUILD_GO
    TEST_REACT --> BUILD_REACT
    BUILD_GO --> RELEASE
    BUILD_REACT --> RELEASE
    RELEASE --> DEPLOY

    style LINT fill:#FFA500,stroke:#333,color:#fff
    style TEST_GO fill:#4A90D9,stroke:#333,color:#fff
    style TEST_REACT fill:#61DAFB,stroke:#333,color:#000
    style BUILD_GO fill:#00ADD8,stroke:#333,color:#fff
    style BUILD_REACT fill:#61DAFB,stroke:#333,color:#000
    style RELEASE fill:#28A745,stroke:#333,color:#fff
    style DEPLOY fill:#6F42C1,stroke:#333,color:#fff
```

---

## Pipeline CI — Detalhado

```mermaid
flowchart TD
    START["Push to main / PR opened"] --> CHECKOUT["Checkout code"]
    CHECKOUT --> PARALLEL_LINT

    subgraph PARALLEL_LINT["Lint (paralelo)"]
        GOLINT["golangci-lint run ./..."]
        ESLINT["npx eslint src/"]
    end

    PARALLEL_LINT --> PARALLEL_TEST

    subgraph PARALLEL_TEST["Testes (paralelo)"]
        direction TB
        GO_UNIT["Go Unit Tests<br/>go test -race -cover ./..."]
        REACT_UNIT["React Tests<br/>vitest run --coverage"]
        GO_INTEG["Go Integration Tests<br/>testcontainers-go + PostgreSQL"]
    end

    PARALLEL_TEST --> BUILD_CHECK{"Todos passaram?"}

    BUILD_CHECK -->|Sim| PARALLEL_BUILD
    BUILD_CHECK -->|Não| FAIL["❌ Pipeline Failed<br/>Notificar desenvolvedor"]

    subgraph PARALLEL_BUILD["Build (paralelo)"]
        BUILD_AGENT["Build Agent<br/>GOOS=windows GOARCH=amd64<br/>go build -o agent.exe"]
        BUILD_SERVER["Build Server<br/>GOOS=linux GOARCH=amd64<br/>go build -o server"]
        BUILD_WEB["Build Dashboard<br/>npm run build"]
    end

    PARALLEL_BUILD --> COVERAGE{"Coverage ≥ threshold?"}
    COVERAGE -->|Sim| SUCCESS["✅ CI Passed"]
    COVERAGE -->|Não| WARN["⚠️ CI Passed (coverage warning)"]

    style FAIL fill:#DC3545,stroke:#333,color:#fff
    style SUCCESS fill:#28A745,stroke:#333,color:#fff
    style WARN fill:#FFC107,stroke:#333,color:#000
```

---

## Pipeline CD — Deploy em Produção

```mermaid
flowchart TD
    TAG["Criar tag v*.*.*<br/>(git tag -a v1.0.0)"] --> CI_PASS{"CI Passed?"}
    CI_PASS -->|Não| ABORT["❌ Abortar release"]
    CI_PASS -->|Sim| RELEASE["Criar GitHub Release<br/>Upload: agent.exe, server, checksums"]
    
    RELEASE --> DEPLOY_DECIDE{"Deploy automático<br/>ou manual?"}
    
    DEPLOY_DECIDE -->|Manual| MANUAL_DEPLOY
    DEPLOY_DECIDE -->|Automático| AUTO_DEPLOY

    subgraph MANUAL_DEPLOY["Deploy Manual (Fase 1)"]
        SSH["SSH no servidor"]
        GIT_PULL["cd /opt/inventory<br/>git fetch --tags<br/>git checkout v*.*.*"]
        BACKUP["Executar backup<br/>scripts/backup.sh"]
        COMPOSE_UP["docker compose up -d --build"]
        HEALTH["Verificar health<br/>curl /healthz + /readyz"]
        MONITOR["Monitorar 30min"]
    end

    subgraph AUTO_DEPLOY["Deploy Automático (Fase 2+)"]
        WEBHOOK["GitHub Webhook"]
        PULL_AUTO["git pull automático"]
        COMPOSE_AUTO["docker compose up -d --build"]
        HEALTH_AUTO["Health check automático"]
        ROLLBACK_AUTO["Rollback se unhealthy"]
    end

    SSH --> GIT_PULL
    GIT_PULL --> BACKUP
    BACKUP --> COMPOSE_UP
    COMPOSE_UP --> HEALTH
    HEALTH --> MONITOR

    MONITOR --> VERIFY{"Tudo OK?"}
    VERIFY -->|Sim| DONE["✅ Deploy completo"]
    VERIFY -->|Não| ROLLBACK

    subgraph ROLLBACK["Rollback"]
        GIT_PREV["git checkout v<anterior>"]
        COMPOSE_PREV["docker compose up -d --build"]
        RESTORE_DB["Restaurar backup se necessário"]
    end

    style ABORT fill:#DC3545,stroke:#333,color:#fff
    style DONE fill:#28A745,stroke:#333,color:#fff
```

---

## Deploy do Agent — Estratégias

```mermaid
flowchart TD
    AGENT_RELEASE["Nova versão do agent<br/>disponível na release"] --> STRATEGY{"Estratégia de deploy?"}

    STRATEGY -->|"1-5 estações"| MANUAL
    STRATEGY -->|"6+ estações"| MASS

    subgraph MANUAL["Deploy Manual"]
        COPY_BIN["Copiar agent.exe<br/>para cada estação"]
        STOP_SVC["Stop-Service InventoryAgent"]
        REPLACE["Substituir binário"]
        START_SVC["Start-Service InventoryAgent"]
        VERIFY_M["Verificar log + dashboard"]
    end

    subgraph MASS["Deploy em Massa"]
        direction TB
        GPO["Opção A: GPO<br/>Startup Script"]
        PSREMOTE["Opção B: PowerShell<br/>Invoke-Command -ComputerName"]
        SCCM["Opção C: SCCM/MECM<br/>Package deployment"]
    end

    COPY_BIN --> STOP_SVC --> REPLACE --> START_SVC --> VERIFY_M

    GPO --> VERIFY_MASS
    PSREMOTE --> VERIFY_MASS
    SCCM --> VERIFY_MASS

    VERIFY_MASS["Verificar no dashboard<br/>que todos os devices<br/>estão com versão nova"]
```

---

## GitHub Actions Workflow — Exemplo

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Go Lint
        uses: golangci/golangci-lint-action@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: React Lint
        working-directory: web
        run: npm ci && npx eslint src/

  test-go:
    runs-on: ubuntu-latest
    needs: lint
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: inventory
          POSTGRES_PASSWORD: test
          POSTGRES_DB: inventory_test
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...
      - name: Check coverage
        run: go tool cover -func=coverage.out

  test-react:
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - working-directory: web
        run: npm ci && npx vitest run --coverage

  build:
    runs-on: ubuntu-latest
    needs: [test-go, test-react]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build Agent (Windows)
        run: GOOS=windows GOARCH=amd64 go build -o agent.exe ./agent
      - name: Build Server (Linux)
        run: GOOS=linux GOARCH=amd64 go build -o server ./server
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Build Dashboard
        working-directory: web
        run: npm ci && npm run build
```

---

## Checklist de Deploy

| # | Verificação | Comando | Esperado |
|---|---|---|---|
| 1 | CI passou | GitHub Actions | ✅ verde |
| 2 | Backup realizado | `scripts/backup.sh` | Arquivo .dump criado |
| 3 | Containers rodando | `docker compose ps` | 3 running |
| 4 | API alive | `curl /healthz` | `{"status":"ok"}` |
| 5 | DB connected | `curl /readyz` | `{"status":"ready"}` |
| 6 | Logs sem erros | `docker compose logs --tail=50 api` | Sem ERROR |
| 7 | Dashboard carrega | Browser → :3000 | Login funciona |
| 8 | Agent reporta | Dashboard → devices | last_seen atualizado |

---

## Referências

- [Gestão de Liberação e Implantação](../03-transicao-de-servico/gestao-de-liberacao-e-implantacao.md)
- [Gestão de Mudanças](../03-transicao-de-servico/gestao-de-mudancas.md)
- [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md)
