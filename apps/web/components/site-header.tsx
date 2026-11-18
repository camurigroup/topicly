"use client";

import { Bell, LogOut, MessagesSquare, Search, Settings, Shield, User as UserIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { useModerationSections } from "@/components/moderation-nav";
import { useNotifications } from "@/components/notifications-provider";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export function SiteHeader() {
  const { user, loading, logout } = useAuth();
  const { unread } = useNotifications();
  const modSections = useModerationSections();
  const router = useRouter();
  const [q, setQ] = useState("");

  async function handleLogout() {
    await logout();
    router.push("/");
  }

  function onSearch(e: React.FormEvent) {
    e.preventDefault();
    const term = q.trim();
    if (term) router.push(`/search?q=${encodeURIComponent(term)}`);
  }

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/80 backdrop-blur">
      <div className="container flex h-16 items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-4">
          <Link href="/" className="flex items-center gap-2 font-semibold">
            <MessagesSquare className="h-5 w-5 text-primary" />
            <span>Forum</span>
          </Link>
          <nav className="hidden items-center gap-4 sm:flex">
            <Link href="/categories" className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground">
              Categorias
            </Link>
            <Link href="/discover" className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground">
              Descobrir
            </Link>
          </nav>
        </div>

        <form onSubmit={onSearch} role="search" className="hidden max-w-xs flex-1 md:block">
          <label htmlFor="global-search" className="sr-only">
            Buscar
          </label>
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="global-search"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Buscar no fórum…"
              className="h-9 pl-9"
              maxLength={100}
            />
          </div>
        </form>

        <nav className="flex shrink-0 items-center gap-1 sm:gap-2">
          <Button variant="ghost" size="icon" aria-label="Buscar" className="md:hidden" asChild>
            <Link href="/search">
              <Search className="h-4 w-4" />
            </Link>
          </Button>
          <ThemeToggle />
          {loading ? null : user ? (
            <>
              <Button
                variant="ghost"
                size="icon"
                aria-label={unread > 0 ? `Notificações (${unread} não lidas)` : "Notificações"}
                className="relative"
                asChild
              >
                <Link href="/notifications">
                  <Bell className="h-4 w-4" />
                  {unread > 0 && (
                    <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-semibold leading-none text-destructive-foreground">
                      {unread > 99 ? "99+" : unread}
                    </span>
                  )}
                </Link>
              </Button>
              {modSections.length > 0 && (
                <Button variant="ghost" size="icon" aria-label="Moderação" title="Moderação" asChild>
                  <Link href={modSections[0].href}>
                    <Shield className="h-4 w-4" />
                  </Link>
                </Button>
              )}
              <Button variant="ghost" size="sm" className="hidden sm:inline-flex" asChild>
                <Link href={`/u/${user.username}`}>
                  <UserIcon className="h-4 w-4" />
                  {user.displayName}
                </Link>
              </Button>
              <Button variant="ghost" size="icon" aria-label="Configurações" asChild>
                <Link href="/settings">
                  <Settings className="h-4 w-4" />
                </Link>
              </Button>
              <Button variant="outline" size="sm" onClick={handleLogout} aria-label="Sair">
                <LogOut className="h-4 w-4" />
                <span className="hidden sm:inline">Sair</span>
              </Button>
            </>
          ) : (
            <>
              <Button variant="ghost" size="sm" asChild>
                <Link href="/login">Entrar</Link>
              </Button>
              <Button size="sm" asChild>
                <Link href="/register">Criar conta</Link>
              </Button>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
