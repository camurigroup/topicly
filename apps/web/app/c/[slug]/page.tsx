"use client";

import { Loader2, Plus } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { TopicRow } from "@/components/topic-row";
import { Button } from "@/components/ui/button";
import { api, type Category, type Topic } from "@/lib/api";

const PAGE_SIZE = 20;

export default function CategoryPage() {
  const { slug } = useParams<{ slug: string }>();
  const { user } = useAuth();

  const [category, setCategory] = useState<Category | null>(null);
  const [subcategories, setSubcategories] = useState<Category[]>([]);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [{ data: cat }, { data: all }, topicPage] = await Promise.all([
        api.category(slug),
        api.categories(),
        api.topics({ category: slug, page, limit: PAGE_SIZE }),
      ]);
      setCategory(cat);
      setSubcategories(all.filter((c) => c.parentId === cat.id));
      setTopics(topicPage.items);
      setTotal(topicPage.meta.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "erro");
    } finally {
      setLoading(false);
    }
  }, [slug, page]);

  useEffect(() => {
    void load();
  }, [load]);

  if (loading && !category) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (error && !category) {
    return <p className="text-center text-destructive">Categoria não encontrada ({error}).</p>;
  }

  if (!category) return null;

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <nav className="text-sm text-muted-foreground">
        <Link href="/categories" className="hover:underline">
          Categorias
        </Link>{" "}
        / <span className="text-foreground">{category.name}</span>
      </nav>

      <div className="flex items-start justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold tracking-tight">
            <span>{category.icon ?? "📁"}</span>
            {category.name}
          </h1>
          {category.description && <p className="mt-1 text-muted-foreground">{category.description}</p>}
        </div>
        {user && (
          <Button asChild>
            <Link href={`/c/${category.slug}/new`}>
              <Plus className="h-4 w-4" />
              Novo tópico
            </Link>
          </Button>
        )}
      </div>

      {subcategories.length > 0 && (
        <ul className="flex flex-wrap gap-2">
          {subcategories.map((sub) => (
            <li key={sub.id}>
              <Link
                href={`/c/${sub.slug}`}
                className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card px-3 py-1 text-sm hover:bg-secondary"
              >
                <span>{sub.icon ?? "📁"}</span>
                {sub.name}
              </Link>
            </li>
          ))}
        </ul>
      )}

      {topics.length === 0 ? (
        <p className="rounded-lg border border-dashed border-border py-12 text-center text-muted-foreground">
          Nenhum tópico nesta categoria ainda.
        </p>
      ) : (
        <ul className="space-y-2">
          {topics.map((t) => (
            <TopicRow key={t.id} topic={t} />
          ))}
        </ul>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4 pt-2">
          <Button variant="outline" size="sm" disabled={page <= 1 || loading} onClick={() => setPage((p) => p - 1)}>
            Anterior
          </Button>
          <span className="text-sm text-muted-foreground">
            Página {page} de {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages || loading}
            onClick={() => setPage((p) => p + 1)}
          >
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
}
