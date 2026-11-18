"use client";

import { Folder, Loader2, Search as SearchIcon, SlidersHorizontal } from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

import { RichText } from "@/components/rich-text";
import { TopicRow } from "@/components/topic-row";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { UserAvatar } from "@/components/user-avatar";
import { api, type Category, type SearchParams, type SearchResult, type SearchSort, type SearchType } from "@/lib/api";
import { timeAgo } from "@/lib/utils";

const PAGE_SIZE = 20;

const TYPES: { value: SearchType; label: string }[] = [
  { value: "all", label: "Tudo" },
  { value: "topics", label: "Tópicos" },
  { value: "replies", label: "Respostas" },
  { value: "users", label: "Usuários" },
  { value: "categories", label: "Categorias" },
];

const SORTS: { value: SearchSort; label: string }[] = [
  { value: "relevance", label: "Relevância" },
  { value: "recent", label: "Mais recentes" },
  { value: "activity", label: "Atividade recente" },
  { value: "popular", label: "Popularidade" },
  { value: "replies", label: "Nº de respostas" },
  { value: "likes", label: "Nº de curtidas" },
];

const FILTER_KEYS = ["q", "type", "category", "author", "from", "to", "minReplies", "minLikes", "sort", "page"] as const;

function readParams(sp: URLSearchParams): SearchParams {
  const num = (k: string) => (sp.get(k) ? Number(sp.get(k)) || undefined : undefined);
  return {
    q: sp.get("q") ?? "",
    type: (sp.get("type") as SearchType) || "all",
    category: sp.get("category") ?? "",
    author: sp.get("author") ?? "",
    from: sp.get("from") ?? "",
    to: sp.get("to") ?? "",
    minReplies: num("minReplies"),
    minLikes: num("minLikes"),
    sort: (sp.get("sort") as SearchSort) || "relevance",
    page: num("page") ?? 1,
  };
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="flex justify-center py-20 text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      }
    >
      <SearchView />
    </Suspense>
  );
}

function SearchView() {
  const router = useRouter();
  const sp = useSearchParams();
  const params = readParams(new URLSearchParams(sp.toString()));

  const [form, setForm] = useState<SearchParams>(params);
  const [showFilters, setShowFilters] = useState(
    Boolean(params.category || params.author || params.from || params.to || params.minReplies || params.minLikes),
  );
  const [categories, setCategories] = useState<Category[]>([]);
  const [result, setResult] = useState<SearchResult | null>(null);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const key = sp.toString();

  useEffect(() => {
    api.categories().then(({ data }) => setCategories(data)).catch(() => {});
  }, []);

  useEffect(() => {
    const p = readParams(new URLSearchParams(key));
    setForm(p);
    if (!p.q && !p.author && !p.category) {
      setResult(null);
      setError(null);
      return;
    }
    let active = true;
    setLoading(true);
    api
      .search({ ...p, limit: PAGE_SIZE })
      .then(({ data, meta }) => {
        if (!active) return;
        setResult(data);
        setTotal(meta?.total ?? 0);
        setError(null);
      })
      .catch((err) => active && setError(err instanceof Error ? err.message : "erro"))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [key]);

  function navigate(next: SearchParams) {
    const q = new URLSearchParams();
    for (const k of FILTER_KEYS) {
      const v = next[k];
      if (v === undefined || v === "" || v === null) continue;
      if (k === "type" && v === "all") continue;
      if (k === "sort" && v === "relevance") continue;
      if (k === "page" && v === 1) continue;
      q.set(k, String(v));
    }
    router.push(`/search${q.toString() ? `?${q}` : ""}`);
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    navigate({ ...form, page: 1 });
  }

  const set = <K extends keyof SearchParams>(k: K, v: SearchParams[K]) => setForm((f) => ({ ...f, [k]: v }));
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const page = params.page ?? 1;

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Buscar</h1>

      <form onSubmit={onSubmit} className="space-y-4" role="search">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <SearchIcon className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <label htmlFor="q" className="sr-only">
              Termo de busca
            </label>
            <Input
              id="q"
              value={form.q ?? ""}
              onChange={(e) => set("q", e.target.value)}
              placeholder="Tópicos, respostas, usuários, categorias…"
              className="pl-9"
              maxLength={100}
              autoFocus
            />
          </div>
          <Button type="submit">Buscar</Button>
          <Button
            type="button"
            variant={showFilters ? "secondary" : "outline"}
            size="icon"
            aria-label="Filtros"
            aria-expanded={showFilters}
            onClick={() => setShowFilters((v) => !v)}
          >
            <SlidersHorizontal className="h-4 w-4" />
          </Button>
        </div>

        {showFilters && (
          <div className="grid gap-4 rounded-xl border border-border bg-card p-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="f-category">Categoria</Label>
              <select
                id="f-category"
                value={form.category ?? ""}
                onChange={(e) => set("category", e.target.value)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">Todas</option>
                {categories
                  .filter((c) => c.parentId === null)
                  .map((root) => [
                    <option key={root.id} value={root.slug}>
                      {root.name}
                    </option>,
                    ...categories
                      .filter((c) => c.parentId === root.id)
                      .map((sub) => (
                        <option key={sub.id} value={sub.slug}>
                          — {sub.name}
                        </option>
                      )),
                  ])}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="f-author">Autor (usuário)</Label>
              <Input id="f-author" value={form.author ?? ""} onChange={(e) => set("author", e.target.value)} placeholder="ex.: maria" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="f-from">De</Label>
              <Input id="f-from" type="date" value={form.from ?? ""} onChange={(e) => set("from", e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="f-to">Até</Label>
              <Input id="f-to" type="date" value={form.to ?? ""} onChange={(e) => set("to", e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="f-replies">Mínimo de respostas</Label>
              <Input
                id="f-replies"
                type="number"
                min={0}
                value={form.minReplies ?? ""}
                onChange={(e) => set("minReplies", e.target.value ? Number(e.target.value) : undefined)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="f-likes">Mínimo de curtidas</Label>
              <Input
                id="f-likes"
                type="number"
                min={0}
                value={form.minLikes ?? ""}
                onChange={(e) => set("minLikes", e.target.value ? Number(e.target.value) : undefined)}
              />
            </div>
            <div className="space-y-2 sm:col-span-2">
              <Label htmlFor="f-sort">Ordenar por</Label>
              <select
                id="f-sort"
                value={form.sort ?? "relevance"}
                onChange={(e) => set("sort", e.target.value as SearchSort)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                {SORTS.map((s) => (
                  <option key={s.value} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex gap-2 sm:col-span-2">
              <Button type="submit" size="sm">
                Aplicar filtros
              </Button>
              <Button type="button" size="sm" variant="ghost" onClick={() => navigate({ q: form.q, type: form.type })}>
                Limpar filtros
              </Button>
            </div>
          </div>
        )}
      </form>

      <div className="flex flex-wrap gap-2" role="tablist" aria-label="Tipo de resultado">
        {TYPES.map((t) => (
          <Button
            key={t.value}
            role="tab"
            aria-selected={params.type === t.value}
            variant={params.type === t.value ? "default" : "outline"}
            size="sm"
            onClick={() => navigate({ ...params, type: t.value, page: 1 })}
          >
            {t.label}
          </Button>
        ))}
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {loading ? (
        <div className="flex justify-center py-12 text-muted-foreground">
          <Loader2 className="h-5 w-5 animate-spin" />
        </div>
      ) : !result ? (
        <p className="py-12 text-center text-sm text-muted-foreground">Digite um termo ou use os filtros de autor/categoria.</p>
      ) : (
        <Results
          result={result}
          onMore={(type) => navigate({ ...params, type, page: 1 })}
        />
      )}

      {result && result.type !== "all" && totalPages > 1 && !loading && (
        <div className="flex items-center justify-center gap-4">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => navigate({ ...params, page: page - 1 })}>
            Anterior
          </Button>
          <span className="text-sm text-muted-foreground">
            Página {page} de {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => navigate({ ...params, page: page + 1 })}
          >
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
}

function Results({ result, onMore }: { result: SearchResult; onMore: (t: SearchType) => void }) {
  const all = result.type === "all";
  const sections = [result.topics, result.replies, result.users, result.categories];
  const nothing = sections.every((s) => !s || s.total === 0);
  if (nothing) {
    return (
      <p className="rounded-lg border border-dashed border-border py-12 text-center text-sm text-muted-foreground">
        Nenhum resultado{result.query ? ` para “${result.query}”` : ""}.
      </p>
    );
  }

  const header = (title: string, total: number, type: SearchType) =>
    all && (
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
          {title} <span className="font-normal">({total})</span>
        </h2>
        {total > 5 && (
          <button type="button" onClick={() => onMore(type)} className="text-sm text-primary hover:underline">
            Ver todos
          </button>
        )}
      </div>
    );

  return (
    <div className="space-y-8">
      {!all && (
        <p className="text-sm text-muted-foreground">
          {sections.find(Boolean)?.total ?? 0} resultado(s){result.query ? ` para “${result.query}”` : ""}
        </p>
      )}

      {result.topics && result.topics.total > 0 && (
        <section className="space-y-3">
          {header("Tópicos", result.topics.total, "topics")}
          <ul className="space-y-2">
            {result.topics.items.map((t) => (
              <TopicRow key={t.id} topic={t} />
            ))}
          </ul>
        </section>
      )}

      {result.replies && result.replies.total > 0 && (
        <section className="space-y-3">
          {header("Respostas", result.replies.total, "replies")}
          <ul className="space-y-2">
            {result.replies.items.map((r) => (
              <li key={r.id} className="rounded-lg border border-border bg-card p-4 text-card-foreground">
                <div className="flex flex-wrap items-center gap-x-2 text-xs text-muted-foreground">
                  <Link href={`/u/${r.author.username}`} className="font-medium text-foreground hover:underline">
                    {r.author.displayName}
                  </Link>
                  <span>em</span>
                  <Link href={`/t/${r.topicId}#reply-${r.id}`} className="font-medium text-foreground hover:underline">
                    {r.topicTitle}
                  </Link>
                  <span>· {timeAgo(r.createdAt)}</span>
                </div>
                <RichText text={r.content} className="mt-2 line-clamp-3 text-sm" />
              </li>
            ))}
          </ul>
        </section>
      )}

      {result.users && result.users.total > 0 && (
        <section className="space-y-3">
          {header("Usuários", result.users.total, "users")}
          <ul className="grid gap-2 sm:grid-cols-2">
            {result.users.items.map((u) => (
              <li key={u.id}>
                <Link href={`/u/${u.username}`} className="flex items-center gap-3 rounded-lg border border-border bg-card p-3 hover:bg-secondary/60">
                  <UserAvatar name={u.displayName} url={u.avatarUrl} />
                  <span className="min-w-0">
                    <span className="block truncate font-medium">{u.displayName}</span>
                    <span className="block truncate text-xs text-muted-foreground">
                      @{u.username} · {u.reputation} rep.
                    </span>
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      )}

      {result.categories && result.categories.total > 0 && (
        <section className="space-y-3">
          {header("Categorias", result.categories.total, "categories")}
          <ul className="grid gap-2 sm:grid-cols-2">
            {result.categories.items.map((c) => (
              <li key={c.id}>
                <Link href={`/c/${c.slug}`} className="flex items-start gap-3 rounded-lg border border-border bg-card p-3 hover:bg-secondary/60">
                  <span className="text-lg leading-none">{c.icon ?? <Folder className="h-4 w-4" />}</span>
                  <span className="min-w-0">
                    <span className="block font-medium">{c.name}</span>
                    {c.description && <span className="block truncate text-xs text-muted-foreground">{c.description}</span>}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  );
}
