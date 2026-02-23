# AGENT CONTEXT — IT Asset Inventory System

> **Last Updated:** 2026-02-16
> **Phase:** 1 — Windows Inventory (Implemented)
> **Service ID:** SVC-INV-001

This document is the authoritative context for any development agent working on this project.
It reflects the **actual state of the codebase**, not aspirational/planned features.
Discrepancies between ITIL documentation (`docs/`) and code are noted in [Known Gaps](#19-known-gaps-docs-vs-code).

---

## 1. LANGUAGE RULES

- All source code must be written entirely in **English**.
- All variable names, function names, structs, database tables, fields, and comments must be in **English**.
- All documentation inside the repository must be written in **English**.
- All chat explanations and reasoning responses must be written in **Portuguese (pt-BR)**.
- Do **NOT** mix Portuguese inside the code.

---

## 2. PROJECT IDENTITY

| Attribute | Value |
|---|---|
| **Service Name** | Sistema de Inventário de Ativos de TI |
| **Service ID** | SVC-INV-001 |
| **Owner** | Equipe de Infraestrutura de TI |
| **Current Phase** | Phase 1 — Windows Asset Inventory |
| **Target Scale** | 100–500 Windows workstations |
| **Criticality** | Medium |
| **Service Type** | Internal IT Service |
| **Repository** | https://github.com/jhow-043/agent-inventory |

### Problem Statement

Organizations with 100–500 Windows machines lack centralized visibility into hardware assets, installed software, Windows licensing status, and network configuration — leading to manual audits, compliance risk, and poor capacity planning.

### Solution

An automated inventory system composed of:
1. **Windows Agent** — collects hardware, software, network, license, and remote access tool data via WMI + Registry + CLI
2. **Central API** — receives, stores, and serves inventory data
3. **PostgreSQL Database** — normalized schema with full device snapshots
4. **Web Dashboard** — dark-themed SPA with sidebar navigation for asset visualization

---

## 3. ARCHITECTURE OVERVIEW

### 3.1 Architectural Style

**Modular monolith** — each component is an independent unit (Go module or React app), deployed via Docker Compose on a single server.

**Rationale (ADR-001):** Solo developer, 100–500 devices, on-premises deployment. Microservices would be premature optimization.

### 3.2 Diagram

```
                          INTERNAL NETWORK (HTTP)
    ┌──────────────────────────────────────────────────────────┐
    │                                                          │
    │  ┌───────────────┐      HTTP/JSON        ┌─────────────────────────────┐
    │  │ Windows Agent  │ ──────────────────►   │      Docker Host            │
    │  │ (Go binary)    │  Bearer: device-token │                             │
    │  │ Win Service    │                       │  ┌───────────────────────┐  │
    │  └───────────────┘                       │  │  API Container (Go)   │  │
    │                                          │  │  Gin + sqlx + JWT     │  │
    │  ┌───────────────┐                       │  │  :8081                │  │
    │  │ ...N agents    │                       │  └──────────┬──────────┘  │
    │  └───────────────┘                       │             │ TCP:5432    │
    │                                          │  ┌──────────▼──────────┐  │
    │  ┌───────────────┐      HTTP/JSON        │  │ PostgreSQL 16       │  │
    │  │ Browser        │ ──────────────────►   │  │ Volume persistente  │  │
    │  │ (Dashboard)    │  Cookie: session=JWT  │  └─────────────────────┘  │
    │  └───────────────┘                       └─────────────────────────────┘
    │         ▲
    │         │ http://localhost:5173 (dev) or Nginx (prod)
    │  ┌──────┴────────┐
    │  │ React SPA      │
    │  │ Vite + Tailwind│
    │  └───────────────┘
    └──────────────────────────────────────────────────────────┘
```

### 3.3 Communication Pattern

- **Agent → API only** (agent never accepts inbound connections)
- Agent initiates all communication: enrollment, then periodic inventory submission
- Dashboard communicates with API via REST, authenticated with httpOnly JWT cookie
- All payloads are JSON

---

## 4. TECHNOLOGY STACK

### 4.1 Exact Versions (from go.mod / package.json / Dockerfile)

| Component | Technology | Version |
|---|---|---|
| Go (workspace) | Go | 1.24.0 |
| Go (shared module) | Go | 1.22.0 |
| Gin | gin-gonic/gin | v1.11.0 |
| JWT | golang-jwt/jwt/v5 | v5.3.1 |
| Migrations | golang-migrate/migrate/v4 | v4.19.1 |
| UUID | google/uuid | v1.6.0 |
| pgx | jackc/pgx/v5 | v5.8.0 |
| sqlx | jmoiron/sqlx | v1.4.0 |
| bcrypt | golang.org/x/crypto | v0.48.0 |
| WMI (agent) | yusufpapurcu/wmi | v1.2.4 |
| Windows svc (agent) | golang.org/x/sys | v0.41.0 |
| React | react | ^19.1.1 |
| React DOM | react-dom | ^19.1.1 |
| React Router | react-router-dom | ^7.13.0 |
| TanStack Query | @tanstack/react-query | ^5.90.21 |
| Tailwind CSS | tailwindcss | ^4.1.18 |
| @tailwindcss/vite | @tailwindcss/vite | ^4.1.18 |
| Vite | vite | ^7.1.7 |
| TypeScript | typescript | ~5.8.3 |
| PostgreSQL (Docker) | postgres | 16-alpine |
| Dockerfile build | golang | 1.24-alpine |
| Dockerfile runtime | alpine | 3.20 |

### 4.2 Go Workspace

```
go.work
├── shared/    (go 1.22.0)   — domain models + DTOs
├── server/    (go 1.24.0)   — API, depends on shared via replace directive
└── agent/     (go 1.24.0)   — Windows agent, depends on shared via replace directive
```

---

## 5. PROJECT STRUCTURE

```
Inventario/
├── .env                               # Environment variables (gitignored)
├── .env.example                       # Template for .env
├── .gitignore
├── agent.md                           # THIS FILE — agent context (gitignored)
├── docker-compose.yml                 # PostgreSQL + API containers
├── go.work                            # Go workspace (shared, server, agent)
├── go.work.sum
├── Makefile                           # Build/run targets
├── README.md                          # Project documentation
│
├── agent/                             # ── WINDOWS AGENT (Go) ──
│   ├── go.mod / go.sum
│   ├── config.example.json            # Template for agent config
│   ├── config.json                    # Active config (gitignored)
│   ├── inventory-agent.exe            # Built binary (gitignored)
│   ├── data/device.token              # Persisted device token (gitignored)
│   ├── cmd/agent/
│   │   └── main.go                    # Entry point: service + CLI (run/collect/install/start/stop/uninstall/version)
│   └── internal/
│       ├── client/
│       │   └── client.go              # HTTP client: Enroll(), SubmitInventory(), SubmitWithRetry(), IsAuthError()
│       ├── collector/
│       │   ├── collector.go           # Orchestrator: Collect() → *dto.InventoryRequest
│       │   ├── system.go              # WMI: Win32_OperatingSystem, Win32_BIOS, Win32_ComputerSystem
│       │   ├── hardware.go            # WMI: Win32_Processor, Win32_PhysicalMemory, Win32_BaseBoard
│       │   ├── disk.go                # WMI: Win32_DiskDrive
│       │   ├── network.go             # WMI: Win32_NetworkAdapter + Go net.Interfaces()
│       │   ├── software.go            # Registry: Uninstall keys (HKLM+HKCU, 32+64bit)
│       │   ├── license.go             # WMI: SoftwareLicensingProduct
│       │   └── remote.go              # Registry+files+CLI: TeamViewer, AnyDesk, RustDesk
│       ├── config/
│       │   └── config.go              # JSON config loader with defaults
│       └── token/
│           └── store.go               # File-based token persistence (0600 perms)
│
├── server/                            # ── CENTRAL API (Go) ──
│   ├── go.mod / go.sum
│   ├── Dockerfile                     # Multi-stage: golang:1.24-alpine → alpine:3.20
│   ├── cmd/api/
│   │   └── main.go                    # Entry point + "create-user" CLI subcommand
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go              # Env vars loader (DATABASE_URL, JWT_SECRET, etc.)
│   │   ├── database/
│   │   │   └── database.go            # Connect() + RunMigrations() with embedded SQL
│   │   ├── handler/
│   │   │   ├── auth.go                # Enroll(), Login(), Logout()
│   │   │   ├── device.go              # ListDevices(), GetDevice()
│   │   │   ├── health.go              # Healthz(), Readyz()
│   │   │   └── inventory.go           # SubmitInventory()
│   │   ├── middleware/
│   │   │   ├── auth.go                # DeviceAuth (Bearer token + SHA256), JWTAuth (cookie), SHA256Hex()
│   │   │   ├── cors.go                # CORS with origin whitelist + credentials
│   │   │   ├── logging.go             # Structured slog request logging
│   │   │   └── requestid.go           # UUID X-Request-Id header
│   │   ├── repository/
│   │   │   ├── device.go              # List, GetByID, GetBySerialNumber, GetHardware, GetDisks, GetNetworkInterfaces, GetInstalledSoftware, GetRemoteTools
│   │   │   ├── inventory.go           # Save() — transactional upsert of full snapshot
│   │   │   ├── token.go               # GetByHash, Create, DeleteByDeviceID
│   │   │   └── user.go                # GetByUsername, Create
│   │   ├── router/
│   │   │   └── router.go              # Gin engine + all route definitions
│   │   └── service/
│   │       ├── auth.go                # Enroll (tx), Login (JWT), CreateUser (bcrypt)
│   │       ├── device.go              # ListDevices, GetDeviceDetail
│   │       └── inventory.go           # ProcessInventory
│   └── migrations/
│       ├── embed.go                   # //go:embed *.sql
│       ├── 001_init.up.sql            # 7 tables: users, devices, device_tokens, hardware, disks, network_interfaces, installed_software
│       ├── 001_init.down.sql
│       ├── 002_remote_tools.up.sql    # 1 table: remote_tools
│       └── 002_remote_tools.down.sql
│
├── shared/                            # ── SHARED MODULE ──
│   ├── go.mod / go.sum
│   ├── models/
│   │   └── models.go                  # DB entities: Device, DeviceToken, Hardware, Disk, NetworkInterface, InstalledSoftware, RemoteTool, User
│   └── dto/
│       ├── requests.go                # EnrollRequest, InventoryRequest, HardwareData, DiskData, NetworkData, SoftwareData, RemoteToolData, LoginRequest, CreateUserRequest
│       └── responses.go               # ErrorResponse, MessageResponse, HealthResponse, ReadyResponse, EnrollResponse, DeviceListResponse, DeviceDetailResponse
│
├── frontend/                          # ── WEB DASHBOARD (React) ──
│   ├── package.json / package-lock.json
│   ├── tsconfig.json / tsconfig.app.json / tsconfig.node.json
│   ├── eslint.config.js
│   ├── vite.config.ts                 # Proxy: /api → http://localhost:8081
│   ├── index.html
│   └── src/
│       ├── main.tsx                   # QueryClient + AuthProvider + App
│       ├── App.tsx                    # BrowserRouter + route definitions
│       ├── index.css                  # Tailwind v4 @theme + dark theme CSS vars + scrollbar
│       ├── vite-env.d.ts
│       ├── api/
│       │   ├── client.ts              # fetch wrapper: credentials:'include', auto-redirect 401
│       │   ├── auth.ts                # login(), logout()
│       │   └── devices.ts             # getDevices(), getDevice(id)
│       ├── components/
│       │   ├── Layout.tsx             # Sidebar (240px) + nav items + Outlet
│       │   └── ProtectedRoute.tsx     # Auth guard → redirect /login
│       ├── hooks/
│       │   └── useAuth.tsx            # AuthContext + localStorage flag
│       ├── pages/
│       │   ├── Login.tsx              # Username/password form
│       │   ├── Dashboard.tsx          # 3 stat cards: Total / Online / Offline
│       │   ├── DeviceList.tsx         # Table with hostname+OS filters, online badges
│       │   ├── DeviceDetail.tsx       # All sections + remote tools table + copy-to-clipboard
│       │   └── Settings.tsx           # Placeholder page
│       └── types/
│           └── index.ts               # TS interfaces matching Go DTOs
│
└── docs/                              # ── ITIL v4 DOCUMENTATION (30 docs, gitignored) ──
    ├── README.md
    ├── 01-estrategia-de-servico/      # Service strategy (4 docs)
    ├── 02-desenho-de-servico/         # Service design (7 docs)
    ├── 03-transicao-de-servico/       # Service transition (5 docs)
    ├── 04-operacao-de-servico/        # Service operation (5 docs)
    ├── 05-melhoria-continua/          # CSI (2 docs)
    └── 06-anexos/                     # Annexes: glossary, RACI matrix, 5 diagrams
```

---

## 6. DATABASE SCHEMA

**8 tables** across 2 migrations. PostgreSQL 16 with `uuid-ossp` extension.

### Migration 001 — Initial Schema (7 tables)

#### `users` — Dashboard user accounts
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| username | VARCHAR(100) | NOT NULL, UNIQUE |
| password_hash | VARCHAR(255) | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

#### `devices` — Managed Windows workstations
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| hostname | VARCHAR(255) | NOT NULL |
| serial_number | VARCHAR(255) | NOT NULL, **UNIQUE** |
| os_name | VARCHAR(100) | NOT NULL, DEFAULT '' |
| os_version | VARCHAR(100) | NOT NULL, DEFAULT '' |
| os_build | VARCHAR(50) | NOT NULL, DEFAULT '' |
| os_arch | VARCHAR(20) | NOT NULL, DEFAULT '' |
| last_boot_time | TIMESTAMPTZ | nullable |
| logged_in_user | VARCHAR(255) | NOT NULL, DEFAULT '' |
| agent_version | VARCHAR(50) | NOT NULL, DEFAULT '' |
| license_status | VARCHAR(100) | NOT NULL, DEFAULT '' |
| last_seen | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

#### `device_tokens` — Agent authentication tokens (SHA-256 hashed)
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| device_id | UUID | NOT NULL, UNIQUE, FK → devices(id) ON DELETE CASCADE |
| token_hash | VARCHAR(64) | NOT NULL, UNIQUE |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

#### `hardware` — CPU, RAM, motherboard, BIOS (1:1 with device)
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| device_id | UUID | NOT NULL, **UNIQUE**, FK → devices(id) ON DELETE CASCADE |
| cpu_model | VARCHAR(255) | NOT NULL, DEFAULT '' |
| cpu_cores | INTEGER | NOT NULL, DEFAULT 0 |
| cpu_threads | INTEGER | NOT NULL, DEFAULT 0 |
| ram_total_bytes | BIGINT | NOT NULL, DEFAULT 0 |
| motherboard_manufacturer | VARCHAR(255) | NOT NULL, DEFAULT '' |
| motherboard_product | VARCHAR(255) | NOT NULL, DEFAULT '' |
| motherboard_serial | VARCHAR(255) | NOT NULL, DEFAULT '' |
| bios_vendor | VARCHAR(255) | NOT NULL, DEFAULT '' |
| bios_version | VARCHAR(255) | NOT NULL, DEFAULT '' |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

#### `disks` — Physical disk drives (1:N with device)
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| device_id | UUID | NOT NULL, FK → devices(id) ON DELETE CASCADE |
| model | VARCHAR(255) | NOT NULL, DEFAULT '' |
| size_bytes | BIGINT | NOT NULL, DEFAULT 0 |
| media_type | VARCHAR(20) | NOT NULL, DEFAULT '' |
| serial_number | VARCHAR(255) | NOT NULL, DEFAULT '' |
| interface_type | VARCHAR(20) | NOT NULL, DEFAULT '' |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

#### `network_interfaces` — Network adapters (1:N with device)
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| device_id | UUID | NOT NULL, FK → devices(id) ON DELETE CASCADE |
| name | VARCHAR(255) | NOT NULL, DEFAULT '' |
| mac_address | VARCHAR(17) | NOT NULL, DEFAULT '' |
| ipv4_address | VARCHAR(15) | NOT NULL, DEFAULT '' |
| ipv6_address | VARCHAR(45) | NOT NULL, DEFAULT '' |
| speed_mbps | BIGINT | nullable |
| is_physical | BOOLEAN | NOT NULL, DEFAULT true |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

#### `installed_software` — Installed applications (1:N with device)
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| device_id | UUID | NOT NULL, FK → devices(id) ON DELETE CASCADE |
| name | VARCHAR(500) | NOT NULL |
| version | VARCHAR(100) | NOT NULL, DEFAULT '' |
| vendor | VARCHAR(255) | NOT NULL, DEFAULT '' |
| install_date | VARCHAR(20) | NOT NULL, DEFAULT '' |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

### Migration 002 — Remote Tools

#### `remote_tools` — Remote access tools (1:N with device)
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT uuid_generate_v4() |
| device_id | UUID | NOT NULL, FK → devices(id) ON DELETE CASCADE |
| tool_name | VARCHAR(100) | NOT NULL, DEFAULT '' |
| remote_id | VARCHAR(255) | NOT NULL, DEFAULT '' |
| version | VARCHAR(100) | NOT NULL, DEFAULT '' |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() |

### Indexes

| Index | Table | Column(s) |
|---|---|---|
| idx_devices_hostname | devices | hostname |
| idx_devices_last_seen | devices | last_seen |
| idx_disks_device_id | disks | device_id |
| idx_network_interfaces_device_id | network_interfaces | device_id |
| idx_installed_software_device_id | installed_software | device_id |
| idx_installed_software_device_name | installed_software | (device_id, name) |
| idx_remote_tools_device_id | remote_tools | device_id |
| *(plus implicit PK/UNIQUE indexes)* | | |

---

## 7. API ROUTES

| Method | Path | Handler | Middleware | Auth |
|---|---|---|---|---|
| GET | `/healthz` | HealthHandler.Healthz | RequestID, Logging, CORS | None |
| HEAD | `/healthz` | HealthHandler.Healthz | RequestID, Logging, CORS | None (Docker healthcheck) |
| GET | `/readyz` | HealthHandler.Readyz | RequestID, Logging, CORS | None |
| POST | `/api/v1/enroll` | AuthHandler.Enroll | RequestID, Logging, CORS | `X-Enrollment-Key` header |
| POST | `/api/v1/inventory` | InventoryHandler.SubmitInventory | RequestID, Logging, CORS, **DeviceAuth** | Bearer device-token |
| POST | `/api/v1/auth/login` | AuthHandler.Login | RequestID, Logging, CORS | Public (username/password body) |
| POST | `/api/v1/auth/logout` | AuthHandler.Logout | RequestID, Logging, CORS | None (clears cookie) |
| GET | `/api/v1/devices` | DeviceHandler.ListDevices | RequestID, Logging, CORS, **JWTAuth** | session cookie (JWT) |
| GET | `/api/v1/devices/:id` | DeviceHandler.GetDevice | RequestID, Logging, CORS, **JWTAuth** | session cookie (JWT) |

### Query Parameters — `GET /api/v1/devices`

| Param | Type | Description |
|---|---|---|
| `hostname` | string | Filter by hostname (ILIKE) |
| `os` | string | Filter by os_name OR os_version (ILIKE) |

### Response Structures

- **Device list:** `{ "devices": [...], "total": N }`
- **Device detail:** `{ "device": {...}, "hardware": {...}, "disks": [...], "network_interfaces": [...], "installed_software": [...], "remote_tools": [...] }`
- **Error:** `{ "error": "message" }`
- **Health:** `{ "status": "ok" }`

---

## 8. AUTHENTICATION FLOWS

### 8.1 Agent Enrollment

```
Agent (no token) → POST /api/v1/enroll
  Header: X-Enrollment-Key: <enrollment_key from config.json>
  Body:   { "hostname": "...", "serial_number": "..." }

Server:
  1. Validate key == ENROLLMENT_KEY env var
  2. BEGIN TX
  3. Upsert device by serial_number (INSERT or UPDATE hostname+last_seen)
  4. Delete old token for device (if exists)
  5. Generate UUID token → SHA256 hash → store hash in device_tokens
  6. COMMIT
  7. Return { "device_id": "uuid", "token": "raw-uuid" }

Agent:
  8. Persist raw token to data/device.token (file perms 0600, dir 0700)
```

### 8.2 Device Authentication (Inventory Submission)

```
Agent → POST /api/v1/inventory
  Header: Authorization: Bearer <raw-token>

Middleware (DeviceAuth):
  1. Extract token from Authorization header
  2. SHA256(raw-token) → lookup hash in device_tokens table
  3. Set device_id in Gin context
  4. Handler extracts device_id from context
```

### 8.3 Dashboard Login

```
Browser → POST /api/v1/auth/login
  Body: { "username": "admin", "password": "..." }

Server:
  1. Lookup user by username
  2. bcrypt.Compare(password, password_hash) — cost 12
  3. Generate JWT HS256: { sub: user_id, username, iat, exp: +24h }
  4. Set cookie: session=<JWT>; httpOnly=true; Secure=false; Path=/; Max-Age=86400

Frontend:
  5. Set localStorage.authenticated = "true"
```

### 8.4 JWT Auth (Dashboard → API)

```
Browser → GET /api/v1/devices
  Cookie: session=<JWT>

Middleware (JWTAuth):
  1. Extract JWT from cookie named "session"
  2. Validate HS256 signature using JWT_SECRET
  3. Check exp claim (24h)
  4. Set user_id and username in Gin context
```

### 8.5 Logout

```
Browser → POST /api/v1/auth/logout
Server: Set cookie session with Max-Age=-1 (deletes it)
Frontend: Remove localStorage.authenticated
```

### 8.6 Automatic Re-enrollment

```
If inventory submit returns 401/403:
  1. client.IsAuthError() detects "status 401" or "status 403" in error string
  2. Agent deletes stored token (store.Delete())
  3. Next cycle: no token found → Enroll() automatically
```

---

## 9. AGENT PIPELINE

### 9.1 Execution Modes

| Command | Description |
|---|---|
| `inventory-agent.exe run -config config.json` | Foreground mode: enroll + collect + submit + sleep |
| `inventory-agent.exe collect` | Dry-run: collect and print JSON to stdout (no server) |
| `inventory-agent.exe install -config "C:\path\config.json"` | Install as Windows Service |
| `inventory-agent.exe start` | Start the Windows Service |
| `inventory-agent.exe stop` | Stop the Windows Service |
| `inventory-agent.exe uninstall` | Remove the Windows Service |
| `inventory-agent.exe version` | Print version string |

### 9.2 Collection Pipeline

```
main.runAgent() → ticker loop (interval_hours) → runCycle()

runCycle:
  1. collector.Collect() — sequential:
     ├── collectSystem()       → WMI: Win32_OperatingSystem, Win32_BIOS, Win32_ComputerSystem + os.Hostname()
     ├── collectHardware()     → WMI: Win32_Processor, Win32_PhysicalMemory, Win32_BaseBoard, Win32_BIOS
     ├── collectDisks()        → WMI: Win32_DiskDrive (model, size, mediaType, serial, interface)
     ├── collectNetwork()      → WMI: Win32_NetworkAdapter (PhysicalAdapter=TRUE) + Go net.Interfaces() for IPs
     ├── collectSoftware()     → Registry: Uninstall keys (HKLM+HKCU, 32+64bit), skips SystemComponent=1
     ├── collectLicense()      → WMI: SoftwareLicensingProduct (Windows ApplicationID)
     └── collectRemoteTools()  → Registry + config files + CLI:
         ├── TeamViewer: HKLM\SOFTWARE\TeamViewer → ClientID (DWORD) + Version
         ├── AnyDesk: %ProgramData%\AnyDesk\system.conf → ad.anynet.id= ; also %APPDATA%; uninstall registry for version
         └── RustDesk: Registry → ID; TOML config → id= ; CLI fallback → rustdesk.exe --get-id (v1.4+)

  2. If no token → Enroll() → persist token
  3. SubmitWithRetry(inventory, maxRetries=5)
     - Exponential backoff: 2^attempt seconds + random jitter (0-1000ms)
     - On 401/403: clear token → re-enroll next cycle
```

### 9.3 RustDesk ID Detection (v1.4+ Support)

RustDesk v1.4+ changed config format from plain `id` to `enc_id` (encrypted).
The agent uses a 3-tier fallback strategy:

1. **Registry**: `HKLM/HKCU\SOFTWARE\RustDesk` → `ID` string value
2. **TOML config**: `%APPDATA%\RustDesk\config\RustDesk.toml` → `id = value` line
3. **CLI fallback**: `rustdesk.exe --get-id` subprocess (10s timeout) — finds exe from:
   - `%ProgramFiles%\RustDesk\rustdesk.exe`
   - `%ProgramFiles(x86)%\RustDesk\rustdesk.exe`
   - Uninstall registry `InstallLocation` field

### 9.4 Soft-fail Behavior

Collectors for hardware, disks, network, software, license, and remote_tools **log warnings but do not abort** if they fail. Only system collection failure returns an error. This ensures partial data is still submitted.

### 9.5 Agent Configuration (`config.json`)

| Field | Type | Default | Required |
|---|---|---|---|
| `server_url` | string | — | **Yes** |
| `enrollment_key` | string | — | **Yes** |
| `interval_hours` | int | 1 (if ≤ 0) | No |
| `data_dir` | string | `<exe_dir>/data` | No |
| `log_level` | string | `info` | No |
| `insecure_skip_verify` | bool | false | No |

---

## 10. FRONTEND ARCHITECTURE

### 10.1 Routes

| Path | Component | Auth | Layout |
|---|---|---|---|
| `/login` | Login | Public (redirect to `/` if authed) | None |
| `/` | Dashboard | Protected | Layout (sidebar) |
| `/devices` | DeviceList | Protected | Layout |
| `/devices/:id` | DeviceDetail | Protected | Layout |
| `/settings` | Settings | Protected | Layout |
| `*` | Navigate to `/` | — | — |

### 10.2 Key Components

| Component | Location | Purpose |
|---|---|---|
| **Layout** | `components/Layout.tsx` | 240px fixed sidebar (Dashboard, Devices, Settings nav items) + logout + Outlet |
| **ProtectedRoute** | `components/ProtectedRoute.tsx` | Redirect to `/login` if `!isAuthenticated` |
| **AuthProvider** | `hooks/useAuth.tsx` | Context: `isAuthenticated`, `login()`, `logout()`. State from `localStorage.authenticated` flag |
| **API Client** | `api/client.ts` | Centralized `fetch()` wrapper: `credentials: 'include'`, auto-redirect on 401, Content-Type: JSON. Custom `ApiError` class |
| **QueryClient** | `main.tsx` | TanStack Query: `retry: 1`, `refetchOnWindowFocus: false`, `staleTime: 30_000` |

### 10.3 Pages Detail

| Page | Key Behavior |
|---|---|
| **Login** | Form → `POST /api/v1/auth/login` → `AuthContext.login()` → redirect `/`. Logo icon + "Inventory" branding. Error message display. |
| **Dashboard** | 3 stat cards (Total / Online / Offline). Online = `last_seen < 1 hour ago`. Calculated client-side from full device list. Link to devices page. |
| **DeviceList** | Table with hostname + OS search filters (debounced via TanStack Query keys). Status badges (Online/Offline with colored dots). Row hostname click → device detail. Columns: Hostname, OS, Agent, Last Seen, Status. |
| **DeviceDetail** | Sections: Remote Access (table with aligned columns: colored dot, tool name, version, ID, copy button), System, Hardware, Disks, Network Interfaces, Software (scrollable max-h-96). Sub-components: `Section`, `Grid`, `Field`, `Th`, `Td`, `RemoteToolRow`. Brand colors: TeamViewer=blue, AnyDesk=red, RustDesk=orange. |
| **Settings** | Placeholder: "Settings will be available in a future update" with gear icon. |

### 10.4 Vite Proxy (Development)

```ts
// vite.config.ts
proxy: { '/api': { target: 'http://localhost:8081' } }
```

---

## 11. DARK THEME SYSTEM

### 11.1 CSS Custom Properties (Tailwind v4 `@theme`)

```css
@theme {
  --color-bg-primary: #0f1117;       /* Page background */
  --color-bg-secondary: #1a1d27;     /* Cards, sidebar */
  --color-bg-tertiary: #232733;      /* Inputs, table headers, nested elements */
  --color-border-primary: #2a2d3a;   /* Main borders */
  --color-border-secondary: #353849; /* Subtle borders */
  --color-text-primary: #e2e8f0;     /* Main text */
  --color-text-secondary: #94a3b8;   /* Secondary text */
  --color-text-muted: #64748b;       /* Muted/disabled text */
  --color-accent: #3b82f6;           /* Primary accent (blue-500) */
  --color-accent-hover: #2563eb;     /* Accent hover (blue-600) */
  --color-success: #22c55e;          /* Online badges, success states */
  --color-danger: #ef4444;           /* Error states, destructive actions */
}
```

### 11.2 Usage Pattern

Tailwind v4 auto-generates utility classes from `@theme`:
- Backgrounds: `bg-bg-primary`, `bg-bg-secondary`, `bg-bg-tertiary`
- Borders: `border-border-primary`, `border-border-secondary`
- Text: `text-text-primary`, `text-text-secondary`, `text-text-muted`
- Accents: `bg-accent`, `text-accent`, `bg-accent/10` (with opacity)
- Status: `bg-success/15`, `text-success`, `text-danger`

### 11.3 Additional Styles

- `html { color-scheme: dark; }` — native dark mode for scrollbars and form controls
- Custom scrollbar styling (`::-webkit-scrollbar`) with theme colors
- Body: `bg-bg-primary text-text-primary`

---

## 12. CODE CONVENTIONS

### 12.1 Go Naming

| Pattern | Convention | Example |
|---|---|---|
| Packages | lowercase single-word | `handler`, `service`, `repository`, `collector`, `middleware` |
| Files | named by domain entity or concern | `device.go`, `auth.go`, `cors.go`, `remote.go` |
| Constructors | `NewXxxYyy(deps)` | `NewDeviceRepository(db)`, `NewAuthService(db, userRepo, tokenRepo, jwtSecret)` |
| Struct tags | json + db + binding | `json:"snake_case" db:"snake_case" binding:"required"` |
| Errors | `fmt.Errorf(...)` | No typed domain errors (yet) |

### 12.2 TypeScript Naming

| Pattern | Convention |
|---|---|
| Interfaces | PascalCase, matching Go struct names exactly |
| Files | camelCase for modules, PascalCase for React components |
| API functions | camelCase verbs: `getDevices()`, `login()` |

### 12.3 Architectural Patterns

```
Handler → Service → Repository (clean architecture, simplified)

Handler:
  - Parse request (ShouldBindJSON, query params, path params)
  - Call service method
  - Return JSON response (dto.ErrorResponse or data)

Service:
  - Business logic, may use DB transactions
  - Calls one or more repositories
  - Returns domain models or errors

Repository:
  - Raw SQL via sqlx (GetContext, SelectContext, ExecContext, NamedExecContext)
  - Maps to/from shared/models structs
  - Returns []T{} (never nil) when no results
```

### 12.4 DTO Flow

```
Agent collectors → dto.InventoryRequest → JSON → HTTP POST
                                                      ↓
Server: JSON → dto.InventoryRequest → repository.Save(deviceID, req) → SQL upserts
                                                      ↓
Server: SQL → models.Device/Hardware/etc → dto.DeviceDetailResponse → JSON
                                                      ↓
Frontend: JSON → TypeScript interfaces (types/index.ts) → React components
```

### 12.5 Inventory Transaction Pattern

```sql
BEGIN TX
  1. UPSERT devices ON CONFLICT (id) SET all columns
  2. UPSERT hardware ON CONFLICT (device_id) SET all columns
  3. DELETE disks WHERE device_id = $1 → INSERT new disks
  4. DELETE network_interfaces WHERE device_id = $1 → INSERT new
  5. DELETE installed_software WHERE device_id = $1 → INSERT new
  6. DELETE remote_tools WHERE device_id = $1 → INSERT new
COMMIT
```

Full snapshot replacement on every collection (no delta sync in Phase 1).

### 12.6 Other Conventions

- **Dependency injection:** Manual in `main.go` (no DI framework)
- **Logging:** `log/slog` (stdlib), JSON handler, structured key-value pairs
- **Graceful shutdown:** Signal listener → context cancel → `srv.Shutdown(ctx)` with 10s timeout
- **Nil-safe slices:** Repositories return `[]T{}` instead of nil when no results found
- **Agent version:** Hardcoded const `version = "0.1.0"` in `agent/cmd/agent/main.go` and `AgentVersion = "0.1.0"` in `collector.go`
- **HTTP client timeout:** 30s for agent HTTP client

---

## 13. CONFIGURATION

### 13.1 Server — Environment Variables

| Variable | Description | Default | Required |
|---|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://inventory:changeme@localhost:5432/inventory?sslmode=disable` | Yes |
| `POSTGRES_PASSWORD` | PostgreSQL password (Docker Compose) | `changeme` | Yes |
| `SERVER_PORT` | API host listen port (docker maps to 8081 inside) | `8081` | No |
| `JWT_SECRET` | JWT HS256 signing key (≥32 chars, exits if empty) | — | **Yes** |
| `ENROLLMENT_KEY` | Agent enrollment key (exits if empty) | — | **Yes** |
| `CORS_ORIGINS` | Comma-separated allowed origins | `http://localhost:3000` | No |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` | No |

### 13.2 Current Dev Values (`.env`)

```env
SERVER_PORT=8081
JWT_SECRET=super-secret-jwt-key-for-dev-32chars!
ENROLLMENT_KEY=dev-enrollment-key-2026
CORS_ORIGINS=http://localhost:5173,http://localhost:3000
```

### 13.3 Database Connection Pool

| Setting | Value |
|---|---|
| MaxOpenConns | 25 |
| MaxIdleConns | 5 |
| ConnMaxLifetime | 5 minutes |

### 13.4 HTTP Server Timeouts

| Setting | Value |
|---|---|
| ReadTimeout | 15s |
| WriteTimeout | 30s |
| IdleTimeout | 60s |
| Graceful shutdown | 10s |

---

## 14. DOCKER & DEPLOYMENT

### 14.1 docker-compose.yml — 2 Services

| Service | Image | Port | Healthcheck | Restart |
|---|---|---|---|---|
| **postgres** | postgres:16-alpine | 5432:5432 | `pg_isready -U inventory` (10s interval, 5 retries) | unless-stopped |
| **api** | inventario-api (built) | ${SERVER_PORT:-8081}:8081 | `wget -qO/dev/null http://localhost:8081/healthz` (30s interval, 3 retries, 15s start_period) | unless-stopped |

- API depends_on postgres (condition: service_healthy)
- Volume: `postgres-data` for persistent database
- Logging: json-file driver, max-size 10m, max-file 5

### 14.2 server/Dockerfile (Multi-stage)

**Stage 1 — Builder:**
```dockerfile
FROM golang:1.24-alpine
# git, copy shared/ + server/, GOWORK=off, go mod download
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server ./cmd/api
```

**Stage 2 — Runtime:**
```dockerfile
FROM alpine:3.20
# ca-certificates, wget (healthcheck), tzdata
# COPY --from=builder /app/server
# EXPOSE 8081 → CMD ["./server"]
```

Note: Migrations are embedded via `//go:embed *.sql` in the binary — no separate migration files in the container.

### 14.3 No Frontend Container

Frontend runs separately via `npm run dev` (Vite dev server) in development. Production deployment with Nginx is documented in ITIL docs but NOT implemented.

---

## 15. ARCHITECTURAL DECISIONS (ADR)

| ADR | Decision | Rationale |
|---|---|---|
| ADR-001 | Modular monolith | Solo dev, 100-500 devices, simplicity |
| ADR-002 | HTTP in Phase 1 | On-premises internal network; HTTPS planned for Phase 2 |
| ADR-003 | Gin over Fiber | `net/http` compatibility, 12-year maturity, rich middleware |
| ADR-004 | sqlx over GORM | Explicit SQL control, clean architecture alignment |
| ADR-005 | Vite over CRA | CRA deprecated, 10-100x faster dev builds |
| ADR-006 | httpOnly cookies over localStorage | XSS protection for JWT |
| ADR-007 | Enrollment key + device token | Practical for 100-500 devices vs pre-generating tokens |
| ADR-008 | log/slog over zerolog/zap | Built-in stdlib, zero external dependency |

---

## 16. SLA TARGETS (from ITIL docs)

| Metric | Target |
|---|---|
| API uptime | >= 99.5% monthly (~3.6h downtime allowed/month) |
| Inventory latency p95 | < 500ms |
| Device list latency p95 | < 200ms |
| Error rate (5xx) | < 0.1% |
| Agent collection success | >= 99% |
| Data freshness | last_seen < 2x interval |
| Database uptime | >= 99.9% |

---

## 17. IMPLEMENTED FEATURES (Phase 1 — Current State)

### Agent
- [x] System data collection (hostname, OS, serial, boot time, logged-in user)
- [x] Hardware collection (CPU, RAM, motherboard, BIOS)
- [x] Disk collection (model, size, type SSD/HDD, serial, interface)
- [x] Network collection (adapters, MAC, IPv4/IPv6, speed, physical flag)
- [x] Software collection (installed programs via Registry: name, version, vendor, date)
- [x] License collection (Windows activation status via WMI)
- [x] Remote tools detection (TeamViewer ID, AnyDesk ID, RustDesk ID with v1.4+ CLI fallback)
- [x] Windows Service mode (install/start/stop/uninstall)
- [x] Foreground mode + collect-only dry-run CLI
- [x] Token persistence (file-based, 0600 permissions)
- [x] Exponential backoff retry (5 retries)
- [x] Automatic re-enrollment on token invalidation
- [x] TLS skip verify option for dev

### API
- [x] Agent enrollment (POST /api/v1/enroll)
- [x] Inventory submission with transactional upsert (POST /api/v1/inventory)
- [x] Device list with hostname+OS filters (GET /api/v1/devices)
- [x] Device detail with all related data (GET /api/v1/devices/:id)
- [x] Dashboard login via JWT httpOnly cookie (POST /api/v1/auth/login)
- [x] Logout (POST /api/v1/auth/logout)
- [x] Health + readiness probes (/healthz, /readyz) — GET + HEAD
- [x] CLI `create-user` subcommand
- [x] Embedded SQL migrations (auto-run on startup)
- [x] CORS with origin whitelist + credentials
- [x] Structured request logging (slog)
- [x] Request ID middleware (UUID)
- [x] Graceful shutdown (signal-based)

### Frontend
- [x] Dark theme (CSS custom properties via Tailwind v4 @theme)
- [x] Sidebar navigation (Dashboard, Devices, Settings) with active state highlight
- [x] Login page with error handling and branding
- [x] Dashboard with stat cards (Total / Online / Offline)
- [x] Device list with hostname+OS search, online badges
- [x] Device detail: System, Hardware, Disks, Network, Software, Remote Access sections
- [x] Remote Access table with aligned columns, copy-to-clipboard, brand color dots
- [x] Settings placeholder page
- [x] Auth guard (ProtectedRoute)
- [x] Auto-redirect to /login on 401

### Infrastructure
- [x] Docker Compose (PostgreSQL 16 + API)
- [x] Multi-stage Dockerfile (golang:1.24-alpine → alpine:3.20)
- [x] Docker healthchecks (pg_isready + wget /healthz)
- [x] Makefile with build/run/lint/test/docker targets
- [x] .env.example template
- [x] ITIL v4 documentation (30 documents)

---

## 18. SECURITY POSTURE

| Aspect | Implementation |
|---|---|
| Device tokens | SHA-256 hashed in DB; raw token never stored server-side |
| Passwords | bcrypt cost 12 |
| JWT | HS256, 24h expiration, httpOnly cookie |
| Cookies | httpOnly=true, Secure=false (Phase 1), Path=/ |
| CORS | Whitelist-based with credentials support |
| Agent auth | Bearer token → SHA256 → DB lookup |
| Enrollment | Shared key via X-Enrollment-Key header |
| PostgreSQL | Accessible via Docker internal network + exposed on 5432 for dev |
| Agent TLS | Configurable `insecure_skip_verify` for dev environments |

---

## 19. KNOWN GAPS (Docs vs Code)

These features are described in the ITIL documentation or README but are **NOT yet implemented** in the codebase:

| Gap | Documentation Reference | Status |
|---|---|---|
| `GET /api/v1/dashboard/stats` endpoint | README, service catalog | **NOT IMPLEMENTED** — Dashboard calculates stats client-side |
| `GET /api/v1/auth/me` endpoint | README | **NOT IMPLEMENTED** — Auth relies on localStorage flag |
| JWT refresh token flow (15min access + 7d refresh) | gestao-de-seguranca.md | **NOT IMPLEMENTED** — Single 24h JWT in cookie |
| Rate limiting | gestao-de-seguranca.md, CSI-006 | **NOT IMPLEMENTED** |
| Security headers (HSTS, X-Frame-Options, etc.) | gestao-de-seguranca.md | **NOT IMPLEMENTED** |
| Docker non-root user | gestao-de-seguranca.md | **NOT IMPLEMENTED** — Dockerfile doesn't set USER |
| Nginx container for dashboard (prod) | arquitetura-da-solucao.md | **NOT IMPLEMENTED** — Frontend runs via `npm run dev` |
| Server-side pagination | CSI-010 | **NOT IMPLEMENTED** — Returns all devices |
| HTTPS/TLS | CSI-001, RISK-SEC-001 | **NOT IMPLEMENTED** — Uses HTTP |

---

## 20. IMPROVEMENT ROADMAP (CSI Register)

### Phase 2 — Consolidation (target: ~2 months after Phase 1 stable)

| Sprint | ID | Feature | Priority | Effort |
|---|---|---|---|---|
| 2.1 | CSI-001 | HTTP → HTTPS (TLS termination) | Alta | Medio |
| 2.1 | CSI-006 | Rate limiting by IP | Media | Baixo |
| 2.2 | CSI-005 | Alert on inactive agents | Alta | Baixo |
| 2.2 | CSI-010 | Server-side pagination | Media | Baixo |
| 2.3 | CSI-003 | CSV export | Media | Baixo |
| 2.3 | CSI-004 | Dashboard charts/evolution | Media | Medio |
| 2.4 | CSI-009 | Prometheus + Grafana monitoring | Media | Medio |
| 2.4 | CSI-013 | Hardware change history/tracking | Media | Medio |

### Phase 3 — Expansion (3+ months)

| Sprint | ID | Feature | Priority | Effort |
|---|---|---|---|---|
| 3.1 | CSI-002 | Delta sync (incremental inventory) | Media | Alto |
| 3.2 | CSI-008 | RBAC (role-based access control) | Media | Medio |
| 3.3 | CSI-007 | Linux agent | Baixa | Alto |
| 3.4 | CSI-011 | Remote agent configuration API | Baixa | Alto |
| 3.4 | CSI-012 | Printer collection | Baixa | Medio |

### Phase 4+ — Maturity

| ID | Feature | Priority | Effort |
|---|---|---|---|
| CSI-014 | Disk full alerts | Baixa | Medio |
| CSI-015 | Network scanning/discovery | Baixa | Alto |
| — | CMDB integration | Baixa | Alto |
| — | OpenAPI/Swagger docs | Baixa | Baixo |
| — | Mobile-responsive dashboard | Baixa | Medio |

---

## 21. OUT OF SCOPE — DO NOT IMPLEMENT

Unless explicitly requested, do **NOT** implement:

- Multi-tenant support
- Remote command execution
- Network scanner / active probing
- ITSM / ticketing integration
- RMM features (remote desktop, patch management)
- Linux agent (Phase 3)
- External integrations (AD/LDAP, SCCM, etc.)
- Mobile application
- Microservices architecture refactoring

---

## 22. GIT HISTORY

```
08daa0d (HEAD) fix: RustDesk ID detection via --get-id CLI + remote tools list layout
f3c3aa8        feat: dark theme, sidebar navigation, remote access tool detection
9f7f1df        fix: remove ITIL docs, add project README
f769af0        feat: initial implementation - full inventory system
```

---

## 23. DEVELOPMENT WORKFLOW

### Start Backend

```bash
# Copy and fill environment variables
cp .env.example .env
# Edit .env (JWT_SECRET, ENROLLMENT_KEY, etc.)

# Start PostgreSQL + API (auto-runs migrations)
docker compose up -d --build

# Create admin user
docker exec inventory-api ./server create-user --username admin --password <password>
```

### Start Frontend (Development)

```bash
cd frontend
npm install
npm run dev
# Access http://localhost:5173 (or next available port)
```

### Build & Test Agent

```bash
cd agent

# Build
go build -o inventory-agent.exe ./cmd/agent

# Dry-run (collect + print JSON, no server needed)
.\inventory-agent.exe collect

# Run in foreground (enroll + submit + sleep loop)
.\inventory-agent.exe run -config config.json

# Install as Windows Service
.\inventory-agent.exe install -config "C:\full\path\config.json"
.\inventory-agent.exe start
```

### Useful Commands

```bash
# View API logs
docker logs inventory-api --tail 50 -f

# Rebuild after code changes
docker compose up -d --build

# Reset database (drops volume)
docker compose down -v
docker compose up -d --build

# Run Go checks
cd server && go vet ./... && go build ./...
cd agent && go vet ./... && go build ./cmd/agent

# Frontend build check
cd frontend && npm run build

# Go mod tidy all modules
make tidy
```

---

## 24. MINOR CODE NOTES

- **Copy-paste comment in models.go**: The `RemoteTool` struct has the comment `// User represents a dashboard user account` — this is a copy-paste artifact from the `User` struct. Harmless but should be corrected to `// RemoteTool represents a detected remote access tool`.
