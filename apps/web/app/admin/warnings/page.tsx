"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

import { fmtDate, Pager, SearchBox, SortHeader, TableShell, Th, Toolbar, useListState } from "@/components/admin/list-kit";
import { ModerationShell } from "@/components/moderation-nav";
import { adminApi, type AdminWarning } from "@/lib/admin-api";

const LIMIT = 20;

export default function WarningsPage() {
  const list = useListState("created", {});
  const { state } = list;
  const [items, setItems] = useState<AdminWarning[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminApi.warnings({ q: state.q, order: state.order, page: state.page });
      setItems(res.items);
      setTotal(res.meta.total);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao carregar advertências");
    } finally {
      setLoading(false);
    }
  }, [state]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <ModerationShell
      title="Advertências"
      description="Avisos formais enviados aos membros. Para advertir alguém, abra a ficha do usuário."
      perms={["warnings.apply", "users.ban", "users.mute"]}
    >
      <Toolbar>
        <SearchBox value={list.input} onChange={list.setInput} placeholder="Buscar por @usuário ou motivo" />
      </Toolbar>
      <TableShell
        loading={loading}
        empty={items.length === 0}
        colSpan={4}
        error={error}
        head={
          <>
            <Th>Usuário</Th>
            <Th>Motivo</Th>
            <Th>Aplicada por</Th>
            <SortHeader label="Data" sortKey="created" state={state} onSort={list.toggleSort} />
          </>
        }
      >
        {items.map((w) => (
          <tr key={w.id} className="hover:bg-secondary/40">
            <td className="px-4 py-3">
              <Link href={`/admin/users?u=${encodeURIComponent(w.username)}`} className="font-medium hover:underline">
                {w.displayName}
              </Link>
              <span className="block text-xs text-muted-foreground">@{w.username}</span>
            </td>
            <td className="max-w-md px-4 py-3 text-muted-foreground">{w.reason}</td>
            <td className="px-4 py-3 text-muted-foreground">{w.issuedBy ? `@${w.issuedBy}` : "—"}</td>
            <td className="px-4 py-3 text-muted-foreground">{fmtDate(w.createdAt)}</td>
          </tr>
        ))}
      </TableShell>
      <Pager page={state.page} total={total} limit={LIMIT} onPage={list.setPage} />
    </ModerationShell>
  );
}
