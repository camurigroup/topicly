# Política de segurança

Levamos a segurança das comunidades que usam o Topicly a sério. Se você encontrou uma
vulnerabilidade, obrigado por avisar com responsabilidade.

## Versões com suporte

| Versão | Correções de segurança |
|---|---|
| 1.0.x | ✅ |
| < 1.0 | ❌ |

## Como reportar

**Não abra issue pública** nem divulgue a falha antes da correção.

Use o canal privado do repositório oficial: aba **Security → Report a vulnerability**
(GitHub Private Vulnerability Reporting). Inclua, se possível:

- a versão do Topicly e o banco (PostgreSQL ou MariaDB);
- o tipo de falha (ex.: injeção de SQL, XSS, escalonamento de privilégio, desvio de
  autenticação, exposição de dados);
- passos para reproduzir ou uma prova de conceito mínima;
- o impacto que você identificou.

Não use dados reais de terceiros nos testes, não degrade a disponibilidade de
instâncias em produção e não acesse contas que não sejam suas.

## O que acontece depois

| Etapa | Prazo-alvo |
|---|---|
| Confirmação de recebimento | até 3 dias úteis |
| Avaliação inicial e severidade | até 7 dias úteis |
| Correção de falhas críticas e altas | até 30 dias |
| Divulgação pública (aviso + versão corrigida) | após a correção, em comum acordo |

Quem reportar será creditado no aviso de segurança, se desejar.

## Escopo

Dentro do escopo: o código deste repositório (API Go, site Next.js, instalador e
migrations). Fora do escopo: configurações de cada instalação (ex.: segredos fracos
definidos pelo operador, ausência de HTTPS), serviços de terceiros e ataques que
exigem acesso físico ou contas de administrador legitimamente comprometidas.

## Boas práticas para quem opera um Topicly

- Rode com `APP_ENV=production` — a API se recusa a iniciar com segredos JWT fracos.
- Sirva tudo por HTTPS e restrinja `CORS_ALLOWED_ORIGINS` ao domínio do site.
- Mantenha o `.env` fora do controle de versão, com permissão `600`.
- Faça backup do banco e da pasta de uploads; aplique as atualizações de segurança.
- Revise periodicamente os cargos e a auditoria no painel administrativo.
