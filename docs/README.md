# Documentação — Inventário de Ativos de TI

Documentação técnica completa do sistema: como funciona, como foi construído e como usar.

> **Versão atual:** v1.2.0 — Hardening de segurança (rate limiting, body size limit, security headers, bcrypt DoS protection), RBAC completo, audit logs, 10 migrations, 12 tabelas.

## Índice

| # | Documento | O que explica |
|---|-----------|---------------|
| 1 | [Visão Geral e Arquitetura](01-visao-geral.md) | O que é o sistema, componentes, stack, fluxo completo |
| 2 | [Backend — API](02-backend-api.md) | Rotas, handlers, middlewares, cleanup service, autenticação |
| 3 | [Agent Windows](03-agent.md) | Como o agent funciona, collectors WMI, serviço Windows |
| 4 | [Frontend](04-frontend.md) | Páginas, componentes, hooks, API client, temas |
| 5 | [Banco de Dados](05-banco-de-dados.md) | Schema (10 migrações), 12 tabelas, relações, hardware_history granular |
| 6 | [Instalação e Deploy](06-instalacao.md) | Docker, variáveis (.env.example), build, CI/CD, HTTPS |
