# Gestão de Problemas

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Identificar e resolver as causas raiz de incidentes recorrentes ou significativos, prevenindo recorrências e mantendo um registro de erros conhecidos (KEDB).

---

## 2. Escopo

Análise de problemas relacionados a todos os componentes da Fase 1: Agent, API, Dashboard, banco de dados e infraestrutura.

---

## 3. Definições

| Termo | Definição |
|---|---|
| **Incidente** | Interrupção ou degradação do serviço (ver [Gestão de Incidentes](gestao-de-incidentes.md)) |
| **Problema** | Causa raiz desconhecida de um ou mais incidentes |
| **Known Error** | Problema com causa raiz identificada e workaround documentado |
| **KEDB** | Known Error Database — base de dados de erros conhecidos |
| **Workaround** | Solução temporária que restaura o serviço sem corrigir a causa raiz |
| **Root Cause** | Causa fundamental que, se removida, previne a recorrência |

---

## 4. Diferença entre Incidente e Problema

| Aspecto | Incidente | Problema |
|---|---|---|
| **Foco** | Restaurar o serviço rapidamente | Encontrar e eliminar a causa raiz |
| **Urgência** | Alta (SLA de resolução) | Variável (depende do impacto) |
| **Duração** | Horas | Dias a semanas |
| **Resultado** | Serviço restaurado | Causa raiz eliminada |
| **Trigger** | Algo parou de funcionar | Incidente recorrente ou de alto impacto |

---

## 5. Processo de Gestão de Problemas

### 5.1 Proativo (antes de incidentes)

```
1. Analisar tendências de logs e métricas
      │
2. Identificar padrões de falha potencial
      │
3. Registrar como problema proativo
      │
4. Investigar e implementar correção
```

### 5.2 Reativo (após incidentes)

```
1. Incidente P1/P2 resolvido → trigger automático
   OU incidente recorrente detectado (≥ 3 ocorrências)
      │
2. Registrar problema no tracker
      │
3. Análise de causa raiz (RCA)
      │
4. Documentar causa raiz + workaround
      │
5. Atualizar KEDB
      │
6. Implementar correção definitiva
      │
7. Validar que incidente não recorre
      │
8. Fechar problema
```

---

## 6. Técnicas de Análise de Causa Raiz

### 6.1 Cinco Porquês (5 Whys)

```
Exemplo:
- Por que os agents não enviam inventário?
  → Porque a API retorna 503.
- Por que a API retorna 503?
  → Porque não consegue conectar ao PostgreSQL.
- Por que não consegue conectar ao PostgreSQL?
  → Porque o container do PostgreSQL está parado.
- Por que o container está parado?
  → Porque o disco ficou cheio e o container crashou.
- Por que o disco ficou cheio?
  → Porque os logs do PostgreSQL não tinham rotação configurada.

CAUSA RAIZ: Falta de rotação de logs do PostgreSQL.
CORREÇÃO: Configurar log_rotation_age e monitorar uso de disco.
```

### 6.2 Diagrama de Ishikawa (Espinha de Peixe)

```
                     ┌───────────────────────────┐
  Pessoas ──────────→│                           │
  Processo ─────────→│  Agents não enviam        │
  Tecnologia ───────→│  inventário               │
  Infraestrutura ───→│                           │
  Configuração ─────→│                           │
                     └───────────────────────────┘

Pessoas:       - Admin alterou config errado
Processo:      - Deploy sem validação de health check
Tecnologia:    - Bug na API
Infraestrutura: - Rede instável, disco cheio
Configuração:  - Enrollment key incorreta no agent
```

### 6.3 Fault Tree Analysis (para problemas complexos)

Árvore de falhas top-down: evento topo (falha) → portas OR/AND → causas básicas.

---

## 7. Known Error Database (KEDB)

### 7.1 Erros Conhecidos

| ID | Problema | Causa Raiz | Workaround | Correção Definitiva | Status |
|---|---|---|---|---|---|
| KE-001 | Agents com last_seen atrasado após reinício da API | Backoff acumulado quando API fica indisponível por muito tempo | Reiniciar o serviço do agent nas estações | Adicionar lógica de "reset backoff" quando API volta | Planejado |
| KE-002 | Dashboard mostra 0 devices após migration | Migration executou corretamente, mas cache do TanStack Query exibe dados antigos | Ctrl+F5 (hard refresh) no browser | Invalidar cache ao detectar mudança de versão da API | Planejado |
| KE-003 | Erro 500 em POST /inventory com payload muito grande | Software list > 500 programas excede timeout de transação | Aumentar timeout temporariamente | Otimizar bulk insert com COPY ou batch | Planejado |
| KE-004 | Disco cheio no servidor para logs | WAL do PostgreSQL cresce indefinidamente | `CHECKPOINT; SELECT pg_switch_wal();` | Configurar `max_wal_size` e rotação de logs | Planejado |
| KE-005 | Agente não consegue se registrar | Enrollment key incorreta ou expirada | Verificar enrollment_key no .env do server | Endpoint de verificação de enrollment key | Planejado |

### 7.2 Template de Known Error

```markdown
## KE-XXX: [Título]

**Problema:** [Descrição do sintoma]
**Causa Raiz:** [O que causa o problema]
**Impacto:** [Quem/o que é afetado]
**Workaround:** [Solução temporária]
**Correção Definitiva:** [O que precisa ser implementado]
**Status:** Planejado / Em andamento / Resolvido
**Incidentes relacionados:** INC-YYYY-NNN, INC-YYYY-MMM
```

---

## 8. Problemas Antecipados da Fase 1

| Área | Problema Potencial | Probabilidade | Mitigação Preventiva |
|---|---|---|---|
| **HTTP** | Token interceptado em rede não segmentada | Baixa | Documentar requisito de rede segmentada |
| **Escala** | Payload grande com 500+ softwares | Média | Limitar tamanho no middleware |
| **Concorrência** | Race condition no upsert de inventário | Baixa | Transação com lock por device |
| **Disco** | Crescimento de logs não controlado | Média | Configurar rotação no Docker e PostgreSQL |
| **Agent** | WMI timeout em máquinas lentas | Média | Timeout configurável por coletor |
| **Dashboard** | Paginação com muitos devices | Média | Implementar paginação server-side |
| **Auth** | JWT secret fraco | Baixa | Validar comprimento mínimo no startup |

---

## 9. Métricas de Gestão de Problemas

| Métrica | Descrição | Meta |
|---|---|---|
| Problemas abertos | Quantidade de problemas sem resolução | < 5 simultaneamente |
| Tempo médio de resolução | Dias entre abertura e fechamento | < 30 dias |
| Problemas recorrentes | Problemas que geraram > 1 incidente | 0 (após correção) |
| KEDB entries | Quantidade de known errors documentados | Crescente (mais = melhor) |

---

## 10. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Identificação de problemas | Administrador de TI + Desenvolvedor | Análise de incidentes e tendências |
| Análise de causa raiz | Desenvolvedor | Investigação técnica (5 Whys, logs, código) |
| Manutenção da KEDB | Desenvolvedor | Documentar e atualizar known errors |
| Implementação de correções | Desenvolvedor | Código, configuração, infra |

---

## 11. Referências

- [Gestão de Incidentes](gestao-de-incidentes.md)
- [Gestão de Eventos](gestao-de-eventos.md)
- [Plano de Melhoria](../05-melhoria-continua/plano-de-melhoria.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
