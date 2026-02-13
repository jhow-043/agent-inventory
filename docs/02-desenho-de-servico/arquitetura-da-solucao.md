# Arquitetura da Solução

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Documentar a arquitetura técnica completa do Sistema de Inventário de Ativos de TI, incluindo componentes, tecnologias, decisões arquiteturais e padrões de projeto.

---

## 2. Escopo

Arquitetura da Fase 1: coleta de inventário de ativos Windows, API central, banco de dados PostgreSQL e dashboard web.

---

## 3. Visão Geral da Arquitetura

### 3.1 Estilo Arquitetural

**Monólito modular** — cada componente é uma unidade independente (módulo Go ou aplicação React), mas deployado de forma simples via Docker Compose em um único servidor.

**Justificativa:**
- Desenvolvedor solo (sem overhead de orquestração de microsserviços)
- Volume de 100–500 dispositivos (sem necessidade de escala horizontal)
- Deploy on-premises (simplicidade operacional)
- Fase 1 (microsserviços seriam premature optimization)

### 3.2 Diagrama de Arquitetura

```
                          REDE INTERNA (HTTP)
    ┌──────────────────────────────────────────────────────┐
    │                                                      │
    │  ┌───────────────┐      HTTP/JSON      ┌──────────────────────────────┐
    │  │ Windows Agent  │ ─────────────────→  │      Docker Host             │
    │  │ (Go binary)    │  Bearer: device-tkn │                              │
    │  │ Win Service     │                     │  ┌────────────────────────┐  │
    │  └───────────────┘                     │  │  API Container (Go)    │  │
    │                                        │  │  :8080                  │  │
    │  ┌───────────────┐      HTTP/JSON      │  │                        │  │
    │  │ Windows Agent  │ ─────────────────→  │  │  ┌──────────────────┐ │  │
    │  │ (Go binary)    │                     │  │  │ Gin Router       │ │  │
    │  └───────────────┘                     │  │  │  ├─ Middleware    │ │  │
    │                                        │  │  │  ├─ Handlers     │ │  │
    │  ┌───────────────┐                     │  │  │  ├─ Services     │ │  │
    │  │ ...N agents    │                     │  │  │  └─ Repositories │ │  │
    │  └───────────────┘                     │  │  └──────────────────┘ │  │
    │                                        │  └──────────┬───────────┘  │
    │                                        │             │ TCP:5432     │
    │  ┌───────────────┐      HTTP/JSON      │  ┌──────────▼───────────┐  │
    │  │ Browser        │ ─────────────────→  │  │ PostgreSQL Container │  │
    │  │ (Dashboard)    │  Bearer: JWT        │  │ :5432                │  │
    │  └───────────────┘                     │  │ Volume persistente   │  │
    │         ▲                               │  └──────────────────────┘  │
    │         │ HTTP                          │                              │
    │  ┌──────┴────────┐                     │  ┌────────────────────────┐  │
    │  │ Web Container  │                     │  │ Nginx (dashboard)     │  │
    │  │ (React SPA)    │◄────serve───────── │  │ :3000                  │  │
    │  └───────────────┘                     │  └────────────────────────┘  │
    │                                        └──────────────────────────────┘
    └──────────────────────────────────────────────────────┘
```

> Diagrama Mermaid detalhado disponível em [Anexos — Arquitetura Geral](../06-anexos/diagramas/arquitetura-geral.md).

---

## 4. Stack Tecnológica

### 4.1 Backend — API e Agent

| Componente | Tecnologia | Versão Mínima | Justificativa |
|---|---|---|---|
| Linguagem | Go | 1.22+ | Binário único, tipado, excelente performance, bom ecossistema |
| Framework HTTP | Gin | v1.9+ | Maduro (2014+), segue `net/http`, middleware rico, bem documentado |
| Acesso a dados | sqlx + pgx | pgx v5 | Controle explícito de SQL, queries previsíveis, superior a ORMs |
| Migrations | golang-migrate | v4 | Migrations versionadas, suporte a PostgreSQL, CLI disponível |
| Logging | log/slog (stdlib) | Go 1.21+ | Built-in, estruturado, zero dependência |
| Config | caarlos0/env | v11+ | Mapeamento direto env vars → struct, type-safe |
| Validação | go-playground/validator | v10 | Validação por tags, integrado com Gin |
| JWT | golang-jwt/jwt | v5 | Padrão da comunidade Go para JWT |
| Windows Service | kardianos/service | v1 | Abstração cross-platform de serviços do SO |
| WMI | yusufpapurcu/wmi | latest | Fork mantido do StackExchange/wmi para coleta Windows |

#### Gin vs Fiber — Decisão

| Critério | Gin | Fiber |
|---|---|---|
| Suporte a `net/http` | ✅ Nativo | ❌ Usa `fasthttp` (incompatível) |
| Maturidade | Desde 2014 | Desde 2020 |
| Middleware do ecossistema | ✅ Compatível com todos | ⚠️ Só Fiber-específico |
| Performance (500 devices) | Mais que suficiente | Vantagem irrelevante nesse volume |
| **Decisão** | **✅ Escolhido** | Descartado |

#### sqlx vs GORM — Decisão

| Critério | sqlx | GORM |
|---|---|---|
| Controle do SQL | ✅ Total | ❌ Abstrai demais |
| Previsibilidade de queries | ✅ Você escreve o SQL | ❌ ORM decide |
| Fit com Clean Architecture | ✅ Repository com SQL explícito | ⚠️ Acopla camadas |
| Curva de aprendizado | Requer SQL | Mais simples para CRUD |
| **Decisão** | **✅ Escolhido** | Descartado |

### 4.2 Frontend — Dashboard

| Componente | Tecnologia | Versão | Justificativa |
|---|---|---|---|
| Framework | React | 18+ | Ecossistema maduro, suporte amplo, crescimento futuro |
| Linguagem | TypeScript | 5+ | Type safety, melhor DX, menos bugs |
| Build tool | Vite | 5+ | CRA deprecated, Vite 10-100× mais rápido em dev |
| CSS | Tailwind CSS | 3+ | Utility-first, produtivo, sem CSS custom extenso |
| Componentes UI | Shadcn/UI | latest | Componentes copiados (sem dep runtime), customizáveis |
| State (server) | TanStack Query | v5 | Gerenciamento de server state, cache, retry automático |
| Roteamento | React Router | v6+ | Padrão da comunidade React |
| Testes | Vitest + Testing Library | latest | Compatível com Vite, rápido, API familiar |

### 4.3 Banco de Dados

| Componente | Tecnologia | Versão | Justificativa |
|---|---|---|---|
| RDBMS | PostgreSQL | 16+ | UUID nativo, JSONB, maduro, excelente para dados relacionais |
| Driver Go | pgx | v5 | Driver nativo (não usa `lib/pq`), melhor performance |
| Migrations | golang-migrate | v4 | CLI + lib, SQL puro, versionamento sequencial |

### 4.4 Infraestrutura

| Componente | Tecnologia | Justificativa |
|---|---|---|
| Containerização | Docker + Docker Compose | Simplifica deploy on-premises |
| CI/CD | GitHub Actions | Pipeline automatizado sem custo |
| Build automation | Makefile | Standard, cross-platform com WSL |
| Linting Go | golangci-lint | Agrega múltiplos linters |
| Linting JS | ESLint + Prettier | Padrão React/TypeScript |

---

## 5. Estrutura do Projeto

### 5.1 Go Workspace

O projeto utiliza **Go Workspace** (`go.work`) para gerenciar múltiplos módulos Go com dependências distintas:

```
inventario/
├── go.work                       # Go workspace linking modules
│
├── shared/                       # Contratos compartilhados (Agent ↔ API)
│   ├── go.mod
│   └── models/
│       └── inventory.go          # Payload structs
│
├── agent/                        # Windows Agent
│   ├── go.mod
│   ├── cmd/
│   │   └── agent/
│   │       └── main.go           # Entry point
│   └── internal/
│       ├── collector/            # Coletores de dados
│       │   ├── system.go         # Hostname, OS, serial, boot, user
│       │   ├── hardware.go       # CPU, RAM, discos, motherboard, BIOS
│       │   ├── network.go        # Interfaces, IP, MAC
│       │   ├── software.go       # Programas instalados
│       │   └── license.go        # Status de ativação Windows
│       ├── config/
│       │   └── config.go         # Configuração do agent
│       ├── service/
│       │   └── windows.go        # Wrapper Windows Service
│       └── transport/
│           ├── client.go         # HTTP client
│           └── retry.go          # Backoff exponencial
│
├── server/                       # API Central
│   ├── go.mod
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go         # Configuração do server
│   │   ├── domain/
│   │   │   ├── device.go         # Entidade Device
│   │   │   ├── hardware.go       # Entidade Hardware
│   │   │   ├── disk.go           # Entidade Disk
│   │   │   ├── network.go        # Entidade NetworkInterface
│   │   │   ├── software.go       # Entidade InstalledSoftware
│   │   │   ├── user.go           # Entidade User (dashboard auth)
│   │   │   ├── errors.go         # Erros de domínio tipados
│   │   │   └── interfaces.go     # Interfaces (contracts)
│   │   ├── handler/
│   │   │   ├── auth.go           # Login, refresh
│   │   │   ├── device.go         # List, get detail
│   │   │   ├── inventory.go      # Register, submit inventory
│   │   │   ├── dashboard.go      # Stats
│   │   │   └── health.go         # Healthz, readyz
│   │   ├── middleware/
│   │   │   ├── auth.go           # JWT validation
│   │   │   ├── device_auth.go    # Device token validation
│   │   │   ├── cors.go           # CORS configuration
│   │   │   ├── logging.go        # Request logging
│   │   │   ├── ratelimit.go      # Rate limiting
│   │   │   └── requestid.go      # Request ID generation
│   │   ├── repository/
│   │   │   ├── device.go         # Device SQL queries
│   │   │   ├── hardware.go       # Hardware SQL queries
│   │   │   ├── inventory.go      # Inventory upsert (transactional)
│   │   │   ├── software.go       # Software SQL queries
│   │   │   ├── network.go        # Network SQL queries
│   │   │   └── user.go           # User SQL queries
│   │   └── service/
│   │       ├── auth.go           # Auth business logic
│   │       ├── device.go         # Device business logic
│   │       └── inventory.go      # Inventory business logic
│   └── migrations/
│       ├── 000001_create_users.up.sql
│       ├── 000001_create_users.down.sql
│       ├── 000002_create_devices.up.sql
│       ├── 000002_create_devices.down.sql
│       ├── 000003_create_device_tokens.up.sql
│       ├── 000003_create_device_tokens.down.sql
│       ├── 000004_create_hardware.up.sql
│       ├── 000004_create_hardware.down.sql
│       ├── 000005_create_disks.up.sql
│       ├── 000005_create_disks.down.sql
│       ├── 000006_create_network_interfaces.up.sql
│       ├── 000006_create_network_interfaces.down.sql
│       ├── 000007_create_installed_software.up.sql
│       └── 000007_create_installed_software.down.sql
│
├── web/                          # Dashboard React
│   ├── public/
│   ├── src/
│   │   ├── components/           # Componentes reutilizáveis
│   │   │   ├── ui/               # Shadcn/UI components
│   │   │   ├── layout/           # Header, Sidebar, Footer
│   │   │   └── devices/          # DeviceCard, DeviceTable, etc.
│   │   ├── pages/
│   │   │   ├── LoginPage.tsx
│   │   │   ├── DashboardPage.tsx
│   │   │   ├── DeviceListPage.tsx
│   │   │   └── DeviceDetailPage.tsx
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   └── useDevices.ts
│   │   ├── services/
│   │   │   ├── api.ts            # HTTP client base
│   │   │   ├── authService.ts
│   │   │   └── deviceService.ts
│   │   ├── types/
│   │   │   ├── device.ts
│   │   │   └── auth.ts
│   │   ├── lib/
│   │   │   └── utils.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   └── vite.config.ts
│
├── docker-compose.yml            # Produção
├── docker-compose.dev.yml        # Desenvolvimento
├── Makefile
├── .gitignore
├── .github/
│   └── workflows/
│       └── ci.yml
├── README.md
└── docs/                         # Esta documentação
```

**Justificativa do Go Workspace:** Agent e Server possuem dependências distintas (agent depende de WMI/DPAPI exclusivos de Windows, server depende de PostgreSQL/Gin). Módulos separados permitem compilação independente e dependências limpas. O módulo `shared/` evita duplicação dos modelos de payload.

---

## 6. Padrões de Projeto

### 6.1 Clean Architecture (Simplificada)

```
            HTTP Request
                 │
         ┌───────▼───────┐
         │   Handler      │  Parsing, validação de input, serialização
         │   (Controller) │
         └───────┬───────┘
                 │ DTO → Domain
         ┌───────▼───────┐
         │   Service      │  Regras de negócio, orquestração
         │   (Use Case)   │
         └───────┬───────┘
                 │ Domain
         ┌───────▼───────┐
         │  Repository    │  Acesso a dados, SQL queries
         │  (Data Access) │
         └───────┬───────┘
                 │ SQL
         ┌───────▼───────┐
         │  PostgreSQL    │
         └───────────────┘
```

| Camada | Diretório | Responsabilidade | Depende de |
|---|---|---|---|
| **Domain** | `domain/` | Entidades, interfaces (contratos), erros tipados | Nada |
| **Handler** | `handler/` | Parse HTTP, validação, serialização de resposta | Service |
| **Service** | `service/` | Lógica de negócio, orquestração | Repository (via interface) |
| **Repository** | `repository/` | Queries SQL, mapeamento DB ↔ Domain | PostgreSQL |

**Regra de dependência:** As camadas só dependem de suas adjacentes internas. Domain não depende de nada. Repositories implementam interfaces definidas em Domain.

### 6.2 Dependency Injection (Manual)

Sem frameworks DI. Dependências construídas no `main.go` e injetadas via construtores:

```go
// main.go (pseudocódigo)
db := database.Connect(cfg.DatabaseURL)
deviceRepo := repository.NewDeviceRepository(db)
deviceSvc := service.NewDeviceService(deviceRepo)
deviceHandler := handler.NewDeviceHandler(deviceSvc)
```

### 6.3 Interfaces em Domain

```go
// domain/interfaces.go (pseudocódigo)
type DeviceRepository interface {
    FindAll(ctx context.Context, filter DeviceFilter) ([]Device, error)
    FindByID(ctx context.Context, id uuid.UUID) (*Device, error)
    Upsert(ctx context.Context, device *Device) error
}
```

### 6.4 Erros de Domínio Tipados

```go
// domain/errors.go (pseudocódigo)
var (
    ErrDeviceNotFound  = errors.New("device not found")
    ErrInvalidToken    = errors.New("invalid device token")
    ErrDuplicateDevice = errors.New("device already registered")
)
```

### 6.5 DTO Separation

- **Request DTOs:** Estruturas específicas para decode do JSON de entrada
- **Domain Entities:** Estruturas internas com regras de negócio
- **Response DTOs:** Estruturas específicas para serialização da resposta
- **Shared Models:** Contratos compartilhados entre agent e server (module `shared/`)

### 6.6 Transações

O upsert de inventário ocorre em uma única transação SQL:
1. Upsert `devices` (by serial_number)
2. Upsert `hardware` (by device_id)
3. Delete + re-insert `disks` (by device_id)
4. Delete + re-insert `network_interfaces` (by device_id)
5. Delete + re-insert `installed_software` (by device_id)
6. Update `last_seen`

Se qualquer passo falhar, toda a transação é revertida.

### 6.7 Graceful Shutdown

API e Agent interceptam sinais do SO (`SIGTERM`, `SIGINT`, `Ctrl+C`) e executam shutdown graceful:
1. Para de aceitar novas requisições
2. Aguarda requisições em andamento terminarem (timeout 30s)
3. Fecha pool de conexões do banco
4. Registra shutdown nos logs
5. Retorna exit code 0

---

## 7. Modelo de Comunicação

### 7.1 Agent → API

| Aspecto | Detalhe |
|---|---|
| **Protocolo** | HTTP/1.1 (Fase 1) |
| **Formato** | JSON |
| **Direção** | Agent → API (somente outbound, nunca inbound) |
| **Autenticação** | Bearer token (device token) |
| **Retry** | Backoff exponencial: 2s → 4s → 8s → 16s → 32s → max 5min |
| **Timeout** | 30s por request |
| **Jitter** | ±15% no intervalo de coleta |

### 7.2 Dashboard → API

| Aspecto | Detalhe |
|---|---|
| **Protocolo** | HTTP/1.1 (Fase 1) |
| **Formato** | JSON |
| **Autenticação** | JWT (access token via httpOnly cookie) |
| **Token refresh** | Automático via interceptor quando access token expira |
| **CORS** | Restrito às origens permitidas (configurável) |

### 7.3 API → PostgreSQL

| Aspecto | Detalhe |
|---|---|
| **Protocolo** | TCP (porta 5432) |
| **Driver** | pgx v5 (nativo Go) |
| **Connection pool** | Máximo 25 conexões (configurável) |
| **Timeout** | 5s para conexão, 30s para query |

---

## 8. Configuração

### 8.1 API Server — Variáveis de Ambiente

| Variável | Descrição | Exemplo | Obrigatório |
|---|---|---|---|
| `DATABASE_URL` | Connection string do PostgreSQL | `postgres://user:pass@localhost:5432/inventory?sslmode=disable` | Sim |
| `SERVER_PORT` | Porta HTTP da API | `8080` | Não (default: 8080) |
| `JWT_SECRET` | Chave secreta para assinar JWT | `random-32-char-string` | Sim |
| `JWT_ACCESS_TTL` | Duração do access token | `15m` | Não (default: 15m) |
| `JWT_REFRESH_TTL` | Duração do refresh token | `168h` | Não (default: 7d) |
| `ENROLLMENT_KEY` | Chave de registro para novos agents | `random-enrollment-key` | Sim |
| `RATE_LIMIT_RPM` | Requests por minuto por device | `10` | Não (default: 10) |
| `LOG_LEVEL` | Nível de log | `info` | Não (default: info) |
| `CORS_ORIGINS` | Origens permitidas para CORS | `http://localhost:3000` | Não |
| `RUN_MIGRATIONS` | Executar migrations no startup | `true` | Não (default: false) |

### 8.2 Agent — Arquivo de Configuração (YAML)

```yaml
# agent-config.yaml
api_url: "http://192.168.1.100:8080"
enrollment_key: "random-enrollment-key"
collection_interval: "4h"
log_level: "info"
log_file: "C:\\ProgramData\\InventoryAgent\\agent.log"
```

> O token de dispositivo é armazenado separadamente após o primeiro registro.

---

## 9. Decisões Arquiteturais (ADRs)

### ADR-001: Monólito modular sobre microsserviços

- **Contexto:** Sistema Fase 1, desenvolvedor solo, 100-500 devices
- **Decisão:** Monólito modular com módulos Go separados
- **Justificativa:** Simplicidade operacional, sem overhead de orquestração
- **Consequência:** Escala vertical; redesign necessário acima de ~5.000 devices

### ADR-002: HTTP na Fase 1

- **Contexto:** Deploy on-premises em rede interna
- **Decisão:** Comunicação via HTTP puro (sem TLS)
- **Justificativa:** Simplifica setup inicial; rede interna controlada
- **Risco aceito:** Tokens em texto claro no transporte
- **Mitigação:** Rede segmentada, roadmap HTTPS documentado
- **Referência:** [Gestão de Segurança](gestao-de-seguranca.md)

### ADR-003: Gin sobre Fiber

- **Decisão:** Gin como framework HTTP
- **Justificativa:** Compatibilidade `net/http`, maturidade (12 anos), middleware rico

### ADR-004: sqlx sobre GORM

- **Decisão:** sqlx + pgx para acesso a dados
- **Justificativa:** Controle explícito do SQL, alinhado com Clean Architecture

### ADR-005: Vite sobre CRA

- **Decisão:** Vite como build tool do React
- **Justificativa:** CRA deprecated, Vite 10-100× mais rápido em dev

### ADR-006: httpOnly cookies sobre localStorage

- **Decisão:** JWT em httpOnly cookies em vez de localStorage
- **Justificativa:** Proteção contra XSS (JavaScript não pode acessar o token)

### ADR-007: Enrollment key + device token

- **Decisão:** Agent registra-se com uma enrollment key compartilhada e recebe um device token único
- **Justificativa:** Prático para 100-500 devices, melhor que pré-gerar tokens

### ADR-008: log/slog sobre zerolog/zap

- **Decisão:** Usar log/slog (stdlib)
- **Justificativa:** Built-in desde Go 1.21, zero dependência, suficiente para Fase 1

---

## 10. Referências

- [Gestão de Segurança](gestao-de-seguranca.md)
- [Gestão de Capacidade](gestao-de-capacidade.md)
- [Gestão de Disponibilidade](gestao-de-disponibilidade.md)
- [Diagrama — Arquitetura Geral](../06-anexos/diagramas/arquitetura-geral.md)
- [Diagrama — Fluxo de Comunicação](../06-anexos/diagramas/fluxo-de-comunicacao.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
