"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

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
import { UserAvatar } from "@/components/user-avatar";
import { adminApi, type AdminUser } from "@/lib/admin-api";
import { modApi, type RoleDetail } from "@/lib/moderation-api";

const LIMIT = 20;

/** Admin users list: search, role/status filters, sortable columns, pages. */
export function UsersTable() {
  const { hasPermission } = useAuth();
  const list = useListState("created", { role: "", status: "" });
  const { state } = list;
  const [items, setItems] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [roles, setRoles] = useState<RoleDetail[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const canRoles = hasPermission("roles.manage");

  useEffect(() => {
    if (!canRoles) return;
    modApi
      .roles()
      .then(({ data }) => setRoles(data))
      .catch(() => setRoles([]));
  }, [canRoles]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    adminApi
      .users({ q: state.q, sort: state.sort, order: state.order, page: state.page, ...state.filters })
      .then((res) => {
        if (!active) return;
        setItems(res.items);
        setTotal(res.meta.total);
        setError(null);
      })
      .catch((err) => active && setError(err instanceof Error ? err.message : "falha ao carregar usuários"))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [state]);

  return (
    <div className="space-y-4">
      <Toolbar>
        <SearchBox value={list.input} onChange={list.setInput} placeholder="Buscar por nome, @usuário ou e-mail" />
        <FilterSelect
          label="Cargo"
          value={state.filters.role}
          onChange={(v) => list.setFilter("role", v)}
          options={[
            { value: "", label: "Todos os cargos" },
            { value: "member", label: "Só membros" },
            ...roles.map((r) => ({ value: r.slug, label: r.name })),
          ]}
        />
        <FilterSelect
          label="Situação"
          value={state.filters.status}
          onChange={(v) => list.setFilter("status", v)}
          options={[
            { value: "", label: "Todas" },
            { value: "active", label: "Sem banimento" },
            { value: "banned", label: "Banidos" },
            { value: "muted", label: "Silenciados" },
          ]}
        />
      </Toolbar>

      <TableShell
        loading={loading}
        empty={items.length === 0}
        colSpan={6}
        error={error}
        head={
          <>
            <SortHeader label="Usuário" sortKey="username" state={state} onSort={list.toggleSort} />
            <Th>Cargos</Th>
            <SortHeader label="Tópicos" sortKey="topics" state={state} onSort={list.toggleSort} className="text-right" />
            <SortHeader label="Reputação" sortKey="reputation" state={state} onSort={list.toggleSort} className="text-right" />
            <SortHeader label="Último acesso" sortKey="active" state={state} onSort={list.toggleSort} />
            <SortHeader label="Cadastro" sortKey="created" state={state} onSort={list.toggleSort} />
          </>
        }
      >
        {items.map((u) => (
          <tr key={u.id} className="hover:bg-secondary/40">
            <td className="px-4 py-3">
              <Link href={`/admin/users?u=${encodeURIComponent(u.username)}`} className="flex items-center gap-3">
                <UserAvatar name={u.displayName} url={u.avatarUrl} />
                <span className="min-w-0">
                  <span className="block truncate font-medium hover:underline">{u.displayName}</span>
                  <span className="block truncate text-xs text-muted-foreground">
                    @{u.username} · {u.email}
                  </span>
                </span>
              </Link>
            </td>
            <td className="px-4 py-3">
              <div className="flex flex-wrap gap-1">
                {u.roles.length === 0 && <Badge>membro</Badge>}
                {u.roles.map((r) => (
                  <Badge key={r} tone="info">
                    {r}
                  </Badge>
                ))}
                {u.isBanned && <Badge tone="danger">banido</Badge>}
                {u.isMuted && <Badge tone="warn">silenciado</Badge>}
              </div>
            </td>
            <td className="px-4 py-3 text-right tabular-nums">
              {u.topics}
              <span className="block text-xs text-muted-foreground">{u.replies} resp.</span>
            </td>
            <td className="px-4 py-3 text-right tabular-nums">{u.reputation}</td>
            <td className="px-4 py-3 text-muted-foreground">{fmtDate(u.lastActiveAt)}</td>
            <td className="px-4 py-3 text-muted-foreground">{fmtDate(u.createdAt)}</td>
          </tr>
        ))}
      </TableShell>
      <Pager page={state.page} total={total} limit={LIMIT} onPage={list.setPage} />
    </div>
  );
}
