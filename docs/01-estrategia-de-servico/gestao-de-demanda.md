# Gestão de Demanda

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Documentar a demanda atual e projetada do Sistema de Inventário de Ativos de TI, analisar padrões de utilização e planejar a capacidade necessária para atender a demanda ao longo do tempo.

---

## 2. Escopo

Este documento cobre a demanda da Fase 1 (100–500 dispositivos Windows) e projeções para fases futuras.

---

## 3. Demanda Atual

### 3.1 Volume de Dispositivos

| Métrica | Valor |
|---|---|
| **Dispositivos alvo (mínimo)** | 100 |
| **Dispositivos alvo (máximo)** | 500 |
| **Planejamento inicial** | 100–500 dispositivos Windows |
| **Crescimento anual estimado** | 10–20% |

### 3.2 Perfil dos Dispositivos

| Tipo | Proporção Estimada |
|---|---|
| Desktops (Windows 10/11) | 70% |
| Notebooks (Windows 10/11) | 25% |
| Servidores Windows | 5% |

---

## 4. Padrões de Demanda

### 4.1 Padrão de Coleta (Agent → API)

A demanda gerada pelos agentes segue um padrão **periódico e previsível**:

| Parâmetro | Valor Padrão | Configurável |
|---|---|---|
| Intervalo de envio | 4 horas | Sim (`collection_interval` no config do agent) |
| Tipo de envio | Snapshot completo | Não (delta planejado para fases futuras) |
| Tamanho médio do payload | 5–15 KB (JSON) | Varia com quantidade de software instalado |
| Janela de envio | 24×7 | Sim (possível configurar horário comercial) |

### 4.2 Cálculo de Requests por Hora

| Cenário | Dispositivos | Intervalo | Requests/hora | Requests/dia |
|---|---|---|---|---|
| Mínimo | 100 | 4h | 25 | 600 |
| Médio | 300 | 4h | 75 | 1.800 |
| Máximo | 500 | 4h | 125 | 3.000 |
| Pico (1h interval) | 500 | 1h | 500 | 12.000 |

> **Nota:** Mesmo no cenário de pico, o volume é trivial para uma API Go + PostgreSQL (<1 request/segundo).

### 4.3 Padrão de Uso do Dashboard

| Métrica | Estimativa |
|---|---|
| Usuários concorrentes | 1–5 |
| Sessões por dia | 10–30 |
| Requests por sessão (GET) | 5–20 |
| Horário de pico | 08:00–12:00 (horário comercial) |
| Requests/hora (dashboard) | 10–50 |

### 4.4 Padrão de Carga Total

```
Requests/hora (total) = Requests(agents) + Requests(dashboard)
Caso médio: 75 + 30 = ~105 requests/hora
Caso pico:  500 + 50 = ~550 requests/hora
```

---

## 5. Sazonalidade e Picos

### 5.1 Picos Previsíveis

| Evento | Impacto | Frequência |
|---|---|---|
| **Boot simultâneo (início do expediente)** | Agentes enviam inventory quase ao mesmo tempo | Diário, ~08:00 |
| **Após manutenção da API** | Agentes acumulados fazem retry simultâneo | Ocasional |
| **Inclusão em massa de novos dispositivos** | Múltiplos enrollment + first inventory | Raro |

### 5.2 Mitigação de Picos

| Estratégia | Implementação |
|---|---|
| **Jitter no intervalo** | Agent adiciona variação aleatória (±15%) ao intervalo de coleta, evitando thundering herd |
| **Backoff exponencial** | Em caso de falha, agent espaça retries (2s → 4s → 8s → ... → max 5min) |
| **Rate limiting na API** | Máximo de 10 requests/minuto por device token |
| **Connection pooling** | Pool de conexões do PostgreSQL gerenciado pela API |

---

## 6. Projeção de Crescimento

### 6.1 Projeção de 3 Anos

| Ano | Dispositivos | Requests/hora | Espaço em disco (DB) | Observação |
|---|---|---|---|---|
| Ano 1 | 100–500 | 25–125 | ~50–250 MB | Fase 1, capacidade confortável |
| Ano 2 | 200–750 | 50–190 | ~100–375 MB | Possível crescimento orgânico |
| Ano 3 | 300–1.000 | 75–250 | ~150–500 MB | Pode requerer otimizações |

### 6.2 Cálculo de Espaço em Disco

| Dado | Tamanho Estimado | Por Device | 500 Devices/ano |
|---|---|---|---|
| Registro do device | ~500 bytes | 500 B | ~250 KB |
| Hardware | ~200 bytes | 200 B | ~100 KB |
| Discos (avg 3) | ~150 bytes × 3 | 450 B | ~225 KB |
| Interfaces de rede (avg 2) | ~100 bytes × 2 | 200 B | ~100 KB |
| Software instalado (avg 100) | ~100 bytes × 100 | 10 KB | ~5 MB |
| **Total por device** | | **~11.3 KB** | **~5.7 MB** |

> **Nota:** Como snapshots são substituídos (upsert), o tamanho do banco cresce **linearmente com o número de dispositivos**, não com o número de submissões. Para 500 devices: ~5.7 MB de dados + overhead de índices ≈ **~50 MB total** (incluindo indexes, WAL, overhead).

---

## 7. Thresholds para Ação

| Métrica | Alerta (Amarelo) | Crítico (Vermelho) | Ação |
|---|---|---|---|
| Dispositivos gerenciados | 400 | 500 | Revisar capacity plan |
| Uso de disco do PostgreSQL | 70% do volume | 85% do volume | Expandir volume, limpar dados antigos |
| Tempo de resposta P95 | > 500ms | > 1s | Investigar queries lentas, add indexes |
| Taxa de falha dos agents | > 5% | > 15% | Investigar rede, verificar API health |
| Conexões simultâneas DB | > 80% do pool | > 95% do pool | Aumentar pool size |

---

## 8. Pontos de Decisão para Escalabilidade

| Volume | Infraestrutura Atual | Ação Necessária |
|---|---|---|
| < 500 devices | Servidor único, Docker Compose | Nenhuma — capacidade confortável |
| 500–1.000 devices | Servidor único | Otimizar queries, aumentar conexões DB |
| 1.000–2.000 devices | Servidor único (beefy) | Considerar escala vertical (mais CPU/RAM) |
| > 2.000 devices | Limite do servidor único | Considerar separação de API e DB, load balancer |
| > 5.000 devices | Arquitetura distribuída | Requer redesign: clustering, réplica de leitura, etc. |

---

## 9. Demanda por Funcionalidades (Roadmap)

### 9.1 Demandas Fase 1 (Atual)

| Funcionalidade | Prioridade | Status |
|---|---|---|
| Coleta de inventário completo | Alta | Em desenvolvimento |
| Dashboard com lista de dispositivos | Alta | Em desenvolvimento |
| Filtros por hostname e OS | Média | Em desenvolvimento |
| Status online/offline | Média | Em desenvolvimento |

### 9.2 Demandas Identificadas para Fases Futuras

| Funcionalidade | Demanda | Prioridade Estimada | Fase |
|---|---|---|---|
| HTTPS (TLS) | Alta | Alta | Fase 2 |
| Delta sync (envio diferencial) | Média | Média | Fase 2 |
| Histórico de mudanças | Média | Média | Fase 2 |
| Agente Linux | Baixa | Baixa | Fase 2+ |
| Relatórios exportáveis (CSV/PDF) | Média | Média | Fase 2 |
| RBAC (controle de acesso granular) | Baixa | Baixa | Fase 3 |
| Monitoramento (Prometheus/Grafana) | Média | Média | Fase 2 |
| Integração com ITSM | Baixa | Baixa | Fase 3+ |

---

## 10. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Gestão de demanda | Gestor de TI | Monitorar crescimento, priorizar funcionalidades |
| Gestão de capacidade | Administrador de TI | Monitorar métricas de infraestrutura |
| Implementação | Desenvolvedor | Implementar otimizações e novas funcionalidades |

---

## 11. Referências

- [Visão Geral do Serviço](visao-geral-do-servico.md)
- [Gestão de Capacidade](../02-desenho-de-servico/gestao-de-capacidade.md)
- [Plano de Melhoria](../05-melhoria-continua/plano-de-melhoria.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
