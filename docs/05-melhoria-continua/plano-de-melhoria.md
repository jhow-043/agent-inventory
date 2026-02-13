# Plano de Melhoria Contínua

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir o framework de melhoria contínua do Sistema de Inventário de Ativos de TI, utilizando o modelo CSI (Continual Service Improvement) da ITIL v4 para garantir evolução constante em qualidade, eficiência e valor entregue.

---

## 2. Escopo

Melhoria contínua de todos os aspectos do sistema: funcionalidades, desempenho, segurança, processos operacionais e experiência do usuário.

---

## 3. Modelo de Melhoria em 7 Passos (CSI)

```
┌─────────────────────────────────────────────┐
│  1. Identificar a estratégia de melhoria    │
│     ↓                                       │
│  2. Definir o que será medido               │
│     ↓                                       │
│  3. Coletar os dados                        │
│     ↓                                       │
│  4. Processar os dados                      │
│     ↓                                       │
│  5. Analisar informações e tendências       │
│     ↓                                       │
│  6. Apresentar e usar a informação          │
│     ↓                                       │
│  7. Implementar a melhoria                  │
└─────────────────┬───────────────────────────┘
                  │
                  └──→ Voltar ao passo 1
```

### Aplicação Prática

| Passo | Atividade | Frequência |
|---|---|---|
| 1 | Revisar objetivos do serviço vs. estado atual | Mensal |
| 2 | Definir métricas relevantes (ver [Métricas e KPIs](metricas-e-kpis.md)) | A cada ciclo |
| 3 | Coletar dados de logs, métricas, feedback | Contínuo |
| 4 | Consolidar dados em relatórios | Mensal |
| 5 | Identificar gaps, tendências e oportunidades | Mensal |
| 6 | Priorizar melhorias no backlog | Mensal |
| 7 | Implementar, testar e validar | Sprint / ciclo |

---

## 4. Registro de Melhoria Contínua (CSI Register)

### 4.1 Backlog de Melhorias

| ID | Melhoria | Categoria | Prioridade | Esforço | Fase | Status |
|---|---|---|---|---|---|---|
| CSI-001 | Migrar HTTP → HTTPS | Segurança | Alta | Médio | 2 | Planejado |
| CSI-002 | Delta sync (enviar apenas mudanças) | Desempenho | Média | Alto | 2 | Planejado |
| CSI-003 | Relatório de software com export CSV | Funcionalidade | Média | Baixo | 2 | Planejado |
| CSI-004 | Dashboard com gráficos de evolução | UX | Média | Médio | 2 | Planejado |
| CSI-005 | Notificação de agents inativos | Operação | Alta | Baixo | 2 | Planejado |
| CSI-006 | Rate limiting por IP | Segurança | Média | Baixo | 2 | Planejado |
| CSI-007 | Agent para Linux | Funcionalidade | Baixa | Alto | 3 | Planejado |
| CSI-008 | RBAC (roles no dashboard) | Segurança | Média | Médio | 3 | Planejado |
| CSI-009 | Prometheus + Grafana | Observabilidade | Média | Médio | 2 | Planejado |
| CSI-010 | Paginação server-side no dashboard | Desempenho | Média | Baixo | 2 | Planejado |
| CSI-011 | API de configuração remota do agent | Funcionalidade | Baixa | Alto | 3 | Planejado |
| CSI-012 | Coleta de informações de impressoras | Funcionalidade | Baixa | Médio | 3 | Planejado |
| CSI-013 | Histórico de mudanças de hardware | Funcionalidade | Média | Médio | 2 | Planejado |
| CSI-014 | Alertas de disco cheio via agent | Operação | Baixa | Médio | 3 | Planejado |
| CSI-015 | Scan de rede (descoberta passiva) | Funcionalidade | Baixa | Alto | 3+ | Planejado |

### 4.2 Critérios de Priorização

| Critério | Peso | Escala |
|---|---|---|
| Impacto na segurança | 5 | 1 (baixo) a 5 (crítico) |
| Solicitação de usuários | 4 | 1 (nenhuma) a 5 (frequente) |
| Custo de implementação | 3 | 1 (alto custo) a 5 (rápido/fácil) |
| Impacto operacional | 3 | 1 (baixo) a 5 (elimina trabalho manual) |
| Alinhamento estratégico | 2 | 1 (baixo) a 5 (essencial) |

**Score = Σ(Peso × Nota)**

---

## 5. Roadmap de Evolução

### Fase 2 — Consolidação (após Fase 1 estável, ~2 meses)

| Sprint | Melhorias | Objetivo |
|---|---|---|
| 2.1 | CSI-001 (HTTPS), CSI-006 (rate limiting) | Segurança |
| 2.2 | CSI-005 (alertas inativos), CSI-010 (paginação) | Operação + UX |
| 2.3 | CSI-003 (export CSV), CSI-004 (gráficos) | Funcionalidade |
| 2.4 | CSI-009 (Prometheus/Grafana), CSI-013 (histórico HW) | Observabilidade |

### Fase 3 — Expansão (3+ meses)

| Sprint | Melhorias | Objetivo |
|---|---|---|
| 3.1 | CSI-002 (delta sync) | Desempenho/Escala |
| 3.2 | CSI-008 (RBAC) | Segurança |
| 3.3 | CSI-007 (agent Linux) | Cobertura |
| 3.4 | CSI-011 (config remota), CSI-012 (impressoras) | Funcionalidade |

### Fase 4+ — Maturidade

- CSI-014 (alertas de disco)
- CSI-015 (scan de rede)
- Integração com CMDB corporativo
- API pública com documentação OpenAPI
- App mobile para consultas

---

## 6. Ciclo de Revisão

### 6.1 Revisão Mensal (30 min)

**Agenda:**
1. Revisar métricas do mês (ver [Métricas e KPIs](metricas-e-kpis.md))
2. Analisar incidentes e problemas do período
3. Coletar feedback de usuários
4. Atualizar prioridades do CSI Register
5. Planejar próximo ciclo

**Participantes:** Desenvolvedor + Gestor de TI (quando disponível)

**Output:** Atualização do CSI Register + decisão sobre próximas melhorias

### 6.2 Revisão Trimestral (1h)

**Agenda:**
1. Análise de tendências (3 meses)
2. Revisão de SLOs vs. resultados
3. Avaliação de satisfação dos usuários
4. Revisão do roadmap
5. Planejamento do próximo trimestre

### 6.3 Revisão Anual (2h)

**Agenda:**
1. Balanço geral do serviço
2. Análise de custo vs. valor
3. Definição de objetivos para o próximo ano
4. Revisão de arquitetura e stack tecnológica
5. Planejamento de investimentos (se aplicável)

---

## 7. Template de Proposta de Melhoria

```markdown
## CSI-XXX: [Título da Melhoria]

**Data:** YYYY-MM-DD
**Proponente:** [Nome]
**Categoria:** Segurança / Desempenho / Funcionalidade / UX / Operação / Observabilidade

### Problema / Oportunidade
[Descrição do problema atual ou da oportunidade identificada]

### Proposta de Solução
[O que será implementado]

### Benefício Esperado
[Quantificar quando possível: tempo ganho, risco reduzido, etc.]

### Esforço Estimado
- **Desenvolvimento:** X horas/dias
- **Testes:** X horas
- **Deploy:** X horas
- **Documentação:** X horas

### Riscos
[Potenciais riscos da implementação]

### Métricas de Sucesso
[Como saberemos que a melhoria funcionou]

### Dependências
[Se depende de outra melhoria ou recurso externo]
```

---

## 8. Fontes de Melhoria

| Fonte | Como Capturar | Frequência |
|---|---|---|
| Incidentes e problemas | Post-mortems, KEDB | Contínuo |
| Feedback de usuários | Conversas, observação | Contínuo |
| Métricas e KPIs | Análise de dados | Mensal |
| Tendências de mercado | Pesquisa | Trimestral |
| Atualizações de dependências | Dependabot, changelogs | Contínuo |
| Auditorias | Revisão de segurança | Semestral |

---

## 9. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Dono do CSI Register | Desenvolvedor | Manter e priorizar o backlog |
| Aprovador de melhorias | Gestor de TI | Aprovar melhorias de alto impacto |
| Executor | Desenvolvedor | Implementar |
| Coletor de feedback | Administrador de TI | Capturar necessidades dos usuários |

---

## 10. Referências

- [Métricas e KPIs](metricas-e-kpis.md)
- [Gestão de Problemas](../04-operacao-de-servico/gestao-de-problemas.md)
- [Visão Geral do Serviço](../01-estrategia-de-servico/visao-geral-do-servico.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
