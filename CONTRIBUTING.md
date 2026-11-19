# Contribuindo com o Topicly

Obrigado pelo interesse! Este guia explica como propor mudanças de um jeito que dê
para revisar e aceitar rápido.

> Encontrou uma falha de segurança? **Não abra issue pública** — siga o [SECURITY.md](SECURITY.md).

## Licença das contribuições

Ao enviar uma contribuição, você concorda que ela é distribuída sob a
[AGPL-3.0](LICENSE), a mesma licença do projeto, e declara que tem o direito de
enviá-la (o código é seu ou compatível com a AGPL-3.0). Marque cada commit com
`Signed-off-by` (`git commit -s`), conforme o [Developer Certificate of Origin](https://developercertificate.org/).

## Antes de começar

- Para mudanças grandes (novo módulo, mudança de banco, nova dependência), abra uma
  issue antes para combinar a abordagem.
- Confira o [ROADMAP.md](ROADMAP.md) e as issues abertas para evitar trabalho duplicado.
- Ambiente: Go 1.24+, Node 20+ e PostgreSQL ou MariaDB. Veja [docs/INSTALL.md](docs/INSTALL.md).

## Regras do projeto

1. **Um commit por micro-alteração.** Cada commit faz uma coisa só e deixa o projeto
   compilando e com os testes passando.
2. **Conventional Commits** em português: `feat(topic): ...`, `fix(auth): ...`,
   `docs: ...`, `refactor`, `test`, `perf`, `style`, `chore`.
3. **Testes primeiro** para regras de negócio: escreva o teste que falha, depois o código.
4. **Os dois bancos importam.** Todo SQL deve funcionar em PostgreSQL e MariaDB; use
   `?` como placeholder (a API reescreve para `$1` no PostgreSQL). Mudanças de schema
   exigem migration nas duas pastas (`migrations/postgres` e `migrations/mariadb`) com
   o mesmo número.
5. **Segurança por padrão:** SQL parametrizado, ordenação só por lista permitida,
   permissões verificadas no servidor, validação de toda entrada, nada de `innerHTML`.
6. **Arquitetura por domínio:** `handler → service → repository → model` dentro do
   módulo; nada de lógica de negócio nos handlers.
7. **Textos de interface em português**, código e comentários em inglês.

## Checagens antes do pull request

```bash
cd apps/api && gofmt -l . && go vet ./... && go test ./...
cd apps/web && npx tsc --noEmit && npm run build
scripts/e2e.sh http://localhost:8080        # com a API rodando
node scripts/gen-openapi.mjs                # se você mudou rotas
```

## Pull request

- Descreva **o problema** e **a solução**, com passos para testar.
- Inclua capturas de tela para mudanças visuais (claro e escuro, desktop e celular).
- Atualize a documentação (`README.md`, `docs/`) quando o comportamento mudar.
- Mantenha o PR focado: refatorações não relacionadas vão em outro PR.

## Reportando bugs

Abra uma issue com: versão, banco (PostgreSQL ou MariaDB), passos para reproduzir,
resultado esperado, resultado obtido e logs relevantes (sem senhas ou tokens).

## Conduta

Seja respeitoso e construtivo. Críticas são ao código, nunca às pessoas.
Comportamento abusivo leva ao bloqueio no projeto.
