# Instalação e Deploy

## Pré-requisitos

- Docker + Docker Compose
- Go 1.24 (para desenvolvimento/build local)
- Node.js 20 (para desenvolvimento frontend)

## Quick Start com Docker

### 1. Configurar variáveis de ambiente

```bash
cp .env.example .env
```

Edite o `.env`:

```dotenv
# Senha do PostgreSQL
POSTGRES_PASSWORD=uma-senha-forte

# Porta da API (default: 8081)
SERVER_PORT=8081

# Nível de log: debug, info, warn, error
LOG_LEVEL=info

# OBRIGATÓRIO: chave para assinar JWT (min 32 caracteres)
JWT_SECRET=gerar-uma-string-aleatoria-de-32-chars-minimo

# OBRIGATÓRIO: chave que os agents usam para se registrar
ENROLLMENT_KEY=uma-chave-secreta-para-agents

# Origens permitidas pelo CORS (separadas por vírgula)
# Para acesso na rede local, adicione http://<SEU-IP>:5173
CORS_ORIGINS=http://localhost:5173,http://localhost:3000

# Cleanup automático (opcional)
RETENTION_DAYS=90         # Dias para reter logs
INACTIVE_DAYS=30          # Dias sem comunicação → dispositivo inativo
CLEANUP_INTERVAL=24h      # Intervalo entre execuções
```

### 2. Subir os serviços

```bash
docker compose up -d --build
```

Isso cria:
- **inventory-postgres** — PostgreSQL 16 com volume persistente
- **inventory-api** — API Go compilada em Alpine Linux (non-root)

A API espera o PostgreSQL estar healthy (healthcheck via `pg_isready`) antes de iniciar.

### 3. Criar o primeiro usuário admin

```bash
docker compose exec api ./server create-user --username admin --password senha_segura
```

Ou via Makefile:

```bash
make create-user USERNAME=admin PASSWORD=senha_segura
```

### 4. Acessar o sistema

- API: `http://localhost:8081`
- Health check: `http://localhost:8081/healthz`
- Readiness: `http://localhost:8081/readyz`

## Docker Compose — Detalhes

```yaml
services:
  postgres:
    image: postgres:16-alpine
    healthcheck:
      test: pg_isready -U inventory    # Verifica se o banco está pronto
      interval: 10s
    volumes:
      - postgres-data:/var/lib/postgresql/data   # Dados persistidos
    restart: unless-stopped

  api:
    build: server/Dockerfile
    ports: "8081:8081"
    depends_on:
      postgres:
        condition: service_healthy   # Espera o banco estar saudável
    healthcheck:
      test: wget --tries=1 -qO/dev/null http://localhost:8081/healthz
      interval: 30s
      start_period: 15s              # 15s para inicialização
    restart: unless-stopped
```

Ambos os serviços têm limite de log: 10MB por arquivo, máximo 5 arquivos.

## Build Manual (sem Docker)

### Servidor

```bash
cd server
go build -o ../bin/server ./cmd/api
```

Ou via Makefile: `make build-server`

### Agent (cross-compile para Windows)

```bash
cd agent
GOOS=windows GOARCH=amd64 go build -o ../bin/agent.exe ./cmd/agent
```

Ou via Makefile: `make build-agent`

### Frontend

```bash
cd frontend
npm install
npm run build    # Gera dist/ com os assets estáticos
```

## Deploy do Agent em Windows

### 1. Preparar arquivos

Copie para uma pasta no Windows (ex: `C:\Program Files\InventoryAgent\`):

```
inventory-agent.exe
config.json
```

### 2. Criar config.json

```json
{
  "server_url": "http://SEU-SERVIDOR:8081",
  "enrollment_key": "mesma-chave-do-servidor",
  "interval_hours": 1,
  "data_dir": "data",
  "log_level": "info",
  "insecure_skip_verify": false
}
```

| Campo | Valor |
|-------|-------|
| `server_url` | URL da API (ex: `http://192.168.1.100:8081`) |
| `enrollment_key` | Mesma `ENROLLMENT_KEY` configurada no servidor |
| `interval_hours` | Intervalo entre coletas em horas |
| `insecure_skip_verify` | `true` apenas para HTTPS com certificado auto-assinado |

### 3. Testar coleta (sem servidor)

```cmd
inventory-agent.exe collect
```

Mostra o JSON que seria enviado. Útil para verificar se o WMI está funcionando.

### 4. Testar envio (modo foreground)

```cmd
inventory-agent.exe run
```

Roda em primeiro plano, mostra logs no console. Ctrl+C para parar.

### 5. Instalar como serviço

```cmd
# Execute como Administrador
inventory-agent.exe install
inventory-agent.exe start
```

O serviço será criado com inicio automático (inicia com o Windows).

### 6. Gerenciar o serviço

```cmd
inventory-agent.exe stop        # Para o serviço
inventory-agent.exe uninstall   # Remove o serviço
```

Ou via `services.msc` — procure por "Inventory Agent".

## Go Workspace

O projeto usa Go Workspace (`go.work`) para compartilhar código entre server e agent:

```go
go 1.24.0

use (
    ./agent
    ./server
    ./shared
)
```

O módulo `shared` contém models e DTOs que ambos server e agent importam. O workspace permite que os 3 módulos se referenciem sem require/replace no go.mod.

**Nota:** no Dockerfile, `GOWORK=off` é setado porque o build copia apenas os diretórios necessários.

## Makefile

```bash
make help            # Lista todos os targets
make build-server    # Compila a API
make build-agent     # Cross-compile do agent para Windows
make run             # Roda a API localmente (go run)
make test            # Testes com race detection + cobertura
make lint            # golangci-lint
make docker-up       # docker compose up -d --build
make docker-down     # docker compose down
make docker-logs     # docker compose logs -f
make create-user     # Cria usuário (USERNAME=x PASSWORD=y)
make tidy            # go mod tidy em todos os módulos
```

## CI/CD — GitHub Actions

Pipeline executado em push para `main` e `feature/*`, e PRs em `main`.

### Jobs (6 total)

| Job | Runner | O que faz |
|-----|--------|-----------|
| **lint** | ubuntu-latest | golangci-lint no server (timeout 5min) |
| **build-server** | ubuntu-latest | `go build` do server |
| **build-agent** | ubuntu-latest | `GOOS=windows go build` do agent |
| **test** | ubuntu-latest | Testes com PostgreSQL real (service container), race detection, cobertura → Codecov |
| **frontend** | ubuntu-latest | `npm ci` → lint → build |
| **docker** | ubuntu-latest | Build da imagem Docker com cache GitHub Actions (depende de build-server e build-agent) |

O job de **test** sobe um PostgreSQL 16 como service container com healthcheck, roda os testes com cobertura e envia para Codecov.

## Frontend — Desenvolvimento

### Proxy de Desenvolvimento

O Vite está configurado para proxy de API:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': 'http://localhost:8081',
  },
}
```

Toda requisição para `/api/*` é redirecionada para o backend em `localhost:8081`.

### Rodar em desenvolvimento

```bash
cd frontend
npm install
npm run dev       # Inicia Vite dev server (hot reload)
```

## HTTPS (Produção)

Para produção com HTTPS, use um reverse proxy (Nginx, Caddy, Traefik) na frente da API.

### Exemplo com Nginx

```nginx
server {
    listen 443 ssl;
    server_name inventario.empresa.com;

    ssl_certificate     /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    # Frontend estático
    location / {
        root /var/www/inventario;
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Exemplo com Caddy (mais simples)

```
inventario.empresa.com {
    root * /var/www/inventario
    file_server
    try_files {path} /index.html

    handle /api/* {
        reverse_proxy localhost:8081
    }
}
```

Caddy gera certificados HTTPS automaticamente via Let's Encrypt.

### Configuração do Agent com HTTPS

No `config.json` do agent, mude o `server_url`:

```json
{
  "server_url": "https://inventario.empresa.com",
  "insecure_skip_verify": false
}
```

Se usando certificado auto-assinado (dev):

```json
{
  "server_url": "https://10.0.0.100",
  "insecure_skip_verify": true
}
```

Atualize também `CORS_ORIGINS` no servidor:

```dotenv
CORS_ORIGINS=https://inventario.empresa.com
```
