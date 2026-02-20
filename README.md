# Inventory System

**Versão:** 1.1.0 | **Atualizado:** 20/02/2026

Sistema de inventário automatizado de hardware e software para estações Windows.
Um agente Windows coleta informações via WMI/Registry e envia para uma API centralizada, que alimenta um dashboard web com tema dark/light e gráficos analíticos.

---

## Sumário

1. [Arquitetura](#arquitetura)
2. [Stack Tecnológico](#stack-tecnológico)
3. [Estrutura do Projeto](#estrutura-do-projeto)
4. [Quick Start](#quick-start)
5. [API Endpoints](#api-endpoints)
6. [Dados Coletados pelo Agent](#dados-coletados-pelo-agent)
7. [Database Schema](#database-schema)
8. [Variáveis de Ambiente](#variáveis-de-ambiente)
9. [Makefile](#makefile)

---

## Arquitetura

```
┌──────────────────────────────────────────────────────────────────────────┐
│                            REDE INTERNA                                  │
│                                                                          │
│  ┌──────────────┐                                                        │
│  │  Estação #1  │──┐                                                     │
│  │  Agent.exe   │  │     HTTP/JSON                ┌──────────────────┐   │
│  └──────────────┘  │  ┌─────────────────────┐     │                  │   │
│  ┌──────────────┐  ├─►│    API Server       │     │  PostgreSQL 16   │   │
│  │  Estação #2  │──┤  │    Go 1.24 + Gin    │────►│                  │   │
│  │  Agent.exe   │  │  │    Port <sua porta> │     │  12 tabelas      │   │
│  └──────────────┘  │  └─────────┬───────────┘     │  10 migrations   │   │
│  ┌──────────────┐  │            │                 │                  │   │
│  │  Estação #N  │──┘            │                 └──────────────────┘   │
│  │  Agent.exe   │               │                                        │
│  └──────────────┘      ┌────────┴───────────┐                            │
│                        │   Frontend SPA     │                            │
│  ┌──────────────┐      │   React 19 + Vite 7│                            │
│  │   Browser    │─────►│   Tailwind 4       │                            │
│  │   (Admin)    │      │   Recharts 3.7     │                            │
│  └──────────────┘      └────────────────────┘                            │
└──────────────────────────────────────────────────────────────────────────┘
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
│   ├── cmd/api/main.go                # Entry point + CLI create-user
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
│   │   │   ├── hardware_diff.go       # Diff detalhado de alterações de hardware
│   │   │   ├── department.go          # CRUD departamentos
│   │   │   ├── dashboard.go           # COUNT queries
│   │   │   ├── user.go                # CRUD usuários
│   │   │   ├── token.go               # SHA-256 token lookup
│   │   │   ├── audit.go               # Audit logs
│   │   │   ├── device_activity.go     # Log de atividades do device
│   │   │   └── cleanup.go             # Purga de dados antigos
│   │   ├── service/                   # Lógica de negócio
│   │   │   ├── auth.go                # Enrollment, login (bcrypt + JWT), users
│   │   │   ├── device.go              # Orquestração queries
│   │   │   ├── department.go          # Regras departamentos
│   │   │   ├── dashboard.go           # Cálculo estatísticas
│   │   │   ├── inventory.go           # Processamento inventário
│   │   │   └── cleanup.go             # Cleanup service (background)
│   │   └── router/router.go          # Rotas + grupos de auth
│   └── migrations/                    # SQL (embedded via go:embed)
│       ├── 001_init.up.sql            # 7 tabelas base + índices
│       ├── 002_remote_tools.up.sql    # remote_tools
│       ├── 003_disk_free_space.up.sql # drive_letter, free_space_bytes
│       ├── 004_lifecycle.up.sql       # departments, status, hardware_history
│       ├── 005_add_user_roles.up.sql  # RBAC (admin/viewer)
│       ├── 006_add_audit_logs.up.sql  # audit_logs + 5 índices
│       ├── 007_device_activity_log    # Log de atividades do device
│       ├── 008_hardware_history_det.. # Detalhes no histórico de hardware
│       ├── 009_cleanup_orphan_hist..  # Limpeza de registros órfãos
│       └── 010_add_user_name.up.sql   # Campo nome no usuário
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
│       │   └── remote.go              # TeamViewer, AnyDesk, RustDesk
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
│   ├── vite.config.ts                 # Proxy /api → API server
│   └── src/
│       ├── api/                       # HTTP client + endpoints
│       │   ├── client.ts              # Fetch wrapper, auto-redirect 401
│       │   ├── auth.ts, devices.ts, departments.ts, dashboard.ts
│       │   ├── users.ts, audit.ts
│       ├── components/
│       │   ├── Layout.tsx             # Sidebar colapsável + tema toggle
│       │   ├── GlobalSearch.tsx        # Busca global
│       │   ├── ProtectedRoute.tsx     # Auth guard
│       │   ├── AdminRoute.tsx         # Guard role admin
│       │   ├── ErrorBoundary.tsx      # Captura erros React
│       │   └── ui/                    # Badge, Button, Card, Input, Modal, PageHeader, Select
│       ├── hooks/                     # useAuth, useTheme, useDebounce, useSidebar, useToast
│       ├── pages/
│       │   ├── Login.tsx              # Tela de login
│       │   ├── Dashboard.tsx          # KPIs + gráficos Recharts (pizza + barras)
│       │   ├── DeviceList.tsx         # Tabela + filtros + paginação + CSV export
│       │   ├── DeviceDetail.tsx       # Detalhes completos + hardware history
│       │   ├── Departments.tsx        # CRUD + contagem devices por departamento
│       │   ├── DepartmentDetail.tsx   # Devices vinculados a um departamento
│       │   ├── AuditLogs.tsx          # Visualização de logs de auditoria
│       │   └── Settings.tsx           # Tema dark/light + gerenciamento de usuários
│       └── types/index.ts            # Interfaces TypeScript
│
└── docs/                              # Documentação técnica
```

## Quick Start

### 1. Pré-requisitos

- Docker + Docker Compose
- Node.js 20+ (para frontend)
- Go 1.24+ (para desenvolvimento)

### 2. Configuração

```bash
# Copiar e preencher variáveis de ambiente
cp .env.example .env
```

Edite o `.env` com suas configurações. As variáveis obrigatórias são `JWT_SECRET` (mínimo 32 caracteres) e `ENROLLMENT_KEY`. Defina `SERVER_PORT` com a porta desejada para a API (sugestão: `8081`).

Para rodar em um IP específico da rede, edite o `docker-compose.yml` e o `vite.config.ts` com o IP da sua máquina. Veja a seção [Variáveis de Ambiente](#variáveis-de-ambiente).

### 3. Subir o Backend

```bash
# Subir PostgreSQL + API (migrations automáticas)
docker compose up -d --build

# Criar usuário admin
docker compose exec api ./server create-user --username admin --password SenhaSegura123 --role admin

# Criar usuário somente leitura (opcional)
docker compose exec api ./server create-user --username viewer --password Viewer123 --role viewer
```

### 4. Subir o Frontend

```bash
cd frontend
npm install
npm run dev
```

O Vite abrirá o dashboard no endereço configurado em `vite.config.ts` (host e porta).

### 5. Configurar o Agent (Windows)

```powershell
cd agent
copy config.example.json config.json
# Editar config.json com a URL da API e a enrollment key

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
  "server_url": "http://<IP_DO_SERVIDOR>:<PORTA_DA_API>",
  "enrollment_key": "mesma-chave-do-env",
  "interval_hours": 1
}
```

> **Dica:** Use o IP da máquina onde a API está rodando e a porta definida em `SERVER_PORT` no `.env` (sugestão: `8081`).

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
| **Acesso Remoto** | TeamViewer ID, AnyDesk ID, RustDesk ID | Registry + config files + CLI fallback |

---

## Database Schema

**12 tabelas** em 10 migrations (auto-executadas no startup via `go:embed`).

| Tabela | Relação | Descrição |
|---|---|---|
| `users` | — | Contas dashboard (username, nome, bcrypt hash, role admin/viewer) |
| `devices` | — | Dispositivos (hostname, serial, OS, status, department_id FK) |
| `device_tokens` | 1:1 device | Tokens SHA-256 dos agents |
| `hardware` | 1:1 device | CPU, RAM, motherboard, BIOS |
| `disks` | 1:N device | Discos + partições (modelo, tipo, letra, espaço livre) |
| `network_interfaces` | 1:N device | Interfaces de rede (MAC, IPs, velocidade) |
| `installed_software` | 1:N device | Software instalado |
| `remote_tools` | 1:N device | TeamViewer, AnyDesk, RustDesk |
| `departments` | — | Departamentos organizacionais |
| `hardware_history` | 1:N device | Histórico detalhado de mudanças (componente, campo, valor anterior/novo) |
| `device_activity_log` | 1:N device | Log de atividades do dispositivo |
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
| 007 | `007_device_activity_log.up.sql` | Tabela device_activity_log + índices |
| 008 | `008_hardware_history_details.up.sql` | Colunas de detalhe no hardware_history |
| 009 | `009_cleanup_orphan_history.up.sql` | Limpeza de registros órfãos + NOT NULL |
| 010 | `010_add_user_name.up.sql` | Campo nome no usuário |

---

## Variáveis de Ambiente

Copie `.env.example` para `.env` e preencha os valores.

| Variável | Descrição | Default | Obrigatória |
|---|---|---|---|
| `POSTGRES_PASSWORD` | Senha do PostgreSQL (Docker) | `changeme` | Sim |
| `SERVER_PORT` | Porta da API exposta no host (sugestão: `8081`) | `8080` | Não |
| `JWT_SECRET` | Chave HS256 para JWT (≥32 chars) | — | **Sim** |
| `ENROLLMENT_KEY` | Chave de enrollment dos agents | — | **Sim** |
| `CORS_ORIGINS` | Origens permitidas (vírgula-separadas, ex: `http://<SEU_IP>:5173`) | — | Não |
| `LOG_LEVEL` | Nível de log: debug, info, warn, error | `info` | Não |
| `RETENTION_DAYS` | Dias para reter logs (audit, activity, hardware history) | `90` | Não |
| `INACTIVE_DAYS` | Dias sem comunicação para marcar device como inativo | `30` | Não |
| `CLEANUP_INTERVAL` | Intervalo do cleanup service (ex: `24h`, `12h`) | `24h` | Não |

### Configurando IP e Portas

Para rodar o sistema na rede interna, configure:

1. **`.env`** — `SERVER_PORT` com a porta desejada e `CORS_ORIGINS` com `http://<SEU_IP>:<PORTA_FRONTEND>`
2. **`docker-compose.yml`** — na seção `ports` da API, use `<SEU_IP>:<PORTA>:8080` para restringir o bind
3. **`frontend/vite.config.ts`** — altere `host` para o IP e `proxy` para apontar à API

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

## Documentação

A documentação técnica está em `docs/`:

| Arquivo | Conteúdo |
|---|---|
| [01-visao-geral.md](docs/01-visao-geral.md) | Visão geral do sistema |
| [02-backend-api.md](docs/02-backend-api.md) | API REST — rotas, auth, middleware |
| [03-agent.md](docs/03-agent.md) | Agent Windows — instalação e coleta |
| [04-frontend.md](docs/04-frontend.md) | Frontend — componentes e páginas |
| [05-banco-de-dados.md](docs/05-banco-de-dados.md) | Schema e migrations |
| [06-instalacao.md](docs/06-instalacao.md) | Guia de instalação e deploy |

---

## Licença

Este projeto é licenciado sob a [MIT License](LICENSE).
