# Inventory System

Sistema de inventário automatizado de hardware e software para estações Windows.  
Um agente Windows coleta informações da máquina via WMI e envia para uma API centralizada, que alimenta um dashboard web.

## Arquitetura

```
┌──────────────┐        HTTPS/JSON        ┌──────────────┐       ┌────────────┐
│  Agent (Win) │ ───────────────────────►  │   API (Go)   │ ───►  │ PostgreSQL │
│  WMI + Svc   │  enrollment / inventory   │   Gin + JWT  │       │    16      │
└──────────────┘                           └──────┬───────┘       └────────────┘
                                                  │
                                           ┌──────┴───────┐
                                           │   Frontend   │
                                           │ React + Vite │
                                           └──────────────┘
```

## Stack

| Componente | Tecnologia |
|------------|------------|
| **API** | Go 1.24 · Gin · sqlx · pgx v5 · golang-migrate · JWT |
| **Agent** | Go 1.24 · WMI · Windows Service (x/sys/windows/svc) |
| **Frontend** | React 18 · TypeScript · Vite · Tailwind CSS v4 · TanStack Query |
| **Database** | PostgreSQL 16 |
| **Deploy** | Docker Compose |

## Estrutura do Projeto

```
├── server/              # API REST (Go)
│   ├── cmd/api/         # Entrypoint + CLI create-user
│   ├── internal/
│   │   ├── config/      # Variáveis de ambiente
│   │   ├── database/    # Conexão + migrations
│   │   ├── handler/     # HTTP handlers
│   │   ├── middleware/   # Auth, CORS, logging, request ID
│   │   ├── repository/  # Camada de dados (sqlx)
│   │   ├── router/      # Rotas Gin
│   │   └── service/     # Lógica de negócio
│   └── migrations/      # SQL migrations (embed)
│
├── agent/               # Agent Windows (Go)
│   ├── cmd/agent/       # Entrypoint (service + CLI)
│   └── internal/
│       ├── client/      # HTTP client com retry/backoff
│       ├── collector/   # Coletores WMI (system, hw, disk, net, sw, license)
│       ├── config/      # Loader JSON config
│       └── token/       # Persistência do token (arquivo)
│
├── shared/              # Módulo compartilhado
│   ├── models/          # Structs do domínio
│   └── dto/             # Request/Response DTOs
│
├── frontend/            # Dashboard Web (React)
│   └── src/
│       ├── api/         # HTTP client + endpoints
│       ├── components/  # Layout, ProtectedRoute
│       ├── hooks/       # useAuth context
│       ├── pages/       # Login, Dashboard, DeviceList, DeviceDetail
│       └── types/       # TypeScript interfaces
│
├── docker-compose.yml   # PostgreSQL + API
├── Makefile             # Targets de build/run
└── .env.example         # Template de variáveis
```

## Quick Start

### 1. Pré-requisitos

- Docker + Docker Compose
- Node.js 18+ (para frontend local)
- Go 1.24+ (para desenvolvimento)

### 2. Subir o Backend

```bash
# Copiar e preencher variáveis de ambiente
cp .env.example .env
# Editar .env com seus valores (JWT_SECRET, ENROLLMENT_KEY, etc)

# Subir PostgreSQL + API
docker compose up -d --build

# Criar usuário admin para o dashboard
docker compose exec api ./server create-user --username admin --password sua-senha
```

### 3. Subir o Frontend

```bash
cd frontend
npm install
npm run dev
# Acesse http://localhost:5173
```

### 4. Configurar o Agent (Windows)

```bash
# Copiar config de exemplo
cd agent
copy config.example.json config.json
# Editar config.json com o endereço do servidor e enrollment key

# Build
go build -o inventory-agent.exe ./cmd/agent

# Testar coleta local (sem servidor)
inventory-agent.exe collect

# Executar em foreground
inventory-agent.exe run -config config.json

# Instalar como Windows Service
inventory-agent.exe install -config "C:\caminho\completo\config.json"
inventory-agent.exe start
```

## API Endpoints

### Autenticação (Dashboard)

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/api/v1/auth/login` | Login (retorna JWT via cookie httpOnly) |
| `POST` | `/api/v1/auth/logout` | Logout (limpa cookie) |
| `GET` | `/api/v1/auth/me` | Dados do usuário logado |

### Dispositivos (Dashboard - requer JWT)

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/api/v1/devices` | Listar dispositivos (com paginação e busca) |
| `GET` | `/api/v1/devices/:id` | Detalhes completos do dispositivo |
| `GET` | `/api/v1/dashboard/stats` | Estatísticas para o dashboard |

### Agent (requer Bearer Token)

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/api/v1/enroll` | Enrollment (gera token do device) |
| `POST` | `/api/v1/inventory` | Envio de inventário |

### Health

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/healthz` | Health check |

## Dados Coletados pelo Agent

- **Sistema**: Hostname, OS, build, arquitetura, uptime, usuário logado
- **Hardware**: CPU (modelo, cores, threads), RAM total, placa-mãe, BIOS
- **Discos**: Modelo, tamanho, tipo (SSD/HDD), serial, interface
- **Rede**: Adaptadores, MAC, IPv4/IPv6, velocidade, tipo (físico/virtual)
- **Software**: Programas instalados (nome, versão, fabricante, data)
- **Licença**: Status de ativação do Windows

## Variáveis de Ambiente

| Variável | Descrição | Default |
|----------|-----------|---------|
| `DATABASE_URL` | Connection string PostgreSQL | — |
| `POSTGRES_PASSWORD` | Senha do PostgreSQL | `changeme` |
| `SERVER_PORT` | Porta da API (host) | `8080` |
| `JWT_SECRET` | Chave secreta para tokens JWT (min 32 chars) | — |
| `ENROLLMENT_KEY` | Chave de enrollment dos agents | — |
| `CORS_ORIGINS` | Origens permitidas (comma-separated) | `http://localhost:3000` |
| `LOG_LEVEL` | Nível de log (`debug`, `info`, `warn`, `error`) | `info` |

## Database Schema

7 tabelas: `users`, `devices`, `device_tokens`, `hardware`, `disks`, `network_interfaces`, `installed_software`. Migrations automáticas via embed no binário.

## Makefile

```bash
make help           # Listar targets disponíveis
make build-server   # Compilar API
make build-agent    # Compilar Agent (Windows)
make run            # Rodar API localmente
make docker-up      # docker compose up -d --build
make docker-down    # docker compose down
make docker-logs    # docker compose logs -f
make tidy           # go mod tidy em todos os módulos
make create-user USERNAME=admin PASSWORD=secret
```

## Licença

Uso interno.
