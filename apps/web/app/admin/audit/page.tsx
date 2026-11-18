"use client";

import { Loader2 } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

import { FilterSelect, SearchBox } from "@/components/admin/list-kit";
import { ModerationShell } from "@/components/moderation-nav";
import { Button } from "@/components/ui/button";
import { modApi, type AuditEntry } from "@/lib/moderation-api";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 30;

const ACTION_LABEL: Record<string, string> = {
  "user.ban": "Banimento",
  "user.mute": "Silenciamento",
  "user.warn": "Advertência",
  "sanction.revoke": "Sanção revogada",
  "user.roles": "Cargos do usuário",
  "role.create": "Cargo criado",
  "role.update": "Cargo editado",
  "role.delete": "Cargo excluído",
  "role.permissions": "Permissões do cargo",
  "permission.create": "Permissão criada",
  "permission.toggle": "Permissão (de)ativada",
  "permission.delete": "Permissão excluída",
  "report.status": "Denúncia atualizada",
  "topic.edit": "Tópico editado (moderação)",
  "topic.delete": "Tópico excluído (moderação)",
  "reply.edit": "Resposta editada (moderação)",
  "reply.delete": "Resposta excluída (moderação)",
  "topic.pin": "Tópico fixado/desafixado",
  "topic.lock": "Tópico bloqueado/desbloqueado",
  "topic.feature": "Destaque de tópico",
  "category.create": "Categoria criada",
  "category.update": "Categoria editada",
  "category.delete": "Categoria excluída",
  "site.settings": "Configurações do site",
  "notification.broadcast": "Aviso geral",
};

/** Severity tint per action family. */
function tone(action: string): string {
  if (action === "user.ban" || action.endsWith(".delete")) return "bg-destructive";
  if (action === "user.mute" || action === "user.warn") return "bg-amber-500";
  if (action.startsWith("role.") || action.startsWith("permission.") || action === "user.roles") return "bg-violet-500";
  return "bg-sky-500";
}

const fmt = (iso: string) => new Date(iso).toLocaleString("pt-BR", { dateStyle: "short", timeStyle: "medium" });

function summarize(details: Record<string, unknown> | null): string {
  if (!details) return "";
  return Object.entries(details)
    .filter(([k]) => !["method", "path", "authorId"].includes(k))
    .map(([k, v]) => `${k}: ${typeof v === "object" ? JSON.stringify(v) : String(v)}`)
    .join(" · ");
}

export default function AuditPage() {
  const [action, setAction] = useState("");
  const [actions, setActions] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<AuditEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [q, setQ] = useState("");
  const [order, setOrder] = useState<"asc" | "desc">("desc");

  useEffect(() => {
    const t = setTimeout(() => {
      setQ(search.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [search]);

  useEffect(() => {
    modApi
      .auditActions()
      .then(({ data }) => setActions(data))
      .catch(() => setActions([]));
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await modApi.audit({ action: action || undefined, page, q: q || undefined, order });
      setItems(res.items);
      setTotal(res.meta.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "erro ao carregar auditoria");
    } finally {
      setLoading(false);
    }
  }, [action, page, q, order]);

  useEffect(() => {
    void load();
  }, [load]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <ModerationShell
      title="Auditoria"
      description="Registro de toda ação administrativa: quem fez, o quê, em quem, quando e de onde."
      perms={["logs.view"]}
      actions={
        <label className="flex items-center gap-2 text-sm text-muted-foreground">
          Ação
          <select
            value={action}
            onChange={(e) => {
              setAction(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-input bg-background px-2 text-sm text-foreground"
          >
            <option value="">Todas</option>
            {actions.map((a) => (
              <option key={a} value={a}>
                {ACTION_LABEL[a] ?? a}
              </option>
            ))}
          </select>
        </label>
      }
    >
      <div className="flex flex-wrap items-center gap-2">
        <SearchBox value={search} onChange={setSearch} placeholder="Buscar por @autor ou ação" />
        <FilterSelect
          label="Ordem"
          value={order}
          onChange={(v) => {
            setOrder(v as "asc" | "desc");
            setPage(1);
          }}
          options={[
            { value: "desc", label: "Mais recentes" },
            { value: "asc", label: "Mais antigas" },
          ]}
        />
      </div>
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {loading ? (
        <div className="flex justify-center py-16 text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : items.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border py-16 text-center text-sm text-muted-foreground">
          Nenhum registro.
        </p>
      ) : (
        <ol className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
          {items.map((e) => (
            <li key={e.id} className="flex gap-4 px-4 py-3 text-sm">
              <span className={cn("mt-1.5 h-2 w-2 shrink-0 rounded-full", tone(e.action))} aria-hidden />
              <div className="min-w-0 flex-1">
                <p>
                  {e.actor ? (
                    <Link href={`/u/${e.actor.username}`} className="font-semibold hover:underline">
                      @{e.actor.username}
                    </Link>
                  ) : (
                    <span className="font-semibold text-muted-foreground">sistema</span>
                  )}{" "}
                  <span className="text-muted-foreground">·</span> {ACTION_LABEL[e.action] ?? e.action}
                  {e.targetType && (
                    <span className="ml-1 font-mono text-xs text-muted-foreground">
                      [{e.targetType}
                      {e.targetId ? ` ${e.targetId.slice(0, 8)}` : ""}]
                    </span>
                  )}
                </p>
                {summarize(e.details) && (
                  <p className="mt-0.5 break-words text-xs text-muted-foreground">{summarize(e.details)}</p>
                )}
              </div>
              <div className="shrink-0 text-right text-xs tabular-nums text-muted-foreground">
                <time dateTime={e.createdAt}>{fmt(e.createdAt)}</time>
                {e.ip && <p className="font-mono">{e.ip}</p>}
              </div>
            </li>
          ))}
        </ol>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Anterior
          </Button>
          <span className="text-sm tabular-nums text-muted-foreground">
            {page} / {totalPages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Próxima
          </Button>
        </div>
      )}
    </ModerationShell>
  );
}
