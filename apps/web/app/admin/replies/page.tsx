"use client";

import { Trash2 } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

import {
  Badge,
  FilterSelect,
  fmtDate,
  Pager,
  SearchBox,
  SortHeader,
  TableShell,
  Th,
  Toolbar,
  useListState,
} from "@/components/admin/list-kit";
import { useAuth } from "@/components/auth-provider";
import { ModerationShell } from "@/components/moderation-nav";
import { api } from "@/lib/api";
import { adminApi, type AdminReply } from "@/lib/admin-api";

const LIMIT = 20;

export default function AdminRepliesPage() {
  const { hasPermission } = useAuth();
  const list = useListState("created", { state: "visible" });
  const { state } = list;
  const [items, setItems] = useState<AdminReply[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminApi.replies({ q: state.q, sort: state.sort, order: state.order, page: state.page, ...state.filters });
      setItems(res.items);
      setTotal(res.meta.total);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao carregar respostas");
    } finally {
      setLoading(false);
    }
  }, [state]);

  useEffect(() => {
    void load();
  }, [load]);

  async function remove(r: AdminReply) {
    if (!window.confirm("Excluir esta resposta?")) return;
    try {
      await api.deleteReply(r.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao excluir");
    }
  }

  return (
    <ModerationShell
      title="Respostas"
      description="Respostas de todos os tópicos. Excluir conteúdo de outra pessoa fica registrado na auditoria."
      perms={["replies.edit", "replies.delete"]}
    >
      <Toolbar>
        <SearchBox value={list.input} onChange={list.setInput} placeholder="Buscar no conteúdo" />
        <FilterSelect
          label="Estado"
          value={state.filters.state}
          onChange={(v) => list.setFilter("state", v)}
          options={[
            { value: "visible", label: "Visíveis" },
            { value: "deleted", label: "Excluídas" },
            { value: "", label: "Todas" },
          ]}
        />
      </Toolbar>

      <TableShell
        loading={loading}
        empty={items.length === 0}
        colSpan={4}
        error={error}
        head={
          <>
            <Th>Resposta</Th>
            <SortHeader label="Curtidas" sortKey="likes" state={state} onSort={list.toggleSort} className="text-right" />
            <SortHeader label="Data" sortKey="created" state={state} onSort={list.toggleSort} />
            <Th className="text-right">Ações</Th>
          </>
        }
      >
        {items.map((r) => (
          <tr key={r.id} className="hover:bg-secondary/40">
            <td className="max-w-xl px-4 py-3">
              <p className={r.isDeleted ? "text-muted-foreground line-through" : ""}>{r.excerpt || "—"}</p>
              <p className="mt-1 text-xs text-muted-foreground">
                @{r.author} em{" "}
                <Link href={`/t/${r.topicId}#reply-${r.id}`} target="_blank" className="hover:underline">
                  “{r.topicTitle}”
                </Link>{" "}
                {r.isDeleted && <Badge tone="danger">excluída</Badge>}
              </p>
            </td>
            <td className="px-4 py-3 text-right tabular-nums">{r.likes}</td>
            <td className="px-4 py-3 text-muted-foreground">{fmtDate(r.createdAt)}</td>
            <td className="px-4 py-3 text-right">
              {hasPermission("replies.delete") && !r.isDeleted && (
                <button
                  type="button"
                  aria-label="Excluir resposta"
                  title="Excluir"
                  onClick={() => remove(r)}
                  className="rounded-md p-1.5 text-muted-foreground hover:bg-secondary hover:text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              )}
            </td>
          </tr>
        ))}
      </TableShell>
      <Pager page={state.page} total={total} limit={LIMIT} onPage={list.setPage} />
    </ModerationShell>
  );
}
