# Forum Framework — Plugin in Play · Roadmap de Execução

> Framework de fórum **white-label**, modular e pronto para uso.
> Stack: **Go** (API REST) · **Next.js (App Router) + TypeScript + Tailwind + Magic UI** · **MariaDB/PostgreSQL** · Redis opcional · Local/MinIO storage.
>
> Princípio central: **sem protótipo visual vazio**. Toda funcionalidade exibida tem lógica real, persistência e integração com a API. O backend é a fonte de verdade.

---

## Convenções do roadmap

Cada etapa tem:
- **Objetivo** — o que a etapa entrega.
- **Entregáveis** — itens concretos de código/config.
- **Critério de pronto (DoD)** — como validamos antes de avançar.

**Regra de ouro:** ao fim de cada etapa, o código compila, os testes passam e as funcionalidades estão integradas de ponta a ponta. Não avançar com botão sem lógica ou dado fictício no lugar de função real.

---

## Decisões técnicas fixadas (base de todas as etapas)

| Área | Decisão |
|------|---------|
| Monorepo | `apps/api` (Go) + `apps/web` (Next.js), mais `infrastructure/`, `docs/`, `docker-compose.yml`, `.env.example` |
| Arquitetura API | Modular por domínio: `handlers → services → repositories → models`, `middleware`, `config`, `shared` |
| Persistência | Interface de repositório única; implementações por dialeto. SQL compatível com MariaDB **e** PostgreSQL; migrations separadas por banco quando a sintaxe divergir |
| Auth | Access token (curto) + refresh token (rotação + revogação). Argon2id para senhas |
| Validação | Em todo boundary (schema-based), sempre no backend |
| Permissões | RBAC dinâmico em banco (nada fixo no código); verificação sempre server-side; proteção contra escalonamento de privilégio |
| Cache/infra | Redis **opcional** — sempre há fallback interno (ex.: rate limit em memória, contadores em banco) |
| Storage | Provider configurável: Local **ou** MinIO (S3). Nunca `localStorage` do browser para imagens compartilhadas |
| API | Envelope de resposta consistente (`success`, `data`, `error`, `meta` de paginação). OpenAPI/Swagger. Paginação sempre no backend |
| Tempo real | Notificações via SSE (fallback) ou WebSocket quando viável |
| Docs | OpenAPI + guia de instalação + guia de configuração |

---

## Etapa 1 — Fundação ⚙️

**Objetivo:** esqueleto do monorepo rodando localmente via Docker, com API Go e web Next.js conversando, bancos configuráveis e migrations funcionando.

**Entregáveis**
- Estrutura do monorepo (`apps/api`, `apps/web`, `infrastructure`, `docs`).
- API Go: servidor HTTP, roteador, `config` por env, healthcheck, envelope de resposta padrão, middleware base (logging, recover, CORS, request-id).
- Camada de banco: abstração de conexão + seleção MariaDB/PostgreSQL por env; runner de migrations; primeira migration (ex.: tabela de versão/instalação).
- Web Next.js: App Router, TypeScript, Tailwind, Magic UI, Lucide; tema claro/escuro; layout base; cliente HTTP para a API.
- `Dockerfile` (api), `Dockerfile` (web), `docker-compose.yml` (api + web + mariadb + postgres + redis + minio opcionais), `.env.example`.

**DoD**
- `docker compose up` sobe tudo; web acessa a API; healthcheck verde nos dois bancos.
- Migration aplica em MariaDB **e** PostgreSQL sem alterar lógica.
- `go build` e `next build` passam.

---

## Etapa 2 — Autenticação e usuários 🔐

**Objetivo:** cadastro, login, sessões seguras, perfis e uploads, com a base de permissões.

**Entregáveis**
- Cadastro, login, logout; access + refresh token com rotação e revogação; proteção anti-spam/flood e anti-conta-automatizada.
- Hash Argon2id; validação de entrada; rate limiting (com fallback sem Redis).
- Modelo de usuário + perfil: avatar, banner, display name, username único, bio, datas, reputação (base).
- Upload de arquivos: provider Local/MinIO, validação de MIME, limite de tamanho.
- Esqueleto do RBAC: tabelas de cargos/permissões e verificação server-side (cargos iniciais: Dono/Admin/Moderador).
- Páginas web: cadastro, login, perfil público (abas base), configurações pessoais.

**DoD**
- Fluxo completo cadastro → login → refresh → logout funciona ponta a ponta.
- Upload persiste no provider escolhido e é servido corretamente.
- Rotas protegidas rejeitam sem/ com token inválido; permissões checadas no backend.

---

## Etapa 3 — Fórum (núcleo) 💬

**Objetivo:** estrutura hierárquica e interações centrais do fórum.

**Entregáveis**
- Categorias e subcategorias (CRUD, ordem, ícone, capa, descrição, ativar/desativar, permissões por categoria, múltiplos níveis).
- Tópicos: criar/editar/excluir, fixar, bloquear respostas, destacar, favoritar, seguir, denunciar, compartilhar; contadores (views, respostas, curtidas, seguidores); status de publicação.
- Respostas: responder tópico e respostas aninhadas, editar/excluir próprias, menções `@`, imagens/links/formatação, histórico de edição.
- Curtidas em tópicos/respostas com prevenção de duplicata (constraint única).
- Paginação eficiente no backend; índices adequados; transações onde há consistência; sem N+1.

**DoD**
- Criar categoria → subcategoria → tópico → resposta → curtida funciona ponta a ponta e aparece na web.
- Curtida duplicada é impedida no banco; contadores batem.
- Listagens paginadas no backend; queries revisadas contra N+1.

---

## Etapa 4 — Social, notificações e busca 🔔

**Objetivo:** recursos de comunidade, notificações e descoberta.

**Entregáveis**
- Seguir usuários, categorias, subcategorias e tópicos; listas de seguidores/seguindo; contadores no perfil.
- Notificações internas (novo seguidor, curtidas, resposta em tópico seguido/criado, menção, punições, avisos admin); marcar lida/todas, excluir, contador não-lido; tempo real via SSE/WebSocket.
- Busca global (tópicos, respostas, usuários, categorias, subcategorias) com filtros (data, popularidade, nº respostas, nº curtidas, autor, categoria).
- Página de descoberta: populares, recentes, em alta.
- Home personalizável: banner, apresentação, categorias, tópicos recentes/populares, usuários em destaque, estatísticas, últimas atividades — todos os textos administráveis.

**DoD**
- Seguir gera notificação real para o seguido; contador atualiza (em tempo real quando viável).
- Busca retorna resultados corretos com filtros aplicados no backend.
- Textos da home vêm do banco/admin, não do código.

---

## Etapa 5 — Moderação e permissões dinâmicas 🛡️

**Objetivo:** controle de acesso completo e ferramentas de moderação com auditoria.

**Entregáveis**
- RBAC dinâmico: Dono cria/remove/ativa permissões, cria cargos, define hierarquia e escopo; permissões associadas a cargos; Dono imutável por outros cargos; proteção contra escalonamento.
- Banimentos (temp/perma, motivo, expiração, responsável, histórico, revogação autorizada).
- Mutações (temp/perma; impedir tópicos/respostas/interações; motivo e duração).
- Advertências (motivo, responsável, data, histórico).
- Denúncias (usuários/tópicos/respostas/conteúdo) com central e status: pendente/em análise/resolvida/rejeitada.
- Auditoria: toda ação administrativa gera registro.

**DoD**
- Dono reconfigura permissões de Admin/Moderador e o efeito é imediato no backend.
- Ban/mute/advertência/denúncia fluem ponta a ponta e geram log de auditoria.
- Tentativa de escalonamento de privilégio é bloqueada.

---

## Etapa 6 — Painel administrativo 📊

**Objetivo:** painel profissional separado da experiência pública, com métricas reais.

**Entregáveis**
- Layout admin com navegação lateral; dashboard com métricas reais do banco (usuários, ativos, novos, categorias, subcategorias, tópicos, respostas, curtidas, denúncias, banidos, crescimento) e gráficos diário/semanal/mensal.
- Páginas de gerenciamento: usuários, cargos/permissões, categorias, subcategorias, tópicos, respostas, denúncias, banimentos, mutações, advertências, config gerais, personalização visual, armazenamento, logs/auditoria.
- Filtros, pesquisa, ordenação e paginação em todas as listagens.
- Personalização white-label: nome, logo, favicon, descrição, cores, redes sociais, textos (home/boas-vindas/rodapé), links institucionais, termos, privacidade.

**DoD**
- Todas as métricas vêm do banco (zero mock).
- Cada listagem tem filtro/busca/ordenação/paginação funcionando.
- Alterar identidade no admin reflete no fórum público sem tocar no código.

---

## Etapa 7 — Finalização e instalação 🚀

**Objetivo:** produto instalável, testado, seguro e documentado.

**Entregáveis**
- Instalador inicial: nome do fórum, URL, banco (MariaDB/PostgreSQL) + credenciais, provider de storage + config, conta do Dono. Trava contra reexecução sem reinstalação autorizada explícita.
- Seeds iniciais; migrations consolidadas.
- Testes automatizados (unit + integração + E2E de fluxos críticos) e meta de cobertura; revisão de segurança (SQLi/XSS/CSRF, uploads, sessões, rotas admin, logs).
- Otimização (índices, N+1, paginação, payloads).
- Documentação: instalação, configuração, API (OpenAPI).
- Teste de instalação limpa do zero.

**DoD**
- Instalação limpa: configurar banco → storage → conta Dono → personalizar → receber usuários, sem desenvolver nada extra.
- Testes passam; revisão de segurança sem itens críticos; instalador não roda de novo sozinho.

---

## Resultado esperado

Uma pessoa instala, escolhe o banco, define o armazenamento, cria a conta de Dono, personaliza a identidade e lança a comunidade — **o fórum já está construído; basta configurar e abrir**.

---

## Como vamos trabalhar

- Uma etapa por vez; só avança com a anterior compilando, testada e integrada.
- Dentro de cada etapa: pesquisa/reuso → plano → TDD → código → review → integração.
- Dúvidas de escopo são resolvidas antes de implementar o item.
