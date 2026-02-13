# Gestão de Disponibilidade

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Garantir que o Sistema de Inventário de Ativos de TI atinja os níveis de disponibilidade definidos nos SLOs, identificando e mitigando pontos de falha.

---

## 2. Escopo

Disponibilidade de todos os componentes da Fase 1: API, banco de dados, dashboard e agentes.

---

## 3. Meta de Disponibilidade

| Componente | SLO | Downtime máximo/mês | Downtime máximo/ano |
|---|---|---|---|
| API | 99.5% | 3h 36min | 43h 48min |
| PostgreSQL | 99.9% | 43min | 8h 46min |
| Dashboard | 99.5% | 3h 36min | 43h 48min |
| Agent (coleta) | 99% | 7h 12min | 87h 36min |

---

## 4. Análise de Pontos Únicos de Falha (SPOF)

### 4.1 Mapa de SPOFs

| SPOF | Componente Afetado | Impacto | Severidade |
|---|---|---|---|
| **Servidor físico** | Todos | Sistema completamente indisponível | Crítica |
| **Container PostgreSQL** | API + Dashboard | Dados inacessíveis, API retorna 503 | Alta |
| **Container API** | Agents + Dashboard | Agents acumulam retries; dashboard sem dados | Alta |
| **Container Dashboard** | Dashboard | Visualização indisponível; coleta não afetada | Média |
| **Rede interna** | Agents | Agents não enviam dados; acumulam retries | Alta |
| **Disco do servidor** | PostgreSQL | Perda de dados em caso de falha de disco | Crítica |
| **DNS/resolução** | Agents | Agents não localizam o servidor da API | Média |

### 4.2 Mapa de Impacto

```
Servidor físico ─┬─→ PostgreSQL ─→ API ─→ Agents (retry)
                 │                    └─→ Dashboard (inacessível)
                 ├─→ Docker Engine ─→ Todos os containers
                 └─→ Disco ─→ Dados perdidos (se sem backup)
```

---

## 5. Estratégias de Mitigação

### 5.1 Para Falha do Servidor

| Estratégia | Implementação | RPO | RTO |
|---|---|---|---|
| Backup diário do PostgreSQL | pg_dump com cron, armazenamento remoto | 24h | 2-4h |
| Documentação de rebuild | Runbook com todos os passos | — | 2h (com backup) |
| Docker Compose como IaC | `docker-compose.yml` versionado no Git | 0 (código) | 30min (apenas o compose) |

### 5.2 Para Falha de Containers

| Estratégia | Implementação | Configuração |
|---|---|---|
| **Restart automático** | Docker restart policy | `restart: unless-stopped` |
| **Health checks** | Docker HEALTHCHECK | API: `/healthz`, PostgreSQL: `pg_isready` |
| **Dependência de startup** | `depends_on` com condition | API depende de PostgreSQL ser healthy |

Exemplo de configuração Docker:

```yaml
services:
  api:
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U inventory"]
      interval: 10s
      timeout: 5s
      retries: 5
```

### 5.3 Para Falha de Rede

| Estratégia | Implementação |
|---|---|
| **Retry com backoff** | Agent retenta envio: 2s → 4s → 8s → ... → max 5min |
| **Jitter** | Variação aleatória no intervalo evita thundering herd pós-outage |
| **Dados não perdidos** | Agent mantém o snapshot em memória até conseguir enviar |
| **IP estático para API** | Evitar dependência de DNS; usar IP direto no config do agent |

### 5.4 Para Falha de Disco

| Estratégia | Implementação |
|---|---|
| **Volume Docker separado** | Dados do PostgreSQL em volume mapeado para disco dedicado |
| **Backup externo** | pg_dump copiado para outro servidor/disco/NAS |
| **Monitoramento de disco** | Alerta quando uso > 70% |

---

## 6. Comportamento de Cada Componente em Falha

### 6.1 Agent — API Indisponível

```
1. Agent tenta enviar inventory → timeout/connection refused
2. Log de erro: "failed to send inventory, will retry"
3. Aguarda 2 segundos, tenta novamente
4. Falha: aguarda 4s, depois 8s, 16s, 32s, ..., max 5min
5. Quando API volta: envio bem-sucedido, reset do backoff
6. Se agent reinicia: tenta imediatamente
```

- **Dados perdidos:** Nenhum (próximo envio é snapshot completo)
- **Impacto:** Dados ficam desatualizados até a API voltar
- **Resolução automática:** Sim

### 6.2 API — PostgreSQL Indisponível

```
1. Request chega na API
2. API tenta query no PostgreSQL → connection error
3. API retorna HTTP 503 Service Unavailable
4. Log de erro: "database connection failed"
5. Próximo request: tenta novamente (connection pool reconecta)
6. Quando PostgreSQL volta: requests voltam ao normal
```

- **Dados perdidos:** Nenhum (agents vão reenviar)
- **Impacto:** API retorna 503; dashboard mostra erro; agents fazem retry
- **Resolução automática:** Sim (com restart policy do Docker)

### 6.3 Dashboard — API Indisponível

```
1. Usuário acessa o dashboard
2. Dashboard tenta GET /api/v1/devices → timeout/error
3. Dashboard exibe mensagem: "Não foi possível carregar os dados"
4. Retry automático pelo TanStack Query (3 tentativas)
5. Quando API volta: dados carregam normalmente
```

- **Dados perdidos:** Nenhum
- **Impacto:** Apenas visual — dashboard sem dados
- **Resolução automática:** Sim

---

## 7. Janela de Manutenção

### 7.1 Manutenção Planejada

| Tipo | Janela Preferencial | Notificação |
|---|---|---|
| Atualização da API | Fora do horário comercial (22:00–06:00) | 24h de antecedência |
| Atualização do banco (migration) | Fora do horário comercial | 48h de antecedência |
| Atualização do agent | Horário comercial (para validar imediatamente) | 24h de antecedência |
| Manutenção do servidor | Fora do horário comercial | 48h de antecedência |
| Backup manual (se necessário) | Qualquer momento | Sem notificação |

### 7.2 Impacto da Manutenção

| Atividade | Downtime Esperado | Impacto |
|---|---|---|
| `docker compose pull && up -d` | ~30s–2min | Agents fazem retry; dashboard reconecta |
| Migration de schema | ~5s–30s (tabelas pequenas) | API indisponível momentaneamente |
| Restart do PostgreSQL | ~10s–30s | API retorna 503; reconecta automaticamente |
| Reboot do servidor | ~3–10min | Tudo indisponível; tudo auto-inicia |

---

## 8. Procedimento de Failover

### 8.1 Falha de Container (Automática)

```
1. Container para (crash, OOM kill, etc.)
2. Docker restart policy detecta (1-15s dependendo de retries)
3. Container é reiniciado automaticamente
4. Health check confirma saúde
5. Serviço restaurado
```

**Tempo de recuperação:** ~15–60 segundos

### 8.2 Falha do Servidor (Manual)

```
1. Servidor fica inacessível
2. Administrador identifica o problema
3a. Se corrigível: reiniciar servidor → containers auto-iniciam
3b. Se hardware: provisionar novo servidor
4. Instalar Docker
5. Restaurar docker-compose.yml (do Git)
6. Restaurar backup do PostgreSQL (pg_restore)
7. docker compose up -d
8. Verificar healthz
9. Agents reconectam automaticamente
```

**Tempo de recuperação:** 30min (restart) a 4h (novo servidor + restore)

---

## 9. Testes de Disponibilidade

| Teste | Frequência | Procedimento |
|---|---|---|
| Restart de containers | Mensal | `docker compose restart` — verificar auto-recovery |
| Kill de container | Trimestral | `docker kill api` — verificar restart policy |
| Simulação de rede | Trimestral | Desconectar agent da rede — verificar retry |
| Restore de backup | Trimestral | Restaurar pg_dump em ambiente de teste |
| Reboot do servidor | Semestral | Reboot completo — verificar auto-start |

---

## 10. Referências

- [Requisitos de Nível de Serviço](requisitos-de-nivel-de-servico.md)
- [Gestão de Continuidade](gestao-de-continuidade.md)
- [Gestão de Incidentes](../04-operacao-de-servico/gestao-de-incidentes.md)
- [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
