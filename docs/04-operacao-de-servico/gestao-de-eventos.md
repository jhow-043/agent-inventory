# Gestão de Eventos

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir a estratégia de monitoramento, logging e alertas do Sistema de Inventário de Ativos de TI, garantindo detecção proativa de problemas e observabilidade adequada.

---

## 2. Escopo

Eventos monitorados em todos os componentes da Fase 1: API, Agent, PostgreSQL, Dashboard e infraestrutura Docker.

---

## 3. Classificação de Eventos

| Tipo | Descrição | Ação | Exemplo |
|---|---|---|---|
| **Informativo** | Operação normal, registro para auditoria | Nenhuma | Agent enviou inventário com sucesso |
| **Alerta** | Situação que requer atenção | Investigar em horário comercial | Disco em 70%; latência P95 > 200ms |
| **Exceção** | Falha que afeta o serviço | Ação imediata | API retorna 5xx; banco desconectado |

---

## 4. Logging Estruturado

### 4.1 Biblioteca

- **Go (API + Agent):** `log/slog` (stdlib, Go 1.21+)
- **Formato:** JSON (legível por máquina, filtrável com `jq`)
- **Output:** stdout (containers Docker coletam automaticamente)

### 4.2 Campos Padrão

| Campo | Tipo | Descrição |
|---|---|---|
| `time` | string (ISO 8601) | Timestamp do evento |
| `level` | string | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `msg` | string | Mensagem do evento |
| `component` | string | `api`, `agent`, `middleware`, `repository` |
| `request_id` | string (UUID) | Identificador único da requisição |
| `device_id` | string (UUID) | ID do dispositivo (quando aplicável) |
| `method` | string | HTTP method |
| `path` | string | HTTP path |
| `status` | int | HTTP status code |
| `duration_ms` | float | Duração da requisição em milissegundos |
| `ip` | string | IP do cliente |
| `error` | string | Mensagem de erro (quando aplicável) |

### 4.3 Exemplos de Log

**Request bem-sucedido:**
```json
{
  "time": "2026-02-13T10:30:00Z",
  "level": "INFO",
  "msg": "request completed",
  "component": "middleware",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "method": "POST",
  "path": "/api/v1/inventory",
  "status": 200,
  "duration_ms": 45.2,
  "ip": "192.168.1.50",
  "device_id": "d1e2f3a4-b5c6-7890-abcd-ef1234567890"
}
```

**Erro de autenticação:**
```json
{
  "time": "2026-02-13T10:31:00Z",
  "level": "WARN",
  "msg": "authentication failed",
  "component": "middleware",
  "request_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
  "method": "POST",
  "path": "/api/v1/inventory",
  "status": 401,
  "ip": "192.168.1.51",
  "error": "invalid device token"
}
```

**Erro de banco:**
```json
{
  "time": "2026-02-13T10:32:00Z",
  "level": "ERROR",
  "msg": "database query failed",
  "component": "repository",
  "request_id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
  "error": "connection refused: dial tcp 127.0.0.1:5432: connect: connection refused"
}
```

### 4.4 Níveis de Log

| Nível | Quando Usar | Exemplo |
|---|---|---|
| **DEBUG** | Detalhes úteis para desenvolvimento | Payload recebido, queries SQL |
| **INFO** | Operações normais do sistema | Request completado, agent registrado |
| **WARN** | Situação inesperada que não afeta o serviço | Token expirado, rate limit quase no limite |
| **ERROR** | Falha que afeta funcionalidade | Banco desconectado, migration falhou |

**Configuração por ambiente:**
- Desenvolvimento: `DEBUG`
- Produção: `INFO` (default)
- Troubleshooting produção: `DEBUG` (temporário)

### 4.5 Regras de Segurança em Logs

| ❌ NUNCA logar | ✅ OK logar |
|---|---|
| Tokens de dispositivo | Device ID |
| Senhas (mesmo hash) | Username |
| JWT completo | JWT claims (sub, exp) |
| Enrollment key | Que enrollment ocorreu |
| Payloads completos em produção | Tamanho do payload |

---

## 5. Request ID Tracing

### 5.1 Implementação

```
1. Request chega na API
2. Middleware gera UUID v4 como request_id
3. request_id é passado em context.Context
4. Todos os logs da requisição incluem request_id
5. Response header X-Request-Id retorna o ID ao cliente
```

### 5.2 Benefício

Permite rastrear toda a jornada de uma requisição nos logs:

```bash
# Encontrar todos os logs de uma requisição específica
docker compose logs api | jq 'select(.request_id == "a1b2c3d4-...")'
```

---

## 6. Eventos Monitorados

### 6.1 API

| Evento | Tipo | Log Level | Ação |
|---|---|---|---|
| Request completado (2xx) | Informativo | INFO | — |
| Request com erro de cliente (4xx) | Informativo | WARN | — |
| Request com erro de servidor (5xx) | Exceção | ERROR | Investigar |
| Rate limit excedido | Alerta | WARN | Verificar se ataque ou config errada |
| Banco desconectado | Exceção | ERROR | Restart PostgreSQL |
| Migration executada | Informativo | INFO | — |
| Migration falhou | Exceção | ERROR | Corrigir e redeployar |
| Startup completo | Informativo | INFO | — |
| Shutdown graceful | Informativo | INFO | — |

### 6.2 Agent

| Evento | Tipo | Log Level | Ação |
|---|---|---|---|
| Coleta concluída | Informativo | INFO | — |
| Envio bem-sucedido | Informativo | INFO | — |
| Envio falhou (retry) | Alerta | WARN | Verificar se persistente |
| Envio falhou (max retries) | Exceção | ERROR | Verificar rede/API |
| Registro bem-sucedido (enrollment) | Informativo | INFO | — |
| Coleta WMI falhou | Alerta | WARN | Verificar permissões |
| Service started | Informativo | INFO | — |
| Service stopped | Informativo | INFO | — |

### 6.3 PostgreSQL

| Evento | Tipo | Indicador | Ação |
|---|---|---|---|
| Conexão recusada | Exceção | `pg_isready` falha | Restart container |
| Slow query (> 1s) | Alerta | `log_min_duration_statement` | Otimizar query/índice |
| Conexões no limite | Alerta | `pg_stat_activity` count | Aumentar pool |
| Disco cheio | Exceção | `df -h` | Expandir volume, limpar WAL |
| Deadlock | Exceção | Log do PostgreSQL | Investigar transações |

### 6.4 Docker / Infraestrutura

| Evento | Tipo | Indicador | Ação |
|---|---|---|---|
| Container reiniciou | Alerta | `docker ps` restart count > 0 | Verificar logs |
| OOM Kill | Exceção | `docker inspect` OOMKilled=true | Aumentar memory limit |
| CPU > 80% sustentado | Alerta | `docker stats` | Investigar causa |
| Disco > 70% | Alerta | `df -h` | Limpar/expandir |
| Disco > 85% | Exceção | `df -h` | Ação imediata |

---

## 7. Health Checks

### 7.1 Endpoints

| Endpoint | Propósito | O que verifica | Response |
|---|---|---|---|
| `GET /healthz` | Liveness | API está respondendo | `{"status": "ok"}` |
| `GET /readyz` | Readiness | API + banco conectados | `{"status": "ready", "database": "ok"}` |

### 7.2 Docker Health Checks

```yaml
# docker-compose.yml
services:
  api:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

  postgres:
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U inventory"]
      interval: 10s
      timeout: 5s
      retries: 5
```

### 7.3 Monitoramento Externo (Opcional)

Script cron no servidor para verificar health e alertar:

```bash
#!/bin/bash
# /opt/inventory/scripts/health-check.sh
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/readyz)

if [ "$RESPONSE" != "200" ]; then
  echo "$(date -Iseconds) ALERT: API readyz returned $RESPONSE" >> /var/log/inventory-health.log
  # Enviar email/notificação se configurado
fi
```

---

## 8. Consulta e Análise de Logs

### 8.1 Ferramentas (Fase 1)

| Ferramenta | Uso |
|---|---|
| `docker compose logs` | Ver logs de containers |
| `jq` | Filtrar e formatar logs JSON |
| `grep` | Busca textual simples |
| `tail -f` | Acompanhar logs em tempo real |

### 8.2 Queries Úteis

```bash
# Todos os erros da API na última hora
docker compose logs --since 1h api | jq 'select(.level == "ERROR")'

# Requests lentos (> 500ms)
docker compose logs api | jq 'select(.duration_ms > 500)'

# Logs de um dispositivo específico
docker compose logs api | jq 'select(.device_id == "d1e2f3a4-...")'

# Rastrear uma requisição
docker compose logs api | jq 'select(.request_id == "a1b2c3d4-...")'

# Contagem de status codes
docker compose logs api | jq '.status' | sort | uniq -c | sort -rn

# Taxa de erro (5xx) no último dia
docker compose logs --since 24h api | jq 'select(.status >= 500)' | wc -l
```

### 8.3 Rotação de Logs

```yaml
# docker-compose.yml
services:
  api:
    logging:
      driver: json-file
      options:
        max-size: "10m"    # Máximo 10 MB por arquivo de log
        max-file: "5"       # Manter 5 arquivos (total: 50 MB)
```

---

## 9. Evolução Futura (Fases 2+)

| Ferramenta | Propósito | Fase |
|---|---|---|
| **Prometheus** | Coleta de métricas | Fase 2 |
| **Grafana** | Dashboards de monitoramento | Fase 2 |
| **Loki** | Agregação de logs | Fase 2 |
| **Alertmanager** | Alertas automatizados | Fase 2 |
| **OpenTelemetry** | Distributed tracing | Fase 3 |

---

## 10. Referências

- [Gestão de Incidentes](gestao-de-incidentes.md)
- [Gestão de Problemas](gestao-de-problemas.md)
- [Requisitos de Nível de Serviço](../02-desenho-de-servico/requisitos-de-nivel-de-servico.md)
- [Métricas e KPIs](../05-melhoria-continua/metricas-e-kpis.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
