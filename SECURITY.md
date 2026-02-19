# Política de Segurança

## Versões Suportadas

| Versão | Suporte |
|--------|---------|
| 1.1.x  | ✅ Ativo |
| 1.0.x  | ⚠️ Apenas correções críticas |
| < 1.0  | ❌ Sem suporte |

## Reportando Vulnerabilidades

Se você encontrar uma vulnerabilidade de segurança, **NÃO abra uma issue pública**.

Envie um e-mail para o mantenedor do projeto com:

1. Descrição da vulnerabilidade
2. Passos para reprodução
3. Impacto potencial
4. Sugestão de correção (se houver)

Responderemos em até 48 horas com um plano de ação.

## Práticas de Segurança do Projeto

- Autenticação JWT com cookies httpOnly
- Senhas com hash bcrypt (timing-safe comparison)
- Rate limiting por IP nos endpoints de autenticação
- Enrollment keys com hash bcrypt
- Content Security Policy e Security Headers
- Tokens de dispositivo com hash SHA-256
- Prepared statements para prevenção de SQL injection
- CORS configurável por variável de ambiente
