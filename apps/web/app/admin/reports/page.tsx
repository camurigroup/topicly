"use client";

import { ExternalLink, Gavel, Loader2 } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

import { FilterSelect, SearchBox } from "@/components/admin/list-kit";
import { ModerationShell } from "@/components/moderation-nav";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  modApi,
  REPORT_REASONS,
  REPORT_STATUS_LABEL,
  type Report,
  type ReportStatus,
  type ReportTargetType,
} from "@/lib/moderation-api";
import { cn, timeAgo } from "@/lib/utils";

const STATUSES: ReportStatus[] = ["pending", "reviewing", "resolved", "rejected"];
const TYPE_LABEL: Record<ReportTargetType, string> = { user: "Usuário", topic: "Tópico", reply: "Resposta" };
const STATUS_TONE: Record<ReportStatus, string> = {
  pending: "bg-amber-500/15 text-amber-700 dark:text-amber-300",
  reviewing: "bg-sky-500/15 text-sky-700 dark:text-sky-300",
  resolved: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300",
  rejected: "bg-muted text-muted-foreground",
};
const EMPTY_MESSAGE: Record<ReportStatus, string> = {
  pending: "Nenhuma denúncia pendente. Tudo em dia!",
  reviewing: "Nenhuma denúncia em análise.",
  resolved: "Nenhuma denúncia resolvida ainda.",
  rejected: "Nenhuma denúncia rejeitada.",
};
const reasonLabel =(r: string) => REPORT_REASONS.find((x) => x.value === r)?.label ?? r;
const PAGE_SIZE = 20;

export default function ReportsPage() {
  const [status, setStatus] = useState<ReportStatus>("pending");
  const [type, setType] = useState<ReportTargetType | "">("");
  const [page, setPage] = useState(1);
  const [items, setItems] = useState<Report[]>([]);
  const [total, setTotal] = useState(0);
  const [counts, setCounts] = useState<Record<ReportStatus, number> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [q, setQ] = useState("");
  const [order, setOrder] = useState<"" | "asc" | "desc">("");

  useEffect(() => {
    const t = setTimeout(() => {
      setQ(search.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [search]);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [list, c] = await Promise.all([
        modApi.reports({ status, targetType: type || undefined, page, q: q || undefined, order: order || undefined }),
        modApi.reportCounts(),
      ]);
      setItems(list.items);
      setTotal(list.meta.total);
      setCounts(c.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "erro ao carregar denúncias");
    } finally {
      setLoading(false);
    }
  }, [status, type, page, q, order]);

  useEffect(() => {
    void load();
  }, [load]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <ModerationShell
      title="Central de denúncias"
      description="Analise o que a comunidade sinalizou. Fechar uma denúncia avisa quem a enviou."
      perms={["reports.manage"]}
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div role="tablist" aria-label="Status" className="flex flex-wrap gap-1 rounded-lg bg-secondary p-1">
          {STATUSES.map((s) => (
            <button
              key={s}
              role="tab"
              aria-selected={status === s}
              onClick={() => {
                setStatus(s);
                setPage(1);
              }}
              className={cn(
                "inline-flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-all",
                status === s ? "bg-background shadow-sm" : "text-muted-foreground hover:text-foreground",
              )}
            >
              {REPORT_STATUS_LABEL[s]}
              {counts && (
                <span className={cn("rounded-full px-1.5 text-xs tabular-nums", STATUS_TONE[s])}>{counts[s]}</span>
              )}
            </button>
          ))}
        </div>
        <SearchBox value={search} onChange={setSearch} placeholder="Buscar motivo, detalhes ou @denunciante" />
        <FilterSelect
          label="Ordem"
          value={order}
          onChange={(v) => {
            setOrder(v as "" | "asc" | "desc");
            setPage(1);
          }}
          options={[
            { value: "", label: "Padrão da fila" },
            { value: "desc", label: "Mais recentes" },
            { value: "asc", label: "Mais antigas" },
          ]}
        />
        <label className="flex items-center gap-2 text-sm text-muted-foreground">
          Tipo
          <select
            value={type}
            onChange={(e) => {
              setType(e.target.value as ReportTargetType | "");
              setPage(1);
            }}
            className="h-9 rounded-md border border-input bg-background px-2 text-sm text-foreground"
          >
            <option value="">Todos</option>
            <option value="topic">Tópicos</option>
            <option value="reply">Respostas</option>
            <option value="user">Usuários</option>
          </select>
        </label>
      </div>

      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}

      {loading ? (
        <div className="flex justify-center py-16 text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border py-16 text-center">
          <p className="font-medium">Nada por aqui.</p>
          <p className="mt-1 text-sm text-muted-foreground">{EMPTY_MESSAGE[status]}</p>
        </div>
      ) : (
        <ul className="space-y-3">
          {items.map((r) => (
            <ReportCard key={r.id} report={r} onChanged={load} />
          ))}
        </ul>
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

function ReportCard({ report: r, onChanged }: { report: Report; onChanged: () => Promise<void> }) {
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const open = r.status === "pending" || r.status === "reviewing";

  async function move(status: ReportStatus) {
    setBusy(true);
    setError(null);
    try {
      await modApi.updateReport(r.id, { status, note: note.trim() || undefined });
      await onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao atualizar");
      setBusy(false);
    }
  }

  return (
    <li className="group relative overflow-hidden rounded-xl border border-border bg-card p-5 text-card-foreground transition-shadow hover:shadow-md">
      <span className={cn("absolute inset-y-0 left-0 w-1", open ? "bg-amber-500" : "bg-border")} aria-hidden />
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2 text-xs">
            <span className="rounded-md border border-border px-1.5 py-0.5 font-medium uppercase tracking-wide">
              {TYPE_LABEL[r.target.type]}
            </span>
            <span className="font-semibold text-destructive">{reasonLabel(r.reason)}</span>
            <span className={cn("rounded-full px-2 py-0.5", STATUS_TONE[r.status])}>{REPORT_STATUS_LABEL[r.status]}</span>
            <span className="text-muted-foreground">· {timeAgo(r.createdAt)}</span>
          </div>
          {r.target.exists ? (
            <Link href={r.target.link} target="_blank" className="mt-2 inline-flex items-center gap-1.5 font-semibold hover:underline">
              {r.target.title}
              <ExternalLink className="h-3.5 w-3.5 text-muted-foreground" />
            </Link>
          ) : (
            <p className="mt-2 font-semibold text-muted-foreground line-through">{r.target.title || "Conteúdo removido"}</p>
          )}
          {r.target.excerpt && (
            <blockquote className="mt-2 line-clamp-3 border-l-2 border-border pl-3 text-sm text-muted-foreground">
              {r.target.excerpt}
            </blockquote>
          )}
          {r.details && <p className="mt-3 text-sm">“{r.details}”</p>}
          <p className="mt-2 text-xs text-muted-foreground">
            Enviada por {r.reporter ? `@${r.reporter.username}` : "usuário removido"}
            {r.handledBy && ` · tratada por @${r.handledBy.username}`}
            {r.resolutionNote && ` · nota: ${r.resolutionNote}`}
          </p>
        </div>
        {r.target.author && (
          <Button variant="outline" size="sm" asChild>
            <Link href={`/admin/users?u=${encodeURIComponent(r.target.author)}`}>
              <Gavel className="h-4 w-4" />
              Moderar @{r.target.author}
            </Link>
          </Button>
        )}
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-border pt-4">
        <Input
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="Nota da moderação (opcional)"
          maxLength={1000}
          className="h-9 min-w-0 flex-1 basis-56"
          aria-label="Nota da moderação"
        />
        {r.status === "pending" && (
          <Button size="sm" variant="outline" disabled={busy} onClick={() => move("reviewing")}>
            Analisar
          </Button>
        )}
        {open ? (
          <>
            <Button size="sm" disabled={busy} onClick={() => move("resolved")}>
              Resolver
            </Button>
            <Button size="sm" variant="ghost" disabled={busy} onClick={() => move("rejected")}>
              Rejeitar
            </Button>
          </>
        ) : (
          <Button size="sm" variant="outline" disabled={busy} onClick={() => move("pending")}>
            Reabrir
          </Button>
        )}
        {busy && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
      </div>
      {error && <p role="alert" className="mt-2 text-sm text-destructive">{error}</p>}
    </li>
  );
}
