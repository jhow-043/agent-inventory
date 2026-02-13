# Métricas e KPIs

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir as métricas e indicadores-chave de desempenho (KPIs) para medir a eficácia, eficiência e saúde do Sistema de Inventário de Ativos de TI.

---

## 2. Escopo

Métricas de serviço (operação), métricas de desenvolvimento (qualidade do software) e métricas de processo (ITIL). Fase 1 foca em coleta manual/SQL; Fase 2+ automatiza com Prometheus/Grafana.

---

## 3. Métricas de Serviço

### 3.1 Disponibilidade

| Métrica | Definição | Meta (SLO) | Coleta |
|---|---|---|---|
| **API Uptime** | % de tempo que `GET /healthz` retorna 200 | ≥ 99.5% mensal | Health check externo |
| **DB Uptime** | % de tempo que `pg_isready` retorna OK | ≥ 99.5% mensal | Health check Docker |
| **Dashboard Uptime** | % de tempo que porta 3000 responde | ≥ 99.5% mensal | Health check externo |

**Cálculo:**
$$
\text{Uptime} = \frac{\text{Tempo total} - \text{Tempo de indisponibilidade}}{\text{Tempo total}} \times 100
$$

**Fase 1 — Coleta manual:**
```sql
-- Não há tabela de uptime na Fase 1
-- Registrar manualmente em planilha quando ocorrer downtime
-- Mês de 720h (30 dias): 99.5% = máx 3.6h de downtime
```

### 3.2 Desempenho

| Métrica | Definição | Meta | Coleta |
|---|---|---|---|
| **API Latency P50** | Mediana do tempo de resposta | < 100ms | Logs (duration_ms) |
| **API Latency P95** | Percentil 95 do tempo de resposta | < 500ms | Logs (duration_ms) |
| **API Latency P99** | Percentil 99 do tempo de resposta | < 1000ms | Logs (duration_ms) |

**Fase 1 — Coleta via logs:**
```bash
# P50, P95, P99 das últimas 24h
docker compose logs --since 24h api | \
  jq '.duration_ms' | sort -n | \
  awk '{all[NR]=$1} END{
    print "P50:", all[int(NR*0.5)];
    print "P95:", all[int(NR*0.95)];
    print "P99:", all[int(NR*0.99)]
  }'
```

### 3.3 Taxa de Erro

| Métrica | Definição | Meta | Coleta |
|---|---|---|---|
| **Error Rate (5xx)** | % de requests com status 500-599 | < 1% | Logs |
| **Auth Failure Rate** | % de requests com 401/403 | < 5% | Logs |

**Fase 1 — Coleta:**
```bash
# Taxa de erro 5xx nas últimas 24h
TOTAL=$(docker compose logs --since 24h api | jq '.status' | wc -l)
ERRORS=$(docker compose logs --since 24h api | jq 'select(.status >= 500)' | wc -l)
echo "Error rate: $(echo "scale=2; $ERRORS * 100 / $TOTAL" | bc)%"
```

### 3.4 Cobertura de Inventário

| Métrica | Definição | Meta | Coleta |
|---|---|---|---|
| **Device Coverage** | % de dispositivos conhecidos com agent instalado | 100% | SQL |
| **Data Freshness** | % de devices com last_seen < 8h | ≥ 95% | SQL |
| **Stale Devices** | Devices com last_seen > 24h | 0 (investigar) | SQL |

**Coleta SQL:**
```sql
-- Freshness: devices ativos nas últimas 8h
SELECT
  count(*) FILTER (WHERE last_seen > NOW() - INTERVAL '8 hours') AS active,
  count(*) AS total,
  round(
    count(*) FILTER (WHERE last_seen > NOW() - INTERVAL '8 hours')::numeric /
    NULLIF(count(*), 0) * 100, 1
  ) AS freshness_pct
FROM devices;

-- Stale devices (>24h)
SELECT hostname, last_seen,
  EXTRACT(EPOCH FROM NOW() - last_seen) / 3600 AS hours_since_last_seen
FROM devices
WHERE last_seen < NOW() - INTERVAL '24 hours'
ORDER BY last_seen;
```

### 3.5 Volume de Dados

| Métrica | Definição | Threshold | Coleta |
|---|---|---|---|
| **Total Devices** | Quantidade de devices registrados | — (informativo) | SQL |
| **Total Software** | Registros em installed_software | — | SQL |
| **DB Size** | Tamanho total do banco | Alertar > 10 GB | SQL |
| **Requests/hora** | Volume de requisições | — | Logs |

```sql
-- Estatísticas gerais
SELECT
  (SELECT count(*) FROM devices) AS total_devices,
  (SELECT count(*) FROM installed_software) AS total_software,
  (SELECT count(*) FROM device_tokens) AS total_tokens,
  pg_size_pretty(pg_database_size('inventory')) AS db_size;
```

---

## 4. Métricas de Desenvolvimento

### 4.1 Qualidade de Código

| Métrica | Definição | Meta | Ferramenta |
|---|---|---|---|
| **Test Coverage (Backend)** | % de linhas cobertas por testes | ≥ 80% (service), ≥ 70% (repo) | `go test -cover` |
| **Test Coverage (Frontend)** | % de linhas cobertas por testes | ≥ 70% | Vitest |
| **Lint Warnings** | Warnings de linter | 0 | golangci-lint, ESLint |
| **Build Success Rate** | % de builds CI que passam | ≥ 95% | GitHub Actions |

### 4.2 Velocidade de Entrega

| Métrica | Definição | Meta | Coleta |
|---|---|---|---|
| **Lead Time** | Tempo do commit ao deploy em produção | < 1 dia | GitHub Actions |
| **Deploy Frequency** | Deploys por semana | ≥ 1 (quando em desenvolvimento ativo) | Git tags |
| **MTTR (change)** | Mean time to recover de um deploy falho | < 30 min | Runbook RB-003 |
| **Change Failure Rate** | % de deploys que causam incidentes | < 10% | Histórico |

### 4.3 Dívida Técnica

| Métrica | Definição | Meta | Coleta |
|---|---|---|---|
| **TODO count** | Quantidade de `TODO`/`FIXME` no código | < 10 | `grep -r "TODO\|FIXME" --include="*.go" --include="*.ts"` |
| **Dependency Age** | Idade da dependência mais antiga | < 6 meses | Dependabot |
| **Known Vulnerabilities** | CVEs em dependências | 0 (críticas) | `govulncheck`, `npm audit` |

---

## 5. Métricas de Processo ITIL

| Métrica | Definição | Meta | Coleta |
|---|---|---|---|
| **Incidentes P1/mês** | Quantidade de incidentes críticos | ≤ 1 | Tracker |
| **MTTR (incidentes)** | Tempo médio de resolução de incidentes | P1: < 4h, P2: < 8h | Tracker |
| **MTBF** | Mean time between failures | > 30 dias | Tracker |
| **Problemas resolvidos/mês** | Known errors corrigidos | Crescente | KEDB |
| **Mudanças com sucesso** | % de mudanças sem rollback | ≥ 90% | Git/Tracker |
| **CSI items implementados** | Melhorias entregues no trimestre | ≥ 3 | CSI Register |
| **SLO compliance** | % dos SLOs cumpridos no mês | 100% | Medições |

---

## 6. Dashboard de Métricas (Fase 1)

### 6.1 Relatório Mensal — Template

```markdown
## Relatório Mensal de Métricas — [Mês/Ano]

### Disponibilidade
| Serviço | Uptime | Meta | Status |
|---|---|---|---|
| API | XX.X% | 99.5% | ✅/❌ |
| PostgreSQL | XX.X% | 99.5% | ✅/❌ |
| Dashboard | XX.X% | 99.5% | ✅/❌ |

### Desempenho
| Métrica | Valor | Meta | Status |
|---|---|---|---|
| API P50 | XXms | <100ms | ✅/❌ |
| API P95 | XXms | <500ms | ✅/❌ |
| Error Rate | X.X% | <1% | ✅/❌ |

### Cobertura
| Métrica | Valor | Meta | Status |
|---|---|---|---|
| Devices ativos | XXX | 100% | ✅/❌ |
| Data freshness | XX% | ≥95% | ✅/❌ |
| Stale devices | X | 0 | ✅/❌ |

### Incidentes
| Severidade | Quantidade | MTTR |
|---|---|---|
| P1 | X | Xh |
| P2 | X | Xh |
| P3 | X | Xd |

### Desenvolvimento
| Métrica | Valor |
|---|---|
| Deploys no mês | X |
| Build success rate | X% |
| Test coverage (Go) | X% |
| Test coverage (React) | X% |

### Ações
- [ ] [Ação 1 para o próximo mês]
- [ ] [Ação 2 para o próximo mês]
```

### 6.2 Script de Coleta Automatizada

```bash
#!/bin/bash
# /opt/inventory/scripts/monthly-metrics.sh
# Gera relatório básico de métricas

echo "=== RELATÓRIO DE MÉTRICAS ==="
echo "Data: $(date -I)"
echo ""

echo "--- COBERTURA ---"
docker exec inventory-postgres psql -U inventory -d inventory -c "
SELECT
  count(*) AS total_devices,
  count(*) FILTER (WHERE last_seen > NOW() - INTERVAL '8 hours') AS active_8h,
  count(*) FILTER (WHERE last_seen < NOW() - INTERVAL '24 hours') AS stale_24h,
  round(
    count(*) FILTER (WHERE last_seen > NOW() - INTERVAL '8 hours')::numeric /
    NULLIF(count(*), 0) * 100, 1
  ) AS freshness_pct
FROM devices;
"

echo ""
echo "--- VOLUME ---"
docker exec inventory-postgres psql -U inventory -d inventory -c "
SELECT
  (SELECT count(*) FROM devices) AS devices,
  (SELECT count(*) FROM installed_software) AS software_records,
  pg_size_pretty(pg_database_size('inventory')) AS db_size;
"

echo ""
echo "--- ERROS (últimas 24h) ---"
TOTAL=$(docker compose logs --since 24h api 2>/dev/null | jq -r '.status' 2>/dev/null | wc -l)
ERRORS=$(docker compose logs --since 24h api 2>/dev/null | jq 'select(.status >= 500)' 2>/dev/null | wc -l)
echo "Total requests: $TOTAL"
echo "5xx errors: $ERRORS"
if [ "$TOTAL" -gt 0 ]; then
  echo "Error rate: $(echo "scale=2; $ERRORS * 100 / $TOTAL" | bc)%"
fi

echo ""
echo "--- BACKUP ---"
ls -lht /opt/inventory/backups/ 2>/dev/null | head -3

echo ""
echo "=== FIM DO RELATÓRIO ==="
```

---

## 7. Thresholds e Alertas

| Métrica | Verde (OK) | Amarelo (Alerta) | Vermelho (Ação) |
|---|---|---|---|
| API Uptime | ≥ 99.5% | 99.0% - 99.5% | < 99.0% |
| API P95 Latency | < 500ms | 500ms - 1000ms | > 1000ms |
| Error Rate (5xx) | < 1% | 1% - 5% | > 5% |
| Data Freshness | ≥ 95% | 85% - 95% | < 85% |
| Disco | < 70% | 70% - 85% | > 85% |
| DB Size | < 5 GB | 5 - 10 GB | > 10 GB |
| Stale Devices | 0 | 1-3 | > 3 |

---

## 8. Evolução (Fases 2+)

### Fase 2: Prometheus + Grafana

```yaml
# Métricas exportadas pela API (endpoint /metrics)
# inventory_api_requests_total{method, path, status}
# inventory_api_request_duration_seconds{method, path}
# inventory_devices_total
# inventory_devices_active (last_seen < 8h)
# inventory_db_connections_active
# inventory_agent_collections_total{status}
```

**Dashboards Grafana planejados:**
1. **Overview:** Devices ativos, uptime, error rate
2. **API Performance:** Latency, throughput, status codes
3. **Devices:** Freshness, stale devices, top software
4. **Infrastructure:** CPU, RAM, disco, containers

---

## 9. Referências

- [Requisitos de Nível de Serviço](../02-desenho-de-servico/requisitos-de-nivel-de-servico.md)
- [Plano de Melhoria](plano-de-melhoria.md)
- [Gestão de Eventos](../04-operacao-de-servico/gestao-de-eventos.md)
- [Gestão de Incidentes](../04-operacao-de-servico/gestao-de-incidentes.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
