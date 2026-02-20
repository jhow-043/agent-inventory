# Backend — API REST

## Estrutura do Projeto

```
server/
├── cmd/api/main.go            # Entry point + CLI create-user
├── internal/
│   ├── config/config.go       # Variáveis de ambiente
│   ├── database/database.go   # Conexão PostgreSQL + migrações
│   ├── dto/                   # Request/Response structs
│   ├── handler/               # Handlers HTTP (Gin)
│   ├── middleware/            # Middlewares (auth, cors, rate limit, etc.)
│   ├── migrations/            # SQL migrations (embedded)
│   ├── repository/            # Queries SQL (sqlx)
│   │   ├── inventory.go       # Upsert transacional de inventário
│   │   ├── hardware_diff.go   # Comparação granular de hardware
│   │   ├── cleanup.go         # Purge, inativos, vacuum
│   │   └── ...
│   ├── router/router.go       # Definição de todas as rotas
│   └── service/               # Lógica de negócio
└── Dockerfile                 # Multi-stage build
```

## Inicialização

O `main.go` faz na ordem:

1. Checa se é sub-comando `create-user` (CLI para criar usuários)
2. Carrega config (variáveis de ambiente)
3. Configura logger JSON com `slog`
4. Conecta ao PostgreSQL (pool: 25 open, 5 idle, 5min lifetime)
5. Roda migrações automaticamente (embedded SQL)
6. Cria repositórios → services → handlers
7. Configura rotas
8. Inicia cleanup service (background: purge de logs, marcação de inativos)
9. Starta HTTP server com timeouts (read: 15s, write: 30s, idle: 60s)
10. Graceful shutdown em SIGINT/SIGTERM (cleanup service + HTTP server, 10s timeout)

## Configuração

| Variável | Obrigatória | Default | Descrição |
|----------|-------------|---------|-----------|
| `DATABASE_URL` | Não | `postgres://inventory:changeme@localhost:5432/inventory?sslmode=disable` | Connection string PostgreSQL |
| `SERVER_PORT` | Não | `8080` | Porta HTTP |
| `LOG_LEVEL` | Não | `info` | Nível de log: `debug`, `info`, `warn`, `error` |
| `JWT_SECRET` | **Sim** | — | Chave para assinar JWT (min 32 chars recomendado) |
| `ENROLLMENT_KEY` | **Sim** | — | Chave que os agents usam para se registrar |
| `CORS_ORIGINS` | Não | `http://localhost:3000` | Origens permitidas, separadas por vírgula |
| `RETENTION_DAYS` | Não | `90` | Dias para reter logs (audit, activity, hardware_history) |
| `INACTIVE_DAYS` | Não | `30` | Dias sem comunicação para marcar device como inativo |
| `CLEANUP_INTERVAL` | Não | `24h` | Intervalo entre execuções do cleanup automático |

Se `JWT_SECRET` ou `ENROLLMENT_KEY` estiverem vazias, o servidor recusa iniciar (`os.Exit(1)`).

## Rotas

### Cadeia de Middlewares Global

Toda requisição passa por estes middlewares na ordem:

```
Recovery → RequestID → Logging → SecurityHeaders → CORS → [Handler]
```

### Mapa Completo de Rotas

#### Públicas (sem autenticação)

| Método | Path | Middleware Extra | Handler | Descrição |
|--------|------|-----------------|---------|-----------|
| GET/HEAD | `/healthz` | — | `Healthz` | Retorna `{"status":"ok"}` sempre |
| GET | `/readyz` | — | `Readyz` | Faz ping no DB. Se OK: `{"status":"ready","database":"ok"}`. Se falhar: 503 |

#### Agent

| Método | Path | Middleware Extra | Handler | Descrição |
|--------|------|-----------------|---------|-----------|
| POST | `/api/v1/enroll` | RateLimit(10/min) | `Enroll` | Agent se registra, recebe token |
| POST | `/api/v1/inventory` | DeviceAuth | `SubmitInventory` | Agent envia inventário completo |

#### Usuário Autenticado (JWT)

| Método | Path | Handler | Descrição |
|--------|------|---------|-----------|
| GET | `/api/v1/auth/me` | `Me` | Retorna `{id, username, role}` do usuário logado |
| POST | `/api/v1/auth/logout` | `Logout` | Limpa cookie de sessão |
| GET | `/api/v1/dashboard/stats` | `GetStats` | Estatísticas: total, online, offline, inactive |
| GET | `/api/v1/devices` | `ListDevices` | Lista devices com filtros/sort/paginação |
| GET | `/api/v1/devices/export` | `ExportCSV` | Exporta devices em CSV (sem paginação) |
| GET | `/api/v1/devices/:id` | `GetDevice` | Device completo com hardware, discos, rede, software |
| GET | `/api/v1/devices/:id/hardware-history` | `GetHardwareHistory` | Histórico de mudanças de hardware |
| GET | `/api/v1/departments` | `ListDepartments` | Lista todos os departamentos |
| GET | `/api/v1/users` | `ListUsers` | Lista todos os usuários (sem password_hash) |

#### Admin Only (JWT + role=admin)

| Método | Path | Handler | Descrição |
|--------|------|---------|-----------|
| PATCH | `/api/v1/devices/:id/status` | `UpdateStatus` | Muda status: active/inactive |
| PATCH | `/api/v1/devices/:id/department` | `UpdateDepartment` | Atribui department (ou null) |
| POST | `/api/v1/departments` | `CreateDepartment` | Cria departamento |
| PUT | `/api/v1/departments/:id` | `UpdateDepartment` | Atualiza departamento |
| DELETE | `/api/v1/departments/:id` | `DeleteDepartment` | Deleta departamento |
| POST | `/api/v1/users` | `CreateUser` | Cria usuário (default: viewer) |
| DELETE | `/api/v1/users/:id` | `DeleteUser` | Deleta usuário (não pode deletar a si mesmo) |
| GET | `/api/v1/audit-logs` | `ListAuditLogs` | Logs de auditoria (filtráveis) |
| GET | `/api/v1/audit-logs/:type/:id` | `GetResourceAuditLogs` | Logs de um recurso específico |

## Middlewares

### RequestID

Gera UUID para cada request e seta no header `X-Request-Id`. Usado para rastreabilidade nos logs.

### Logging

Log estruturado em JSON com: `request_id`, `method`, `path`, `status`, `duration_ms`, `ip`, `device_id` (se existir).

Nível de log: Info para 2xx/3xx, Warn para 4xx, Error para 5xx.

### Security Headers

Headers de segurança em toda resposta. O `connect-src` do CSP é construído dinamicamente a partir do `CORS_ORIGINS`:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'
```

### CORS

Whitelist de origens (`CORS_ORIGINS`). Para cada request:
- Verifica se `Origin` está na whitelist
- Se sim, seta `Access-Control-Allow-Origin`, `Allow-Methods`, `Allow-Headers`, `Allow-Credentials: true`
- Requests `OPTIONS` retornam 204 direto (preflight)
- `Max-Age: 86400` (24h cache do preflight)

### Rate Limiting

Implementação própria in-memory com sliding window por IP:

```
Enrollment: 10 requests/minuto por IP
Login:      5 requests/minuto por IP
```

- Usa `map[string]*visitor` com mutex
- Goroutine de cleanup a cada 1 minuto remove entradas expiradas
- Retorna 429 se exceder o limite

### DeviceAuth (autenticação de agents)

```
Request: Authorization: Bearer <token-raw>
    │
    ▼
SHA-256(token-raw) → hash
    │
    ▼
SELECT * FROM device_tokens WHERE token_hash = hash
    │
    ▼
Seta device_id no contexto da request
```

- O token raw nunca é armazenado — apenas o hash SHA-256
- Se token inválido ou não encontrado: 401

### JWTAuth (autenticação de usuários)

```
Request: Cookie: session=<JWT>
    │
    ▼
jwt.Parse(JWT, secretKey, HS256)
    │
    ▼
Extrai claims: sub (user_id), username, role
    │
    ▼
Seta user_id, username, user_role no contexto
```

- Se cookie ausente ou JWT inválido/expirado: 401
- Role padrão: `viewer` (se campo ausente no JWT)

### RequireRole (RBAC)

```go
// Uso: middleware.RequireRole("admin")
// Lê user_role do contexto → compara com roles permitidos
// Se não autorizado: 403 Forbidden
```

### Audit Logger

Registra ações admin de forma assíncrona (goroutine separada):

- **Ações logadas:** login, logout, create/update/delete de departamentos, create/delete de usuários, mudança de status/departamento de devices
- **Dados salvos:** user_id, username, action, resource_type, resource_id, details (JSON), ip, user_agent, timestamp
- **Filtros disponíveis:** user_id, action, resource_type, resource_id, limit/offset (max 100)

## Handlers — Lógica de Negócio

### Enrollment

1. Lê header `X-Enrollment-Key`
2. Compara com `ENROLLMENT_KEY` usando `subtle.ConstantTimeCompare()` (previne timing attack)
3. Recebe `{hostname, serial_number}` no body
4. Busca device pelo `serial_number`:
   - Se existe: atualiza hostname e last_seen
   - Se não existe: cria novo device
5. Deleta tokens antigos do device
6. Gera UUID como token, salva SHA-256(token) na tabela `device_tokens`
7. Retorna `{device_id, token}` (201 Created)

### Login

1. Recebe `{username, password}`
2. Busca user pelo username no banco
3. Compara password com bcrypt hash
4. Gera JWT HS256 com claims: `sub`, `username`, `role`, `iat`, `exp` (24h)
5. Seta cookie `session` (httpOnly, 86400s, path `/`)
6. Loga evento de auditoria
7. Retorna `{message: "login successful"}`

### Processamento de Inventário

Este é o handler mais complexo. Ocorre numa **transação única**:

```
1. Upsert Device
   INSERT ... ON CONFLICT (id) DO UPDATE
   Atualiza: hostname, OS, agent_version, license_status, last_seen, logged_user

2. Detecção Granular de Mudanças de Hardware
   SELECT hardware atual + discos + NICs do banco
   Compara campo a campo (CPU, RAM, placa-mãe, BIOS)
   Compara discos por serial (fallback: model+size+type)
   Compara NICs por MAC address (fallback: nome)
   Cada mudança → INSERT individual em hardware_history com component/field/old/new

3. Upsert Hardware
   INSERT ... ON CONFLICT (device_id) DO UPDATE
   Campos: cpu_model, cpu_cores, cpu_threads, ram_total, mobo_manufacturer,
           mobo_model, mobo_serial, bios_vendor, bios_version

4. Replace Discos (DELETE + INSERT)
   Campos: model, size, type (HDD/SSD), interface_type, serial_number,
           partitions (JSONB com letter, size, free_space, file_system)

5. Replace Network Interfaces (DELETE + INSERT)
   Campos: name, mac_address, ipv4, ipv6, speed, status

6. Replace Software (DELETE + INSERT em chunks de 200)
   Campos: name, version, vendor, install_date
   Chunk de 200 porque PostgreSQL tem limite de parâmetros por query

7. Replace Remote Tools (DELETE + INSERT)
   Campos: tool_name, tool_id, version, is_installed
```

Se qualquer passo falhar, a transação inteira faz rollback — dados ficam consistentes.

### Lista de Devices

Suporta filtros, sorting e paginação via query params:

| Parâmetro | Tipo | Descrição |
|-----------|------|-----------|
| `page` | int | Página atual (default: 1) |
| `limit` | int | Items por página (default: 50) |
| `hostname` | string | Filtro ILIKE (case-insensitive) |
| `os` | string | Filtro por nome do OS |
| `status` | string | `online`, `offline`, `inactive` |
| `department_id` | UUID | Filtro por departamento |
| `sort` | string | Campo de ordenação |
| `order` | string | `asc` ou `desc` |

**Lógica de status online/offline:**

```sql
-- Online:   status = 'active' AND last_seen > NOW() - INTERVAL '1 hour'
-- Offline:  status = 'active' AND last_seen <= NOW() - INTERVAL '1 hour'
-- Inactive: status = 'inactive'
```

Um device é "online" se reportou inventário na última hora.

### Export CSV

Mesmos filtros da listagem, mas sem paginação. Gera CSV streaming com colunas:

```
Hostname, Serial Number, OS, OS Version, OS Build, Architecture,
Logged In User, Agent Version, License Status, Status, Department,
Last Seen, Created At
```

### Dashboard Stats

Retorna 4 contadores:
- **Total:** devices ativos
- **Online:** ativos que reportaram na última hora
- **Offline:** ativos que não reportaram na última hora
- **Inactive:** devices desativados

### Detalhes de Device

Chamada única retorna o device completo com todos os dados relacionados: hardware, discos com partições, interfaces de rede, software instalado, ferramentas de acesso remoto.

### Histórico de Hardware

Retorna histórico granular de mudanças de hardware, filtrado por componente (cpu, ram, motherboard, bios, disk, network). Cada registro inclui: component, change_type (added/removed/changed), field, old_value, new_value e snapshot JSONB do estado anterior. Suporta paginação (limit/offset).

## CLI — Criar Usuário

```bash
server create-user --username admin --password senha_segura --role admin
```

| Flag | Obrigatória | Default | Descrição |
|------|-------------|---------|-----------|
| `--username` | Sim | — | Nome do usuário |
| `--password` | Sim | — | Senha (min 8 caracteres) |
| `--role` | Não | `admin` | `admin` ou `viewer` |

Note que via CLI o role padrão é `admin`, mas via API (POST /users) o padrão é `viewer`.

## Conexão com o Banco

```go
db.SetMaxOpenConns(25)      // Máximo de conexões simultâneas
db.SetMaxIdleConns(5)       // Conexões ociosas mantidas no pool
db.SetConnMaxLifetime(5 * time.Minute)  // Recicla conexões a cada 5 min
```

Driver: `pgx` (implementação PostgreSQL nativa em Go, sem usar `libpq`).

Migrações são executadas automaticamente no startup usando `golang-migrate` com SQL files embedados no binário via `embed.FS`.
