# Inventory System — Documentação Técnica Completa

**Versão:** 1.0.0 | **Data:** 18/02/2026 | **Service ID:** SVC-INV-001

Sistema de inventário automatizado de hardware e software para estações Windows.
Um agente Windows coleta informações via WMI/Registry e envia para uma API centralizada, que alimenta um dashboard web com tema dark/light e gráficos analíticos.

---

## Sumário

1. [Arquitetura](#arquitetura)
2. [Stack Tecnológico](#stack-tecnológico)
3. [Estrutura do Projeto](#estrutura-do-projeto)
4. [Quick Start](#quick-start)
5. [API Endpoints](#api-endpoints)
6. [Modelo de Segurança](#modelo-de-segurança)
7. [Dados Coletados pelo Agent](#dados-coletados-pelo-agent)
8. [Database Schema](#database-schema)
9. [Variáveis de Ambiente](#variáveis-de-ambiente)
10. [CI/CD](#cicd)
11. [Makefile](#makefile)

---

## Arquitetura

```
┌────────────────────────────────────────────────────────────────────────┐
│                          REDE CORPORATIVA                              │
│                                                                        │
│  ┌──────────────┐                                                      │
│  │  Estação #1  │──┐                                                   │
│  │  Agent.exe   │  │     HTTP/JSON               ┌──────────────────┐  │
│  └──────────────┘  │  ┌────────────────────┐     │                  │  │
│  ┌──────────────┐  ├─►│   API Server       │     │   PostgreSQL 16  │  │
│  │  Estação #2  │──┤  │   Go 1.24 + Gin    │────►│                  │  │
│  │  Agent.exe   │  │  │   Port 8080        │     │  10 tabelas      │  │
│  └──────────────┘  │  └────────┬───────────┘     │  6 migrations    │  │
│  ┌──────────────┐  │           │                  │                  │  │
│  │  Estação #N  │──┘           │                  └──────────────────┘  │
│  │  Agent.exe   │              │                                       │
│  └──────────────┘     ┌────────┴───────────┐                           │
│                       │   Frontend SPA     │                           │
│  ┌──────────────┐     │   React 19 + Vite 7│                           │
│  │   Browser    │────►│   Tailwind 4       │                           │
│  │   (Admin)    │     │   Recharts 3.7     │                           │
│  └──────────────┘     └────────────────────┘                           │
└────────────────────────────────────────────────────────────────────────┘
```

**Padrão arquitetural:** Monolito modular — cada componente é uma unidade independente (módulo Go ou React SPA) deployada via Docker Compose.

**Comunicação:** Agent → API only (agent nunca aceita conexões inbound). Dashboard comunica com API via REST com cookie JWT httpOnly.

## Stack Tecnológico

| Componente | Tecnologia | Versão |
|---|---|---|
| **API** | Go, Gin, sqlx, pgx v5, golang-migrate, JWT HS256 | Go 1.24 |
| **Agent** | Go, WMI, Windows Service (`x/sys/windows/svc`) | Go 1.24 |
| **Frontend** | React, TypeScript, Vite, Tailwind CSS, TanStack Query, Recharts | React 19, Vite 7, Tailwind 4 |
| **Database** | PostgreSQL | 16-alpine |
| **Deploy** | Docker Compose | — |
| **CI/CD** | GitHub Actions (6 jobs) | — |
| **Linting** | golangci-lint, ESLint | — |

## Estrutura do Projeto

```
Inventario/
├── docker-compose.yml                 # PostgreSQL + API containers
├── Makefile                           # Build/run targets
├── .env.example                       # Template de variáveis
├── go.work                            # Go workspace (shared + server + agent)
│
├── server/                            # ── API REST (Go) ──
│   ├── Dockerfile                     # Multi-stage: golang:1.24-alpine → alpine:3.20
│   ├── cmd/api/main.go                # Entry point + CLI create-user --role
│   ├── internal/
│   │   ├── config/config.go           # Env vars loader
│   │   ├── database/database.go       # Connect + pool + migrations
│   │   ├── handler/                   # Controladores HTTP
│   │   │   ├── auth.go                # Enroll, Login, Logout, Me
│   │   │   ├── device.go              # ListDevices, GetDevice, ExportCSV, HardwareHistory, UpdateStatus, UpdateDepartment
│   │   │   ├── department.go          # CRUD departamentos
│   │   │   ├── dashboard.go           # GetStats
│   │   │   ├── user.go                # CRUD usuários
│   │   │   ├── audit.go               # Audit logs query
│   │   │   ├── health.go              # Healthz, Readyz
│   │   │   └── inventory.go           # SubmitInventory
│   │   ├── middleware/                # Pipeline HTTP
│   │   │   ├── auth.go                # DeviceAuth (Bearer) + JWTAuth (cookie)
│   │   │   ├── rbac.go                # RequireRole("admin")
│   │   │   ├── ratelimit.go           # Rate limiting por IP
│   │   │   ├── security.go            # Security headers (CSP, X-Frame-Options)
│   │   │   ├── audit.go               # Audit logger
│   │   │   ├── cors.go                # CORS whitelist
│   │   │   ├── logging.go             # Structured slog JSON
│   │   │   └── requestid.go           # UUID X-Request-Id
│   │   ├── repository/               # Camada de dados (SQL)
│   │   │   ├── device.go              # Queries com filtros + paginação + sort
│   │   │   ├── inventory.go           # UPSERT transacional + hardware history
│   │   │   ├── department.go          # CRUD departamentos
│   │   │   ├── dashboard.go           # COUNT queries
│   │   │   ├── user.go                # CRUD usuários
│   │   │   ├── token.go               # SHA-256 token lookup
│   │   │   └── audit_log.go           # Audit logs
│   │   ├── service/                   # Lógica de negócio
│   │   │   ├── auth.go                # Enrollment, login (bcrypt + JWT), users
│   │   │   ├── device.go              # Orquestração queries
│   │   │   ├── department.go          # Regras departamentos
│   │   │   ├── dashboard.go           # Cálculo estatísticas
│   │   │   └── inventory.go           # Processamento inventário
│   │   └── router/router.go          # Rotas + grupos de auth
│   └── migrations/                    # SQL (embedded via go:embed)
│       ├── 001_init.up.sql            # 7 tabelas base + índices
│       ├── 002_remote_tools.up.sql    # remote_tools
│       ├── 003_disk_free_space.up.sql # drive_letter, free_space_bytes
│       ├── 004_lifecycle.up.sql       # departments, status, hardware_history
│       ├── 005_add_user_roles.up.sql  # RBAC (admin/viewer)
│       └── 006_add_audit_logs.up.sql  # audit_logs + 5 índices
│
├── agent/                             # ── WINDOWS AGENT (Go) ──
│   ├── config.example.json            # Template config.json
│   ├── cmd/agent/main.go              # Service + CLI (run/collect/install/start/stop/uninstall/version)
│   └── internal/
│       ├── client/client.go           # HTTP client com retry + backoff exponencial
│       ├── collector/                 # Coletores WMI + Registry
│       │   ├── collector.go           # Orquestrador
│       │   ├── system.go              # OS, hostname, serial, usuário logado
│       │   ├── hardware.go            # CPU, RAM, motherboard, BIOS
│       │   ├── disk.go                # Discos + partições + espaço livre
│       │   ├── network.go             # Interfaces de rede (físicas + IPs)
│       │   ├── software.go            # Software instalado (Registry HKLM/HKCU)
│       │   ├── license.go             # Licença Windows (WMI)
│       │   └── remote.go              # TeamViewer, AnyDesk, RustDesk (v1.4+)
│       ├── config/config.go           # JSON config loader
│       └── token/store.go             # Token persistence (0600)
│
├── shared/                            # ── MÓDULO COMPARTILHADO ──
│   ├── models/models.go              # Entidades DB
│   └── dto/                           # Request/Response DTOs
│       ├── requests.go
│       └── responses.go
│
├── frontend/                          # ── DASHBOARD WEB (React) ──
│   ├── vite.config.ts                 # Proxy /api → localhost:8081
│   └── src/
│       ├── api/                       # HTTP client + endpoints
│       │   ├── client.ts              # Fetch wrapper, auto-redirect 401
│       │   ├── auth.ts, devices.ts, departments.ts, dashboard.ts, users.ts
│       ├── components/
│       │   ├── Layout.tsx             # Sidebar colapsável + tema toggle
│       │   ├── ProtectedRoute.tsx     # Auth guard
│       │   ├── ErrorBoundary.tsx      # Captura erros React
│       │   └── ui/                    # Badge, Button, Card, Input, Modal, PageHeader, Select
│       ├── hooks/                     # useAuth, useTheme, useDebounce, useSidebar
│       ├── pages/
│       │   ├── Login.tsx              # Glass-morphism, gradientes
│       │   ├── Dashboard.tsx          # 4 KPIs + gráficos Recharts (pizza + barras)
│       │   ├── DeviceList.tsx         # Tabela + filtros + paginação + CSV export
│       │   ├── DeviceDetail.tsx       # Detalhes completos + hardware history
│       │   ├── Departments.tsx        # CRUD + contagem devices por departamento
│       │   ├── DepartmentDetail.tsx   # Devices vinculados a um departamento
│       │   └── Settings.tsx           # Tema dark/light + gerenciamento usuários
│       └── types/index.ts            # Interfaces TypeScript
│
└── docs/                              # Documentação ITIL v4 (30+ documentos)
```

## Quick Start

### 1. Pré-requisitos

- Docker + Docker Compose
- Node.js 20+ (para frontend)
- Go 1.24+ (para desenvolvimento)

### 2. Subir o Backend

```bash
# Copiar e preencher variáveis de ambiente
cp .env.example .env
# Editar .env: JWT_SECRET (min 32 chars), ENROLLMENT_KEY, SERVER_PORT

# Subir PostgreSQL + API (migrations automáticas)
docker compose up -d --build

# Criar usuário admin
docker compose exec api ./server create-user --username admin --password SenhaSegura123 --role admin

# Criar usuário somente leitura (opcional)
docker compose exec api ./server create-user --username viewer --password Viewer123 --role viewer
```

### 3. Subir o Frontend

```bash
cd frontend
npm install
npm run dev
# Acesse http://localhost:5173
```

### 4. Configurar o Agent (Windows)

```powershell
cd agent
copy config.example.json config.json
# Editar config.json: server_url, enrollment_key

# Build
go build -o inventory-agent.exe ./cmd/agent

# Testar coleta local (sem servidor)
.\inventory-agent.exe collect

# Executar em foreground (debug)
.\inventory-agent.exe run -config config.json

# Instalar como Windows Service (produção)
.\inventory-agent.exe install -config "C:\ProgramData\InventoryAgent\config.json"
.\inventory-agent.exe start
```

**Config do agent (`config.json`):**
```json
{
  "server_url": "http://SERVER_IP:8081",
  "enrollment_key": "mesma-chave-do-env",
  "interval_hours": 1
}
```

---

## API Endpoints

### Saúde do Serviço (sem auth)

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/readyz` | Readiness probe (testa conexão DB) |

### Agent (auth própria)

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| `POST` | `/api/v1/enroll` | Header `X-Enrollment-Key` | Registra dispositivo, retorna token |
| `POST` | `/api/v1/inventory` | Bearer Token (device) | Envia inventário completo |

### Dashboard — Leitura (JWT: qualquer role)

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/v1/auth/login` | Login (cookie httpOnly JWT, rate limit: 5/min) |
| `GET` | `/api/v1/auth/me` | Dados do usuário logado (id, username, role) |
| `POST` | `/api/v1/auth/logout` | Logout (limpa cookie) |
| `GET` | `/api/v1/dashboard/stats` | Estatísticas: total, online, offline, inactive |
| `GET` | `/api/v1/devices` | Listar dispositivos (filtros, paginação, sort) |
| `GET` | `/api/v1/devices/export` | Exportar CSV |
| `GET` | `/api/v1/devices/:id` | Detalhe completo do dispositivo |
| `GET` | `/api/v1/devices/:id/hardware-history` | Histórico de mudanças de hardware |
| `GET` | `/api/v1/departments` | Listar departamentos |
| `GET` | `/api/v1/users` | Listar usuários |

### Dashboard — Admin Only (JWT: role admin)

| Método | Rota | Descrição |
|---|---|---|
| `PATCH` | `/api/v1/devices/:id/status` | Alterar status (active/inactive) |
| `PATCH` | `/api/v1/devices/:id/department` | Atribuir departamento |
| `POST` | `/api/v1/departments` | Criar departamento |
| `PUT` | `/api/v1/departments/:id` | Renomear departamento |
| `DELETE` | `/api/v1/departments/:id` | Excluir departamento |
| `POST` | `/api/v1/users` | Criar usuário (com role) |
| `DELETE` | `/api/v1/users/:id` | Excluir usuário |
| `GET` | `/api/v1/audit-logs` | Logs de auditoria |
| `GET` | `/api/v1/audit-logs/:type/:id` | Logs de recurso específico |

### Query Params — `GET /api/v1/devices`

| Param | Tipo | Default | Descrição |
|---|---|---|---|
| `hostname` | string | — | Filtro ILIKE por hostname |
| `os` | string | — | Filtro ILIKE por OS |
| `status` | string | — | `online`, `offline`, `inactive` |
| `department_id` | UUID | — | Filtro por departamento |
| `sort` | string | `hostname` | Coluna: hostname, os, last_seen, status |
| `order` | string | `asc` | Direção: asc, desc |
| `page` | int | `1` | Página |
| `limit` | int | `50` | Itens/página (máx: 100) |

---

## Modelo de Segurança

### Autenticação

| Ator | Mecanismo | Armazenamento | Validade |
|---|---|---|---|
| **Usuário (Dashboard)** | JWT HS256 via cookie httpOnly | Cookie `session` | 24 horas |
| **Agent (Device)** | Token UUID via Bearer header | SHA-256 hash no PostgreSQL | Indefinido |

### RBAC (Role-Based Access Control)

| Role | Permissões |
|---|---|
| **admin** | Tudo: visualizar + criar/editar/excluir dispositivos, departamentos, usuários, audit logs |
| **viewer** | Apenas visualizar: dashboard, dispositivos, departamentos, lista de usuários |

### Medidas de Segurança

| Medida | Detalhes |
|---|---|
| Senhas | bcrypt cost 12 (~250ms/hash) |
| Enrollment key | Comparação constant-time (`crypto/subtle`) |
| Rate limiting | Por IP — Login: 5/min, Enrollment: 10/min |
| Security headers | CSP, X-Frame-Options: DENY, X-Content-Type-Options: nosniff |
| CORS | Whitelist configurável, credentials permitidos |
| Audit logging | Todas as ações admin registradas (quem, o quê, quando, IP) |
| Graceful shutdown | 10s para finalizar requisições ativas |

### Ações Auditadas

`auth.login`, `auth.logout`, `device.status.update`, `device.department.update`, `department.create`, `department.update`, `department.delete`, `user.create`, `user.delete`

---

## Dados Coletados pelo Agent

| Categoria | Dados | Fonte |
|---|---|---|
| **Sistema** | Hostname, Serial, OS (nome/versão/build/arch), uptime, usuário logado | WMI: `Win32_OperatingSystem`, `Win32_BIOS`, `Win32_ComputerSystem` |
| **CPU** | Modelo, cores físicos, threads lógicos | WMI: `Win32_Processor` |
| **RAM** | Total em bytes | WMI: `Win32_PhysicalMemory` |
| **Placa-Mãe** | Fabricante, modelo, serial | WMI: `Win32_BaseBoard` |
| **BIOS** | Fabricante, versão SMBIOS | WMI: `Win32_BIOS` |
| **Discos** | Modelo, tamanho, tipo (SSD/HDD), serial, interface, letra, espaço livre | WMI: `Win32_DiskDrive`, `Win32_LogicalDisk` |
| **Rede** | Nome, MAC, IPv4, IPv6, velocidade, tipo (físico/virtual) | WMI: `Win32_NetworkAdapter` + Go `net.Interfaces()` |
| **Software** | Nome, versão, fabricante, data instalação | Registry: `HKLM/HKCU\SOFTWARE\...\Uninstall` (x86+x64) |
| **Licença** | Status ativação Windows | WMI: `SoftwareLicensingProduct` |
| **Acesso Remoto** | TeamViewer ID, AnyDesk ID, RustDesk ID (v1.4+ support) | Registry + config files + CLI fallback |

---

## Database Schema

**10 tabelas** em 6 migrations (auto-executadas no startup via `go:embed`).

| Tabela | Relação | Descrição |
|---|---|---|
| `users` | — | Contas dashboard (username, bcrypt hash, role admin/viewer) |
| `devices` | — | Dispositivos (hostname, serial, OS, status, department_id FK) |
| `device_tokens` | 1:1 device | Tokens SHA-256 dos agents |
| `hardware` | 1:1 device | CPU, RAM, motherboard, BIOS |
| `disks` | 1:N device | Discos + partições (modelo, tipo, letra, espaço livre) |
| `network_interfaces` | 1:N device | Interfaces de rede (MAC, IPs, velocidade) |
| `installed_software` | 1:N device | Software instalado |
| `remote_tools` | 1:N device | TeamViewer, AnyDesk, RustDesk |
| `departments` | — | Departamentos organizacionais |
| `hardware_history` | 1:N device | Snapshots JSONB de mudanças de hardware |
| `audit_logs` | N:1 user | Registro de ações admin |

### Migrations

| # | Arquivo | O que faz |
|---|---|---|
| 001 | `001_init.up.sql` | 7 tabelas base + índices |
| 002 | `002_remote_tools.up.sql` | Tabela remote_tools |
| 003 | `003_disk_free_space.up.sql` | drive_letter, free_space em disks |
| 004 | `004_lifecycle.up.sql` | departments, hardware_history, status em devices |
| 005 | `005_add_user_roles.up.sql` | RBAC (campo role + constraint) |
| 006 | `006_add_audit_logs.up.sql` | audit_logs + 5 índices |

---

## Variáveis de Ambiente

| Variável | Descrição | Default | Obrigatória |
|---|---|---|---|
| `DATABASE_URL` | Connection string PostgreSQL | `postgres://inventory:changeme@localhost:5432/inventory?sslmode=disable` | Sim |
| `POSTGRES_PASSWORD` | Senha do PostgreSQL (Docker) | `changeme` | Sim |
| `SERVER_PORT` | Porta da API no host | `8080` | Não |
| `JWT_SECRET` | Chave HS256 para JWT (≥32 chars) | — | **Sim** |
| `ENROLLMENT_KEY` | Chave de enrollment dos agents | — | **Sim** |
| `CORS_ORIGINS` | Origens permitidas (vírgula-separadas) | `http://localhost:3000` | Não |
| `LOG_LEVEL` | Nível de log: debug, info, warn, error | `info` | Não |

---

## CI/CD

GitHub Actions com **6 jobs** (trigger: push em `main`/`feature/*`, PRs para `main`):

| Job | Descrição |
|---|---|
| **lint** | golangci-lint no código Go |
| **build-server** | Compilação do backend |
| **build-agent** | Cross-compilação Windows amd64 |
| **test** | Testes Go com PostgreSQL real, race detection, cobertura → Codecov |
| **frontend** | ESLint + build Vite/TypeScript (Node.js 20) |
| **docker** | Build da imagem Docker com GHA cache |

---

## Makefile

```bash
make help           # Listar targets
make build-server   # Compilar API
make build-agent    # Compilar Agent (Windows amd64)
make run            # Rodar API localmente
make test           # Testes com race detection + coverage
make lint           # golangci-lint
make docker-up      # docker compose up -d --build
make docker-down    # docker compose down
make docker-logs    # docker compose logs -f
make create-user USERNAME=admin PASSWORD=secret  # Criar usuário
make tidy           # go mod tidy em todos os módulos
```

---

## Documentação Completa

A documentação ITIL v4 completa está em `docs/`:

| Seção | Conteúdo |
|---|---|
| [01-estrategia-de-servico](docs/01-estrategia-de-servico/) | Visão geral, catálogo, demanda, análise financeira |
| [02-desenho-de-servico](docs/02-desenho-de-servico/) | Arquitetura, segurança, SLAs, disponibilidade, capacidade |
| [03-transicao-de-servico](docs/03-transicao-de-servico/) | Mudanças, releases, configuração, conhecimento, testes |
| [04-operacao-de-servico](docs/04-operacao-de-servico/) | Runbooks, incidentes, problemas, eventos |
| [05-melhoria-continua](docs/05-melhoria-continua/) | Plano de melhoria, métricas e KPIs |
| [06-anexos](docs/06-anexos/) | Glossário, matriz RACI, diagramas |
| [07-guia-https-ssl](docs/07-guia-https-ssl/) | Deploy com HTTPS: Nginx, Caddy, certificados |

---

## Licença

Este projeto é licenciado sob a [MIT License](LICENSE).
