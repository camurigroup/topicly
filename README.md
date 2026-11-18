# Forum Framework — Plugin in Play

Framework de fórum **white-label**, modular e pronto para uso.

- **API:** Go (REST, arquitetura modular por domínio)
- **Web:** Next.js (App Router) + TypeScript + Tailwind + Magic UI + Lucide
- **Banco:** MariaDB **ou** PostgreSQL (selecionável por configuração)
- **Cache:** Redis opcional (há fallback interno)
- **Storage:** Local **ou** MinIO (S3)

O plano completo por etapas está em [ROADMAP.md](ROADMAP.md).

## Estrutura

```
apps/
  api/            API Go
    cmd/server/   entrypoint
    internal/     config, server, middleware, database, shared, modules
    migrations/   migrations por banco (postgres/, mariadb/)
  web/            frontend Next.js
infrastructure/   infra auxiliar
docs/             documentação
```

## Desenvolvimento local

Pré-requisitos: Go 1.24+, Node 20+, e um banco (Postgres ou MariaDB).

```bash
cp .env.example .env     # ajuste os valores

# API
cd apps/api
go run ./cmd/server

# Web (noutro terminal)
cd apps/web
npm install
npm run dev
```

Com Docker: `docker compose up` (sobe api, web e serviços de apoio).

## Status

Em construção — ver [ROADMAP.md](ROADMAP.md).

- ✅ **Etapa 1 — Fundação** (monorepo, API Go, web Next.js, migrations, Docker)
- ✅ **Etapa 2 — Autenticação e usuários** (registro/login/refresh, perfis, uploads, RBAC base, rate limit) — verificada ponta a ponta
- ✅ **Etapa 3 — Fórum**: backend (categorias/subcategorias, tópicos, respostas, likes, follows) + frontend — navegação de categorias, tópicos paginados, respostas aninhadas, curtir/seguir, moderação (fixar/bloquear/destacar) e painel de categorias.
- ✅ **Etapa 4 — Social, notificações e busca**: seguir usuários/categorias/subcategorias/tópicos com listas e contadores; notificações internas (seguidor, curtidas, respostas, tópicos seguidos, menções `@`, novos tópicos, avisos admin) com leitura/exclusão e tempo real via SSE; busca global com filtros no backend; descoberta (em alta, populares, recentes); home com textos e seções administráveis (`/admin/home`). Verificada em PostgreSQL e MariaDB.
  - Tempo real usa um hub em memória (sem dependência de Redis); com várias instâncias da API, cada uma alcança só as próprias conexões.
- ✅ **Etapa 5 — Moderação e permissões dinâmicas**: cargos personalizados com hierarquia e permissões editáveis (efeito imediato), catálogo de permissões (ativar/desativar/criar), anti-escalonamento (só gerencia cargos abaixo do seu, só concede o que possui, Dono imutável); banimentos e silenciamentos temporários/permanentes com revogação, advertências e histórico; denúncias de usuários/tópicos/respostas com central e status; auditoria de toda ação administrativa (ator, alvo, detalhes, IP). Telas em `/admin/reports`, `/admin/users`, `/admin/roles`, `/admin/audit`. Verificada em PostgreSQL.
  - Banimento vale na hora (inclusive com token ainda válido) e encerra as sessões; silenciado continua lendo, mas não publica, curte nem segue.
  - O estado de ban/mute fica em cache por até 15 s por instância; alterações feitas pela própria instância invalidam o cache na hora.
