# Gestão de Capacidade

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Garantir que a infraestrutura do Sistema de Inventário de Ativos de TI tenha capacidade suficiente para atender a demanda atual e futura, com custos otimizados.

---

## 2. Escopo

Planejamento de capacidade para Fase 1 (100–500 dispositivos Windows), com projeções para crescimento futuro.

---

## 3. Dimensionamento da Infraestrutura

### 3.1 Requisitos Mínimos do Servidor (100–500 devices)

| Recurso | Mínimo | Recomendado | Justificativa |
|---|---|---|---|
| **CPU** | 2 vCPU | 4 vCPU | API Go é eficiente; PostgreSQL se beneficia de mais cores |
| **RAM** | 4 GB | 8 GB | PostgreSQL: 2-4 GB shared_buffers; API: ~50 MB; containers: overhead |
| **Disco** | 20 GB SSD | 50 GB SSD | DB ~50 MB dados + WAL + backups + logs + images Docker |
| **Rede** | 100 Mbps | 1 Gbps | Payloads pequenos (~10 KB), volume baixo |
| **SO** | Linux com Docker | Linux com Docker | Qualquer distro com Docker CE 24+ |

### 3.2 Distribuição de Recursos por Container

| Container | CPU (limit) | RAM (limit) | Disco |
|---|---|---|---|
| **API (Go)** | 1 CPU | 512 MB | ~100 MB (binário + config) |
| **PostgreSQL** | 2 CPU | 2 GB | ~20 GB (dados + WAL + backups) |
| **Dashboard (Nginx)** | 0.5 CPU | 256 MB | ~50 MB (assets estáticos) |
| **Total** | 3.5 CPU | 2.75 GB | ~20 GB |

---

## 4. Análise de Carga

### 4.1 Carga de Rede

| Cenário | Devices | Payload | Intervalo | Tráfego/hora | Tráfego/dia |
|---|---|---|---|---|---|
| Mínimo | 100 | 10 KB | 4h | ~250 KB | ~6 MB |
| Médio | 300 | 10 KB | 4h | ~750 KB | ~18 MB |
| Máximo | 500 | 15 KB | 4h | ~1.9 MB | ~45 MB |
| Pico (1h) | 500 | 15 KB | 1h | ~7.5 MB | ~180 MB |

> O tráfego é **trivial** mesmo no pior cenário. Não há gargalo de rede.

### 4.2 Carga de CPU

| Operação | CPU por request | Requests/min (max) | CPU utilizada |
|---|---|---|---|
| POST /inventory (upsert) | ~5ms | ~2 | < 1% de 1 core |
| GET /devices (lista) | ~2ms | ~1 | < 0.1% de 1 core |
| GET /devices/:id (detalhe) | ~3ms | ~1 | < 0.1% de 1 core |
| **Total API** | | | **< 5% de 1 core** |

> A API Go ficará essencialmente ociosa na maioria do tempo.

### 4.3 Carga de Memória

| Componente | Uso Base | Uso sob Carga | Máximo |
|---|---|---|---|
| API Go | ~20 MB | ~50 MB | ~100 MB (GC pressure) |
| PostgreSQL | ~200 MB | ~500 MB | ~2 GB (shared_buffers) |
| Nginx (dashboard) | ~10 MB | ~30 MB | ~50 MB |
| Docker overhead | ~100 MB | ~150 MB | ~200 MB |
| **Total** | **~330 MB** | **~730 MB** | **~2.35 GB** |

### 4.4 Carga de Disco

| Dado | Tamanho por Device | 500 Devices |
|---|---|---|
| Tabela devices | ~500 B | ~250 KB |
| Tabela hardware | ~200 B | ~100 KB |
| Tabela disks (avg 3/device) | ~450 B | ~225 KB |
| Tabela network_interfaces (avg 2) | ~200 B | ~100 KB |
| Tabela installed_software (avg 100) | ~10 KB | ~5 MB |
| Índices | ~3 KB | ~1.5 MB |
| **Total dados** | **~11.3 KB** | **~7.2 MB** |
| PostgreSQL WAL | — | ~100 MB |
| Backups (30 dias × daily) | — | ~300 MB |
| Docker images | — | ~500 MB |
| Logs (30 dias, rotação) | — | ~200 MB |
| **Total em disco** | | **~1.1 GB** |

> Com 50 GB de disco, há headroom de **~98%** para crescimento.

---

## 5. Benchmarks de Referência

### 5.1 API Go + Gin (estimativas conservadoras)

| Métrica | Valor | Condição |
|---|---|---|
| Requests/segundo (simples) | > 10.000 | GET, sem DB |
| Requests/segundo (com DB) | > 2.000 | POST com upsert transacional |
| Latência P50 (inventory) | < 10ms | Em rede local |
| Latência P95 (inventory) | < 100ms | Em rede local, sob carga |
| Latência P99 (inventory) | < 500ms | Pico, com GC |

### 5.2 PostgreSQL (estimativas para o volume)

| Métrica | Valor | Condição |
|---|---|---|
| INSERT/segundo | > 5.000 | Batch insert, dados pequenos |
| SELECT/segundo | > 10.000 | Queries indexadas |
| Conexões simultâneas | 200 (default) | Configurável |

> O sistema opera a **< 1% da capacidade** teórica do PostgreSQL para o volume projetado.

---

## 6. Limites e Gargalos

### 6.1 Gargalos Potenciais

| Componente | Gargalo | Limite | Mitigação |
|---|---|---|---|
| **PostgreSQL** | Connection pool | Default 25 conexões | Aumentar `max_connections` |
| **PostgreSQL** | Tabela installed_software | Cresce rápido (100 linhas/device) | Índices otimizados |
| **API** | Upsert transacional | Lock por device_id | Baixo risco com 500 devices |
| **Rede** | Thundering herd | Boot simultâneo de agents | Jitter no intervalo |
| **Disco** | WAL growth | Alto write rate | Ajustar `checkpoint_completion_target` |

### 6.2 Limites Operacionais

| Recurso | Limite Suave | Limite Rígido | Ação |
|---|---|---|---|
| Dispositivos | 500 | ~2.000 | Escala vertical |
| Software por device | 500 programas | ~2.000 | Otimizar bulk insert |
| Payload size | 100 KB | 1 MB | Limitar na API |
| Requests/min/device | 10 | 30 | Rate limit |
| Conexões DB | 25 | 100 | Pool tuning |

---

## 7. Plano de Escala

### 7.1 Escala Vertical (Up)

| Fase | Dispositivos | Servidor | Ação |
|---|---|---|---|
| Atual | 100–500 | 4 vCPU, 8 GB RAM, 50 GB SSD | Setup inicial |
| Crescimento 1 | 500–1.000 | 8 vCPU, 16 GB RAM, 100 GB SSD | Upgrade de hardware |
| Crescimento 2 | 1.000–2.000 | 16 vCPU, 32 GB RAM, 200 GB SSD | Upgrade + tuning de DB |

### 7.2 Escala Horizontal (Out) — Futuro

| Dispositivos | Estratégia |
|---|---|
| > 2.000 | Separar PostgreSQL em servidor dedicado |
| > 5.000 | Réplica de leitura para dashboard; API com load balancer |
| > 10.000 | Requer redesign: sharding, filas (NATS/RabbitMQ), múltiplas instâncias |

---

## 8. Monitoramento de Capacidade

### 8.1 Métricas a Monitorar

| Métrica | Alerta (Amarelo) | Crítico (Vermelho) | Comando de Verificação |
|---|---|---|---|
| CPU do servidor | > 70% sustentado | > 90% sustentado | `docker stats` |
| RAM do servidor | > 75% | > 90% | `docker stats` |
| Disco usado | > 70% | > 85% | `df -h` |
| Conexões PostgreSQL | > 80% do pool | > 95% do pool | `SELECT count(*) FROM pg_stat_activity` |
| Tamanho do banco | > 5 GB | > 10 GB | `SELECT pg_database_size('inventory')` |
| Latência P95 | > 200ms | > 500ms | Análise de logs |

### 8.2 Frequência de Revisão

| Atividade | Frequência |
|---|---|
| Verificação de métricas | Semanal |
| Relatório de capacidade | Mensal |
| Revisão do plano de escala | Trimestral |
| Teste de carga | A cada release major |

---

## 9. Tuning Recomendado — PostgreSQL

Configurações otimizadas para o perfil de carga (write-heavy, leituras simples):

| Parâmetro | Valor Sugerido | Default | Justificativa |
|---|---|---|---|
| `shared_buffers` | 2 GB | 128 MB | 25% da RAM disponível |
| `effective_cache_size` | 6 GB | 4 GB | 75% da RAM |
| `work_mem` | 64 MB | 4 MB | Sorts e joins em memória |
| `maintenance_work_mem` | 512 MB | 64 MB | VACUUM e CREATE INDEX |
| `max_connections` | 50 | 100 | Suficiente para pool de 25 |
| `checkpoint_completion_target` | 0.9 | 0.5 | Spread de WAL writes |
| `wal_buffers` | 64 MB | -1 (auto) | Buffer de WAL |
| `random_page_cost` | 1.1 | 4.0 | SSD: acesso aleatório ≈ sequencial |

---

## 10. Referências

- [Gestão de Demanda](../01-estrategia-de-servico/gestao-de-demanda.md)
- [Gestão de Disponibilidade](gestao-de-disponibilidade.md)
- [Arquitetura da Solução](arquitetura-da-solucao.md)
- [Métricas e KPIs](../05-melhoria-continua/metricas-e-kpis.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
