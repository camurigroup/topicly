#!/usr/bin/env bash
# End-to-end checks for Etapas 5 and 6 against a live API.
# Usage: scripts/e2e.sh http://localhost:8080   (DB populado ou recem-instalado)
# Requer: bash, curl e node. Usa as contas de teste dono/membro/moder (senha de teste local).
B="${1:-http://localhost:8099}/api/v1"
PASS=0; FAIL=0
j() { node -e "let d='';process.stdin.on('data',c=>d+=c).on('end',()=>{try{const o=JSON.parse(d);const v=$1;console.log(typeof v==='object'?JSON.stringify(v):v)}catch(e){console.log('ERR')}})"; }
req() { curl -s -X "$1" "$B$2" -H "Content-Type: application/json" ${3:+-H "Authorization: Bearer $3"} ${4:+-d "$4"}; }
code() { curl -s -o /dev/null -w "%{http_code}" -X "$1" "$B$2" -H "Content-Type: application/json" ${3:+-H "Authorization: Bearer $3"} ${4:+-d "$4"}; }
login() { req POST /auth/login "" "{\"identifier\":\"$1\",\"password\":\"senha12345\"}" | j "o.data.tokens.accessToken"; }
check() { if [ "$2" = "$3" ]; then PASS=$((PASS+1)); echo "PASS $1"; else FAIL=$((FAIL+1)); echo "FAIL $1: got '$2' want '$3'"; fi; }

for u in dono membro moder; do
  req POST /auth/register "" "{\"username\":\"$u\",\"email\":\"$u@example.test\",\"password\":\"senha12345\"}" >/dev/null
done
OWNER=$(login dono); MEMBER=$(login membro); MOD=$(login moder)
check "login dono" "$( [ ${#OWNER} -gt 20 ] && echo ok)" ok

# Seed content when the DB is empty.
CAT=$(req GET /categories "" | j 'o.data[0]?o.data[0].id:""')
if [ -z "$CAT" ]; then CAT=$(req POST /categories "$OWNER" '{"name":"Geral","description":"Conversas gerais"}' | j o.data.id); fi
TOPIC=$(req POST /topics "$MEMBER" "{\"categoryId\":\"$CAT\",\"title\":\"Topico de teste e2e\",\"content\":\"conteudo do teste\"}" | j o.data.id)
check "membro cria topico" "$( [ ${#TOPIC} -eq 36 ] && echo ok)" ok
req POST /topics/$TOPIC/replies "$OWNER" '{"content":"resposta do dono"}' >/dev/null
req POST /topics/$TOPIC/like "$OWNER" >/dev/null

echo "== Etapa 5: RBAC"
check "dono atribui moderador" "$(req PUT /admin/users/moder/roles "$OWNER" '{"roles":["moderator"]}' | j 'o.data.map(r=>r.slug).join()')" moderator
MOD=$(login moder)
check "membro sem acesso a cargos" "$(code GET /admin/roles "$MEMBER")" 403
check "moderador nao se promove" "$(code PUT /admin/users/moder/roles "$MOD" '{"roles":["admin"]}')" 403
check "dono nao altera a si mesmo" "$(code PUT /admin/users/dono/roles "$OWNER" '{"roles":[]}')" 403
ROLES=$(req GET /admin/roles "$OWNER")
OWNER_ROLE=$(echo "$ROLES" | j 'o.data.find(r=>r.slug==="owner").id')
MOD_ROLE=$(echo "$ROLES" | j 'o.data.find(r=>r.slug==="moderator").id')
check "cargo Dono imutavel" "$(code PUT /admin/roles/$OWNER_ROLE/permissions "$OWNER" '{"permissions":[]}')" 403
check "cargo no nivel do Dono bloqueado" "$(code POST /admin/roles "$OWNER" '{"slug":"super-e2e","name":"Super","rankOrder":100,"permissions":[]}')" 403
MODPERMS=$(echo "$ROLES" | j 'o.data.find(r=>r.slug==="moderator").permissions')
LESS=$(echo "$ROLES" | j 'o.data.find(r=>r.slug==="moderator").permissions.filter(p=>p!=="reports.manage")')
req PUT /admin/roles/$MOD_ROLE/permissions "$OWNER" "{\"permissions\":$LESS}" >/dev/null
check "efeito imediato (perde permissao)" "$(code GET /admin/reports "$MOD")" 403
req PUT /admin/roles/$MOD_ROLE/permissions "$OWNER" "{\"permissions\":$MODPERMS}" >/dev/null
check "permissao restaurada" "$(code GET /admin/reports "$MOD")" 200
PM=$(req GET /admin/permissions "$OWNER" | j 'o.data.find(p=>p.slug==="permissions.manage").id')
check "permissao essencial nao desativa" "$(code PATCH /admin/permissions/$PM "$OWNER" '{"isActive":false}')" 403

echo "== Etapa 5: moderacao"
check "moderador nao bane o dono" "$(code POST /admin/users/dono/ban "$MOD" '{"reason":"teste","durationHours":1}')" 403
check "advertencia" "$(code POST /admin/users/membro/warnings "$MOD" '{"reason":"aviso do e2e"}')" 201
check "mute" "$(code POST /admin/users/membro/mute "$MOD" '{"reason":"flood","durationHours":2}')" 201
check "silenciado nao responde" "$(code POST /topics/$TOPIC/replies "$MEMBER" '{"content":"oi"}')" 403
check "silenciado ainda le" "$(code GET /topics/$TOPIC "$MEMBER")" 200
MUTE=$(req GET "/admin/sanctions?type=mute&active=1&q=membro" "$MOD" | j 'o.data[0].id')
check "revoga mute" "$(code POST /admin/sanctions/$MUTE/revoke "$MOD" '{"reason":"acordo"}')" 200
check "volta a responder" "$(code POST /topics/$TOPIC/replies "$MEMBER" '{"content":"de volta"}')" 201
REFRESH=$(req POST /auth/login "" '{"identifier":"membro","password":"senha12345"}' | j o.data.tokens.refreshToken)
check "ban" "$(code POST /admin/users/membro/ban "$MOD" '{"reason":"spam","durationHours":0}')" 201
check "banido barrado com token valido" "$(code GET /me "$MEMBER")" 403
check "banido nao faz login" "$(code POST /auth/login "" '{"identifier":"membro","password":"senha12345"}')" 403
check "sessoes encerradas" "$(code POST /auth/refresh "" "{\"refreshToken\":\"$REFRESH\"}")" 401
BAN=$(req GET "/admin/sanctions?type=ban&active=1&q=membro" "$OWNER" | j 'o.data[0].id')
check "revoga ban" "$(code POST /admin/sanctions/$BAN/revoke "$OWNER" '{"reason":"recurso aceito"}')" 200
MEMBER=$(login membro)
check "login apos revogacao" "$( [ ${#MEMBER} -gt 20 ] && echo ok)" ok

echo "== Etapa 5: denuncias e auditoria"
check "denuncia" "$(code POST /reports "$MOD" "{\"targetType\":\"topic\",\"targetId\":\"$TOPIC\",\"reason\":\"spam\",\"details\":\"e2e\"}")" 201
check "denuncia duplicada" "$(code POST /reports "$MOD" "{\"targetType\":\"topic\",\"targetId\":\"$TOPIC\",\"reason\":\"spam\"}")" 409
REP=$(req GET "/admin/reports?status=pending&q=e2e" "$OWNER" | j 'o.data[0].id')
check "busca na fila de denuncias" "$( [ ${#REP} -eq 36 ] && echo ok)" ok
check "resolve denuncia" "$(req PATCH /admin/reports/$REP "$OWNER" '{"status":"resolved","note":"ok"}' | j o.data.status)" resolved
ACTIONS=$(req GET /admin/audit/actions "$OWNER" | j 'o.data.join()')
for a in user.roles role.permissions user.warn user.mute user.ban sanction.revoke report.status; do
  case ",$ACTIONS," in *",$a,"*) check "auditoria $a" ok ok;; *) check "auditoria $a" missing ok;; esac
done
check "auditoria busca por autor" "$(req GET "/admin/audit?q=moder&limit=1" "$OWNER" | j 'o.data[0].actor.username')" moder
check "moderador sem logs.view" "$(code GET /admin/audit "$MOD")" 403

echo "== Etapa 6: metricas"
M=$(req GET /admin/metrics "$OWNER")
check "metricas: usuarios >= 3" "$(echo "$M" | j 'o.data.users>=3')" true
check "metricas: topicos >= 1" "$(echo "$M" | j 'o.data.topics>=1')" true
check "metricas: advertencias >= 1" "$(echo "$M" | j 'o.data.warnings>=1')" true
check "metricas: banidos agora = 0" "$(echo "$M" | j 'o.data.banned')" 0
check "serie diaria (30 pts)" "$(req GET "/admin/metrics/series?metric=users&period=daily" "$OWNER" | j 'o.data.length')" 30
check "serie semanal (12 pts)" "$(req GET "/admin/metrics/series?metric=topics&period=weekly" "$OWNER" | j 'o.data.length')" 12
check "serie mensal (12 pts)" "$(req GET "/admin/metrics/series?metric=replies&period=monthly" "$OWNER" | j 'o.data.length')" 12
check "serie diaria inclui hoje" "$(req GET "/admin/metrics/series?metric=topics&period=daily" "$OWNER" | j 'o.data[29].value>=1')" true
check "metrica invalida" "$(code GET "/admin/metrics/series?metric=passwords" "$OWNER")" 422
check "membro sem metricas" "$(code GET /admin/metrics "$MEMBER")" 403

echo "== Etapa 6: listagens"
check "usuarios: busca" "$(req GET "/admin/users?q=membr" "$OWNER" | j 'o.data[0].username')" membro
check "usuarios: filtro cargo" "$(req GET "/admin/users?role=moderator" "$OWNER" | j 'o.data.map(u=>u.username).join()')" moder
check "usuarios: ordena por nome" "$(req GET "/admin/users?sort=username&order=asc&limit=1" "$OWNER" | j 'o.data[0].username')" dono
check "usuarios: paginacao" "$(req GET "/admin/users?limit=1&page=2" "$OWNER" | j 'o.data.length+","+o.meta.page')" "1,2"
check "usuarios: ordenacao invalida" "$(code GET "/admin/users?sort=password_hash" "$OWNER")" 422
check "topicos: busca" "$(req GET "/admin/topics?q=e2e&sort=replies" "$OWNER" | j 'o.data[0].title')" "Topico de teste e2e"
check "respostas: busca" "$(req GET "/admin/replies?q=volta&state=visible" "$OWNER" | j 'o.data[0].author')" membro
check "advertencias: busca" "$(req GET "/admin/warnings?q=membro" "$OWNER" | j 'o.data[0].username')" membro
check "armazenamento" "$(req GET /admin/storage "$OWNER" | j 'o.data.provider')" local

echo "== Etapa 6: white-label"
ORIG=$(req GET /site/settings "" | j 'JSON.stringify({"brand.name":o.data["brand.name"],"theme.primary":o.data["theme.primary"]})')
check "cor invalida rejeitada" "$(code PUT /admin/site/settings "$OWNER" '{"theme.primary":"vermelho"}')" 422
check "link inseguro rejeitado" "$(code PUT /admin/site/settings "$OWNER" '{"footer.links":"x|javascript:alert(1)"}')" 422
req PUT /admin/site/settings "$OWNER" '{"brand.name":"Marca E2E","theme.primary":"#ABC"}' >/dev/null
check "publico reflete identidade" "$(req GET /site/settings "" | j 'o.data["brand.name"]+" "+o.data["theme.primary"]')" "Marca E2E #aabbcc"
req PUT /admin/site/settings "$OWNER" "$ORIG" >/dev/null
printf '\x89PNG\r\n\x1a\n\0\0\0\rIHDR\0\0\0\x01\0\0\0\x01\x08\x06\0\0\0\x1f\x15\xc4\x89\0\0\0\rIDATx\x9cc\xf8\x0f\0\0\x01\x01\0\x05\x18\xd8N\0\0\0\0IEND\xaeB`\x82' > e2e-logo.png
URL=$(curl -s -X POST "$B/admin/site/assets" -H "Authorization: Bearer $OWNER" -F "file=@e2e-logo.png" | j o.data.url)
case "$URL" in /*) URL="${1:-http://localhost:8099}$URL";; esac
check "upload de logo" "$(curl -s -o /dev/null -w "%{http_code}" "$URL")" 200
echo "not an image" > e2e-fake.png
check "upload invalido rejeitado" "$(curl -s -o /dev/null -w "%{http_code}" -X POST "$B/admin/site/assets" -H "Authorization: Bearer $OWNER" -F "file=@e2e-fake.png")" 400
check "membro nao envia marca" "$(curl -s -o /dev/null -w "%{http_code}" -X POST "$B/admin/site/assets" -H "Authorization: Bearer $MEMBER" -F "file=@e2e-logo.png")" 403
rm -f e2e-logo.png e2e-fake.png

echo "== RESULTADO: $PASS passaram, $FAIL falharam"
