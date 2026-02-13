# Cumprimento de Requisições

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir o catálogo de service requests do Sistema de Inventário de Ativos de TI e os procedimentos padronizados para cumprimento de cada requisição.

---

## 2. Escopo

Requisições de serviço relacionadas à operação, configuração e administração do sistema na Fase 1.

---

## 3. Definições

| Termo | Definição |
|---|---|
| **Service Request** | Solicitação formal para uma ação pré-definida (não é incidente nem mudança) |
| **Catálogo de Requisições** | Lista de todas as requisições disponíveis e seus procedimentos |
| **SLA de Atendimento** | Tempo máximo para cumprir a requisição |

---

## 4. Catálogo de Requisições

### SR-001: Registrar Novo Agent (Gerar Enrollment Key)

| Atributo | Valor |
|---|---|
| **ID** | SR-001 |
| **Nome** | Registrar novo agent Windows |
| **Descrição** | Instalar e registrar o agent de inventário em uma nova estação Windows |
| **Solicitante** | Administrador de TI |
| **Aprovação** | Não necessária (pré-aprovada) |
| **SLA** | 4 horas (horário comercial) |
| **Tipo** | Padrão |

**Procedimento:**

1. Verificar que a enrollment key está configurada no `.env` do servidor
2. Copiar `agent.exe` e `agent-config.yaml` para a estação Windows
3. Configurar o `agent-config.yaml` com a URL da API e enrollment key
4. Instalar como serviço: `agent.exe install`
5. Iniciar o serviço: `agent.exe start`
6. Verificar no dashboard que o device apareceu
7. Confirmar: `Get-Service InventoryAgent` → Status: Running

**Referência:** [Runbook RB-002](runbooks-operacionais.md)

---

### SR-002: Revogar Token de Dispositivo

| Atributo | Valor |
|---|---|
| **ID** | SR-002 |
| **Nome** | Revogar token de dispositivo |
| **Descrição** | Invalidar o token de um dispositivo específico (ex: máquina aposentada, comprometida) |
| **Solicitante** | Administrador de TI |
| **Aprovação** | Não necessária |
| **SLA** | 2 horas |
| **Tipo** | Padrão |

**Procedimento:**

1. Identificar o device_id no dashboard (página de detalhes do device)
2. Conectar ao banco:
   ```bash
   docker exec -it inventory-postgres psql -U inventory -d inventory
   ```
3. Revogar o token:
   ```sql
   DELETE FROM device_tokens WHERE device_id = '<device-id>';
   ```
4. Opcionalmente, remover o device:
   ```sql
   -- Remove dados relacionados e o device
   DELETE FROM installed_software WHERE device_id = '<device-id>';
   DELETE FROM network_interfaces WHERE device_id = '<device-id>';
   DELETE FROM disks WHERE device_id = '<device-id>';
   DELETE FROM hardware WHERE device_id = '<device-id>';
   DELETE FROM devices WHERE id = '<device-id>';
   ```
5. O agent na estação receberá 401 e parará de enviar
6. Desinstalar o agent da estação (se aplicável): `agent.exe stop && agent.exe uninstall`

---

### SR-003: Criar Usuário do Dashboard

| Atributo | Valor |
|---|---|
| **ID** | SR-003 |
| **Nome** | Criar novo usuário do dashboard |
| **Descrição** | Criar conta de acesso ao dashboard web |
| **Solicitante** | Gestor de TI |
| **Aprovação** | Gestor de TI |
| **SLA** | 4 horas |
| **Tipo** | Padrão |

**Procedimento:**

1. Receber solicitação com username desejado
2. Executar o comando de seed/criação:
   ```bash
   docker compose exec api ./server create-user --username <user> --password <senha>
   ```
3. Informar credenciais ao solicitante de forma segura
4. Solicitar que o usuário altere a senha no primeiro login (quando implementado)

---

### SR-004: Remover Usuário do Dashboard

| Atributo | Valor |
|---|---|
| **ID** | SR-004 |
| **Nome** | Remover usuário do dashboard |
| **Descrição** | Desativar ou remover conta de acesso ao dashboard |
| **Solicitante** | Gestor de TI |
| **Aprovação** | Gestor de TI |
| **SLA** | 2 horas |
| **Tipo** | Padrão |

**Procedimento:**

1. Conectar ao banco:
   ```bash
   docker exec -it inventory-postgres psql -U inventory -d inventory
   ```
2. Remover o usuário:
   ```sql
   DELETE FROM users WHERE username = '<username>';
   ```
3. Sessões ativas expirarão naturalmente (JWT TTL)

---

### SR-005: Alterar Intervalo de Coleta

| Atributo | Valor |
|---|---|
| **ID** | SR-005 |
| **Nome** | Alterar frequência de coleta de inventário |
| **Descrição** | Modificar o intervalo de envio de inventário dos agents |
| **Solicitante** | Gestor de TI |
| **Aprovação** | Não necessária |
| **SLA** | 8 horas |
| **Tipo** | Padrão |

**Procedimento:**

1. Definir novo intervalo desejado (ex: `2h`, `6h`, `12h`)
2. Alterar `collection_interval` no `agent-config.yaml` de cada estação
3. Reiniciar o serviço do agent em cada estação:
   ```powershell
   Restart-Service InventoryAgent
   ```
4. Verificar no log do agent que o novo intervalo foi aplicado

> **Nota:** Em fases futuras, o intervalo poderá ser controlado centralmente via API.

---

### SR-006: Solicitar Relatório de Inventário

| Atributo | Valor |
|---|---|
| **ID** | SR-006 |
| **Nome** | Gerar relatório de inventário |
| **Descrição** | Extrair dados de inventário para auditoria ou análise |
| **Solicitante** | Gestor de TI, Auditoria |
| **Aprovação** | Não necessária |
| **SLA** | 24 horas |
| **Tipo** | Padrão |

**Procedimento (Fase 1 — manual):**

1. Conectar ao banco:
   ```bash
   docker exec -it inventory-postgres psql -U inventory -d inventory
   ```
2. Executar query conforme necessidade:
   ```sql
   -- Lista completa de devices
   SELECT hostname, serial_number, os_version, last_seen FROM devices ORDER BY hostname;

   -- Software instalado em todos os devices
   SELECT d.hostname, s.name, s.version, s.vendor
   FROM installed_software s
   JOIN devices d ON d.id = s.device_id
   ORDER BY d.hostname, s.name;
   ```
3. Exportar resultado:
   ```bash
   docker exec -it inventory-postgres psql -U inventory -d inventory \
     -c "COPY (SELECT ...) TO STDOUT WITH CSV HEADER" > report.csv
   ```
4. Enviar arquivo ao solicitante

> **Nota:** Em fases futuras, funcionalidade de export integrada ao dashboard.

---

### SR-007: Resetar Senha de Usuário do Dashboard

| Atributo | Valor |
|---|---|
| **ID** | SR-007 |
| **Nome** | Resetar senha de acesso ao dashboard |
| **Descrição** | Redefinir a senha de um usuário que esqueceu suas credenciais |
| **Solicitante** | Qualquer usuário do dashboard |
| **Aprovação** | Gestor de TI |
| **SLA** | 4 horas |
| **Tipo** | Padrão |

**Procedimento:**

1. Confirmar identidade do solicitante
2. Executar reset via CLI:
   ```bash
   docker compose exec api ./server reset-password --username <user> --password <nova-senha>
   ```
3. Informar nova senha ao usuário de forma segura
4. Solicitar que altere imediatamente após login

---

### SR-008: Atualizar Enrollment Key

| Atributo | Valor |
|---|---|
| **ID** | SR-008 |
| **Nome** | Rotacionar enrollment key |
| **Descrição** | Gerar nova enrollment key (ex: após vazamento ou periodicamente) |
| **Solicitante** | Administrador de TI |
| **Aprovação** | Não necessária |
| **SLA** | 4 horas |
| **Tipo** | Normal |

**Procedimento:**

1. Gerar nova key aleatória:
   ```bash
   openssl rand -hex 32
   ```
2. Atualizar `ENROLLMENT_KEY` no `.env` do servidor
3. Reiniciar a API:
   ```bash
   docker compose restart api
   ```
4. Atualizar `enrollment_key` no `agent-config.yaml` de estações **que ainda não se registraram**
5. Agents já registrados **não são afetados** (usam device token, não enrollment key)

---

## 5. Resumo de SLAs

| ID | Requisição | SLA | Aprovação |
|---|---|---|---|
| SR-001 | Registrar novo agent | 4h | Pré-aprovada |
| SR-002 | Revogar token | 2h | Pré-aprovada |
| SR-003 | Criar usuário | 4h | Gestor de TI |
| SR-004 | Remover usuário | 2h | Gestor de TI |
| SR-005 | Alterar intervalo de coleta | 8h | Pré-aprovada |
| SR-006 | Relatório de inventário | 24h | Pré-aprovada |
| SR-007 | Resetar senha | 4h | Gestor de TI |
| SR-008 | Rotacionar enrollment key | 4h | Pré-aprovada |

---

## 6. Responsáveis

| Papel | Responsável | Atribuição |
|---|---|---|
| Atendimento N1 | Administrador de TI | Executar requisições pré-aprovadas |
| Atendimento N2 | Desenvolvedor | Requisições que envolvem código/SQL complexo |
| Aprovador | Gestor de TI | Aprovar requisições que exigem |

---

## 7. Referências

- [Runbooks Operacionais](runbooks-operacionais.md)
- [Gestão de Segurança](../02-desenho-de-servico/gestao-de-seguranca.md)
- [Catálogo de Serviços](../01-estrategia-de-servico/catalogo-de-servicos.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial |
