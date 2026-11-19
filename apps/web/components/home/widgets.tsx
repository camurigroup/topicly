"use client";

import {
  Activity as ActivityIcon,
  Eye,
  Folder,
  Heart,
  MessageSquare,
  MessagesSquare,
  Settings2,
  Sparkles,
  Users,
} from "lucide-react";
import Link from "next/link";

import { TopicRow } from "@/components/topic-row";
import { Button } from "@/components/ui/button";
import { UserAvatar } from "@/components/user-avatar";
import type { Account, HomeData, HomeWidget, Topic } from "@/lib/api";
import { cn, formatCount, timeAgo } from "@/lib/utils";

interface BlockProps {
  w: HomeWidget;
  home: HomeData;
  user: Account | null;
  canCustomize: boolean;
}

/** Renders one configurable home block. */
export function HomeBlock(props: BlockProps) {
  switch (props.w.type) {
    case "hero":
      return <Hero {...props} />;
    case "intro":
      return <Intro {...props} />;
    case "stats":
      return <Stats {...props} />;
    case "categories":
      return <Categories {...props} />;
    case "topics":
      return <Topics {...props} />;
    case "members":
      return <Members {...props} />;
    case "activity":
      return <ActivityFeed {...props} />;
  }
}

function Block({
  title,
  action,
  children,
  className,
}: {
  title?: string;
  action?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("space-y-3", className)}>
      {(title || action) && (
        <div className="flex items-center justify-between gap-3">
          {title && (
            <h2 className="flex items-center gap-2.5 text-lg font-semibold tracking-tight">
              <span aria-hidden className="h-5 w-1 rounded-full bg-primary" />
              {title}
            </h2>
          )}
          {action}
        </div>
      )}
      {children}
    </section>
  );
}

const Empty = ({ text }: { text: string }) => (
  <p className="rounded-xl border border-dashed border-border py-8 text-center text-sm text-muted-foreground">
    {text}
  </p>
);

function Hero({ w, home, user, canCustomize }: BlockProps) {
  const s = home.settings;
  const compact = w.variant === "compact";
  const welcome = user
    ? s["home.welcome.member"]?.replaceAll("{name}", user.displayName)
    : null;
  const image = s["home.banner.image"];
  return (
    <section
      className={cn(
        "relative overflow-hidden rounded-2xl border border-border bg-card text-card-foreground",
        compact ? "px-6 py-8 sm:px-8" : "px-6 py-12 sm:px-10 sm:py-16",
      )}
      style={
        image
          ? {
              backgroundImage: `linear-gradient(to right, hsl(var(--card)) 35%, hsl(var(--card) / 0.6)), url(${JSON.stringify(image)})`,
              backgroundSize: "cover",
              backgroundPosition: "center",
            }
          : undefined
      }
    >
      {!image && (
        <div aria-hidden className="pointer-events-none absolute inset-0">
          <div className="absolute -right-24 -top-24 h-72 w-72 rounded-full bg-primary/25 blur-3xl" />
          <div className="absolute -bottom-32 right-1/3 h-64 w-64 rounded-full bg-primary/10 blur-3xl" />
          <div className="absolute inset-0 bg-[radial-gradient(hsl(var(--foreground)/0.07)_1px,transparent_1px)] [background-size:18px_18px] [mask-image:linear-gradient(to_left,black,transparent_70%)]" />
        </div>
      )}
      {canCustomize && (
        <Button
          variant="outline"
          size="sm"
          className="absolute right-4 top-4 z-10 bg-background/70 backdrop-blur"
          asChild
        >
          <Link href="/admin/home">
            <Settings2 className="h-4 w-4" />
            Personalizar
          </Link>
        </Button>
      )}
      <div className="relative max-w-2xl space-y-4">
        {welcome && (
          <p className="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
            <Sparkles className="h-3.5 w-3.5" />
            {welcome}
          </p>
        )}
        <h1
          className={cn(
            "font-bold tracking-tight",
            compact ? "text-3xl" : "text-4xl sm:text-6xl",
          )}
        >
          {s["home.banner.title"]}
        </h1>
        {s["home.banner.subtitle"] && (
          <p className="text-lg text-muted-foreground">
            {s["home.banner.subtitle"]}
          </p>
        )}
        <div className="flex flex-wrap gap-3 pt-1">
          {s["home.banner.ctaLabel"] && (
            <Button asChild>
              <Link href={s["home.banner.ctaLink"] || "/categories"}>
                {s["home.banner.ctaLabel"]}
              </Link>
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
  );
}

function Intro({ w, home }: BlockProps) {
  const s = home.settings;
  const text = s["home.intro.text"];
  if (!text && !w.title && !s["home.intro.title"]) return null;
  const body = (
    <p className="whitespace-pre-line text-sm leading-relaxed text-muted-foreground">
      {text}
    </p>
  );
  return (
    <Block title={w.title || s["home.intro.title"]}>
      {w.variant === "card" ? (
        <div className="rounded-xl border border-border bg-card p-4">
          {body}
        </div>
      ) : (
        body
      )}
    </Block>
  );
}

function Stats({ w, home }: BlockProps) {
  if (!home.stats) return null;
  const items = [
    { label: "Membros", value: home.stats.users, icon: Users },
    { label: "Tópicos", value: home.stats.topics, icon: MessagesSquare },
    { label: "Respostas", value: home.stats.replies, icon: MessageSquare },
    { label: "Categorias", value: home.stats.categories, icon: Folder },
    { label: "Curtidas", value: home.stats.likes, icon: Heart },
  ];
  return (
    <Block title={w.title}>
      {w.variant === "list" ? (
        <dl className="divide-y divide-border rounded-xl border border-border bg-card">
          {items.map(({ label, value, icon: Icon }) => (
            <div
              key={label}
              className="flex items-center justify-between px-4 py-2.5 text-sm"
            >
              <dt className="flex items-center gap-2 text-muted-foreground">
                <Icon className="h-4 w-4 text-primary" />
                {label}
              </dt>
              <dd className="font-semibold tabular-nums">
                {formatCount(value)}
              </dd>
            </div>
          ))}
        </dl>
      ) : (
        <dl className="grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-border bg-border [&>div:last-child:nth-child(odd)]:col-span-2">
          {items.map(({ label, value, icon: Icon }) => (
            <div key={label} className="bg-card px-4 py-3">
              <dt className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                <Icon className="h-3.5 w-3.5 text-primary" />
                {label}
              </dt>
              <dd className="mt-1 text-2xl font-bold tabular-nums">
                {formatCount(value)}
              </dd>
            </div>
          ))}
        </dl>
      )}
    </Block>
  );
}

function Categories({ w, home }: BlockProps) {
  const roots = home.categories
    .filter((c) => c.parentId === null)
    .slice(0, w.limit);
  const subs = (id: string) =>
    home.categories.filter((c) => c.parentId === id).length;
  const action = (
    <Link href="/categories" className="text-sm text-primary hover:underline">
      Ver todas
    </Link>
  );
  if (roots.length === 0)
    return (
      <Block title={w.title} action={action}>
        <Empty text="Nenhuma categoria ainda." />
      </Block>
    );
  return (
    <Block title={w.title} action={action}>
      {w.variant === "list" ? (
        <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
          {roots.map((c) => (
            <li key={c.id}>
              <Link
                href={`/c/${c.slug}`}
                className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-secondary/60"
              >
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-base">
                  {c.icon ?? "📁"}
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium">
                    {c.name}
                  </span>
                  {subs(c.id) > 0 && (
                    <span className="block text-xs text-muted-foreground">
                      {subs(c.id)} subcategoria(s)
                    </span>
                  )}
                </span>
              </Link>
            </li>
          ))}
        </ul>
      ) : (
        <ul className="grid gap-3 sm:grid-cols-2">
          {roots.map((c) => (
            <li key={c.id}>
              <Link
                href={`/c/${c.slug}`}
                className="group flex h-full items-start gap-3 rounded-xl border border-border bg-card p-4 transition-all hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md"
              >
                <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-xl group-hover:bg-primary/20">
                  {c.icon ?? "📁"}
                </span>
                <span className="min-w-0">
                  <span className="block font-medium">{c.name}</span>
                  {c.description && (
                    <span className="mt-0.5 line-clamp-2 block text-sm text-muted-foreground">
                      {c.description}
                    </span>
                  )}
                </span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </Block>
  );
}

const SOURCE_LINK: Record<string, string> = {
  recent: "/discover?section=recent",
  popular: "/discover?section=popular",
  trending: "/discover",
};

function Topics({ w, home }: BlockProps) {
  const topics = home.feeds[w.id] ?? [];
  const more =
    w.source === "category" && w.category
      ? `/c/${w.category}`
      : w.source
        ? SOURCE_LINK[w.source]
        : undefined;
  const action = more ? (
    <Link href={more} className="text-sm text-primary hover:underline">
      Ver mais
    </Link>
  ) : undefined;
  return (
    <Block title={w.title} action={action}>
      {topics.length === 0 ? (
        <Empty text="Nenhum tópico por aqui ainda." />
      ) : w.variant === "cards" ? (
        <ul className="grid gap-3 sm:grid-cols-2">
          {topics.map((t) => (
            <TopicCard key={t.id} t={t} />
          ))}
        </ul>
      ) : w.variant === "compact" ? (
        <ol className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
          {topics.map((t, i) => (
            <li key={t.id}>
              <Link
                href={`/t/${t.id}`}
                className="flex items-center gap-3 px-4 py-2.5 text-sm hover:bg-secondary/60"
              >
                <span className="w-5 text-right font-semibold tabular-nums text-muted-foreground">
                  {i + 1}
                </span>
                <span className="min-w-0 flex-1 truncate font-medium">
                  {t.title}
                </span>
                <span className="flex items-center gap-1 text-xs text-muted-foreground">
                  <MessageSquare className="h-3.5 w-3.5" />
                  {formatCount(t.replyCount)}
                </span>
              </Link>
            </li>
          ))}
        </ol>
      ) : (
        <ul className="space-y-2">
          {topics.map((t) => (
            <TopicRow key={t.id} topic={t} />
          ))}
        </ul>
      )}
    </Block>
  );
}

function TopicCard({ t }: { t: Topic }) {
  return (
    <li>
      <Link
        href={`/t/${t.id}`}
        className="flex h-full flex-col rounded-xl border border-border bg-card p-4 transition-all hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md"
      >
        <span className="line-clamp-2 font-semibold leading-snug">
          {t.title}
        </span>
        <span className="mt-1.5 line-clamp-3 flex-1 text-sm text-muted-foreground">
          {t.content}
        </span>
        <span className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
          <UserAvatar
            name={t.author.displayName}
            url={t.author.avatarUrl}
            className="h-5 w-5 text-[10px]"
          />
          <span className="truncate">{t.author.displayName}</span>
          <span className="ml-auto flex shrink-0 items-center gap-2.5">
            <span className="flex items-center gap-1">
              <MessageSquare className="h-3 w-3" />
              {formatCount(t.replyCount)}
            </span>
            <span className="flex items-center gap-1">
              <Heart className="h-3 w-3" />
              {formatCount(t.likeCount)}
            </span>
            <span className="flex items-center gap-1">
              <Eye className="h-3 w-3" />
              {formatCount(t.viewCount)}
            </span>
          </span>
        </span>
      </Link>
    </li>
  );
}

function Members({ w, home }: BlockProps) {
  const users = home.featuredUsers.slice(0, w.limit);
  if (users.length === 0) return null;
  return (
    <Block title={w.title}>
      {w.variant === "avatars" ? (
        <ul className="flex flex-wrap gap-2 rounded-xl border border-border bg-card p-4">
          {users.map((u) => (
            <li key={u.id}>
              <Link
                href={`/u/${u.username}`}
                title={u.displayName}
                className="block rounded-full ring-2 ring-transparent transition hover:ring-primary"
              >
                <UserAvatar
                  name={u.displayName}
                  url={u.avatarUrl}
                  className="h-10 w-10"
                />
              </Link>
            </li>
          ))}
        </ul>
      ) : (
        <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
          {users.map((u) => (
            <li key={u.id}>
              <Link
                href={`/u/${u.username}`}
                className="flex items-center gap-3 px-4 py-2.5 hover:bg-secondary/60"
              >
                <UserAvatar
                  name={u.displayName}
                  url={u.avatarUrl}
                  className="h-8 w-8"
                />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium">
                    {u.displayName}
                  </span>
                  <span className="block text-xs text-muted-foreground">
                    @{u.username}
                  </span>
                </span>
                <span className="text-right text-xs tabular-nums text-muted-foreground">
                  {formatCount(u.followers)} seg.
                </span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </Block>
  );
}

function ActivityFeed({ w, home }: BlockProps) {
  const items = home.activity.slice(0, w.limit);
  if (items.length === 0) return null;
  return (
    <Block title={w.title}>
      <ul className="space-y-3 rounded-xl border border-border bg-card p-4">
        {items.map((a) => (
          <li
            key={`${a.kind}-${a.id}`}
            className="flex items-start gap-3 text-sm"
          >
            <UserAvatar
              name={a.actor.displayName}
              url={a.actor.avatarUrl}
              className="h-7 w-7 text-xs"
            />
            <p className="min-w-0 flex-1">
              <Link
                href={`/u/${a.actor.username}`}
                className="font-medium hover:underline"
              >
                {a.actor.displayName}
              </Link>{" "}
              <span className="text-muted-foreground">
                {a.kind === "topic" ? "abriu" : "respondeu em"}
              </span>{" "}
              <Link
                href={`/t/${a.topicId}`}
                className="font-medium hover:underline"
              >
                {a.topicTitle}
              </Link>
              <span className="block text-xs text-muted-foreground">
                {timeAgo(a.createdAt)}
              </span>
            </p>
            <ActivityIcon className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          </li>
        ))}
      </ul>
    </Block>
  );
}
