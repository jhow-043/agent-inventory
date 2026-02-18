# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/),
e este projeto adere ao [Versionamento Semântico](https://semver.org/lang/pt-BR/).

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
