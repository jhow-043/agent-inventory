# Gestão de Continuidade

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir os procedimentos de continuidade e recuperação de desastres (Disaster Recovery — DR) para o Sistema de Inventário de Ativos de TI, assegurando a capacidade de restaurar o serviço em tempo aceitável.

---

## 2. Escopo

Continuidade de todos os componentes da Fase 1: API, banco de dados PostgreSQL, dashboard e agentes Windows.

---

## 3. Objetivos de Recuperação

| Métrica | Definição | Alvo |
|---|---|---|
| **RPO** (Recovery Point Objective) | Máxima perda de dados aceitável | **24 horas** (backup diário) |
| **RTO** (Recovery Time Objective) | Tempo máximo para restaurar o serviço | **4 horas** (rebuild completo) |
| **MTPD** (Maximum Tolerable Period of Disruption) | Tempo máximo tolerável sem o serviço | **48 horas** |

### Justificativa dos Objetivos

- **RPO 24h:** Inventário é reconstruído automaticamente pelos agents na próxima coleta. Mesmo com perda de 24h de dados, a recuperação é automática.
- **RTO 4h:** Tempo suficiente para provisionar servidor, restaurar backup e reiniciar containers.
- **MTPD 48h:** O inventário não é sistema crítico de negócio. A operação continua sem ele; os dados são reconstituídos após restauração.

---

## 4. Análise de Impacto no Negócio (BIA)

### 4.1 Impacto por Tempo de Indisponibilidade

| Duração | Impacto | Severidade |
|---|---|---|
| < 1 hora | Mínimo — agents fazem retry automático | Baixa |
| 1–8 horas | Dados de inventário desatualizados; dashboard inacessível | Média |
| 8–24 horas | Nenhum inventário novo; decisões adiadas | Média |
| 24–48 horas | Inventário antigo; risco em auditorias programadas | Alta |
| > 48 horas | Sem visibilidade do parque; processos manuais necessários | Alta |

### 4.2 Dados Críticos

| Dado | Criticidade | Recuperável sem Backup? |
|---|---|---|
| Registros de dispositivos | Alta | Parcial — agents recriam ao reenviar |
| Tokens de dispositivos | Alta | **Não** — agents precisariam se registrar novamente |
| Usuários do dashboard | Média | Não — precisam ser recriados |
| Snapshots de inventário | Média | Sim — agents reenviam na próxima coleta |
| Schema e migrations | Baixa | Sim — versionados no Git |

### 4.3 Conclusão BIA

O dado mais crítico é a tabela `device_tokens`. Sem ela, todos os agents perdem autenticação e precisam ser re-registrados. O backup do banco é, portanto, **essencial**, mesmo que os dados de inventário sejam reconstituíveis.

---

## 5. Cenários de Desastre

### Cenário 1: Falha de Container

| Aspecto | Detalhe |
|---|---|
| **Descrição** | Um container (API, PostgreSQL ou Dashboard) para |
| **Probabilidade** | Média |
| **Impacto** | Parcial — depende de qual container |
| **Detecção** | Docker health check (automática) |
| **Recuperação** | Automática via `restart: unless-stopped` |
| **RTO** | ~15–60 segundos |
| **Dados perdidos** | Nenhum |

### Cenário 2: Corrupção do Banco de Dados

| Aspecto | Detalhe |
|---|---|
| **Descrição** | Dados corrompidos no PostgreSQL (crash, disco, OOM kill) |
| **Probabilidade** | Baixa |
| **Impacto** | Alto — dados inacessíveis ou incorretos |
| **Detecção** | Erros de query; API retorna 500; health check falha |
| **Recuperação** | Restaurar backup pg_dump |
| **RTO** | 1–2 horas |
| **Dados perdidos** | Até 24h (RPO) |

### Cenário 3: Falha do Servidor Físico

| Aspecto | Detalhe |
|---|---|
| **Descrição** | Servidor host fica completamente inoperante |
| **Probabilidade** | Baixa |
| **Impacto** | Crítico — todos os componentes indisponíveis |
| **Detecção** | Agents não conseguem enviar; dashboard inacessível |
| **Recuperação** | Provisionar novo servidor + restaurar backup |
| **RTO** | 2–4 horas |
| **Dados perdidos** | Até 24h (RPO) |

### Cenário 4: Falha de Disco

| Aspecto | Detalhe |
|---|---|
| **Descrição** | Disco do servidor falha (SSD/HDD failure) |
| **Probabilidade** | Baixa |
| **Impacto** | Crítico — perda de dados se sem backup |
| **Detecção** | Erros de I/O; containers param |
| **Recuperação** | Trocar disco + restaurar backup em novo volume |
| **RTO** | 2–4 horas |
| **Dados perdidos** | Até 24h (RPO) |

### Cenário 5: Erro Humano (Exclusão Acidental)

| Aspecto | Detalhe |
|---|---|
| **Descrição** | DROP TABLE, DELETE sem WHERE, docker volume prune |
| **Probabilidade** | Média |
| **Impacto** | Alto — perda parcial ou total de dados |
| **Detecção** | Dashobard vazio; erros de API |
| **Recuperação** | Restaurar backup pg_dump |
| **RTO** | 30min–1h |
| **Dados perdidos** | Até 24h (RPO) |

### Cenário 6: Ransomware/Comprometimento

| Aspecto | Detalhe |
|---|---|
| **Descrição** | Servidor comprometido, dados criptografados/destruídos |
| **Probabilidade** | Baixa |
| **Impacto** | Crítico |
| **Detecção** | Serviço indisponível; arquivos inacessíveis |
| **Recuperação** | Rebuild completo em servidor limpo + backup offline |
| **RTO** | 4–8 horas |
| **Dados perdidos** | Até 24h (backup offline) |

---

## 6. Estratégia de Backup

### 6.1 Política de Backup

| Aspecto | Definição |
|---|---|
| **Método** | `pg_dump` (dump lógico completo) |
| **Frequência** | Diário, 02:00 (fora do horário comercial) |
| **Retenção** | 30 dias (últimos 30 backups) |
| **Formato** | Custom format (`-Fc`) — compactado, permite restore parcial |
| **Destino primário** | Volume separado no mesmo servidor |
| **Destino secundário** | Cópia para NAS/servidor remoto (recomendado) |
| **Verificação** | Restore automático em ambiente de teste (mensal) |

### 6.2 Script de Backup

```bash
#!/bin/bash
# /opt/inventory/scripts/backup.sh
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/inventory/backups"
RETENTION_DAYS=30

# Executar pg_dump via Docker
docker exec inventory-postgres pg_dump \
  -U inventory \
  -d inventory \
  -Fc \
  -f /tmp/backup_${TIMESTAMP}.dump

# Copiar backup do container
docker cp inventory-postgres:/tmp/backup_${TIMESTAMP}.dump \
  ${BACKUP_DIR}/backup_${TIMESTAMP}.dump

# Limpar backups antigos
find ${BACKUP_DIR} -name "backup_*.dump" -mtime +${RETENTION_DAYS} -delete

# Log
echo "$(date -Iseconds) Backup completed: backup_${TIMESTAMP}.dump" \
  >> ${BACKUP_DIR}/backup.log
```

### 6.3 Cron Job

```
# /etc/cron.d/inventory-backup
0 2 * * * root /opt/inventory/scripts/backup.sh
```

### 6.4 O que NÃO é feito backup (versionado no Git)

- Código-fonte (repositório Git)
- docker-compose.yml (repositório Git)
- Migrations SQL (repositório Git)
- Documentação (repositório Git)
- Dockerfile (repositório Git)

---

## 7. Procedimentos de Recuperação

### 7.1 Restaurar Banco a partir de Backup

```bash
# 1. Parar a API
docker compose stop api web

# 2. Conectar ao PostgreSQL e dropar o banco
docker exec -it inventory-postgres psql -U inventory -c "DROP DATABASE inventory;"
docker exec -it inventory-postgres psql -U inventory -c "CREATE DATABASE inventory;"

# 3. Restaurar o backup
docker cp /opt/inventory/backups/backup_YYYYMMDD_HHMMSS.dump \
  inventory-postgres:/tmp/restore.dump

docker exec -it inventory-postgres pg_restore \
  -U inventory \
  -d inventory \
  -Fc \
  /tmp/restore.dump

# 4. Reiniciar a API e o dashboard
docker compose start api web

# 5. Verificar health
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

### 7.2 Rebuild Completo em Novo Servidor

```bash
# 1. Instalar Docker e Docker Compose no novo servidor
curl -fsSL https://get.docker.com | sh
apt install docker-compose-plugin

# 2. Clonar o repositório
git clone https://github.com/org/inventario.git /opt/inventory
cd /opt/inventory

# 3. Configurar variáveis de ambiente
cp .env.example .env
# Editar .env com as configurações corretas

# 4. Subir os containers
docker compose up -d

# 5. Restaurar o backup do banco (se disponível)
# ... seguir procedimento 7.1

# 6. Se backup indisponível: agents vão reenviar inventário
# Porém, device_tokens foram perdidos — agents precisam se registrar novamente:
# - Atualizar enrollment_key no .env
# - Reinstalar/reconfigurar agents nas estações

# 7. Verificar health
curl http://localhost:8080/healthz
```

---

## 8. Testes de Continuidade

### 8.1 Cronograma de Testes

| Teste | Frequência | Responsável |
|---|---|---|
| Verificação de integridade do backup | Semanal (automático) | Script cron |
| Restore de backup em ambiente de teste | Mensal | Administrador de TI |
| Simulação de falha de container | Trimestral | Administrador de TI |
| Simulação de rebuild completo | Semestral | Desenvolvedor |
| Revisão do plano de DR | Anual | Gestor de TI |

### 8.2 Checklist de Teste de Restore

- [ ] Backup copiado para ambiente de teste
- [ ] pg_restore executado sem erros
- [ ] API conecta ao banco restaurado
- [ ] Dashboard carrega lista de dispositivos
- [ ] Dados estão consistentes (contagem de devices bate)
- [ ] Resultado documentado

---

## 9. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Proprietário do plano de DR | Gestor de TI | Aprovação e revisão |
| Execução de backup | Automático (cron) | Backup diário |
| Monitoramento de backups | Administrador de TI | Verificação semanal |
| Execução de restore | Administrador de TI + Desenvolvedor | Quando necessário |
| Teste de DR | Desenvolvedor | Conforme cronograma |

---

## 10. Referências

- [Gestão de Disponibilidade](gestao-de-disponibilidade.md)
- [Gestão de Segurança](gestao-de-seguranca.md)
- [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
