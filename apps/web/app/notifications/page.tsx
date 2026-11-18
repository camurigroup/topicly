"use client";

import { AtSign, Bell, CheckCheck, Heart, Loader2, Megaphone, MessageSquare, Plus, ShieldAlert, Trash2, UserPlus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { useNotifications } from "@/components/notifications-provider";
import { Button } from "@/components/ui/button";
import { UserAvatar } from "@/components/user-avatar";
import { api, type AppNotification, type NotificationType } from "@/lib/api";
import { cn, timeAgo } from "@/lib/utils";

const PAGE_SIZE = 20;

const ICONS: Record<NotificationType, React.ComponentType<{ className?: string }>> = {
  follow: UserPlus,
  topic_like: Heart,
  reply_like: Heart,
  reply: MessageSquare,
  followed_reply: MessageSquare,
  mention: AtSign,
  new_topic: Plus,
  punishment: ShieldAlert,
  admin: Megaphone,
};

export default function NotificationsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { unread, lastEvent, live } = useNotifications();

  const [items, setItems] = useState<AppNotification[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!authLoading && !user) router.replace("/login");
  }, [authLoading, user, router]);

  const load = useCallback(
    async (silent = false) => {
      if (!silent) setLoading(true);
      try {
        const res = await api.notifications({ page, unread: unreadOnly });
        setItems(res.items);
        setTotal(res.meta.total);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "erro");
      } finally {
        setLoading(false);
      }
    },
    [page, unreadOnly],
  );

  useEffect(() => {
    if (user) void load();
  }, [user, load]);

  // Live updates: prepend new notifications on page 1, resync otherwise.
  useEffect(() => {
    if (!lastEvent) return;
    if (lastEvent.kind === "notification" && lastEvent.notification && page === 1) {
      const n = lastEvent.notification;
      setItems((prev) => (prev.some((p) => p.id === n.id) ? prev : [n, ...prev].slice(0, PAGE_SIZE)));
      setTotal((t) => t + 1);
    } else if (lastEvent.kind === "sync") {
      void load(true);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [lastEvent]);

  async function open(n: AppNotification) {
    if (!n.isRead) {
      setItems((prev) => prev.map((p) => (p.id === n.id ? { ...p, isRead: true } : p)));
      await api.markNotificationRead(n.id).catch(() => {});
    }
    if (n.link) router.push(n.link);
  }

  async function markAll() {
    setItems((prev) => prev.map((p) => ({ ...p, isRead: true })));
    await api.markAllNotificationsRead().catch(() => {});
  }

  async function remove(id: string) {
    setItems((prev) => prev.filter((p) => p.id !== id));
    setTotal((t) => Math.max(0, t - 1));
    await api.deleteNotification(id).catch(() => void load(true));
  }

  if (authLoading || !user) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Notificações</h1>
          <p className="mt-1 flex items-center gap-2 text-sm text-muted-foreground">
            <span className={cn("h-2 w-2 rounded-full", live ? "bg-green-500" : "bg-muted-foreground/40")} />
            {live ? "Atualizando em tempo real" : "Reconectando…"} · {unread} não lida{unread === 1 ? "" : "s"}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={markAll} disabled={unread === 0}>
          <CheckCheck className="h-4 w-4" />
          Marcar todas como lidas
        </Button>
      </div>

      <div className="flex gap-2" role="tablist" aria-label="Filtro">
        {[
          { label: "Todas", value: false },
          { label: "Não lidas", value: true },
        ].map((tab) => (
          <Button
            key={tab.label}
            role="tab"
            aria-selected={unreadOnly === tab.value}
            variant={unreadOnly === tab.value ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setUnreadOnly(tab.value);
              setPage(1);
            }}
          >
            {tab.label}
          </Button>
        ))}
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {loading && items.length === 0 ? (
        <div className="flex justify-center py-12 text-muted-foreground">
          <Loader2 className="h-5 w-5 animate-spin" />
        </div>
      ) : items.length === 0 ? (
        <div className="rounded-lg border border-dashed border-border py-12 text-center text-muted-foreground">
          <Bell className="mx-auto mb-2 h-6 w-6" />
          {unreadOnly ? "Nenhuma notificação não lida." : "Você ainda não tem notificações."}
        </div>
      ) : (
        <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
          {items.map((n) => {
            const Icon = ICONS[n.type] ?? Bell;
            return (
              <li key={n.id} className={cn("group flex items-start gap-3 p-4", !n.isRead && "bg-primary/5")}>
                <div className="relative">
                  {n.actor ? (
                    <UserAvatar name={n.actor.displayName} url={n.actor.avatarUrl} />
                  ) : (
                    <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-muted">
                      <Icon className="h-4 w-4 text-muted-foreground" />
                    </span>
                  )}
                  {n.actor && (
                    <span className="absolute -bottom-1 -right-1 rounded-full border border-card bg-card p-0.5">
                      <Icon className="h-3 w-3 text-primary" />
                    </span>
                  )}
                </div>
                <button type="button" onClick={() => open(n)} className="min-w-0 flex-1 text-left">
                  <p className="text-sm">
                    {n.actor && <strong className="font-semibold">{n.actor.displayName} </strong>}
                    {n.message}
                  </p>
                  <p className="mt-0.5 text-xs text-muted-foreground">{timeAgo(n.createdAt)}</p>
                </button>
                {!n.isRead && <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-primary" aria-label="não lida" />}
                <button
                  type="button"
                  onClick={() => remove(n.id)}
                  aria-label="Excluir notificação"
                  className="shrink-0 rounded-md p-1 text-muted-foreground opacity-60 hover:bg-secondary hover:text-destructive group-hover:opacity-100"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </li>
            );
          })}
        </ul>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Anterior
          </Button>
          <span className="text-sm text-muted-foreground">
            Página {page} de {totalPages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Próxima
          </Button>
        </div>
      )}
    </div>
  );
}
