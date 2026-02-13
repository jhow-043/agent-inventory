# Runbooks Operacionais

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Fornecer procedimentos operacionais detalhados (passo a passo) para as atividades mais comuns de operação do Sistema de Inventário de Ativos de TI.

---

## 2. Escopo

Procedimentos para instalação, manutenção, atualização, backup, restore e troubleshooting de todos os componentes da Fase 1.

---

## 3. Índice de Runbooks

| ID | Runbook | Categoria |
|---|---|---|
| RB-001 | [Instalação inicial do sistema](#rb-001-instalação-inicial-do-sistema) | Instalação |
| RB-002 | [Instalação do agent em estação Windows](#rb-002-instalação-do-agent-em-estação-windows) | Instalação |
| RB-003 | [Atualização da API e Dashboard](#rb-003-atualização-da-api-e-dashboard) | Manutenção |
| RB-004 | [Atualização do agent em estações](#rb-004-atualização-do-agent-em-estações) | Manutenção |
| RB-005 | [Backup e restore do PostgreSQL](#rb-005-backup-e-restore-do-postgresql) | Backup |
| RB-006 | [Rotação de tokens de devices](#rb-006-rotação-de-tokens-de-devices) | Segurança |
| RB-007 | [Troubleshooting — Agent não envia dados](#rb-007-troubleshooting--agent-não-envia-dados) | Troubleshooting |
| RB-008 | [Troubleshooting — API retorna erros](#rb-008-troubleshooting--api-retorna-erros) | Troubleshooting |
| RB-009 | [Troubleshooting — Dashboard não carrega](#rb-009-troubleshooting--dashboard-não-carrega) | Troubleshooting |
| RB-010 | [Verificação de saúde do sistema](#rb-010-verificação-de-saúde-do-sistema) | Monitoramento |
| RB-011 | [Migração HTTP → HTTPS (futuro)](#rb-011-migração-http--https-futuro) | Evolução |

---

## RB-001: Instalação Inicial do Sistema

### Pré-requisitos
- Servidor Linux com Docker CE 24+ e Docker Compose v2+
- Mínimo: 4 vCPU, 8 GB RAM, 50 GB SSD
- Acesso de rede das estações Windows ao servidor (porta 8080)
- Acesso SSH ao servidor

### Procedimento

```bash
# 1. Conectar ao servidor
ssh admin@<server-ip>

# 2. Instalar Docker (se não instalado)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Reconectar SSH para aplicar grupo

# 3. Clonar o repositório
git clone https://github.com/<org>/inventario.git /opt/inventory
cd /opt/inventory

# 4. Criar arquivo de configuração
cp .env.example .env
nano .env
```

**Configurar no `.env`:**
```bash
DATABASE_URL=postgres://inventory:<SENHA-FORTE>@postgres:5432/inventory?sslmode=disable
JWT_SECRET=<GERAR: openssl rand -hex 32>
ENROLLMENT_KEY=<GERAR: openssl rand -hex 16>
SERVER_PORT=8080
LOG_LEVEL=info
CORS_ORIGINS=http://<server-ip>:3000
```

```bash
# 5. Subir os containers
docker compose up -d

# 6. Verificar que todos estão running
docker compose ps

# 7. Verificar health
curl http://localhost:8080/healthz
# Esperado: {"status":"ok"}

curl http://localhost:8080/readyz
# Esperado: {"status":"ready","database":"ok"}

# 8. Criar usuário admin
docker compose exec api ./server create-user --username admin --password <SENHA>

# 9. Acessar o dashboard
# Browser: http://<server-ip>:3000
# Login: admin / <SENHA>

# 10. Configurar backup automático
sudo cp scripts/backup.sh /opt/inventory/scripts/
sudo chmod +x /opt/inventory/scripts/backup.sh
sudo mkdir -p /opt/inventory/backups
echo "0 2 * * * root /opt/inventory/scripts/backup.sh" | sudo tee /etc/cron.d/inventory-backup
```

### Verificação
- [ ] `docker compose ps` — 3 containers running (api, postgres, web)
- [ ] `curl /healthz` retorna 200
- [ ] `curl /readyz` retorna 200
- [ ] Login no dashboard funciona
- [ ] Cron de backup configurado

---

## RB-002: Instalação do Agent em Estação Windows

### Pré-requisitos
- Windows 10/11 ou Windows Server
- Acesso de administrador local
- Conectividade de rede com o servidor da API (porta 8080)
- Binário `agent.exe` (da GitHub Release ou rede compartilhada)

### Procedimento

```powershell
# 1. Criar diretório do agent
New-Item -ItemType Directory -Force -Path "C:\ProgramData\InventoryAgent"

# 2. Copiar o binário
Copy-Item "\\fileserver\inventory\agent.exe" "C:\ProgramData\InventoryAgent\"
# OU: baixar da GitHub Release

# 3. Criar arquivo de configuração
@"
api_url: "http://<server-ip>:8080"
enrollment_key: "<enrollment-key>"
collection_interval: "4h"
log_level: "info"
log_file: "C:\ProgramData\InventoryAgent\agent.log"
"@ | Set-Content -Path "C:\ProgramData\InventoryAgent\agent-config.yaml" -Encoding UTF8

# 4. Instalar como serviço Windows
Set-Location "C:\ProgramData\InventoryAgent"
.\agent.exe install

# 5. Iniciar o serviço
.\agent.exe start

# 6. Verificar status
Get-Service InventoryAgent
# Esperado: Status = Running

# 7. Verificar log
Get-Content "C:\ProgramData\InventoryAgent\agent.log" -Tail 20
# Esperado: "enrollment successful" + "inventory sent successfully"

# 8. Verificar no dashboard
# Acessar dashboard → o device deve aparecer na lista
```

### Verificação
- [ ] `Get-Service InventoryAgent` mostra Running
- [ ] Log mostra enrollment e envio bem-sucedidos
- [ ] Device aparece no dashboard com dados corretos

### Rollback
```powershell
.\agent.exe stop
.\agent.exe uninstall
Remove-Item -Recurse "C:\ProgramData\InventoryAgent"
```

---

## RB-003: Atualização da API e Dashboard

### Pré-requisitos
- Backup do banco realizado (RB-005)
- Nova versão disponível (tag Git ou image Docker)

### Procedimento

```bash
# 1. Fazer backup primeiro!
/opt/inventory/scripts/backup.sh

# 2. Atualizar o código
cd /opt/inventory
git fetch --tags
git checkout v<nova-versão>

# 3. Rebuild e restart
docker compose up -d --build

# 4. Verificar health (aguardar ~30s para startup)
sleep 30
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz

# 5. Verificar logs
docker compose logs --tail=50 api

# 6. Verificar dashboard
# Acessar http://<server-ip>:3000 e confirmar versão

# 7. Monitorar por 30 minutos
# Verificar que agents estão enviando dados normalmente
```

### Rollback
```bash
# Se algo deu errado:
git checkout v<versão-anterior>
docker compose up -d --build

# Se migration causou problema: restaurar backup (RB-005)
```

---

## RB-004: Atualização do Agent em Estações

### Procedimento (por estação)

```powershell
# 1. Parar o serviço
C:\ProgramData\InventoryAgent\agent.exe stop

# 2. Fazer backup da config
Copy-Item "C:\ProgramData\InventoryAgent\agent-config.yaml" "C:\ProgramData\InventoryAgent\agent-config.yaml.bak"

# 3. Substituir o binário
Copy-Item "\\fileserver\inventory\agent-v<nova>.exe" "C:\ProgramData\InventoryAgent\agent.exe" -Force

# 4. Iniciar o serviço
C:\ProgramData\InventoryAgent\agent.exe start

# 5. Verificar
Get-Service InventoryAgent
Get-Content "C:\ProgramData\InventoryAgent\agent.log" -Tail 10
```

### Rollback
```powershell
C:\ProgramData\InventoryAgent\agent.exe stop
Copy-Item "\\fileserver\inventory\agent-v<anterior>.exe" "C:\ProgramData\InventoryAgent\agent.exe" -Force
C:\ProgramData\InventoryAgent\agent.exe start
```

---

## RB-005: Backup e Restore do PostgreSQL

### Backup Manual

```bash
# Gerar backup
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
docker exec inventory-postgres pg_dump \
  -U inventory -d inventory -Fc \
  > /opt/inventory/backups/backup_${TIMESTAMP}.dump

# Verificar integridade
docker exec inventory-postgres pg_restore \
  --list /tmp/backup_test.dump > /dev/null 2>&1 && echo "OK" || echo "CORROMPIDO"

# Listar backups
ls -lh /opt/inventory/backups/
```

### Restore

```bash
# 1. Parar API e Dashboard
docker compose stop api web

# 2. Dropar e recriar banco
docker exec -it inventory-postgres psql -U inventory -c "DROP DATABASE inventory;"
docker exec -it inventory-postgres psql -U inventory -c "CREATE DATABASE inventory;"

# 3. Copiar backup para o container
docker cp /opt/inventory/backups/backup_YYYYMMDD_HHMMSS.dump \
  inventory-postgres:/tmp/restore.dump

# 4. Restaurar
docker exec -it inventory-postgres pg_restore \
  -U inventory -d inventory -Fc /tmp/restore.dump

# 5. Reiniciar API e Dashboard
docker compose start api web

# 6. Verificar
curl http://localhost:8080/readyz
```

---

## RB-006: Rotação de Tokens de Devices

### Cenário: Revogar token de um device específico

```bash
# 1. Identificar o device
docker exec -it inventory-postgres psql -U inventory -d inventory \
  -c "SELECT id, hostname, serial_number FROM devices WHERE hostname = '<hostname>';"

# 2. Revogar token
docker exec -it inventory-postgres psql -U inventory -d inventory \
  -c "DELETE FROM device_tokens WHERE device_id = '<device-id>';"

# 3. O agent receberá 401 no próximo envio e tentará re-registrar
# Se enrollment key estiver configurada, o re-registro é automático
```

---

## RB-007: Troubleshooting — Agent Não Envia Dados

### Checklist de Diagnóstico

```
1. Agent está rodando?
   → Get-Service InventoryAgent
   → Se Stopped: agent.exe start

2. Agent está loggando?
   → Get-Content "C:\ProgramData\InventoryAgent\agent.log" -Tail 50
   → Se vazio: verificar permissões do diretório

3. Agent consegue alcançar a API?
   → Test-NetConnection -ComputerName <server-ip> -Port 8080
   → Se falha: problema de rede/firewall

4. API está respondendo?
   → Invoke-RestMethod http://<server-ip>:8080/healthz
   → Se erro: verificar servidor (RB-008)

5. Token é válido?
   → Verificar log do agent para "401" ou "invalid token"
   → Se token inválido: re-registrar (deletar token local, reiniciar agent)

6. Enrollment key está correta?
   → Comparar agent-config.yaml com ENROLLMENT_KEY no .env do servidor

7. Coleta WMI funciona?
   → Verificar log para erros de WMI
   → Executar manualmente: Get-WmiObject Win32_ComputerSystem
```

---

## RB-008: Troubleshooting — API Retorna Erros

### Checklist

```bash
# 1. API está rodando?
docker compose ps api
# Se Exited: docker compose up -d api

# 2. Health check
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
# Se readyz falha: problema de banco

# 3. Logs da API
docker compose logs --tail=100 api

# 4. PostgreSQL está rodando?
docker compose ps postgres
docker exec inventory-postgres pg_isready -U inventory

# 5. Disco cheio?
df -h
docker system df

# 6. Restart geral (se indefinido)
docker compose restart

# 7. Último deploy causou problema?
# Verificar git log; considerar rollback (RB-003)
```

---

## RB-009: Troubleshooting — Dashboard Não Carrega

### Checklist

```bash
# 1. Container web está rodando?
docker compose ps web
# Se Exited: docker compose up -d web

# 2. Porta está acessível?
curl -I http://localhost:3000

# 3. API está respondendo? (Dashboard depende da API)
curl http://localhost:8080/healthz

# 4. CORS configurado corretamente?
# Verificar CORS_ORIGINS no .env
# Deve incluir http://<server-ip>:3000

# 5. Logs do container web
docker compose logs --tail=50 web

# 6. Cache do browser?
# Ctrl+F5 (hard refresh) no navegador
# Ou: limpar cache do navegador
```

---

## RB-010: Verificação de Saúde do Sistema

### Health Check Completo

```bash
echo "=== INVENTORY SYSTEM HEALTH CHECK ==="
echo ""

# 1. Containers
echo "--- Containers ---"
docker compose ps

# 2. API Health
echo ""
echo "--- API Health ---"
curl -s http://localhost:8080/healthz | jq .
curl -s http://localhost:8080/readyz | jq .

# 3. PostgreSQL
echo ""
echo "--- PostgreSQL ---"
docker exec inventory-postgres pg_isready -U inventory
docker exec inventory-postgres psql -U inventory -d inventory \
  -c "SELECT count(*) as total_devices FROM devices;"

# 4. Devices com last_seen recente
echo ""
echo "--- Devices Ativos (últimas 8h) ---"
docker exec inventory-postgres psql -U inventory -d inventory \
  -c "SELECT count(*) FROM devices WHERE last_seen > NOW() - INTERVAL '8 hours';"

# 5. Devices com last_seen atrasado
echo ""
echo "--- Devices Inativos (>8h) ---"
docker exec inventory-postgres psql -U inventory -d inventory \
  -c "SELECT hostname, last_seen FROM devices WHERE last_seen < NOW() - INTERVAL '8 hours' ORDER BY last_seen;"

# 6. Disco
echo ""
echo "--- Disco ---"
df -h /opt/inventory

# 7. Docker stats
echo ""
echo "--- Recursos ---"
docker stats --no-stream

# 8. Último backup
echo ""
echo "--- Último Backup ---"
ls -lht /opt/inventory/backups/ | head -5

echo ""
echo "=== HEALTH CHECK COMPLETO ==="
```

---

## RB-011: Migração HTTP → HTTPS (Futuro)

### Pré-requisitos
- Certificado TLS válido (CA interna, self-signed ou Let's Encrypt)
- Decisão sobre método: reverse proxy (Nginx/Traefik) ou TLS direto na API

### Procedimento (Opção A: Reverse Proxy com Nginx)

```bash
# 1. Gerar certificados (self-signed para teste)
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
  -days 365 -nodes -subj "/CN=inventory.internal"

# 2. Adicionar Nginx ao docker-compose.yml
# (novo service com volumes para certs e nginx.conf)

# 3. Atualizar porta: Nginx escuta 443, proxy para API:8080

# 4. Atualizar .env
# CORS_ORIGINS=https://<server-ip>

# 5. Atualizar agent-config.yaml em TODAS as estações
# api_url: "https://<server-ip>"

# 6. Setar Secure=true nos cookies JWT
# (atualizar middleware de auth)

# 7. Restart
docker compose up -d

# 8. Testar
curl -k https://<server-ip>/healthz
# -k para aceitar self-signed (remover em produção)

# 9. Desabilitar HTTP (port 8080) ou redirecionar para HTTPS
```

### Verificação
- [ ] API responde em HTTPS
- [ ] Agents enviam via HTTPS com sucesso
- [ ] Dashboard funciona via HTTPS
- [ ] HTTP desabilitado ou redirecionando
- [ ] Certificado válido e não expirado

---

## 4. Responsáveis

| Papel | Responsável | Runbooks |
|---|---|---|
| Administrador de TI | Executor primário | RB-001 a RB-006, RB-010 |
| Desenvolvedor | Executor secundário / suporte | RB-007 a RB-009, RB-011 |

---

## 5. Referências

- [Gestão de Incidentes](gestao-de-incidentes.md)
- [Gestão de Eventos](gestao-de-eventos.md)
- [Gestão de Continuidade](../02-desenho-de-servico/gestao-de-continuidade.md)
- [Gestão de Liberação e Implantação](../03-transicao-de-servico/gestao-de-liberacao-e-implantacao.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
