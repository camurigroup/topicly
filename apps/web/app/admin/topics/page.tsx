"use client";

import { Lock, Pin, Sparkles, Trash2 } from "lucide-react";
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
import { api, type Category } from "@/lib/api";
import { adminApi, type AdminTopic } from "@/lib/admin-api";
import { cn } from "@/lib/utils";

const LIMIT = 20;

export default function AdminTopicsPage() {
  const { hasPermission } = useAuth();
  const list = useListState("created", { category: "", flag: "" });
  const { state } = list;
  const [items, setItems] = useState<AdminTopic[]>([]);
  const [total, setTotal] = useState(0);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .categories()
      .then(({ data }) => setCategories(data))
      .catch(() => setCategories([]));
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminApi.topics({ q: state.q, sort: state.sort, order: state.order, page: state.page, ...state.filters });
      setItems(res.items);
      setTotal(res.meta.total);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao carregar tópicos");
    } finally {
      setLoading(false);
    }
  }, [state]);

  useEffect(() => {
    void load();
  }, [load]);

  async function act(fn: () => Promise<unknown>) {
    try {
      await fn();
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "ação falhou");
    }
  }

  return (
    <ModerationShell
      title="Tópicos"
      description="Todos os tópicos do fórum. Fixar, destacar, bloquear e excluir ficam registrados na auditoria."
      perms={["topics.edit", "topics.delete", "topics.lock", "topics.pin"]}
    >
      <Toolbar>
        <SearchBox value={list.input} onChange={list.setInput} placeholder="Buscar pelo título" />
        <FilterSelect
          label="Categoria"
          value={state.filters.category}
          onChange={(v) => list.setFilter("category", v)}
          options={[{ value: "", label: "Todas as categorias" }, ...categories.map((c) => ({ value: c.slug, label: c.name }))]}
        />
        <FilterSelect
          label="Marcação"
          value={state.filters.flag}
          onChange={(v) => list.setFilter("flag", v)}
          options={[
            { value: "", label: "Qualquer" },
            { value: "pinned", label: "Fixados" },
            { value: "featured", label: "Destacados" },
            { value: "locked", label: "Bloqueados" },
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
            <SortHeader label="Tópico" sortKey="title" state={state} onSort={list.toggleSort} />
            <SortHeader label="Respostas" sortKey="replies" state={state} onSort={list.toggleSort} className="text-right" />
            <SortHeader label="Curtidas" sortKey="likes" state={state} onSort={list.toggleSort} className="text-right" />
            <SortHeader label="Views" sortKey="views" state={state} onSort={list.toggleSort} className="text-right" />
            <SortHeader label="Atividade" sortKey="activity" state={state} onSort={list.toggleSort} />
            <Th className="text-right">Ações</Th>
          </>
        }
      >
        {items.map((t) => (
          <tr key={t.id} className="hover:bg-secondary/40">
            <td className="max-w-md px-4 py-3">
              <Link href={`/t/${t.id}`} target="_blank" className="font-medium hover:underline">
                {t.title}
              </Link>
              <div className="mt-1 flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
                <span>{t.categoryName}</span>
                <span>· @{t.author}</span>
                <span>· {fmtDate(t.createdAt)}</span>
                {t.isPinned && <Badge tone="info">fixado</Badge>}
                {t.isFeatured && <Badge tone="ok">destaque</Badge>}
                {t.isLocked && <Badge tone="warn">bloqueado</Badge>}
              </div>
            </td>
            <td className="px-4 py-3 text-right tabular-nums">{t.replies}</td>
            <td className="px-4 py-3 text-right tabular-nums">{t.likes}</td>
            <td className="px-4 py-3 text-right tabular-nums">{t.views}</td>
            <td className="px-4 py-3 text-muted-foreground">{fmtDate(t.lastActivityAt)}</td>
            <td className="px-4 py-3">
              <div className="flex justify-end gap-1">
                {hasPermission("topics.pin") && (
                  <>
                    <IconAction
                      label={t.isPinned ? "Desafixar" : "Fixar"}
                      on={t.isPinned}
                      onClick={() => act(() => api.setTopicPin(t.id, !t.isPinned))}
                    >
                      <Pin className="h-4 w-4" />
                    </IconAction>
                    <IconAction
                      label={t.isFeatured ? "Remover destaque" : "Destacar"}
                      on={t.isFeatured}
                      onClick={() => act(() => api.setTopicFeature(t.id, !t.isFeatured))}
                    >
                      <Sparkles className="h-4 w-4" />
                    </IconAction>
                  </>
                )}
                {hasPermission("topics.lock") && (
                  <IconAction
                    label={t.isLocked ? "Desbloquear" : "Bloquear"}
                    on={t.isLocked}
                    onClick={() => act(() => api.setTopicLock(t.id, !t.isLocked))}
                  >
                    <Lock className="h-4 w-4" />
                  </IconAction>
                )}
                {hasPermission("topics.delete") && (
                  <IconAction
                    label="Excluir"
                    danger
                    onClick={() =>
                      window.confirm(`Excluir o tópico "${t.title}" e todas as respostas?`) && act(() => api.deleteTopic(t.id))
                    }
                  >
                    <Trash2 className="h-4 w-4" />
                  </IconAction>
                )}
              </div>
            </td>
          </tr>
        ))}
      </TableShell>
      <Pager page={state.page} total={total} limit={LIMIT} onPage={list.setPage} />
    </ModerationShell>
  );
}

interface IconActionProps {
  label: string;
  on?: boolean;
  danger?: boolean;
  onClick: () => void;
  children: React.ReactNode;
}

function IconAction({ label, on, danger, onClick, children }: IconActionProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      aria-pressed={on}
      title={label}
      className={cn(
        "rounded-md p-1.5 transition-colors hover:bg-secondary",
        on ? "text-primary" : "text-muted-foreground",
        danger && "hover:text-destructive",
      )}
    >
      {children}
    </button>
  );
}
