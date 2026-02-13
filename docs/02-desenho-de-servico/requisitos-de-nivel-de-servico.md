# Requisitos de Nível de Serviço

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir os indicadores de nível de serviço (SLI), objetivos de nível de serviço (SLO) e acordos de nível de serviço (SLA) para o Sistema de Inventário de Ativos de TI.

---

## 2. Escopo

SLIs, SLOs e SLAs para a Fase 1, aplicáveis ao ambiente on-premises com 100–500 dispositivos.

---

## 3. Definições

| Termo | Definição |
|---|---|
| **SLI** (Service Level Indicator) | Métrica quantitativa que mede um aspecto específico do serviço |
| **SLO** (Service Level Objective) | Meta interna para cada SLI, definindo o nível desejado de qualidade |
| **SLA** (Service Level Agreement) | Acordo formal entre provedor e consumidor do serviço (com consequências) |

---

## 4. Service Level Indicators (SLIs)

### 4.1 SLIs da API

| ID | SLI | Descrição | Método de Medição |
|---|---|---|---|
| SLI-API-01 | **Disponibilidade** | % de tempo que a API responde com sucesso (HTTP 2xx/4xx) | `healthz` check a cada 60s |
| SLI-API-02 | **Latência (inventory)** | Tempo de resposta do `POST /api/v1/inventory` | Log de request duration |
| SLI-API-03 | **Latência (devices)** | Tempo de resposta do `GET /api/v1/devices` | Log de request duration |
| SLI-API-04 | **Taxa de erro** | % de requests que retornam HTTP 5xx | Log analysis |
| SLI-API-05 | **Throughput** | Requests processados por minuto | Log analysis |

### 4.2 SLIs do Agent

| ID | SLI | Descrição | Método de Medição |
|---|---|---|---|
| SLI-AGT-01 | **Taxa de coleta** | % de envios de inventário bem-sucedidos | Log do agent |
| SLI-AGT-02 | **Tempo de coleta** | Tempo para coletar todos os dados locais | Log do agent |
| SLI-AGT-03 | **Freshness** | Tempo desde o último `last_seen` de cada device | Query no banco |

### 4.3 SLIs do Dashboard

| ID | SLI | Descrição | Método de Medição |
|---|---|---|---|
| SLI-WEB-01 | **Disponibilidade** | % de tempo que o dashboard carrega corretamente | Health check HTTP |
| SLI-WEB-02 | **Tempo de carregamento** | Tempo até a página ficar interativa (TTI) | Medição manual/Lighthouse |

### 4.4 SLIs do Banco de Dados

| ID | SLI | Descrição | Método de Medição |
|---|---|---|---|
| SLI-DB-01 | **Disponibilidade** | % de tempo que o PostgreSQL aceita conexões | Connection check |
| SLI-DB-02 | **Latência de queries** | Tempo médio de queries (p50, p95) | pg_stat_statements |

---

## 5. Service Level Objectives (SLOs)

### 5.1 SLOs da API

| SLI | SLO | Janela | Observação |
|---|---|---|---|
| SLI-API-01 Disponibilidade | **≥ 99.5%** | Mensal | ≈ 3.6h downtime máximo/mês |
| SLI-API-02 Latência (inventory) | **p95 < 500ms** | Mensal | Inclui upsert transacional completo |
| SLI-API-03 Latência (devices) | **p95 < 200ms** | Mensal | Lista de dispositivos |
| SLI-API-04 Taxa de erro | **< 0.1%** | Mensal | Erros 5xx apenas |
| SLI-API-05 Throughput | **≥ 1000 req/min** | Capacidade | Acima da demanda máxima projetada |

### 5.2 SLOs do Agent

| SLI | SLO | Janela | Observação |
|---|---|---|---|
| SLI-AGT-01 Taxa de coleta | **≥ 99%** | Semanal | Considerando retries com backoff |
| SLI-AGT-02 Tempo de coleta | **< 30s** | Por execução | Coleta local de todos os dados |
| SLI-AGT-03 Freshness | **< 2× intervalo** | Contínuo | Se intervalo = 4h, alerta se > 8h |

### 5.3 SLOs do Dashboard

| SLI | SLO | Janela | Observação |
|---|---|---|---|
| SLI-WEB-01 Disponibilidade | **≥ 99.5%** | Mensal | Mesmo que a API |
| SLI-WEB-02 Tempo de carregamento | **< 3s** | Por acesso | Inclui TTI (Time to Interactive) |

### 5.4 SLOs do Banco de Dados

| SLI | SLO | Janela | Observação |
|---|---|---|---|
| SLI-DB-01 Disponibilidade | **≥ 99.9%** | Mensal | ≈ 43min downtime máximo/mês |
| SLI-DB-02 Latência de queries | **p95 < 100ms** | Mensal | Queries individuais |

---

## 6. Error Budget

O conceito de **Error Budget** define quanto downtime é "permitido" dentro do SLO:

| Componente | SLO | Error Budget (mensal) | Error Budget (anual) |
|---|---|---|---|
| API | 99.5% | 3h 36min | 43h 48min |
| Dashboard | 99.5% | 3h 36min | 43h 48min |
| PostgreSQL | 99.9% | 43min | 8h 46min |

### Política de Error Budget

- **Budget > 50%:** Priorizar features e melhorias
- **Budget 20–50%:** Atenção a incidentes, balancear features com estabilidade
- **Budget < 20%:** Freeze de features, foco total em confiabilidade
- **Budget esgotado:** Apenas correções de bugs e melhorias de confiabilidade

---

## 7. Service Level Agreement (SLA)

### 7.1 SLA Interno

Na Fase 1, o SLA é **interno** (não há clientes externos). Ele formaliza o compromisso da equipe de TI consigo mesma.

| Aspecto | Definição |
|---|---|
| **Tipo** | SLA interno (intra-organizacional) |
| **Serviço coberto** | Sistema de Inventário de Ativos de TI |
| **Consumidores** | Equipe de Infraestrutura, Gestão de TI, Auditoria |
| **Provedor** | Desenvolvedor + Administrador de TI |
| **Vigência** | Permanente enquanto o serviço estiver em produção |

### 7.2 Compromissos do SLA

| Compromisso | Meta | Penalidade por Descumprimento |
|---|---|---|
| Disponibilidade geral | 99.5% mensal | Análise de causa raiz obrigatória |
| Tempo de resposta para P1 (Crítico) | Início em ≤ 30min | Post-mortem obrigatório |
| Tempo de resposta para P2 (Alto) | Início em ≤ 2h | Registro no backlog de incidentes |
| Tempo de resposta para P3 (Médio) | Início em ≤ 8h | — |
| Tempo de resposta para P4 (Baixo) | Início em ≤ 24h | — |
| Tempo de resolução para P1 | ≤ 4h | Escalação e post-mortem |
| Tempo de resolução para P2 | ≤ 8h | Registro e análise |
| Janela de manutenção | Notificação com ≥ 24h de antecedência | — |
| Backup do banco | Diário, retenção 30 dias | Auditoria de backup |

---

## 8. Monitoramento e Reporte

### 8.1 Como Medir (Fase 1)

| SLI | Ferramenta | Método |
|---|---|---|
| Disponibilidade API | Script cron + curl | `curl http://api:8080/healthz` a cada 60s |
| Latência | Logs estruturados (slog) | Campo `duration_ms` em cada request log |
| Taxa de erro | Logs | Contagem de `status >= 500` |
| Freshness | Query SQL | `SELECT * FROM devices WHERE last_seen < NOW() - INTERVAL '8 hours'` |
| Disponibilidade DB | Docker health check | `pg_isready` |

### 8.2 Reporte

| Frequência | Conteúdo | Destinatário |
|---|---|---|
| Diário | Health check visual (dashboard /healthz) | Administrador de TI |
| Semanal | Dispositivos com last_seen atrasado | Administrador de TI |
| Mensal | Relatório de SLO (uptime, latência, error rate) | Gestor de TI |

---

## 9. Revisão

| Item | Frequência |
|---|---|
| Revisão dos SLOs | Trimestral |
| Revisão do SLA | A cada release major ou mudança significativa |
| Calibragem de thresholds | Após cada incidente P1/P2 |

---

## 10. Referências

- [Gestão de Disponibilidade](gestao-de-disponibilidade.md)
- [Gestão de Incidentes](../04-operacao-de-servico/gestao-de-incidentes.md)
- [Gestão de Eventos](../04-operacao-de-servico/gestao-de-eventos.md)
- [Métricas e KPIs](../05-melhoria-continua/metricas-e-kpis.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
