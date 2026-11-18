"use client";

import { Folder, Loader2, Settings2 } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { api, type Category } from "@/lib/api";

type State =
  | { kind: "loading" }
  | { kind: "ready"; categories: Category[] }
  | { kind: "error"; message: string };

export default function CategoriesPage() {
  const { hasPermission } = useAuth();
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    let active = true;
    setState({ kind: "loading" });
    api
      .categories()
      .then(({ data }) => active && setState({ kind: "ready", categories: data }))
      .catch(
        (err) => active && setState({ kind: "error", message: err instanceof Error ? err.message : "erro" }),
      );
    return () => {
      active = false;
    };
  }, []);

  if (state.kind === "loading") {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (state.kind === "error") {
    return <p className="text-center text-destructive">Não foi possível carregar as categorias ({state.message}).</p>;
  }

  const roots = state.categories.filter((c) => c.parentId === null);
  const childrenOf = (id: string) => state.categories.filter((c) => c.parentId === id);

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Categorias</h1>
        {hasPermission("categories.manage") && (
          <Button variant="outline" size="sm" asChild>
            <Link href="/admin/categories">
              <Settings2 className="h-4 w-4" />
              Gerenciar
            </Link>
          </Button>
        )}
      </div>

      {roots.length === 0 ? (
        <p className="rounded-lg border border-dashed border-border py-12 text-center text-muted-foreground">
          Nenhuma categoria ainda.
        </p>
      ) : (
        <ul className="space-y-3">
          {roots.map((cat) => {
            const subs = childrenOf(cat.id);
            return (
              <li key={cat.id} className="rounded-xl border border-border bg-card p-5 text-card-foreground">
                <Link href={`/c/${cat.slug}`} className="group flex items-start gap-3">
                  <span className="mt-0.5 text-xl leading-none">{cat.icon ?? "📁"}</span>
                  <span className="min-w-0 flex-1">
                    <span className="font-semibold group-hover:underline">{cat.name}</span>
                    {cat.description && (
                      <span className="mt-0.5 block text-sm text-muted-foreground">{cat.description}</span>
                    )}
                  </span>
                </Link>

                {subs.length > 0 && (
                  <ul className="mt-3 space-y-1 border-t border-border pt-3">
                    {subs.map((sub) => (
                      <li key={sub.id}>
                        <Link
                          href={`/c/${sub.slug}`}
                          className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm text-muted-foreground hover:bg-secondary hover:text-foreground"
                        >
                          <Folder className="h-4 w-4" />
                          {sub.name}
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
