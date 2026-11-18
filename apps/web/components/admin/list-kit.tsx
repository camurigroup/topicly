"use client";

import { ArrowDown, ArrowUp, ArrowUpDown, Loader2, Search } from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export type Order = "asc" | "desc";

export interface ListState<F extends Record<string, string>> {
  q: string;
  sort: string;
  order: Order;
  page: number;
  filters: F;
}

/**
 * Search/sort/filter/page state for admin listings. The search term is
 * debounced; any change other than the page resets to page 1.
 */
export function useListState<F extends Record<string, string>>(defaultSort: string, filters: F) {
  const [input, setInput] = useState("");
  const [state, setState] = useState<ListState<F>>({ q: "", sort: defaultSort, order: "desc", page: 1, filters });

  useEffect(() => {
    const t = setTimeout(() => setState((s) => (s.q === input.trim() ? s : { ...s, q: input.trim(), page: 1 })), 300);
    return () => clearTimeout(t);
  }, [input]);

  return {
    state,
    input,
    setInput,
    setFilter: (key: keyof F, value: string) =>
      setState((s) => ({ ...s, page: 1, filters: { ...s.filters, [key]: value } })),
    toggleSort: (key: string) =>
      setState((s) => ({
        ...s,
        page: 1,
        sort: key,
        order: s.sort === key && s.order === "desc" ? "asc" : "desc",
      })),
    setPage: (page: number) => setState((s) => ({ ...s, page })),
  };
}

export function Toolbar({ children }: { children: ReactNode }) {
  return <div className="flex flex-wrap items-center gap-2">{children}</div>;
}

export function SearchBox({ value, onChange, placeholder }: { value: string; onChange: (v: string) => void; placeholder: string }) {
  return (
    <div className="relative min-w-0 flex-1 basis-56">
      <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="h-9 pl-9"
        aria-label={placeholder}
        maxLength={100}
      />
    </div>
  );
}

interface FilterSelectProps {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
}

export function FilterSelect({ label, value, onChange, options }: FilterSelectProps) {
  return (
    <label className="flex items-center gap-2 text-sm text-muted-foreground">
      <span className="sr-only sm:not-sr-only">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-9 rounded-md border border-input bg-background px-2 text-sm text-foreground"
        aria-label={label}
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </label>
  );
}

interface SortHeaderProps {
  label: string;
  sortKey: string;
  state: { sort: string; order: Order };
  onSort: (key: string) => void;
  className?: string;
}

export function SortHeader({ label, sortKey, state, onSort, className }: SortHeaderProps) {
  const active = state.sort === sortKey;
  const Icon = !active ? ArrowUpDown : state.order === "asc" ? ArrowUp : ArrowDown;
  return (
    <th
      scope="col"
      className={cn("px-4 py-2.5 font-medium", className)}
      aria-sort={active ? (state.order === "asc" ? "ascending" : "descending") : "none"}
    >
      <button
        type="button"
        onClick={() => onSort(sortKey)}
        className={cn("inline-flex items-center gap-1 hover:text-foreground", active && "text-foreground")}
      >
        {label}
        <Icon className="h-3 w-3" />
      </button>
    </th>
  );
}

export function Th({ children, className }: { children?: ReactNode; className?: string }) {
  return (
    <th scope="col" className={cn("px-4 py-2.5 font-medium", className)}>
      {children}
    </th>
  );
}

interface TableShellProps {
  head: ReactNode;
  loading: boolean;
  empty: boolean;
  colSpan: number;
  error?: string | null;
  children: ReactNode;
}

/** Table frame with loading, empty and error states. */
export function TableShell({ head, loading, empty, colSpan, error, children }: TableShellProps) {
  return (
    <div className="overflow-x-auto rounded-xl border border-border bg-card">
      <table className="w-full min-w-[640px] text-sm">
        <thead className="border-b border-border bg-secondary/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
          <tr>{head}</tr>
        </thead>
        <tbody className={cn("divide-y divide-border transition-opacity", loading && !empty && "opacity-50")}>
          {error ? (
            <tr>
              <td colSpan={colSpan} role="alert" className="px-4 py-10 text-center text-destructive">
                {error}
              </td>
            </tr>
          ) : loading && empty ? (
            <tr>
              <td colSpan={colSpan} className="px-4 py-12 text-center text-muted-foreground">
                <Loader2 className="mx-auto h-5 w-5 animate-spin" />
              </td>
            </tr>
          ) : empty ? (
            <tr>
              <td colSpan={colSpan} className="px-4 py-12 text-center text-muted-foreground">
                Nenhum resultado para os filtros atuais.
              </td>
            </tr>
          ) : (
            children
          )}
        </tbody>
      </table>
    </div>
  );
}

export function Pager({ page, total, limit, onPage }: { page: number; total: number; limit: number; onPage: (p: number) => void }) {
  const pages = Math.max(1, Math.ceil(total / Math.max(1, limit)));
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
      <span className="tabular-nums">
        {total} {total === 1 ? "resultado" : "resultados"}
      </span>
      {pages > 1 && (
        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => onPage(page - 1)}>
            Anterior
          </Button>
          <span className="tabular-nums">
            {page} / {pages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= pages} onClick={() => onPage(page + 1)}>
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
}

export function Badge({ children, tone }: { children: ReactNode; tone?: "danger" | "warn" | "info" | "ok" }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium",
        tone === "danger" && "border-destructive/40 bg-destructive/10 text-destructive",
        tone === "warn" && "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
        tone === "info" && "border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-300",
        tone === "ok" && "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
        !tone && "border-border text-muted-foreground",
      )}
    >
      {children}
    </span>
  );
}

export const fmtDate = (iso: string | null) =>
  iso ? new Date(iso).toLocaleDateString("pt-BR", { day: "2-digit", month: "short", year: "numeric" }) : "—";
