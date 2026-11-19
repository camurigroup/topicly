// Generates docs/openapi.yaml from apps/api/internal/server/router.go so the
// API reference always matches the registered routes.
//   node scripts/gen-openapi.mjs
import { readFileSync, writeFileSync } from "node:fs";

const src = readFileSync("apps/api/internal/server/router.go", "utf8");
const permConsts = Object.fromEntries(
  [...readFileSync("apps/api/internal/modules/rbac/model.go", "utf8").matchAll(/(Perm\w+)\s*=\s*"([^"]+)"/g)].map((m) => [m[1], m[2]]),
);

const routes = [];
for (const line of src.split("\n")) {
  const m = line.match(/api\.Handle(?:Func)?\("(GET|POST|PUT|PATCH|DELETE) ([^"]+)",\s*(.*)\)\s*$/);
  if (!m) continue;
  const [, method, path, rest] = m;
  const perms = [...rest.matchAll(/rbac\.(Perm\w+)/g)].map((p) => permConsts[p[1]]).filter(Boolean);
  const auth = /s\.perm\(|s\.anyPerm\(|s\.write\(|\bauth\(|RequireAuth/.test(rest);
  const optional = /\boptional\(/.test(rest);
  const handler = (rest.match(/\.(\w+)\)*\s*$/) || [])[1] || "";
  routes.push({ method, path, perms, auth, optional, handler, anyOf: /anyPerm/.test(rest) });
}

const tagOf = (p) => {
  const seg = p.split("/").filter(Boolean);
  return seg[0] === "admin" ? `admin/${seg[1] ?? ""}` : seg[0];
};
const words = (s) => s.replace(/([a-z])([A-Z])/g, "$1 $2").toLowerCase();

let y = `openapi: 3.0.3
info:
  title: Forum Framework API
  version: "1.0.0"
  description: |
    API REST do Forum Framework. Todas as respostas usam o envelope
    {success, data, error, meta}. Rotas protegidas usam Bearer JWT
    (POST /auth/login). Este arquivo é gerado por scripts/gen-openapi.mjs.
servers:
  - url: http://localhost:8080/api/v1
components:
  securitySchemes:
    bearer:
      type: http
      scheme: bearer
      bearerFormat: JWT
  schemas:
    Envelope:
      type: object
      properties:
        success: { type: boolean }
        data: {}
        error:
          type: object
          nullable: true
          properties:
            code: { type: string }
            message: { type: string }
        meta:
          type: object
          nullable: true
          properties:
            total: { type: integer }
            page: { type: integer }
            limit: { type: integer }
  responses:
    Ok:
      description: Sucesso
      content:
        application/json:
          schema: { $ref: "#/components/schemas/Envelope" }
    Error:
      description: Erro (400 corpo inválido, 401, 403, 404, 409, 422 validação, 429 limite)
      content:
        application/json:
          schema: { $ref: "#/components/schemas/Envelope" }
paths:
`;

const byPath = new Map();
for (const r of routes) {
  const p = r.path.replace(/\{(\w+)\}/g, "{$1}");
  if (!byPath.has(p)) byPath.set(p, []);
  byPath.get(p).push(r);
}
for (const [path, ops] of [...byPath.entries()].sort()) {
  y += `  ${path}:\n`;
  const params = [...path.matchAll(/\{(\w+)\}/g)].map((m) => m[1]);
  for (const r of ops) {
    const desc = [];
    if (r.perms.length) desc.push(`Permissão ${r.anyOf ? "(qualquer uma)" : "exigida"}: ${r.perms.join(", ")}.`);
    if (r.optional) desc.push("Autenticação opcional (enriquece a resposta para usuários logados).");
    y += `    ${r.method.toLowerCase()}:\n`;
    y += `      tags: ["${tagOf(path)}"]\n`;
    y += `      summary: "${words(r.handler)}"\n`;
    y += `      operationId: "${r.method.toLowerCase()}_${path.replace(/[^\w]+/g, "_").replace(/^_|_$/g, "")}"\n`;
    if (desc.length) y += `      description: "${desc.join(" ")}"\n`;
    if (r.auth) y += `      security: [{ bearer: [] }]\n`;
    if (params.length) {
      y += `      parameters:\n`;
      for (const p of params) y += `        - { name: ${p}, in: path, required: true, schema: { type: string } }\n`;
    }
    y += `      responses:\n        "200": { $ref: "#/components/responses/Ok" }\n        default: { $ref: "#/components/responses/Error" }\n`;
  }
}
writeFileSync("docs/openapi.yaml", y);
console.log(`docs/openapi.yaml: ${routes.length} operações em ${byPath.size} caminhos`);
