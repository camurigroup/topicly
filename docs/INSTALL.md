# Instalação e configuração

O Forum Framework já vem pronto: você escolhe o banco, o armazenamento, cria a conta
do Dono e personaliza a identidade — sem escrever código.

## Requisitos

| Componente | Versão |
|---|---|
| Go | 1.24+ (para compilar a API) |
| Node.js | 20+ (para o site) |
| Banco | PostgreSQL 14+ **ou** MariaDB 10.6+ |
| Armazenamento | disco local **ou** MinIO/S3 |
| Redis | opcional (há alternativas internas) |

## 1. Banco de dados

Crie um banco vazio e um usuário com permissão total nele.

```sql
-- PostgreSQL
CREATE ROLE forum LOGIN PASSWORD 'troque-esta-senha';
CREATE DATABASE forum OWNER forum;

-- MariaDB
CREATE DATABASE forum CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'forum'@'%' IDENTIFIED BY 'troque-esta-senha';
GRANT ALL PRIVILEGES ON forum.* TO 'forum'@'%';
```

## 2. Instalador de linha de comando

Na pasta `apps/api`:

```bash
go run ./cmd/installer
```

Ele pergunta banco (PostgreSQL ou MariaDB), credenciais, porta da API, URL do site e o
armazenamento; **testa a conexão**, **aplica as migrations** e grava o `.env` com
segredos JWT fortes e um `INSTALL_TOKEN`. Também aceita tudo por flags:

```bash
go run ./cmd/installer -yes -driver mariadb -db-host 127.0.0.1 -db-user forum \
  -db-password 'troque-esta-senha' -db-name forum -app-url https://forum.exemplo.com
```

Se o `.env` já existir, o instalador **se recusa a rodar** — use `-force` apenas para
uma reinstalação consciente.

## 3. Subir a API e o site

```bash
# API (lê o .env da pasta atual, ou o caminho em ENV_FILE)
cd apps/api && go build -o bin/server ./cmd/server && ./bin/server

# Site (NEXT_PUBLIC_API_URL aponta para a API)
cd apps/web && npm ci && NEXT_PUBLIC_API_URL=http://localhost:8080 npm run build && npm start
```

Ou com Docker: `docker compose up` (PostgreSQL por padrão; MariaDB, Redis e MinIO por
perfis — veja `docker-compose.yml`).

## 4. Assistente de instalação

Abra o site: enquanto não houver Dono, todo acesso é levado para **`/install`**.
Informe o nome do fórum, a descrição, a conta do Dono e o `INSTALL_TOKEN` (mostrado pelo
instalador). Ao concluir:

- a conta vira **Dono** (todas as permissões, imutável por outros cargos);
- nome e descrição são aplicados à identidade visual;
- a categoria **Geral** é criada;
- o instalador **trava**: `/install` deixa de funcionar e o cadastro público é liberado.

Antes da instalação, `POST /auth/register` responde `403 not_installed` — ninguém
consegue “roubar” o cargo de Dono num servidor recém-publicado.

## 5. Personalizar

No painel (`/admin`): **Identidade visual** (nome, logo, favicon, cores, redes sociais,
rodapé, termos e privacidade), **Página inicial**, **Categorias**, **Cargos e
permissões**. Tudo reflete no site público na hora.

## Variáveis de ambiente

| Variável | Padrão | Descrição |
|---|---|---|
| `APP_ENV` | `development` | `production` ativa as validações de segurança |
| `APP_URL` | `http://localhost:3000` | URL pública do site |
| `API_PORT` | `8080` | porta da API |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | origens permitidas (vírgula) |
| `DB_DRIVER` | `postgres` | `postgres` ou `mariadb` |
| `DB_HOST` / `DB_PORT` | `localhost` / `5432` | servidor do banco |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `forum` | credenciais |
| `DB_SSLMODE` | `disable` | só PostgreSQL |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | valores de dev | **obrigatório trocar em produção** (≥ 32 caracteres, diferentes) |
| `JWT_ACCESS_TTL` / `JWT_REFRESH_TTL` | `15m` / `720h` | validade dos tokens |
| `STORAGE_PROVIDER` | `local` | `local` ou `minio` |
| `STORAGE_LOCAL_PATH` | `./storage` | pasta dos uploads (provider local) |
| `STORAGE_PUBLIC_URL` | vazio | vazio = URLs relativas `/files/...` (recomendado) |
| `MINIO_ENDPOINT` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` / `MINIO_BUCKET` / `MINIO_USE_SSL` | — | provider MinIO/S3 |
| `INSTALL_TOKEN` | vazio | se definido, exigido pelo assistente `/install` |
| `ENV_FILE` | `.env` | arquivo de variáveis lido pela API |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | (site) endereço da API; o site encaminha `/files` para ela |

## Checklist de produção

- [ ] `APP_ENV=production` — a API **não inicia** com segredos JWT padrão, curtos ou iguais.
- [ ] HTTPS na frente da API e do site (proxy reverso); `CORS_ALLOWED_ORIGINS` só com o domínio do site.
- [ ] `.env` fora do controle de versão (já está no `.gitignore`) e com permissão `600`.
- [ ] Backup do banco e da pasta `STORAGE_LOCAL_PATH` (ou do bucket).
- [ ] Senha forte para a conta do Dono.

A API já envia cabeçalhos defensivos (`X-Content-Type-Options`, `X-Frame-Options`,
`Referrer-Policy`, `Permissions-Policy`), limita tentativas em login/cadastro/denúncias,
usa argon2id nas senhas e consultas parametrizadas em todo o SQL.

## Testes e referência da API

```bash
cd apps/api && go test ./...                 # testes unitários
scripts/e2e.sh http://localhost:8080          # 62 verificações ponta a ponta (bash, curl, node)
node scripts/gen-openapi.mjs                  # regenera docs/openapi.yaml a partir das rotas
```

A referência completa da API (83 operações) está em [`openapi.yaml`](openapi.yaml).
