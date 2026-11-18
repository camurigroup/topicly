"use client";

import { Bell, X } from "lucide-react";
import Link from "next/link";
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { api, openNotificationStream, type AppNotification, type StreamEvent } from "@/lib/api";

interface NotificationsContextValue {
  unread: number;
  live: boolean;
  /** Last event received from the stream (pages use it to update in place). */
  lastEvent: StreamEvent | null;
  setUnread: (n: number) => void;
}

const NotificationsContext = createContext<NotificationsContextValue | null>(null);

const TOAST_MS = 6000;

export function NotificationsProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const [unread, setUnread] = useState(0);
  const [live, setLive] = useState(false);
  const [lastEvent, setLastEvent] = useState<StreamEvent | null>(null);
  const [toast, setToast] = useState<AppNotification | null>(null);

  useEffect(() => {
    if (!user) {
      setUnread(0);
      setLive(false);
      return;
    }
    const controller = new AbortController();
    api
      .unreadCount()
      .then(({ data }) => setUnread(data.unread))
      .catch(() => {});
    openNotificationStream(
      (ev) => {
        setUnread(ev.unread);
        setLastEvent(ev);
        if (ev.kind === "notification" && ev.notification) setToast(ev.notification);
      },
      controller.signal,
      setLive,
    );
    return () => controller.abort();
  }, [user]);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), TOAST_MS);
    return () => clearTimeout(t);
  }, [toast]);

  const value = useMemo<NotificationsContextValue>(
    () => ({ unread, live, lastEvent, setUnread }),
    [unread, live, lastEvent],
  );

  return (
    <NotificationsContext.Provider value={value}>
      {children}
      {toast && <NotificationToast notification={toast} onClose={() => setToast(null)} />}
    </NotificationsContext.Provider>
  );
}

export function useNotifications() {
  const ctx = useContext(NotificationsContext);
  if (!ctx) throw new Error("useNotifications must be used within NotificationsProvider");
  return ctx;
}

function NotificationToast({ notification, onClose }: { notification: AppNotification; onClose: () => void }) {
  const close = useCallback(() => onClose(), [onClose]);
  const body = (
    <span className="min-w-0 flex-1 text-sm">
      {notification.actor && <strong className="font-semibold">{notification.actor.displayName} </strong>}
      {notification.message}
    </span>
  );
  return (
    <div
      role="status"
      aria-live="polite"
      className="fixed bottom-4 right-4 z-50 flex w-[calc(100%-2rem)] max-w-sm items-start gap-3 rounded-lg border border-border bg-card p-4 text-card-foreground shadow-lg"
    >
      <Bell className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
      {notification.link ? (
        <Link href={notification.link} onClick={close} className="min-w-0 flex-1 hover:underline">
          {body}
        </Link>
      ) : (
        body
      )}
      <button type="button" aria-label="Fechar" onClick={close} className="text-muted-foreground hover:text-foreground">
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}
