# Server (API Backend)

**Versão:** 1.2.0

API REST construída com **Go 1.24** e **Gin**, responsável por receber dados de inventário dos agents, gerenciar dispositivos, departamentos, usuários e prover dados para o dashboard.

## Stack

- **Go 1.24** + **Gin** (HTTP framework)
- **PostgreSQL 17** (banco de dados via Docker)
- **sqlx** + **pgx v5** (driver PostgreSQL)
- **golang-migrate** (migrations embarcadas via `go:embed`)
- **JWT HS256** (autenticação em cookie httpOnly)
- **bcrypt** (hash de senhas)
- **slog** (logging estruturado JSON)

## Arquitetura

```
server/
├── Dockerfile                    # Multi-stage: golang:1.24-alpine → alpine:3.21
├── go.mod
├── cmd/api/main.go               # Entry point + sub-comando create-user
├── internal/
│   ├── config/config.go          # Carregamento de variáveis de ambiente + validação
│   ├── database/database.go      # Conexão + pool + migrations automáticas
│   ├── handler/                  # Controladores HTTP
│   │   ├── auth.go               # Enroll, Login, Logout, Me
│   │   ├── device.go             # CRUD devices, bulk ops, export CSV, history
│   │   ├── department.go         # CRUD departamentos
│   │   ├── dashboard.go          # GetStats
│   │   ├── user.go               # CRUD usuários
│   │   ├── audit.go              # Audit logs query
│   │   ├── health.go             # Healthz, Readyz
│   │   └── inventory.go          # SubmitInventory
│   ├── middleware/               # Pipeline HTTP (9 middlewares)
│   │   ├── auth.go               # DeviceAuth (Bearer token) + JWTAuth (cookie)
│   │   ├── rbac.go               # RequireRole("admin")
│   │   ├── ratelimit.go          # Rate limiting por IP (configurável por rota)
│   │   ├── bodylimit.go          # Limite de body 10MB (anti-OOM)
│   │   ├── security.go           # Security headers (CSP, X-Frame-Options, etc.)
│   │   ├── audit.go              # Audit logger (ações admin)
│   │   ├── cors.go               # CORS whitelist (baseado em CORS_ORIGINS)
│   │   ├── logging.go            # Structured slog JSON com request_id
│   │   └── requestid.go          # UUID X-Request-Id em cada request
│   ├── repository/               # Camada de dados (SQL queries)
│   │   ├── device.go, inventory.go, hardware_diff.go
│   │   ├── department.go, dashboard.go, user.go
│   │   ├── token.go, audit.go, device_activity.go
│   │   └── cleanup.go
│   ├── service/                  # Lógica de negócio
│   │   ├── auth.go, device.go, department.go
│   │   ├── dashboard.go, inventory.go
│   │   └── cleanup.go            # Background cleanup (goroutine)
│   └── router/router.go         # Definição de rotas + grupos de auth
└── migrations/                   # SQL embarcado (10 migrations up/down)
    ├── 001_init.up.sql           # 7 tabelas base
    ├── 002_remote_tools.up.sql
    ├── 003_disk_free_space.up.sql
    ├── 004_lifecycle.up.sql
    ├── 005_add_user_roles.up.sql
    ├── 006_add_audit_logs.up.sql
    ├── 007_device_activity_log.up.sql
    ├── 008_hardware_history_details.up.sql
    ├── 009_cleanup_orphan_history.up.sql
    ├── 010_add_user_name.up.sql
    └── embed.go                  # go:embed *.sql
```

## Pipeline de Middlewares

Ordem de execução (global, aplicado a todas as rotas):

```
Request → Recovery → RequestID → Logging → SecurityHeaders → CORS → DefaultMaxBodySize(10MB) → Router
```

| Middleware | Arquivo | Descrição |
|---|---|---|
| **Recovery** | Gin built-in | Captura panics e retorna 500 |
| **RequestID** | `requestid.go` | Gera UUID único por request (`X-Request-Id`) |
| **Logging** | `logging.go` | Log estruturado JSON via `slog` (method, path, status, latency) |
| **Security Headers** | `security.go` | CSP (dinâmico via `CORS_ORIGINS`), `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy` |
| **CORS** | `cors.go` | Whitelist baseada em `CORS_ORIGINS`, credentials habilitadas |
| **Body Limit** | `bodylimit.go` | `http.MaxBytesReader` de 10 MB (previne OOM/DoS) |
| **Auth (JWT)** | `auth.go` | Valida token JWT do cookie `token` (grupos protegidos) |
| **Auth (Device)** | `auth.go` | Valida Bearer token SHA-256 do agent (rota `/inventory`) |
| **RBAC** | `rbac.go` | `RequireRole("admin")` com `AbortWithStatusJSON` (grupo admin) |
| **Rate Limit** | `ratelimit.go` | Limitação por IP: `/auth/login` (5/min), `/enroll` (10/min) |
| **Audit** | `audit.go` | Log de ações admin em `audit_logs` (insert DB) |

## Segurança

- **Cookie JWT**: `HttpOnly`, `SameSite=Lax`, `Secure` dinâmico (ativo em HTTPS automaticamente)
- **JWT Secret**: mínimo de 32 caracteres — servidor recusa iniciar se menor
- **Senhas**: hash bcrypt com salt automático, input limitado a 200 chars (previne bcrypt DoS)
- **Body size**: limitado a 10 MB por padrão via `MaxBytesReader`
- **Bulk operations**: capadas em 100 itens por requisição
- **Paginação**: capada em 200 itens por página
- **Rate limiting**: `/auth/login` (5 req/min), `/enroll` (10 req/min)
- **Device token**: SHA-256 com proteção contra token vazio
- **SELECT explícito**: queries de usuários nunca retornam `password_hash`
- **Security headers**: CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy

## Configuração

Toda a configuração é realizada via **variáveis de ambiente** (sem flags CLI).

| Variável | Padrão | Obrigatória | Descrição |
|---|---|---|---|
| `DATABASE_URL` | — | **Sim** | Connection string PostgreSQL (ex: `postgres://user:pass@host:5432/db?sslmode=disable`) |
| `SERVER_PORT` | `8081` | Não | Porta HTTP do servidor |
| `JWT_SECRET` | — | **Sim** | Secret HS256 para JWT (mínimo 32 caracteres) |
| `ENROLLMENT_KEY` | — | **Sim** | Chave de enrollment para agents |
| `CORS_ORIGINS` | `http://localhost:3000` | Não | Origens CORS permitidas (vírgula-separadas) |
| `LOG_LEVEL` | `info` | Não | Nível de log: `debug`, `info`, `warn`, `error` |
| `RETENTION_DAYS` | `90` | Não | Dias para reter audit_logs, activity_log, hardware_history |
| `INACTIVE_DAYS` | `30` | Não | Dias sem comunicação para marcar device como inativo |
| `CLEANUP_INTERVAL` | `24h` | Não | Intervalo do background cleanup (ex: `12h`, `24h`) |

> **Nota:** O `DATABASE_URL` é montado automaticamente pelo `docker-compose.yml` a partir das variáveis `POSTGRES_*`. Em deploy manual, defina diretamente.

## Executando

### Via Docker Compose (recomendado)

```bash
docker compose up -d --build
```

### Manualmente

```bash
cd server

export DATABASE_URL="postgres://inventory:senha@localhost:5432/inventory?sslmode=disable"
export JWT_SECRET="sua-chave-com-pelo-menos-32-caracteres-aqui"
export ENROLLMENT_KEY="chave-de-enrollment-dos-agents"
export SERVER_PORT="8081"

go run ./cmd/api
```

## CLI — Gerenciamento de Usuários

O binário do server aceita o sub-comando `create-user` para criar usuários diretamente no banco:

```bash
# Criar usuário admin
./server create-user --username admin --password SenhaSegura123 --role admin

# Criar usuário somente leitura
./server create-user --username viewer --password Viewer123 --role viewer

# Via Docker
docker compose exec api ./server create-user --username admin --password SenhaSegura123 --role admin
```

Roles disponíveis: `admin` (acesso total) e `viewer` (apenas leitura).

## Endpoints

### Saúde do Serviço (sem auth)

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/healthz` | Liveness probe |
| `HEAD` | `/healthz` | Liveness probe (HEAD) |
| `GET` | `/readyz` | Readiness probe (testa conexão DB) |

### Agent (auth própria)

| Método | Rota | Auth | Rate Limit | Descrição |
|---|---|---|---|---|
| `POST` | `/api/v1/enroll` | Header `X-Enrollment-Key` | 10 req/min | Registra dispositivo, retorna token |
| `POST` | `/api/v1/inventory` | Bearer Token (device) | — | Envia inventário completo |

### Autenticação e Leitura (JWT: qualquer role)

| Método | Rota | Rate Limit | Descrição |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | 5 req/min | Login (set cookie JWT httpOnly) |
| `GET` | `/api/v1/auth/me` | — | Dados do usuário logado (id, username, name, role) |
| `POST` | `/api/v1/auth/logout` | — | Logout (limpa cookie) |
| `GET` | `/api/v1/dashboard/stats` | — | Estatísticas: total, online, offline, inactive |
| `GET` | `/api/v1/devices` | — | Listar dispositivos (filtros, paginação, sort, max 200/pg) |
| `GET` | `/api/v1/devices/export` | — | Exportar dispositivos em CSV |
| `GET` | `/api/v1/devices/:id` | — | Detalhe completo do dispositivo |
| `GET` | `/api/v1/devices/:id/hardware-history` | — | Histórico de mudanças de hardware |
| `GET` | `/api/v1/devices/:id/activity` | — | Log de atividades do dispositivo |
| `GET` | `/api/v1/departments` | — | Listar departamentos |
| `GET` | `/api/v1/users` | — | Listar usuários |

### Admin Only (JWT: role admin)

| Método | Rota | Descrição |
|---|---|---|
| `PATCH` | `/api/v1/devices/:id/status` | Alterar status (active/inactive) |
| `PATCH` | `/api/v1/devices/:id/department` | Atribuir departamento ao device |
| `DELETE` | `/api/v1/devices/:id` | Excluir dispositivo |
| `PATCH` | `/api/v1/devices/bulk/status` | Alterar status em lote (max 100) |
| `PATCH` | `/api/v1/devices/bulk/department` | Atribuir departamento em lote (max 100) |
| `POST` | `/api/v1/devices/bulk/delete` | Excluir em lote (max 100) |
| `POST` | `/api/v1/departments` | Criar departamento |
| `PUT` | `/api/v1/departments/:id` | Renomear departamento |
| `DELETE` | `/api/v1/departments/:id` | Excluir departamento |
| `POST` | `/api/v1/users` | Criar usuário (com role) |
| `PUT` | `/api/v1/users/:id` | Atualizar usuário |
| `DELETE` | `/api/v1/users/:id` | Excluir usuário |
| `GET` | `/api/v1/audit-logs` | Logs de auditoria (paginados) |
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
| `limit` | int | `50` | Itens/página (máx: 200) |

## Migrations

As 10 migrations ficam em `server/migrations/` e são executadas automaticamente no startup via `go:embed`:

| # | Arquivo | O que faz |
|---|---|---|
| 001 | `001_init` | 7 tabelas base (devices, hardware, disks, network, software, tokens, users) + índices |
| 002 | `002_remote_tools` | Tabela `remote_tools` (TeamViewer, AnyDesk, RustDesk) |
| 003 | `003_disk_free_space` | Campos `drive_letter` e `free_space_bytes` em disks |
| 004 | `004_lifecycle` | Tabelas `departments` e `hardware_history`, campo `status` em devices |
| 005 | `005_add_user_roles` | RBAC: campo `role` em users + constraint (admin/viewer) |
| 006 | `006_add_audit_logs` | Tabela `audit_logs` + 5 índices |
| 007 | `007_device_activity_log` | Tabela `device_activity_log` + índices |
| 008 | `008_hardware_history_details` | Colunas de detalhe no hardware_history (component, field, old/new) |
| 009 | `009_cleanup_orphan_history` | Limpeza de registros órfãos + constraints NOT NULL |
| 010 | `010_add_user_name` | Campo `name` na tabela users |

Cada migration possui arquivo `.up.sql` (aplicar) e `.down.sql` (reverter).