# Fluxo de Comunicação

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  

---

## Fluxo Principal: Agent Envia Inventário

```mermaid
sequenceDiagram
    participant Agent as Agent Windows
    participant API as API (Go/Gin)
    participant MW as Middleware
    participant SVC as Service Layer
    participant DB as PostgreSQL

    Note over Agent: Coleta WMI executada<br/>Timer: a cada 4h

    Agent->>+API: POST /api/v1/inventory<br/>Header: Authorization: Bearer <device-token><br/>Body: { hostname, hardware, disks, network, software }

    API->>+MW: Request pipeline

    MW->>MW: Generate request_id (UUID)
    MW->>MW: Log: request started
    MW->>MW: Rate limit check

    MW->>+SVC: Validate device token
    SVC->>+DB: SELECT * FROM device_tokens<br/>WHERE token_hash = SHA256(token)
    DB-->>-SVC: Token found → device_id

    SVC->>+DB: BEGIN TRANSACTION
    
    SVC->>DB: INSERT INTO devices ... ON CONFLICT (id) UPDATE<br/>(hostname, os_version, last_seen)
    
    SVC->>DB: INSERT INTO hardware ... ON CONFLICT (device_id) UPDATE<br/>(cpu, ram, motherboard, etc.)
    
    SVC->>DB: DELETE FROM disks WHERE device_id = ?
    SVC->>DB: INSERT INTO disks (device_id, ...) VALUES (...)
    
    SVC->>DB: DELETE FROM network_interfaces WHERE device_id = ?
    SVC->>DB: INSERT INTO network_interfaces (device_id, ...) VALUES (...)
    
    SVC->>DB: DELETE FROM installed_software WHERE device_id = ?
    SVC->>DB: INSERT INTO installed_software (device_id, ...) VALUES (...)
    
    SVC->>DB: COMMIT
    DB-->>-SVC: OK

    SVC-->>-MW: 200 OK

    MW->>MW: Log: request completed<br/>duration_ms, status
    MW-->>-API: Response

    API-->>-Agent: 200 OK<br/>Header: X-Request-Id: <uuid>

    Note over Agent: Próxima coleta em 4h<br/>(± jitter de 15%)
```

---

## Fluxo: Dashboard Consulta Devices

```mermaid
sequenceDiagram
    participant Browser as Browser
    participant Dashboard as Dashboard (React)
    participant API as API (Go/Gin)
    participant MW as Middleware
    participant SVC as Service Layer
    participant DB as PostgreSQL

    Browser->>+Dashboard: Acessar /devices

    Dashboard->>+API: GET /api/v1/devices<br/>Cookie: session=<JWT>

    API->>+MW: Request pipeline
    MW->>MW: Validate JWT (cookie)
    MW->>MW: Extract user claims
    MW-->>-API: Authenticated

    API->>+SVC: GetAllDevices()
    SVC->>+DB: SELECT id, hostname, serial_number,<br/>os_version, last_seen FROM devices<br/>ORDER BY hostname
    DB-->>-SVC: []Device

    SVC-->>-API: []Device

    API-->>-Dashboard: 200 OK<br/>[{ id, hostname, serial_number, ... }]

    Dashboard->>Dashboard: TanStack Query caches data

    Dashboard-->>-Browser: Render device list
```

---

## Fluxo: Dashboard — Detalhe do Device

```mermaid
sequenceDiagram
    participant Dashboard as Dashboard (React)
    participant API as API (Go/Gin)
    participant DB as PostgreSQL

    Dashboard->>+API: GET /api/v1/devices/:id<br/>Cookie: session=<JWT>

    Note over API: JWT validado pelo middleware

    API->>+DB: SELECT * FROM devices WHERE id = :id
    API->>DB: SELECT * FROM hardware WHERE device_id = :id
    API->>DB: SELECT * FROM disks WHERE device_id = :id
    API->>DB: SELECT * FROM network_interfaces WHERE device_id = :id
    API->>DB: SELECT * FROM installed_software WHERE device_id = :id
    DB-->>-API: Device + Hardware + Disks + NICs + Software

    API-->>-Dashboard: 200 OK<br/>{ device, hardware, disks, network, software }

    Dashboard->>Dashboard: Render device detail page
```

---

## Fluxo: Agent Retry com Backoff

```mermaid
sequenceDiagram
    participant Agent as Agent Windows
    participant API as API (Go/Gin)

    Agent->>+API: POST /api/v1/inventory
    API-->>-Agent: ❌ 503 Service Unavailable

    Note over Agent: Retry 1: espera 5s (± jitter)
    Agent->>+API: POST /api/v1/inventory
    API-->>-Agent: ❌ 503 Service Unavailable

    Note over Agent: Retry 2: espera 10s (± jitter)
    Agent->>+API: POST /api/v1/inventory
    API-->>-Agent: ❌ 503 Service Unavailable

    Note over Agent: Retry 3: espera 20s (± jitter)
    Agent->>+API: POST /api/v1/inventory
    API-->>-Agent: ✅ 200 OK

    Note over Agent: Backoff resetado.<br/>Próxima coleta em 4h (± 15% jitter)
```

---

## Resumo de Endpoints

| Método | Endpoint | Autenticação | Descrição |
|---|---|---|---|
| `POST` | `/api/v1/enroll` | Enrollment Key | Registrar novo agent |
| `POST` | `/api/v1/inventory` | Device Token | Enviar dados de inventário |
| `POST` | `/api/v1/auth/login` | Username/Password | Login no dashboard |
| `POST` | `/api/v1/auth/logout` | JWT Cookie | Logout do dashboard |
| `GET` | `/api/v1/devices` | JWT Cookie | Listar todos os devices |
| `GET` | `/api/v1/devices/:id` | JWT Cookie | Detalhe de um device |
| `GET` | `/healthz` | Nenhuma | Liveness check |
| `GET` | `/readyz` | Nenhuma | Readiness check |

---

## Referências

- [Arquitetura da Solução](../02-desenho-de-servico/arquitetura-da-solucao.md)
- [Fluxo de Autenticação](fluxo-de-autenticacao.md)
- [Catálogo de Serviços](../01-estrategia-de-servico/catalogo-de-servicos.md)
