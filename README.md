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
