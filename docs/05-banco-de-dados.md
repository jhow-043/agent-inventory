# Banco de Dados

## PostgreSQL 16

O sistema usa PostgreSQL com o driver `pgx` (nativo Go, sem libpq). Conexão via `sqlx` para queries nomeadas e scan automático em structs.

## Migrações

6 migrações SQL executadas automaticamente no startup da API via `golang-migrate`. Os arquivos `.sql` são embedados no binário com `embed.FS`.

| # | Arquivo | O que faz |
|---|---------|-----------|
| 001 | `001_init` | Esquema base: users, devices, device_tokens, hardware, disks, network_interfaces, installed_software + índices |
| 002 | `002_remote_tools` | Tabela remote_tools para TeamViewer/AnyDesk/RustDesk |
| 003 | `003_disk_free_space` | Adiciona drive_letter, partition_size_bytes, free_space_bytes na tabela disks |
| 004 | `004_lifecycle` | Tabelas departments e hardware_history. Adiciona status e department_id em devices |
| 005 | `005_add_user_roles` | Adiciona coluna role em users (admin/viewer), atualiza existentes para admin |
| 006 | `006_add_audit_logs` | Tabela audit_logs com 5 índices |

Cada migração tem um arquivo `.up.sql` (aplica) e `.down.sql` (reverte).

## Esquema Completo

### users

Usuários do dashboard.

```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          TEXT NOT NULL DEFAULT 'viewer',    -- migração 005
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_role_check CHECK (role IN ('admin', 'viewer'))
);

CREATE INDEX idx_users_role ON users(role);
```

- `password_hash`: bcrypt hash (custo padrão do Go)
- `role`: `admin` (acesso total) ou `viewer` (somente leitura)
- Usuários criados via CLI recebem `admin` por padrão, via API recebem `viewer`

### devices

Dispositivos Windows monitorados.

```sql
CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hostname        VARCHAR(255) NOT NULL,
    serial_number   VARCHAR(255) NOT NULL UNIQUE,
    os_name         VARCHAR(100) NOT NULL DEFAULT '',
    os_version      VARCHAR(100) NOT NULL DEFAULT '',
    os_build        VARCHAR(50) NOT NULL DEFAULT '',
    os_arch         VARCHAR(20) NOT NULL DEFAULT '',
    last_boot_time  TIMESTAMPTZ,
    logged_in_user  VARCHAR(255) NOT NULL DEFAULT '',
    agent_version   VARCHAR(50) NOT NULL DEFAULT '',
    license_status  VARCHAR(100) NOT NULL DEFAULT '',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',    -- migração 004
    department_id   UUID REFERENCES departments(id) ON DELETE SET NULL,  -- migração 004
    last_seen       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_devices_hostname   ON devices(hostname);
CREATE INDEX idx_devices_last_seen  ON devices(last_seen);
CREATE INDEX idx_devices_status     ON devices(status);
CREATE INDEX idx_devices_department ON devices(department_id);
```

- `serial_number`: UNIQUE — identifica o device (vem do Win32_BIOS)
- `status`: `active` ou `inactive` (controlado pelo admin)
- `department_id`: FK opcional para departments (SET NULL ao deletar dept)
- **Online/Offline** não é uma coluna — é calculado em runtime baseado em `last_seen`:
  - `status = 'active' AND last_seen > NOW() - INTERVAL '1 hour'` → Online
  - `status = 'active' AND last_seen <= NOW() - INTERVAL '1 hour'` → Offline

### device_tokens

Tokens de autenticação dos agents.

```sql
CREATE TABLE device_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id  UUID NOT NULL UNIQUE REFERENCES devices(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- `device_id`: UNIQUE — cada device tem no máximo 1 token ativo
- `token_hash`: SHA-256 hex do token raw (64 chars)
- O token raw nunca é armazenado no banco
- CASCADE delete: se o device for deletado, o token vai junto

### hardware

Informações de hardware (CPU, RAM, placa-mãe, BIOS). Relação 1:1 com device.

```sql
CREATE TABLE hardware (
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id                UUID NOT NULL UNIQUE REFERENCES devices(id) ON DELETE CASCADE,
    cpu_model                VARCHAR(255) NOT NULL DEFAULT '',
    cpu_cores                INTEGER NOT NULL DEFAULT 0,
    cpu_threads              INTEGER NOT NULL DEFAULT 0,
    ram_total_bytes          BIGINT NOT NULL DEFAULT 0,
    motherboard_manufacturer VARCHAR(255) NOT NULL DEFAULT '',
    motherboard_product      VARCHAR(255) NOT NULL DEFAULT '',
    motherboard_serial       VARCHAR(255) NOT NULL DEFAULT '',
    bios_vendor              VARCHAR(255) NOT NULL DEFAULT '',
    bios_version             VARCHAR(255) NOT NULL DEFAULT '',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- `device_id`: UNIQUE — relação 1:1
- `ram_total_bytes`: armazenado em bytes (ex: 17179869184 = 16 GB)

### disks

Discos físicos e partições lógicas. Relação 1:N com device.

```sql
CREATE TABLE disks (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id            UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    model                VARCHAR(255) NOT NULL DEFAULT '',
    size_bytes           BIGINT NOT NULL DEFAULT 0,
    media_type           VARCHAR(20) NOT NULL DEFAULT '',
    serial_number        VARCHAR(255) NOT NULL DEFAULT '',
    interface_type       VARCHAR(20) NOT NULL DEFAULT '',
    drive_letter         VARCHAR(5) NOT NULL DEFAULT '',       -- migração 003
    partition_size_bytes BIGINT NOT NULL DEFAULT 0,            -- migração 003
    free_space_bytes     BIGINT NOT NULL DEFAULT 0,            -- migração 003
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_disks_device_id ON disks(device_id);
```

- Discos físicos: `media_type` = HDD/SSD/Removable, sem `drive_letter`
- Partições: `media_type` = Partition, `drive_letter` = C:/D:/etc, com `partition_size_bytes` e `free_space_bytes`
- Dados são substituídos a cada coleta (DELETE + INSERT na transação)

### network_interfaces

Adaptadores de rede físicos. Relação 1:N com device.

```sql
CREATE TABLE network_interfaces (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id    UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL DEFAULT '',
    mac_address  VARCHAR(17) NOT NULL DEFAULT '',
    ipv4_address VARCHAR(15) NOT NULL DEFAULT '',
    ipv6_address VARCHAR(45) NOT NULL DEFAULT '',
    speed_mbps   BIGINT,
    is_physical  BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_network_interfaces_device_id ON network_interfaces(device_id);
```

### installed_software

Software instalado. Relação 1:N com device.

```sql
CREATE TABLE installed_software (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id    UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name         VARCHAR(500) NOT NULL,
    version      VARCHAR(100) NOT NULL DEFAULT '',
    vendor       VARCHAR(255) NOT NULL DEFAULT '',
    install_date VARCHAR(20) NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_installed_software_device_id   ON installed_software(device_id);
CREATE INDEX idx_installed_software_device_name ON installed_software(device_id, name);
```

- `name`: VARCHAR(500) para acomodar nomes longos
- `install_date`: string porque o Registry retorna formato variado (YYYYMMDD)
- Dados substituídos a cada coleta em chunks de 200 (limite de parâmetros do PostgreSQL)

### remote_tools

Ferramentas de acesso remoto detectadas. Relação 1:N com device.

```sql
CREATE TABLE remote_tools (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id  UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    tool_name  VARCHAR(100) NOT NULL DEFAULT '',
    remote_id  VARCHAR(255) NOT NULL DEFAULT '',
    version    VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_remote_tools_device_id ON remote_tools(device_id);
```

- `tool_name`: TeamViewer, AnyDesk ou RustDesk
- `remote_id`: ID de conexão remota (pode ser vazio se não encontrado)

### departments

Departamentos organizacionais.

```sql
CREATE TABLE departments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- `name`: UNIQUE — não permite nomes duplicados

### hardware_history

Snapshots de mudanças de hardware. Quando o agent reporta hardware diferente do atual, o sistema salva o estado anterior como snapshot JSONB.

```sql
CREATE TABLE hardware_history (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id  UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    snapshot   JSONB NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hw_history_device ON hardware_history(device_id);
```

- `snapshot`: JSON completo do hardware anterior (CPU, RAM, mobo, BIOS)
- Exemplo de snapshot:
  ```json
  {
    "cpu_model": "Intel Core i5-10400",
    "cpu_cores": 6,
    "cpu_threads": 12,
    "ram_total_bytes": 8589934592,
    "motherboard_manufacturer": "ASRock",
    "motherboard_product": "B460M Pro4"
  }
  ```

### audit_logs

Log de auditoria de todas as ações administrativas.

```sql
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY,
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    username      TEXT NOT NULL,
    action        TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id   UUID,
    details       JSONB,
    ip_address    TEXT,
    user_agent    TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id    ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action     ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource   ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_composite  ON audit_logs(user_id, created_at DESC);
```

- `user_id`: SET NULL — se o usuário for deletado, o log preserva o username
- `details`: JSONB com detalhes específicos da ação
- 5 índices para permitir consultas eficientes por diferentes critérios
- Exemplo de audit log:
  ```json
  {
    "action": "device.status.update",
    "resource_type": "device",
    "resource_id": "abc-123-...",
    "details": {
      "old_status": "active",
      "new_status": "inactive"
    }
  }
  ```

## Diagrama de Relações

```
users
  │
  └──< audit_logs         (user_id → SET NULL on delete)

departments
  │
  └──< devices             (department_id → SET NULL on delete)
          │
          ├──── device_tokens      (1:1, CASCADE)
          ├──── hardware           (1:1, CASCADE)
          ├────< disks              (1:N, CASCADE)
          ├────< network_interfaces (1:N, CASCADE)
          ├────< installed_software (1:N, CASCADE)
          ├────< remote_tools       (1:N, CASCADE)
          └────< hardware_history   (1:N, CASCADE)
```

Todas as tabelas filhas de `devices` usam CASCADE delete — ao deletar um device, todos os dados relacionados são removidos automaticamente.

## Índices (16 total)

| Tabela | Índice | Colunas |
|--------|--------|---------|
| devices | `idx_devices_hostname` | hostname |
| devices | `idx_devices_last_seen` | last_seen |
| devices | `idx_devices_status` | status |
| devices | `idx_devices_department` | department_id |
| disks | `idx_disks_device_id` | device_id |
| network_interfaces | `idx_network_interfaces_device_id` | device_id |
| installed_software | `idx_installed_software_device_id` | device_id |
| installed_software | `idx_installed_software_device_name` | device_id, name |
| remote_tools | `idx_remote_tools_device_id` | device_id |
| hardware_history | `idx_hw_history_device` | device_id |
| users | `idx_users_role` | role |
| audit_logs | `idx_audit_logs_user_id` | user_id |
| audit_logs | `idx_audit_logs_action` | action |
| audit_logs | `idx_audit_logs_created_at` | created_at DESC |
| audit_logs | `idx_audit_logs_resource` | resource_type, resource_id |
| audit_logs | `idx_audit_logs_composite` | user_id, created_at DESC |

## Estratégia de Atualização de Dados

A cada ciclo do agent, os dados são atualizados assim:

| Tabela | Estratégia | Motivo |
|--------|-----------|--------|
| devices | UPSERT (ON CONFLICT) | Atualiza campos mutáveis (hostname, OS, last_seen) |
| hardware | UPSERT (ON CONFLICT device_id) | Atualiza dados de hardware |
| hardware_history | INSERT condicional | Só insere se hardware mudou (comparação dos campos) |
| disks | DELETE + INSERT | Partições e discos podem mudar completamente |
| network_interfaces | DELETE + INSERT | IPs e adaptadores podem mudar |
| installed_software | DELETE + INSERT (chunks de 200) | Lista completa substituída a cada ciclo |
| remote_tools | DELETE + INSERT | Ferramentas podem ser instaladas/removidas |

Tudo dentro de uma transação única — se qualquer passo falhar, faz rollback completo.
