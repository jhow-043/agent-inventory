# Gestão de Incidentes

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir o processo de gestão de incidentes para o Sistema de Inventário de Ativos de TI, garantindo a restauração rápida do serviço e minimizando o impacto no negócio.

---

## 2. Escopo

Todos os incidentes que afetem a disponibilidade, desempenho ou funcionalidade do sistema em produção.

---

## 3. Definições

| Termo | Definição |
|---|---|
| **Incidente** | Interrupção não planejada ou degradação de qualidade de um serviço de TI |
| **Impacto** | Quantidade de usuários/dispositivos afetados |
| **Urgência** | Velocidade necessária para resolução |
| **Prioridade** | Combinação de impacto + urgência |
| **Workaround** | Solução temporária que restaura o serviço sem corrigir a causa raiz |

---

## 4. Classificação de Severidade

| Prioridade | Severidade | Descrição | Exemplo |
|---|---|---|---|
| **P1** | Crítica | Serviço completamente indisponível | API down, banco de dados inacessível |
| **P2** | Alta | Funcionalidade principal degradada | Agents não conseguem enviar inventário; dashboard sem dados |
| **P3** | Média | Funcionalidade secundária afetada | Dashboard lento; um filtro não funciona |
| **P4** | Baixa | Impacto mínimo, cosmético | Texto incorreto na UI; log com formato errado |

### Matriz de Prioridade

| | Urgência Alta | Urgência Média | Urgência Baixa |
|---|---|---|---|
| **Impacto Alto** (todos os users/devices) | P1 | P1 | P2 |
| **Impacto Médio** (>50% dos users/devices) | P2 | P2 | P3 |
| **Impacto Baixo** (<50% dos users/devices) | P2 | P3 | P4 |

---

## 5. Tempos de Resposta e Resolução

| Prioridade | Tempo de Resposta | Tempo de Resolução | Horário |
|---|---|---|---|
| **P1** | ≤ 30 minutos | ≤ 4 horas | 24×7 |
| **P2** | ≤ 2 horas | ≤ 8 horas | Horário comercial |
| **P3** | ≤ 8 horas | ≤ 3 dias úteis | Horário comercial |
| **P4** | ≤ 24 horas | Próxima release | Horário comercial |

> **Tempo de Resposta:** Início da investigação, não resolução.

---

## 6. Fluxo de Gestão de Incidentes

```
1. DETECÇÃO
   │  ├─ Health check automático (cron/Docker)
   │  ├─ Agent reporta falha nos logs
   │  ├─ Usuário reporta problema no dashboard
   │  └─ Monitoramento (logs/métricas)
   │
2. REGISTRO
   │  └─ Criar issue no GitHub com template de incidente
   │
3. CLASSIFICAÇÃO
   │  └─ Definir severidade (P1-P4) usando a matriz
   │
4. DIAGNÓSTICO
   │  ├─ Verificar health checks: GET /healthz, GET /readyz
   │  ├─ Verificar logs: docker compose logs api
   │  ├─ Verificar banco: psql → pg_isready
   │  ├─ Verificar rede: ping, curl
   │  └─ Consultar runbooks e KEDB
   │
5. RESOLUÇÃO / WORKAROUND
   │  ├─ Aplicar workaround se disponível (restaurar serviço rápido)
   │  ├─ Aplicar fix definitivo
   │  └─ Restaurar backup se necessário
   │
6. RECUPERAÇÃO
   │  ├─ Verificar que serviço está operacional
   │  ├─ Confirmar com health checks
   │  └─ Monitorar por 30min após resolução
   │
7. FECHAMENTO
   │  ├─ Documentar resolução na issue
   │  ├─ Atualizar KEDB se novo known error
   │  └─ Post-mortem (obrigatório para P1/P2)
```

---

## 7. Template de Registro de Incidente

```markdown
## Incidente: [Título breve]

**ID:** INC-YYYY-NNN
**Data/Hora de detecção:** YYYY-MM-DD HH:MM
**Reportado por:** [Nome / Automático]
**Prioridade:** P1 / P2 / P3 / P4

### Descrição
[O que está acontecendo. Qual o impacto observado.]

### Componentes afetados
- [ ] API
- [ ] PostgreSQL
- [ ] Dashboard
- [ ] Agent
- [ ] Rede
- [ ] Servidor

### Impacto
- **Dispositivos afetados:** [Todos / Parcial / Nenhum]
- **Usuários do dashboard afetados:** [Sim / Não]
- **Coleta de inventário afetada:** [Sim / Não]

### Timeline
| Hora | Evento |
|---|---|
| HH:MM | Incidente detectado |
| HH:MM | Diagnóstico iniciado |
| HH:MM | Causa identificada |
| HH:MM | Workaround aplicado |
| HH:MM | Fix definitivo aplicado |
| HH:MM | Serviço restaurado |

### Causa raiz
[Descrição da causa raiz]

### Resolução
[O que foi feito para resolver]

### Ações de follow-up
- [ ] [Ação 1]
- [ ] [Ação 2]

### Lições aprendidas
[O que pode ser melhorado para evitar recorrência]
```

---

## 8. Incidentes Antecipados

### 8.1 Incidentes Comuns e Resolução

| Incidente | Severidade | Causa Provável | Resolução Rápida |
|---|---|---|---|
| API retorna 503 | P1/P2 | PostgreSQL down | `docker compose restart postgres` |
| Dashboard não carrega | P3 | Container web parado | `docker compose restart web` |
| Agents com last_seen atrasado | P2 | API ou rede com problema | Verificar logs API, rede |
| Login falha para todos | P2 | JWT_SECRET alterado/corrompido | Verificar .env, reiniciar API |
| Disco cheio no servidor | P1 | Logs/WAL cresceram demais | Limpar logs antigos, expandir disco |
| Container reiniciando em loop | P2 | Erro de configuração pós-deploy | `docker compose logs api --tail=100` |
| Migration falhou no startup | P2 | SQL com erro | Corrigir migration, reiniciar |
| Token de device inválido | P4 | Token revogado ou corrompido | Re-registrar o agent |

---

## 9. Post-Mortem

### 9.1 Quando Realizar

- **Obrigatório:** Todos os incidentes P1 e P2
- **Opcional:** P3 com recorrência ou aprendizado significativo

### 9.2 Template de Post-Mortem

```markdown
## Post-Mortem: [Título do Incidente]

**ID do Incidente:** INC-YYYY-NNN
**Data do Incidente:** YYYY-MM-DD
**Duração:** X horas Y minutos
**Autor:** [Nome]
**Data do Post-Mortem:** YYYY-MM-DD

### Resumo
[1-2 frases sobre o que aconteceu]

### Impacto
- Duração do impacto: X horas
- Dispositivos afetados: N
- Usuários afetados: N

### Timeline Detalhada
| Hora | Evento | Ação |
|---|---|---|
| HH:MM | ... | ... |

### Causa Raiz
[Análise detalhada — 5 Whys se necessário]

### O que funcionou bem
- [Item 1]
- [Item 2]

### O que não funcionou
- [Item 1]
- [Item 2]

### Ações Corretivas
| Ação | Responsável | Prazo | Status |
|---|---|---|---|
| [Ação 1] | [Nome] | YYYY-MM-DD | Pendente |
| [Ação 2] | [Nome] | YYYY-MM-DD | Pendente |

### Prevenção
[O que faremos para que isso não aconteça novamente]
```

---

## 10. Escalation

### 10.1 Caminhos de Escalação

| Nível | Responsável | Quando Escalar |
|---|---|---|
| **N1** | Administrador de TI | Primeiro contato. Tenta resolver usando runbooks. |
| **N2** | Desenvolvedor | Se N1 não resolve em 30min (P1) ou 2h (P2). Requer análise de código/logs. |
| **N3** | Desenvolvedor (senior/consultor) | Se N2 não resolve no prazo. Requer redesign ou mudança arquitetural. |

> Para equipe solo: N1 e N2 são a mesma pessoa. Utilizar checklist de diagnóstico priorizado.

### 10.2 Checklist de Diagnóstico (ordem priorizada)

1. `curl http://localhost:8080/healthz` → API está respondendo?
2. `curl http://localhost:8080/readyz` → Banco está conectado?
3. `docker compose ps` → Todos os containers estão running?
4. `docker compose logs --tail=50 api` → Erros nos logs?
5. `docker compose logs --tail=50 postgres` → Erros no banco?
6. `docker stats` → CPU/RAM/Disco normais?
7. `df -h` → Disco com espaço?
8. Último deploy: algo mudou? Rollback resolve?

---

## 11. Comunicação

| Audiência | P1 | P2 | P3/P4 |
|---|---|---|---|
| Gestor de TI | Imediato (telefone/mensagem) | Email/chat em 2h | — |
| Administradores de TI | Imediato | Email/chat | — |
| Usuários do dashboard | Se > 30min: aviso no sistema | — | — |

---

## 12. Referências

- [Runbooks Operacionais](runbooks-operacionais.md)
- [Gestão de Problemas](gestao-de-problemas.md)
- [Gestão de Eventos](gestao-de-eventos.md)
- [Gestão de Disponibilidade](../02-desenho-de-servico/gestao-de-disponibilidade.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
