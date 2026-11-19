"use client";

import { Activity as ActivityIcon, Folder, Heart, Loader2, MessageSquare, MessagesSquare, Settings2, Sparkles, Users } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { TopicRow } from "@/components/topic-row";
import { Button } from "@/components/ui/button";
import { UserAvatar } from "@/components/user-avatar";
import { api, type HomeData } from "@/lib/api";
import { formatCount, timeAgo } from "@/lib/utils";

type State = { kind: "loading" } | { kind: "ready"; home: HomeData } | { kind: "error"; message: string };

export default function HomePage() {
  const { user, hasPermission } = useAuth();
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    let active = true;
    api
      .home()
      .then(({ data }) => active && setState({ kind: "ready", home: data }))
      .catch((err) => active && setState({ kind: "error", message: err instanceof Error ? err.message : "erro" }));
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
    return <p className="py-20 text-center text-destructive">Não foi possível carregar a página inicial ({state.message}).</p>;
  }

  const { home } = state;
  const s = home.settings;
  const show = (k: string) => s[`home.show.${k}`] === "true";
  const welcome = user ? s["home.welcome.member"]?.replaceAll("{name}", user.displayName) : null;
  const roots = home.categories.filter((c) => c.parentId === null);
  const subCount = (id: string) => home.categories.filter((c) => c.parentId === id).length;

  return (
    <div className="mx-auto max-w-5xl space-y-12">
      {/* Banner */}
      <section
        className="relative overflow-hidden rounded-2xl border border-border bg-card px-6 py-12 text-card-foreground sm:px-10 sm:py-16"
        style={
          s["home.banner.image"]
            ? { backgroundImage: `linear-gradient(to right, hsl(var(--card)) 35%, hsl(var(--card) / 0.6)), url(${JSON.stringify(s["home.banner.image"])})`, backgroundSize: "cover", backgroundPosition: "center" }
            : undefined
        }
      >
        {hasPermission("settings.manage") && (
          <Button variant="outline" size="sm" className="absolute right-4 top-4 z-10 bg-background/70 backdrop-blur" asChild>
            <Link href="/admin/home">
              <Settings2 className="h-4 w-4" />
              Personalizar
            </Link>
          </Button>
        )}
        {!s["home.banner.image"] && (
          <div aria-hidden className="pointer-events-none absolute inset-0">
            <div className="absolute -right-24 -top-24 h-72 w-72 rounded-full bg-primary/25 blur-3xl" />
            <div className="absolute -bottom-32 right-1/3 h-64 w-64 rounded-full bg-primary/10 blur-3xl" />
            <div className="absolute inset-0 bg-[radial-gradient(hsl(var(--foreground)/0.07)_1px,transparent_1px)] [background-size:18px_18px] [mask-image:linear-gradient(to_left,black,transparent_70%)]" />
          </div>
        )}
        <div className="relative max-w-2xl space-y-5">
          {welcome && (
            <p className="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
              <Sparkles className="h-3.5 w-3.5" />
              {welcome}
            </p>
          )}
          <h1 className="text-4xl font-bold tracking-tight sm:text-6xl">{s["home.banner.title"]}</h1>
          {s["home.banner.subtitle"] && <p className="text-lg text-muted-foreground">{s["home.banner.subtitle"]}</p>}
          <div className="flex flex-wrap gap-3 pt-2">
            {s["home.banner.ctaLabel"] && (
              <Button asChild>
                <Link href={s["home.banner.ctaLink"] || "/categories"}>{s["home.banner.ctaLabel"]}</Link>
              </Button>
            )}
            {!user && (
              <Button variant="outline" asChild>
                <Link href="/register">Criar conta</Link>
              </Button>
            )}
          </div>
        </div>
      </section>

      {/* Apresentação */}
      {(s["home.intro.title"] || s["home.intro.text"]) && (
        <section className="max-w-3xl space-y-2">
          {s["home.intro.title"] && <h2 className="text-xl font-semibold">{s["home.intro.title"]}</h2>}
          {s["home.intro.text"] && <p className="whitespace-pre-line text-muted-foreground">{s["home.intro.text"]}</p>}
        </section>
      )}

      {/* Estatísticas */}
      {show("stats") && home.stats && (
        <section className="space-y-3">
          <SectionTitle>{s["home.section.stats.title"]}</SectionTitle>
          <dl className="grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-border bg-border sm:grid-cols-5">
            {[
              { label: "Membros", value: home.stats.users, icon: Users },
              { label: "Tópicos", value: home.stats.topics, icon: MessagesSquare },
              { label: "Respostas", value: home.stats.replies, icon: MessageSquare },
              { label: "Categorias", value: home.stats.categories, icon: Folder },
              { label: "Curtidas", value: home.stats.likes, icon: Heart },
            ].map(({ label, value, icon: Icon }) => (
              <div key={label} className="bg-card px-5 py-5">
                <dt className="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  <Icon className="h-3.5 w-3.5 text-primary" />
                  {label}
                </dt>
                <dd className="mt-2 text-3xl font-bold tracking-tight tabular-nums">{formatCount(value)}</dd>
              </div>
            ))}
          </dl>
        </section>
      )}

      {/* Destaques */}
      {show("featured") && home.featuredTopics.length > 0 && (
        <section className="space-y-3">
          <SectionTitle icon={<Sparkles className="h-4 w-4 text-primary" />}>{s["home.section.featured.title"]}</SectionTitle>
          <ul className="grid gap-2 md:grid-cols-2">
            {home.featuredTopics.map((t) => (
              <TopicRow key={t.id} topic={t} />
            ))}
          </ul>
        </section>
      )}

      {/* Categorias */}
      {show("categories") && roots.length > 0 && (
        <section className="space-y-3">
          <div className="flex items-center justify-between">
            <SectionTitle>{s["home.section.categories.title"]}</SectionTitle>
            <Link href="/categories" className="text-sm text-primary hover:underline">
              Ver todas
            </Link>
          </div>
          <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {roots.map((c) => (
              <li key={c.id}>
                <Link
                  href={`/c/${c.slug}`}
                  className="group flex h-full items-start gap-3 rounded-xl border border-border bg-card p-4 transition-all hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md"
                >
                  <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-xl leading-none transition-colors group-hover:bg-primary/20">
                    {c.icon ?? "📁"}
                  </span>
                  <span className="min-w-0">
                    <span className="block font-medium">{c.name}</span>
                    {c.description && <span className="mt-0.5 line-clamp-2 block text-sm text-muted-foreground">{c.description}</span>}
                    {subCount(c.id) > 0 && (
                      <span className="mt-1 block text-xs text-muted-foreground">
                        {subCount(c.id)} subcategoria{subCount(c.id) === 1 ? "" : "s"}
                      </span>
                    )}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      )}

      {/* Recentes + Populares */}
      {(show("recent") || show("popular")) && (
        <div className="grid gap-8 lg:grid-cols-2">
          {show("recent") && (
            <TopicColumn title={s["home.section.recent.title"]} topics={home.recentTopics} href="/discover?section=recent" />
          )}
          {show("popular") && (
            <TopicColumn title={s["home.section.popular.title"]} topics={home.popularTopics} href="/discover?section=popular" />
          )}
        </div>
      )}

      {/* Membros + Atividades */}
      {(show("users") || show("activity")) && (
        <div className="grid gap-8 lg:grid-cols-2">
          {show("users") && home.featuredUsers.length > 0 && (
            <section className="space-y-3">
              <SectionTitle>{s["home.section.users.title"]}</SectionTitle>
              <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
                {home.featuredUsers.map((u) => (
                  <li key={u.id}>
                    <Link href={`/u/${u.username}`} className="flex items-center gap-3 p-3 hover:bg-secondary/60">
                      <UserAvatar name={u.displayName} url={u.avatarUrl} />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate font-medium">{u.displayName}</span>
                        <span className="block text-xs text-muted-foreground">@{u.username}</span>
                      </span>
                      <span className="text-right text-xs text-muted-foreground">
                        {formatCount(u.followers)} seguidor{u.followers === 1 ? "" : "es"}
                        <br />
                        {u.reputation} rep.
                      </span>
                    </Link>
                  </li>
                ))}
              </ul>
            </section>
          )}
          {show("activity") && home.activity.length > 0 && (
            <section className="space-y-3">
              <SectionTitle icon={<ActivityIcon className="h-4 w-4 text-primary" />}>{s["home.section.activity.title"]}</SectionTitle>
              <ul className="space-y-3 rounded-xl border border-border bg-card p-4">
                {home.activity.map((a) => (
                  <li key={`${a.kind}-${a.id}`} className="flex items-start gap-3 text-sm">
                    <UserAvatar name={a.actor.displayName} url={a.actor.avatarUrl} className="h-7 w-7 text-xs" />
                    <p className="min-w-0 flex-1">
                      <Link href={`/u/${a.actor.username}`} className="font-medium hover:underline">
                        {a.actor.displayName}
                      </Link>{" "}
                      <span className="text-muted-foreground">{a.kind === "topic" ? "abriu" : "respondeu em"}</span>{" "}
                      <Link
                        href={a.kind === "topic" ? `/t/${a.topicId}` : `/t/${a.topicId}#reply-${a.id}`}
                        className="font-medium hover:underline"
                      >
                        {a.topicTitle}
                      </Link>
                      <span className="block text-xs text-muted-foreground">{timeAgo(a.createdAt)}</span>
                    </p>
                  </li>
                ))}
              </ul>
            </section>
          )}
        </div>
      )}
    </div>
  );
}

function SectionTitle({ children, icon }: { children: React.ReactNode; icon?: React.ReactNode }) {
  return (
    <h2 className="flex items-center gap-2.5 text-lg font-semibold tracking-tight">
      <span aria-hidden className="h-5 w-1 rounded-full bg-primary" />
      {icon}
      {children}
    </h2>
  );
}

function TopicColumn({ title, topics, href }: { title: string; topics: HomeData["recentTopics"]; href: string }) {
  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between">
        <SectionTitle>{title}</SectionTitle>
        <Link href={href} className="text-sm text-primary hover:underline">
          Ver mais
        </Link>
      </div>
      {topics.length === 0 ? (
        <p className="rounded-lg border border-dashed border-border py-8 text-center text-sm text-muted-foreground">Nenhum tópico ainda.</p>
      ) : (
        <ul className="space-y-2">
          {topics.map((t) => (
            <TopicRow key={t.id} topic={t} />
          ))}
        </ul>
      )}
    </section>
  );
}
