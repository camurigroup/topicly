"use client";

import { Ban, VolumeX } from "lucide-react";
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
import { ModerationShell } from "@/components/moderation-nav";
import { modApi, type Sanction, type SanctionType } from "@/lib/moderation-api";

const LIMIT = 20;

export default function SanctionsPage() {
  const list = useListState("created", { type: "", active: "1" });
  const { state } = list;
  const [items, setItems] = useState<Sanction[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await modApi.sanctions({
        q: state.q,
        order: state.order,
        page: state.page,
        type: (state.filters.type || undefined) as SanctionType | undefined,
        active: state.filters.active === "1",
      });
      setItems(res.items);
      setTotal(res.meta.total);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao carregar sanções");
    } finally {
      setLoading(false);
    }
  }, [state]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <ModerationShell
      title="Banimentos e mutes"
      description="Sanções aplicadas pela moderação. Para revogar, abra a ficha do usuário."
      perms={["users.ban", "users.mute"]}
    >
      <Toolbar>
        <SearchBox value={list.input} onChange={list.setInput} placeholder="Buscar por @usuário ou motivo" />
        <FilterSelect
          label="Tipo"
          value={state.filters.type}
          onChange={(v) => list.setFilter("type", v)}
          options={[
            { value: "", label: "Banimentos e mutes" },
            { value: "ban", label: "Só banimentos" },
            { value: "mute", label: "Só mutes" },
          ]}
        />
        <FilterSelect
          label="Situação"
          value={state.filters.active}
          onChange={(v) => list.setFilter("active", v)}
          options={[
            { value: "1", label: "Em vigor" },
            { value: "", label: "Todo o histórico" },
          ]}
        />
      </Toolbar>

      <TableShell
        loading={loading}
        empty={items.length === 0}
        colSpan={5}
        error={error}
        head={
          <>
            <Th>Usuário</Th>
            <Th>Motivo</Th>
            <Th>Vigência</Th>
            <Th>Aplicada por</Th>
            <SortHeader label="Data" sortKey="created" state={state} onSort={list.toggleSort} />
          </>
        }
      >
        {items.map((s) => (
          <tr key={s.id} className="hover:bg-secondary/40">
            <td className="px-4 py-3">
              <Link
                href={`/admin/users?u=${encodeURIComponent(s.user.username)}`}
                className="flex items-center gap-2 font-medium hover:underline"
              >
                {s.type === "ban" ? <Ban className="h-4 w-4 text-destructive" /> : <VolumeX className="h-4 w-4 text-amber-500" />}
                @{s.user.username}
              </Link>
            </td>
            <td className="max-w-xs px-4 py-3 text-muted-foreground">{s.reason}</td>
            <td className="px-4 py-3">
              <div className="flex flex-wrap items-center gap-1.5">
                {s.active ? <Badge tone="danger">em vigor</Badge> : s.revokedAt ? <Badge>revogada</Badge> : <Badge>expirada</Badge>}
                <span className="text-xs text-muted-foreground">{s.expiresAt ? `até ${fmtDate(s.expiresAt)}` : "permanente"}</span>
              </div>
            </td>
            <td className="px-4 py-3 text-muted-foreground">{s.issuedBy ? `@${s.issuedBy.username}` : "—"}</td>
            <td className="px-4 py-3 text-muted-foreground">{fmtDate(s.createdAt)}</td>
          </tr>
        ))}
      </TableShell>
      <Pager page={state.page} total={total} limit={LIMIT} onPage={list.setPage} />
    </ModerationShell>
  );
}
