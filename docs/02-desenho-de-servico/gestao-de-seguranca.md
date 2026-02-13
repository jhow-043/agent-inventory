# Gestão de Segurança

> **Versão:** 1.0.0  
> **Data:** 2026-02-13  
> **Status:** Aprovado  
> **Classificação:** Uso Interno  

---

## 1. Objetivo

Definir as políticas, controles e procedimentos de segurança do Sistema de Inventário de Ativos de TI, incluindo a estratégia de migração de HTTP para HTTPS em fases futuras.

---

## 2. Escopo

Segurança de todos os componentes da Fase 1 (comunicação HTTP) e roadmap para HTTPS.

---

## 3. Classificação de Dados

### 3.1 Dados Tratados pelo Sistema

| Dado | Classificação | Sensibilidade | Justificativa |
|---|---|---|---|
| Hostname | Interno | Baixa | Informação organizacional |
| Serial number | Interno | Média | Identificação de ativos |
| Versão do Windows | Interno | Média | Pode ser usado em ataques direcionados |
| IP addresses | Interno | Média | Topologia de rede |
| MAC addresses | Interno | Média | Identificação de dispositivos |
| Software instalado | Interno | Alta | Revela versões vulneráveis |
| Usuário logado | Interno | Média | Informação de pessoal |
| Device tokens | Confidencial | Alta | Autenticação de dispositivos |
| Senhas de usuários (hash) | Confidencial | Alta | Credenciais de acesso |
| JWT secrets | Confidencial | Crítica | Compromete toda a autenticação |

### 3.2 Classificação de Impacto — CIA

| Dado | Confidencialidade | Integridade | Disponibilidade |
|---|---|---|---|
| Inventário de hardware | Média | Alta | Média |
| Lista de software | Alta | Alta | Média |
| Credenciais/tokens | Alta | Alta | Alta |
| Configuração do sistema | Média | Alta | Alta |

---

## 4. Fase 1: Comunicação HTTP — Análise de Risco

### 4.1 Riscos Aceitos

| Risco | Probabilidade | Impacto | Mitigação | Decisão |
|---|---|---|---|---|
| **Interceptação de tokens** em trânsito | Baixa (rede interna) | Alto | Rede segmentada | **Aceito** |
| **Sniffing de inventário** | Baixa | Médio | Rede interna controlada | **Aceito** |
| **Man-in-the-Middle** | Muito baixa (rede interna) | Alto | Rede confiável | **Aceito** |
| **Replay de requests** | Baixa | Médio | Rate limiting por token | **Aceito** |

### 4.2 Condições para Aceitação

A operação em HTTP é aceitável **somente se**:

1. ✅ O tráfego agent ↔ API **não cruza a internet** (rede interna apenas)
2. ✅ A rede interna é **segmentada** (VLAN dedicada ou sub-rede controlada)
3. ✅ O acesso físico à rede é **controlado** (sem WiFi aberta)
4. ✅ O roadmap HTTPS está **documentado e priorizado** para a próxima fase
5. ✅ O risco é **formalmente aceito** pela gestão de TI

### 4.3 Registro Formal de Aceitação de Risco

| Campo | Valor |
|---|---|
| **ID** | RISK-SEC-001 |
| **Descrição** | Comunicação HTTP sem criptografia de transporte na Fase 1 |
| **Risco residual** | Tokens e dados de inventário transitam em texto claro na rede |
| **Proprietário do risco** | Gestor de TI |
| **Data de aceitação** | YYYY-MM-DD (preencher na implantação) |
| **Revisão** | Na entrega da Fase 2 (HTTPS) |

---

## 5. Controles de Segurança — Fase 1

### 5.1 Autenticação

#### Agent → API: Device Token

| Aspecto | Implementação |
|---|---|
| **Registro** | Agent envia `POST /api/v1/register` com enrollment key |
| **Token** | UUID v4 gerado pelo servidor, retornado ao agent |
| **Armazenamento (agent)** | Arquivo local protegido por ACL do Windows (c:\ProgramData\InventoryAgent\) |
| **Armazenamento (banco)** | Hash SHA-256 do token (token original nunca armazenado) |
| **Uso** | Header `Authorization: Bearer <device-token>` em cada request |
| **Revogação** | DELETE do registro em `device_tokens` |
| **Rotação** | Não implementada na Fase 1 (planejada para Fase 2) |

#### Dashboard → API: JWT

| Aspecto | Implementação |
|---|---|
| **Login** | `POST /api/v1/auth/login` com username + senha |
| **Hash de senha** | bcrypt com cost 12 |
| **Access token** | JWT, expiração 15 minutos |
| **Refresh token** | JWT, expiração 7 dias |
| **Transporte** | httpOnly, Secure=false (Fase 1 HTTP), SameSite=Strict cookies |
| **Claims** | `sub` (user_id), `exp`, `iat`, `role` |
| **Secret** | Variável de ambiente `JWT_SECRET` (≥ 32 caracteres) |
| **Blacklist** | Não implementada na Fase 1 |

### 5.2 Autorização

| Endpoint | Autenticação | Autorização |
|---|---|---|
| `POST /api/v1/register` | Enrollment Key | Qualquer agent com a key |
| `POST /api/v1/inventory` | Device Token | Apenas o device dono do token |
| `POST /api/v1/auth/login` | Pública | — |
| `POST /api/v1/auth/refresh` | Refresh Token | Usuário autenticado |
| `GET /api/v1/devices` | JWT | Role: admin |
| `GET /api/v1/devices/:id` | JWT | Role: admin |
| `GET /api/v1/dashboard/stats` | JWT | Role: admin |

> **Fase 1:** Role único (`admin`). RBAC granular (viewer, operator, admin) planejado para fases futuras.

### 5.3 Validação de Input

| Controle | Implementação |
|---|---|
| **Tamanho de payload** | Máximo 1 MB por request (middleware Gin) |
| **Validação de campos** | `go-playground/validator` — tags de validação em cada struct |
| **Sanitização** | Campos string com limite de tamanho; reject de caracteres especiais onde não esperados |
| **Content-Type** | Apenas `application/json` aceito |
| **Query parameters** | Validação de tipos e ranges |

### 5.4 Rate Limiting

| Escopo | Limite | Ação ao Exceder |
|---|---|---|
| Por device token | 10 requests/minuto | HTTP 429 Too Many Requests |
| Por IP (endpoints públicos) | 20 requests/minuto | HTTP 429 |
| Login endpoint | 5 tentativas/minuto por IP | HTTP 429 |

### 5.5 Headers de Segurança HTTP

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 0
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

### 5.6 CORS

```
Access-Control-Allow-Origin: http://localhost:3000 (configurável por env var)
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
```

---

## 6. Segurança da Infraestrutura

### 6.1 Containers Docker

| Controle | Implementação |
|---|---|
| **Usuário non-root** | `USER 1000:1000` no Dockerfile |
| **Read-only filesystem** | `read_only: true` onde possível (API, Dashboard) |
| **No new privileges** | `security_opt: - no-new-privileges:true` |
| **Imagem base mínima** | `golang:alpine` para build, `alpine` para runtime |
| **Scan de vulnerabilidades** | `docker scout` ou `trivy` na pipeline CI |
| **Sem capabilities extras** | `cap_drop: ALL` |

### 6.2 PostgreSQL

| Controle | Implementação |
|---|---|
| **Credenciais** | Via variáveis de ambiente, nunca hardcoded |
| **Acesso restrito** | Apenas containers na mesma Docker network |
| **Porta não exposta** | Porta 5432 NÃO exposta para o host (apenas rede interna Docker) |
| **Senha forte** | Mínimo 20 caracteres, gerada aleatoriamente |
| **Volume separado** | Dados em volume Docker persistente dedicado |

### 6.3 Rede Docker

```yaml
networks:
  inventory-net:
    driver: bridge
    internal: false  # API precisa ser acessível externamente
```

O PostgreSQL se comunica apenas pela rede Docker interna. Nunca expor a porta 5432 para a rede externa.

---

## 7. Segurança do Agent Windows

| Controle | Implementação |
|---|---|
| **Execução como serviço** | Conta `LocalSystem` ou conta de serviço dedicada |
| **Sem portas abertas** | Agent nunca escuta conexões de entrada |
| **Token protegido** | Armazenado com permissões restritas (ACL: apenas SYSTEM e Administrators) |
| **Log seguro** | Logs não contêm tokens ou dados sensíveis |
| **Binário assinado** | Planejado para fases futuras (code signing) |
| **Sem GUI** | Nenhuma interface gráfica; somente service + logs |

---

## 8. Gestão de Secrets

| Secret | Armazenamento | Rotação |
|---|---|---|
| `JWT_SECRET` | `.env` no servidor (não versionado no Git) | A cada release major |
| `ENROLLMENT_KEY` | `.env` + config do agent | Quando necessário |
| `DATABASE_URL` (senha) | `.env` no servidor | Anual ou após incidente |
| Device tokens | Hash no banco, original no agent (arquivo local) | Não implementada na Fase 1 |
| Senhas de usuários | Hash bcrypt no banco | Política de expiração futura |

### Regras de Secrets

1. **Nunca** commitar secrets no Git
2. `.env` está no `.gitignore`
3. `.env.example` existe com valores de exemplo (sem secrets reais)
4. Secrets de produção são configurados manualmente no servidor
5. Secrets com ≥ 32 caracteres aleatórios

---

## 9. Roadmap HTTP → HTTPS

### 9.1 Pontos de Configuração

| Componente | Mudança Necessária | Arquivo |
|---|---|---|
| **Agent** | `api_url: https://...` | `agent-config.yaml` |
| **API** | TLS listener ou reverse proxy | `docker-compose.yml` ou config Nginx |
| **Dashboard** | URL da API com HTTPS | `.env` ou `vite.config.ts` |
| **Cookies JWT** | `Secure: true` | Middleware de auth |
| **CORS** | Origem com `https://` | Env var `CORS_ORIGINS` |

### 9.2 Opções de Implementação

#### Opção A: TLS Termination com Reverse Proxy (Recomendada)

```
Agent ──HTTPS──→ Nginx/Traefik ──HTTP──→ API Container
Browser ──HTTPS──→ Nginx/Traefik ──HTTP──→ Dashboard Container
```

- Certificado configurado apenas no Nginx/Traefik
- API e Dashboard continuam HTTP internamente
- Renovação automática com Let's Encrypt (se acessível externamente) ou certs manuais

#### Opção B: TLS Direto na API Go

```
Agent ──HTTPS──→ API Container (TLS)
```

- API carrega certificado e chave privada
- Mais simples (sem proxy adicional), porém API gerencia TLS
- Requer restart para renovar certificados

### 9.3 Certificados para Ambiente On-Premises

| Tipo | Uso | Geração |
|---|---|---|
| **CA interna** | Ambiente corporativo com PKI | `openssl` ou Microsoft CA |
| **Self-signed** | Testes e ambientes isolados | `openssl req -x509 -newkey ...` |
| **Let's Encrypt** | Se o servidor for acessível pela internet | `certbot` |

### 9.4 Checklist de Migração HTTP → HTTPS

- [ ] Escolher opção de TLS (reverse proxy vs direto)
- [ ] Gerar/obter certificados
- [ ] Configurar TLS no proxy ou na API
- [ ] Atualizar `api_url` no config de TODOS os agents
- [ ] Setar `Secure: true` nos cookies JWT
- [ ] Atualizar CORS origins para `https://`
- [ ] Testar: agent envia inventory via HTTPS com sucesso
- [ ] Testar: dashboard funciona via HTTPS
- [ ] Testar: certificado válido (não expirado, hostname correto)
- [ ] Documentar procedimento de renovação de certificados
- [ ] Desabilitar HTTP (ou redirecionar 301 para HTTPS)
- [ ] Atualizar este documento e fechar RISK-SEC-001

---

## 10. Gestão de Vulnerabilidades

### 10.1 Dependências

| Ferramenta | Escopo | Frequência |
|---|---|---|
| `go mod verify` | Integridade dos módulos Go | Em cada build (CI) |
| `govulncheck` | Vulnerabilidades conhecidas em deps Go | Semanal (CI) |
| `npm audit` | Vulnerabilidades em deps JavaScript | Semanal (CI) |
| `docker scout` / `trivy` | Vulnerabilidades em imagens Docker | Em cada build |
| Dependabot/Renovate | Atualizações automáticas de deps | Contínuo |

### 10.2 Política de Patches

| Severidade | Ação | Prazo |
|---|---|---|
| Crítica (CVSS ≥ 9.0) | Patch emergencial | 24 horas |
| Alta (CVSS 7.0–8.9) | Patch prioritário | 7 dias |
| Média (CVSS 4.0–6.9) | Próxima release | 30 dias |
| Baixa (CVSS < 4.0) | Backlog | Próximo ciclo |

---

## 11. Auditoria de Segurança

### 11.1 Logs de Auditoria

Eventos registrados nos logs da API:

| Evento | Dados Logados |
|---|---|
| Login bem-sucedido | user_id, IP, timestamp |
| Login falho | username tentado, IP, timestamp |
| Registro de device | device_id, hostname, serial, IP |
| Submissão de inventário | device_id, request_id, duration |
| Acesso a dados de device | user_id, device_id, request_id |
| Erro de autenticação | tipo (JWT expirado, token inválido), IP |
| Rate limit excedido | IP/token, endpoint |

### 11.2 Revisão de Segurança

| Atividade | Frequência |
|---|---|
| Revisão de logs de acesso | Semanal |
| Scan de vulnerabilidades | A cada build |
| Revisão de permissões | Mensal |
| Teste de penetração (básico) | A cada release major |
| Revisão deste documento | Trimestral |

---

## 12. Referências

- [Gestão de Disponibilidade](gestao-de-disponibilidade.md)
- [Gestão de Continuidade](gestao-de-continuidade.md)
- [Gestão de Incidentes](../04-operacao-de-servico/gestao-de-incidentes.md)
- [Runbooks Operacionais](../04-operacao-de-servico/runbooks-operacionais.md)

---

## Controle de Versões

| Versão | Data | Autor | Descrição |
|---|---|---|---|
| 1.0.0 | 2026-02-13 | — | Documento inicial — comunicação HTTP, roadmap HTTPS |
