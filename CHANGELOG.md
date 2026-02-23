# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/),
e este projeto adere ao [Versionamento Semântico](https://semver.org/lang/pt-BR/).

## [1.2.0] - 2026-02-23

### Segurança
- **Cookie hardening** — session cookie agora usa `SameSite=Lax` e `Secure` dinâmico (ativo em HTTPS) via `http.SetCookie`
- **JWT secret mínimo enforced** — servidor recusa iniciar se `JWT_SECRET` tiver menos de 32 caracteres (antes era apenas warning)
- **Body size limit** — novo middleware `MaxBodySize` limita request body a 10 MB (previne OOM/DoS)
- **Bulk operations cap** — operações em massa limitadas a 100 itens por request (`max=100` nas binding tags)
- **Pagination cap** — `?limit=` limitado a 200 em todos os endpoints paginados
- **Password max length** — `LoginRequest.Password` limitado a 200 caracteres (previne bcrypt DoS)
- **Response size limit (Agent)** — respostas HTTP limitadas a 1 MB via `io.LimitReader` (previne OOM)
- **Agent token validation** — rejeita token vazio após enrollment

### Corrigido
- **Type assertions seguras** — 7 locais corrigidos com check `ok` para evitar panic em runtime (`auth.go`, `inventory.go`, `user.go`, `audit.go`)
- **RBAC middleware** — `c.JSON()` + `c.Abort()` substituídos por `c.AbortWithStatusJSON()` (evita execução de handlers após rejeição)
- **Erros 404 vs 500** — device handler agora distingue `sql.ErrNoRows` (404) de erros internos (500) via helper `isNotFound()`
- **SELECT explícito** — `ListUsers` usa colunas nomeadas em vez de `SELECT *`
- **Frontend 204 handling** — `request()` trata respostas `204 No Content` sem tentar parse JSON
- **Frontend logout resiliente** — `logout()` limpa estado local mesmo se a chamada API falhar
- **Agent `IsAuthError`** — usa `errors.As()` tipado em vez de string parsing

### Melhorado
- **Agent coleta resiliente** — queries WMI de `Win32_OperatingSystem` e `Win32_BIOS` não são mais fatais; campos ficam vazios se falharem
- Campos `binding:"required"` relaxados em `InventoryRequest` para suportar coleta parcial
- Documentação e porta padrão atualizados de 8080 para 8081

## [1.1.0] - 2026-02-19

### Adicionado
- **Histórico detalhado de hardware** — detecção automática de alterações em CPU, RAM, placa-mãe, BIOS, discos e interfaces de rede
- Cada alteração registra componente, campo, valor anterior e valor novo (ex: "RAM Total: 16 GB → 8 GB")
- Detecção de discos adicionados, removidos ou com tamanho/tipo alterado
- Detecção de interfaces de rede adicionadas ou removidas (por MAC address)
- Timeline visual no frontend com badges coloridos por componente (CPU, RAM, Disco, etc.)
- Filtro por componente no histórico de hardware (CPU, RAM, Placa-mãe, BIOS, Disco, Rede)
- Paginação no endpoint `GET /devices/:id/hardware-history` com suporte a `?component=&page=&limit=`
- Migration 008 — colunas `component`, `change_type`, `field`, `old_value`, `new_value` em `hardware_history`
- **Log de atividades do device** — rastreamento de login, boot, atualização de OS, software instalado/removido
- Migration 007 — tabela `device_activity_log` com 3 índices
- **Cleanup service** — serviço em background que purga logs antigos, marca devices inativos e executa VACUUM
- Variáveis de ambiente configuráveis: `RETENTION_DAYS`, `INACTIVE_DAYS`, `CLEANUP_INTERVAL`
- Migration 009 — limpeza de registros órfãos + NOT NULL em `hardware_history`

### Melhorado
- Chave de comparação de discos usa composite key (`model+size+type`) quando serial está vazio, evitando colisões
- Lógica de diff de hardware extraída para `hardware_diff.go` (inventory.go: 495 → 308 linhas)
- CSP `connect-src` agora é dinâmico, derivado de `CORS_ORIGINS` (sem IP hardcoded)
- `formatBytesGo(0)` retorna `"N/A"` em vez de `"0 B"`
- `.env.example` atualizado com todas as variáveis documentadas
- Documentação atualizada para v1.1.0

## [1.0.0] - 2026-02-18

### Adicionado
- **Dashboard** com visão geral de dispositivos por status e departamento
- **Listagem de dispositivos** com paginação, ordenação, filtros e exportação CSV
- **Detalhes do dispositivo** com hardware, discos, interfaces de rede, software instalado e ferramentas remotas
- **Ações em massa** — selecionar múltiplos dispositivos para desativar, excluir ou atribuir departamento
- **Gerenciamento de departamentos** — criar, editar, excluir e visualizar dispositivos por departamento
- **Gerenciamento de usuários** — criar, editar e remover usuários (admin only)
- **RBAC** — controle de acesso baseado em roles (admin/viewer)
- **Audit log** — registro de todas as ações administrativas com IP e user-agent
- **Autenticação** — login com JWT via cookies httpOnly (dashboard) e Bearer token (agente)
- **Agent Windows** — serviço Windows que coleta inventário automaticamente e envia ao servidor
- **Instalador** — Inno Setup com enrollment key para provisionamento seguro do agente
- **CI/CD** — GitHub Actions com build, test e lint para server, agent e frontend
- **Tema escuro/claro** — toggle no layout com persistência em localStorage

### Segurança
- Proteção contra timing attacks na comparação de senhas e tokens
- Rate limiting por IP nos endpoints de autenticação
- Content Security Policy e Security Headers
- Enrollment keys com hash bcrypt para provisionamento de agentes

### Infraestrutura
- Docker Compose com PostgreSQL 16 e API Go
- Migrations SQL versionadas (6 arquivos)
- Frontend React + Vite + TypeScript + Tailwind CSS
- Backend Go 1.24 + Gin + sqlx
