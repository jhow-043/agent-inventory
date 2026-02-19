# Visão Geral e Arquitetura

## O que é

Sistema de inventário de ativos de TI que coleta automaticamente informações de hardware e software de computadores Windows e centraliza tudo em um dashboard web.

## Componentes

O sistema tem 3 partes que se comunicam:

```
┌──────────────────┐         ┌──────────────────────────────────────┐
│   Agent Windows  │  HTTP   │           Servidor                   │
│                  │────────>│                                      │
│  - Roda como     │         │  ┌────────────┐    ┌─────────────┐  │
│    serviço       │         │  │  API (Go)   │───>│ PostgreSQL  │  │
│  - Coleta WMI    │         │  │  Gin :8080  │    │    :5432    │  │
│  - Envia dados   │         │  └─────┬──────┘    └─────────────┘  │
│    periodicamente│         │        │                             │
└──────────────────┘         └────────┼─────────────────────────────┘
                                      │
                             ┌────────┴─────────┐
                             │  Frontend (React) │
                             │  SPA no browser   │
                             │  Vite dev :3000    │
                             └──────────────────┘
```

| Componente | Linguagem | Diretório | Função |
|------------|-----------|-----------|--------|
| API | Go 1.24 + Gin | `server/` | Recebe dados, serve API REST, autenticação |
| Agent | Go 1.24 + WMI | `agent/` | Coleta inventário Windows e envia pra API |
| Frontend | React 18 + TypeScript | `frontend/` | Dashboard web, gestão de devices |
| Shared | Go | `shared/` | Models e DTOs compartilhados entre server e agent |

## Stack Completa

**Backend:** Go 1.24, Gin, sqlx, pgx v5, golang-migrate, golang-jwt/v5, bcrypt

**Agent:** Go 1.24, go-ole/WMI, golang.org/x/sys (Windows Service)

**Frontend:** React 18, TypeScript, Vite, TailwindCSS, TanStack Query, Recharts, React Router v6, Lucide icons

**Banco:** PostgreSQL 16

**Infra:** Docker Compose, GitHub Actions CI/CD

## Fluxo Completo

### 1. Agent se registra (Enrollment)

```
Agent                              API
  │                                 │
  │ POST /api/v1/enroll             │
  │ Header: X-Enrollment-Key       │
  │ Body: {hostname, serial_number} │
  │────────────────────────────────>│
  │                                 │ Valida enrollment key (constant-time)
  │                                 │ Cria device no banco (ou atualiza se já existe)
  │                                 │ Gera token UUID, salva hash SHA-256 no banco
  │  {device_id, token}            │
  │<────────────────────────────────│
  │                                 │
  │ Salva token em data/device.token│
```

- O agent precisa de uma `enrollment_key` que é a mesma configurada no servidor
- O servidor identifica o device pelo `serial_number` (número de série da BIOS)
- Se o device já existe, atualiza e gera novo token
- O token retornado é salvo em disco e usado nas próximas requisições

### 2. Agent envia inventário

```
Agent                              API
  │                                 │
  │ Coleta dados via WMI/Registry   │
  │                                 │
  │ POST /api/v1/inventory          │
  │ Header: Authorization: Bearer   │
  │ Body: {hardware, disks,         │
  │   network, software, remote...} │
  │────────────────────────────────>│
  │                                 │ Middleware DeviceAuth:
  │                                 │   SHA-256(token) → busca no banco
  │                                 │   Seta device_id no contexto
  │                                 │
  │                                 │ Transação no banco:
  │                                 │   1. Upsert device (hostname, OS, last_seen)
  │                                 │   2. Compara hardware → se mudou, salva snapshot
  │                                 │   3. Upsert hardware (CPU, RAM, BIOS, mobo)
  │                                 │   4. Replace disks (delete + insert)
  │                                 │   5. Replace network interfaces
  │                                 │   6. Replace software (chunks de 200)
  │                                 │   7. Replace remote tools
  │  {message: "ok"}               │
  │<────────────────────────────────│
```

- O agent coleta tudo de uma vez e envia o payload completo
- A API faz tudo numa transação única — se algo falhar, nada é salvo parcialmente
- Hardware changes são detectados automaticamente e salvos no histórico
- Software é inserido em lotes de 200 para não estourar o limite de parâmetros do PostgreSQL

### 3. Ciclo do Agent

```
Inicia → Carrega config.json → Carrega token salvo
  │
  ├─ Se sem token → Faz enrollment
  │
  ├─ Coleta inventário (WMI + Registry)
  │
  ├─ Envia pra API (com retry: 5 tentativas, exponential backoff + jitter)
  │     Se 401/403 → Limpa token → Re-enrollment no próximo ciclo
  │
  └─ Espera interval_hours → Repete
```

### 4. Usuário acessa o dashboard

```
Browser                            API
  │                                 │
  │ POST /api/v1/auth/login         │
  │ {username, password}            │
  │────────────────────────────────>│
  │                                 │ bcrypt.Compare(password, hash)
  │                                 │ Gera JWT (24h, HMAC-SHA256)
  │  Set-Cookie: session=JWT        │
  │  (httpOnly, SameSite=Lax)       │
  │<────────────────────────────────│
  │                                 │
  │ GET /api/v1/dashboard/stats     │
  │ Cookie: session=JWT             │
  │────────────────────────────────>│
  │                                 │ Middleware JWTAuth:
  │                                 │   Lê cookie → valida JWT
  │                                 │   Seta user_id, username, role
  │  {total, online, offline, ...}  │
  │<────────────────────────────────│
```

- JWT é enviado como cookie httpOnly (não acessível por JavaScript)
- O frontend só sabe se está autenticado via `localStorage.authenticated`
- Se a API retorna 401, o frontend redireciona para `/login`

## Autenticação e Autorização

### Dois mecanismos independentes

| Quem | Como se autentica | O que pode fazer |
|------|-------------------|------------------|
| **Agent** | Header `Authorization: Bearer <token>` | Enviar enrollment e inventário |
| **Usuário** | Cookie `session` com JWT | Acessar o frontend e API |

### RBAC (Role-Based Access Control)

| Role | Permissões |
|------|------------|
| `admin` | Tudo: ver devices, mudar status, gerenciar departamentos, criar/deletar usuários, ver audit logs |
| `viewer` | Somente leitura: dashboard, listar devices, ver detalhes, listar departamentos |

O role é verificado pelo middleware `RequireRole` que lê o campo `role` do JWT.

## Dados Coletados pelo Agent

| Categoria | Dados | Fonte |
|-----------|-------|-------|
| **Sistema** | Hostname, Serial Number, OS (nome/versão/build/arch), último boot, usuário logado | WMI `Win32_OperatingSystem`, `Win32_BIOS`, `Win32_ComputerSystem`, `os.Hostname()` |
| **CPU** | Modelo, cores, threads | WMI `Win32_Processor` |
| **RAM** | Total em bytes (soma dos pentes) | WMI `Win32_PhysicalMemory` |
| **Placa-mãe** | Fabricante, modelo, serial | WMI `Win32_BaseBoard` |
| **BIOS** | Vendor, versão | WMI `Win32_BIOS` |
| **Discos** | Modelo, tamanho, tipo (HDD/SSD), interface, serial, partições com espaço livre | WMI `Win32_DiskDrive` + `Win32_LogicalDisk` |
| **Rede** | Adaptadores físicos: nome, MAC, IPv4, IPv6, velocidade | WMI `Win32_NetworkAdapter` + Go `net.Interfaces()` |
| **Software** | Nome, versão, vendor, data de instalação | Registry `Uninstall` (HKLM + HKCU, x86 + x64) |
| **Licença Windows** | Status de ativação (Licensed, Unlicensed, etc.) | WMI `SoftwareLicensingProduct` |
| **Acesso Remoto** | TeamViewer (ID, versão), AnyDesk (ID, versão), RustDesk (ID, versão) | Registry + arquivos de config |

## Segurança

| Medida | Implementação |
|--------|---------------|
| Senhas | bcrypt hash (nunca em texto plano) |
| Enrollment key | Comparação constant-time (previne timing attack) |
| Device token | UUID → SHA-256 hash no banco (token raw nunca é salvo) |
| JWT | HMAC-SHA256, cookie httpOnly/SameSite (sem acesso via JS) |
| Rate limiting | Enrollment: 10/min por IP, Login: 5/min por IP |
| RBAC | Middleware `RequireRole` em rotas admin |
| Security headers | X-Frame-Options: DENY, CSP, X-Content-Type-Options, etc. |
| Audit log | Toda ação admin é registrada com user, IP, detalhes |
| Request ID | UUID por request para rastreabilidade nos logs |
| CORS | Whitelist de origens configurável |
