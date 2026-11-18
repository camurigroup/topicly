"use client";

import { Loader2 } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { UserAvatar } from "@/components/user-avatar";
import { api, type UserSummary } from "@/lib/api";

const PAGE_SIZE = 20;

/** Paginated list of a user's followers or of who they follow. */
export function FollowList({ username, kind }: { username: string; kind: "followers" | "following" }) {
  const [items, setItems] = useState<UserSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    const req = kind === "followers" ? api.followers(username, page) : api.following(username, page);
    req
      .then((res) => {
        if (!active) return;
        setItems(res.items);
        setTotal(res.meta.total);
        setError(null);
      })
      .catch((err) => active && setError(err instanceof Error ? err.message : "erro"))
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [username, kind, page]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const title = kind === "followers" ? "Seguidores" : "Seguindo";

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <nav className="text-sm text-muted-foreground">
        <Link href={`/u/${username}`} className="hover:underline">
          @{username}
        </Link>{" "}
        / <span className="text-foreground">{title}</span>
      </nav>

      <div className="flex items-center gap-2">
        {(["followers", "following"] as const).map((k) => (
          <Button key={k} variant={k === kind ? "default" : "outline"} size="sm" asChild>
            <Link href={`/u/${username}/${k}`}>{k === "followers" ? "Seguidores" : "Seguindo"}</Link>
          </Button>
        ))}
        <span className="ml-auto text-sm text-muted-foreground">{total} no total</span>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {loading ? (
        <div className="flex justify-center py-12 text-muted-foreground">
          <Loader2 className="h-5 w-5 animate-spin" />
        </div>
      ) : items.length === 0 ? (
        <p className="rounded-lg border border-dashed border-border py-12 text-center text-sm text-muted-foreground">
          {kind === "followers" ? "Ninguém segue este usuário ainda." : "Este usuário ainda não segue ninguém."}
        </p>
      ) : (
        <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
          {items.map((u) => (
            <li key={u.id}>
              <Link href={`/u/${u.username}`} className="flex items-center gap-3 p-4 hover:bg-secondary/60">
                <UserAvatar name={u.displayName} url={u.avatarUrl} />
                <span className="min-w-0 flex-1">
                  <span className="block font-medium">{u.displayName}</span>
                  <span className="block truncate text-xs text-muted-foreground">
                    @{u.username}
                    {u.bio ? ` · ${u.bio}` : ""}
                  </span>
                </span>
                <span className="text-xs text-muted-foreground">{u.reputation} rep.</span>
              </Link>
            </li>
          ))}
        </ul>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Anterior
          </Button>
          <span className="text-sm text-muted-foreground">
            {page} / {totalPages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
}
