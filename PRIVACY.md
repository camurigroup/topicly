# Política de telemetria e privacidade

## Resumo

**O Topicly não tem telemetria.** O software não envia dados de uso, métricas,
relatórios de erro, identificadores de instalação nem qualquer outra informação para
os autores do projeto ou para terceiros. Não há "phone home", verificação de licença
nem rastreadores embutidos.

Cada instalação é independente: os dados ficam no banco e no armazenamento escolhidos
por quem opera o fórum.

## Quem é o responsável pelos dados

Quem instala e opera uma comunidade com o Topicly é o **controlador** dos dados dos
membros dela (para fins da LGPD e de leis semelhantes). Os autores do Topicly não têm
acesso a nenhuma instalação. A política de privacidade apresentada aos membros é a
página `/privacy` de cada fórum, editável no painel administrativo.

## Dados que o Topicly armazena (na própria instalação)

| Dado | Para quê |
|---|---|
| Usuário, e-mail, nome de exibição, bio, avatar e capa | conta e perfil público (o e-mail nunca aparece no perfil) |
| Senha | guardada apenas como hash **argon2id** — nunca em texto |
| Sessões (refresh tokens) com navegador (user agent) e IP | manter o login, permitir encerrar sessões e banir |
| Tópicos, respostas, curtidas, seguidores e notificações | funcionamento do fórum |
| Denúncias, advertências e sanções | moderação |
| Registro de auditoria com autor, ação e IP | rastrear ações administrativas |
| Configurações do site | identidade visual e layout |

O Topicly não usa cookies de rastreamento: a sessão do site fica no armazenamento
local do navegador e serve apenas para autenticar o membro na própria comunidade.

## Comunicação externa

A API só se comunica com o banco de dados e com o armazenamento configurados
(disco local ou um MinIO/S3 indicado pelo operador). O site busca fontes e recursos
apenas do próprio servidor. Se o operador apontar imagens ou links para serviços
externos (ex.: logo em uma CDN), esses serviços seguem as próprias políticas.

## Recomendações a quem opera um fórum

- Publique em `/privacy` quais dados você coleta, por quanto tempo guarda e como o
  membro pode pedir acesso, correção ou exclusão.
- Defina por quanto tempo manter registros de auditoria e sessões antigas.
- Use HTTPS e mantenha backups protegidos.

## Mudanças nesta política

Se algum recurso que envie dados para fora da instalação for proposto no futuro, ele
será **desativado por padrão**, documentado aqui e só funcionará com consentimento
explícito do operador.
