# Diagrama de Arquitetura Geral

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  

---

## Arquitetura de Alto Nível — Fase 1

```mermaid
graph TB
    subgraph "Estações Windows (100-500)"
        A1["Agent Windows<br/>Go Binary<br/>Windows Service"]
        A2["Agent Windows<br/>Go Binary<br/>Windows Service"]
        AN["Agent Windows<br/>Go Binary<br/>Windows Service"]
    end

    subgraph "Servidor Linux (Docker Compose)"
        subgraph "Container: API"
            API["API REST<br/>Go + Gin<br/>Porta 8080"]
        end
        subgraph "Container: PostgreSQL"
            DB[("PostgreSQL 16<br/>Porta 5432<br/>Volume persistente")]
        end
        subgraph "Container: Dashboard"
            WEB["React + Vite<br/>Serve estático<br/>Porta 3000"]
        end
    end

    subgraph "Usuários"
        U1["Administrador de TI<br/>Browser"]
        U2["Gestor de TI<br/>Browser"]
    end

    A1 -->|"HTTP POST /api/v1/inventory<br/>Device Token"| API
    A2 -->|"HTTP POST /api/v1/inventory<br/>Device Token"| API
    AN -->|"HTTP POST /api/v1/inventory<br/>Device Token"| API

    API -->|"SQL (pgx/sqlx)"| DB
    WEB -->|"HTTP GET /api/v1/*<br/>JWT Cookie"| API

    U1 -->|"HTTP :3000"| WEB
    U2 -->|"HTTP :3000"| WEB

    style API fill:#4A90D9,stroke:#333,color:#fff
    style DB fill:#336791,stroke:#333,color:#fff
    style WEB fill:#61DAFB,stroke:#333,color:#000
    style A1 fill:#00ADD8,stroke:#333,color:#fff
    style A2 fill:#00ADD8,stroke:#333,color:#fff
    style AN fill:#00ADD8,stroke:#333,color:#fff
```

---

## Arquitetura Interna da API (Clean Architecture)

```mermaid
graph LR
    subgraph "Handler Layer"
        H1["inventory_handler"]
        H2["auth_handler"]
        H3["device_handler"]
        H4["health_handler"]
    end

    subgraph "Service Layer (Business Logic)"
        S1["inventory_service"]
        S2["auth_service"]
        S3["device_service"]
    end

    subgraph "Repository Layer (Data Access)"
        R1["device_repo"]
        R2["hardware_repo"]
        R3["software_repo"]
        R4["network_repo"]
        R5["user_repo"]
        R6["token_repo"]
    end

    subgraph "Infrastructure"
        DB[("PostgreSQL")]
    end

    H1 --> S1
    H2 --> S2
    H3 --> S3
    H4 -.->|"direct"| DB

    S1 --> R1
    S1 --> R2
    S1 --> R3
    S1 --> R4
    S2 --> R5
    S2 --> R6
    S3 --> R1

    R1 --> DB
    R2 --> DB
    R3 --> DB
    R4 --> DB
    R5 --> DB
    R6 --> DB

    style DB fill:#336791,stroke:#333,color:#fff
```

---

## Estrutura do Go Workspace

```mermaid
graph TD
    ROOT["inventario/"] --> GO_WORK["go.work"]
    ROOT --> SHARED["shared/"]
    ROOT --> AGENT["agent/"]
    ROOT --> SERVER["server/"]
    ROOT --> DOCKER["docker-compose.yml"]
    ROOT --> DOCS["docs/"]

    SHARED --> SH_MODELS["models/<br/>device.go<br/>hardware.go<br/>software.go"]
    SHARED --> SH_DTO["dto/<br/>inventory_request.go<br/>inventory_response.go"]

    AGENT --> AG_MAIN["main.go"]
    AGENT --> AG_COLLECTOR["collector/<br/>hardware.go<br/>software.go<br/>network.go"]
    AGENT --> AG_CLIENT["client/<br/>api_client.go"]
    AGENT --> AG_SERVICE["service/<br/>agent_service.go"]

    SERVER --> SV_MAIN["main.go"]
    SERVER --> SV_HANDLER["handler/<br/>inventory.go<br/>auth.go<br/>device.go"]
    SERVER --> SV_SERVICE["service/<br/>inventory.go<br/>auth.go"]
    SERVER --> SV_REPO["repository/<br/>device.go<br/>hardware.go"]
    SERVER --> SV_MW["middleware/<br/>auth.go<br/>logging.go<br/>cors.go"]
    SERVER --> SV_MIG["migrations/<br/>001_init.up.sql"]
```

---

## Notas

- Toda comunicação Agent → API utiliza **HTTP** na Fase 1 (ver [Gestão de Segurança](../02-desenho-de-servico/gestao-de-seguranca.md))
- Dashboard consome apenas endpoints da API (não acessa o banco diretamente)
- PostgreSQL é acessível apenas internamente via rede Docker
- O fluxo de dependência é sempre de fora para dentro: Handler → Service → Repository

---

## Referências

- [Arquitetura da Solução](../02-desenho-de-servico/arquitetura-da-solucao.md)
- [Gestão de Configuração e Ativos](../03-transicao-de-servico/gestao-de-configuracao-e-ativos.md)
