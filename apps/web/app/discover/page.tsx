"use client";

import { Clock, Flame, Loader2, TrendingUp } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";

import { TopicRow } from "@/components/topic-row";
import { Button } from "@/components/ui/button";
import { api, type DiscoverSection, type Topic } from "@/lib/api";

const PAGE_SIZE = 20;

const SECTIONS: { value: DiscoverSection; label: string; hint: string; icon: React.ComponentType<{ className?: string }> }[] = [
  { value: "trending", label: "Em alta", hint: "Mais curtidas e respostas nos últimos 7 dias.", icon: Flame },
  { value: "popular", label: "Populares", hint: "Maior engajamento de todos os tempos.", icon: TrendingUp },
  { value: "recent", label: "Recentes", hint: "Os tópicos publicados mais recentemente.", icon: Clock },
];

export default function DiscoverPage() {
  return (
    <Suspense
      fallback={
        <div className="flex justify-center py-20 text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      }
    >
      <DiscoverView />
    </Suspense>
  );
}

function DiscoverView() {
  const router = useRouter();
  const sp = useSearchParams();
  const raw = sp.get("section");
  const section: DiscoverSection = raw === "popular" || raw === "recent" ? raw : "trending";
  const page = Math.max(1, Number(sp.get("page")) || 1);

  const [topics, setTopics] = useState<Topic[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    api
      .discover(section, page, PAGE_SIZE)
      .then((res) => {
        if (!active) return;
        setTopics(res.items);
        setTotal(res.meta.total);
        setError(null);
      })
      .catch((err) => active && setError(err instanceof Error ? err.message : "erro"))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [section, page]);

  const go = (s: DiscoverSection, p = 1) => router.push(`/discover?section=${s}${p > 1 ? `&page=${p}` : ""}`);
  const current = SECTIONS.find((s) => s.value === section)!;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Descobrir</h1>
        <p className="mt-1 text-sm text-muted-foreground">{current.hint}</p>
      </div>

      <div className="flex flex-wrap gap-2" role="tablist" aria-label="Seção">
        {SECTIONS.map(({ value, label, icon: Icon }) => (
          <Button
            key={value}
            role="tab"
            aria-selected={section === value}
            variant={section === value ? "default" : "outline"}
            size="sm"
            onClick={() => go(value)}
          >
            <Icon className="h-4 w-4" />
            {label}
          </Button>
        ))}
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {loading ? (
        <div className="flex justify-center py-12 text-muted-foreground">
          <Loader2 className="h-5 w-5 animate-spin" />
        </div>
      ) : topics.length === 0 ? (
        <p className="rounded-lg border border-dashed border-border py-12 text-center text-sm text-muted-foreground">
          {section === "trending" ? "Nada em alta nos últimos 7 dias." : "Nenhum tópico ainda."}
        </p>
      ) : (
        <ol className="space-y-2">
          {topics.map((t) => (
            <TopicRow key={t.id} topic={t} />
          ))}
        </ol>
      )}

      {totalPages > 1 && !loading && (
        <div className="flex items-center justify-center gap-4">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => go(section, page - 1)}>
            Anterior
          </Button>
          <span className="text-sm text-muted-foreground">
            Página {page} de {totalPages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => go(section, page + 1)}>
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
}
