"use client";

import { CalendarDays, Loader2, Star } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { FollowButton } from "@/components/follow-button";
import { ReportButton } from "@/components/report-button";
import { RichText } from "@/components/rich-text";
import { TopicRow } from "@/components/topic-row";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { api, type PublicProfile, type ReplyHit, type Topic, type UserStats } from "@/lib/api";
import { formatCount, timeAgo } from "@/lib/utils";

type State =
  | { kind: "loading" }
  | { kind: "ready"; profile: PublicProfile; stats: UserStats }
  | { kind: "error"; message: string };

type Tab = "topics" | "replies";

export default function ProfilePage() {
  const params = useParams<{ username: string }>();
  const username = params.username;
  const [state, setState] = useState<State>({ kind: "loading" });
  const [tab, setTab] = useState<Tab>("topics");

  useEffect(() => {
    let active = true;
    setState({ kind: "loading" });
    Promise.all([api.publicProfile(username), api.userStats(username)])
      .then(([p, s]) => active && setState({ kind: "ready", profile: p.data, stats: s.data }))
      .catch((err) =>
        active && setState({ kind: "error", message: err instanceof Error ? err.message : "erro" }),
      );
    return () => {
      active = false;
    };
  }, [username]);

  if (state.kind === "loading") {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (state.kind === "error") {
    return (
      <div className="mx-auto max-w-xl text-center text-muted-foreground">
        <p>Perfil não encontrado.</p>
      </div>
    );
  }

  const { profile, stats } = state;
  const joined = new Date(profile.createdAt).toLocaleDateString("pt-BR", {
    year: "numeric",
    month: "long",
  });

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <Card className="overflow-hidden">
        <div className="h-28 w-full bg-secondary">
          {profile.bannerUrl && (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={profile.bannerUrl} alt="" className="h-full w-full object-cover" />
          )}
        </div>
        <CardContent className="pt-0">
          <div className="-mt-10 flex flex-wrap items-end justify-between gap-4">
            <div className="flex items-end gap-4">
              <div className="h-20 w-20 overflow-hidden rounded-full border-4 border-card bg-muted">
                {profile.avatarUrl ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={profile.avatarUrl} alt={profile.displayName} className="h-full w-full object-cover" />
                ) : (
                  <div className="flex h-full w-full items-center justify-center text-2xl font-semibold text-muted-foreground">
                    {profile.displayName.charAt(0).toUpperCase()}
                  </div>
                )}
              </div>
              <div className="pb-1">
                <h1 className="text-xl font-bold leading-tight">{profile.displayName}</h1>
                <p className="text-sm text-muted-foreground">@{profile.username}</p>
              </div>
            </div>
            {!stats.isSelf && (
              <div className="flex items-center gap-1">
                <ReportButton targetType="user" targetId={profile.id} ownerId={profile.id} />
                <FollowButton
                  following={stats.isFollowing}
                  onToggle={() => api.followUser(profile.username)}
                  onChange={(res) =>
                    setState({ kind: "ready", profile, stats: { ...stats, isFollowing: res.following, followers: res.count } })
                  }
                />
              </div>
            )}
          </div>

          {profile.bio && <p className="mt-4 text-sm">{profile.bio}</p>}

          <div className="mt-4 flex flex-wrap gap-x-5 gap-y-2 text-sm">
            <Link href={`/u/${profile.username}/followers`} className="hover:underline">
              <strong>{formatCount(stats.followers)}</strong>{" "}
              <span className="text-muted-foreground">seguidor{stats.followers === 1 ? "" : "es"}</span>
            </Link>
            <Link href={`/u/${profile.username}/following`} className="hover:underline">
              <strong>{formatCount(stats.following)}</strong> <span className="text-muted-foreground">seguindo</span>
            </Link>
            <span>
              <strong>{formatCount(stats.topics)}</strong>{" "}
              <span className="text-muted-foreground">tópico{stats.topics === 1 ? "" : "s"}</span>
            </span>
            <span>
              <strong>{formatCount(stats.replies)}</strong>{" "}
              <span className="text-muted-foreground">resposta{stats.replies === 1 ? "" : "s"}</span>
            </span>
          </div>

          <div className="mt-3 flex flex-wrap gap-4 text-sm text-muted-foreground">
            <span className="inline-flex items-center gap-1.5">
              <Star className="h-4 w-4" /> {profile.reputation} de reputação
            </span>
            <span className="inline-flex items-center gap-1.5">
              <CalendarDays className="h-4 w-4" /> Entrou em {joined}
            </span>
          </div>
        </CardContent>
      </Card>

      <div className="flex gap-2" role="tablist" aria-label="Conteúdo do perfil">
        {(
          [
            ["topics", "Tópicos"],
            ["replies", "Respostas"],
          ] as const
        ).map(([value, label]) => (
          <Button
            key={value}
            role="tab"
            aria-selected={tab === value}
            variant={tab === value ? "default" : "outline"}
            size="sm"
            onClick={() => setTab(value)}
          >
            {label}
          </Button>
        ))}
      </div>

      <AuthorContent username={profile.username} tab={tab} />
    </div>
  );
}

const PAGE_SIZE = 10;

function AuthorContent({ username, tab }: { username: string; tab: Tab }) {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [replies, setReplies] = useState<ReplyHit[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => setPage(1), [tab]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    api
      .search({ type: tab, author: username, sort: "recent", page, limit: PAGE_SIZE })
      .then(({ data, meta }) => {
        if (!active) return;
        setTopics(data.topics?.items ?? []);
        setReplies(data.replies?.items ?? []);
        setTotal(meta?.total ?? 0);
      })
      .catch(() => {})
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [username, tab, page]);

  if (loading) {
    return (
      <div className="flex justify-center py-8 text-muted-foreground">
        <Loader2 className="h-5 w-5 animate-spin" />
      </div>
    );
  }

  const empty = tab === "topics" ? topics.length === 0 : replies.length === 0;
  if (empty) {
    return (
      <p className="rounded-lg border border-dashed border-border py-10 text-center text-sm text-muted-foreground">
        {tab === "topics" ? "Nenhum tópico publicado." : "Nenhuma resposta publicada."}
      </p>
    );
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  return (
    <div className="space-y-4">
      {tab === "topics" ? (
        <ul className="space-y-2">
          {topics.map((t) => (
            <TopicRow key={t.id} topic={t} />
          ))}
        </ul>
      ) : (
        <ul className="space-y-2">
          {replies.map((r) => (
            <li key={r.id} className="rounded-lg border border-border bg-card p-4 text-card-foreground">
              <Link href={`/t/${r.topicId}#reply-${r.id}`} className="text-sm font-medium hover:underline">
                em “{r.topicTitle}”
              </Link>
              <span className="ml-2 text-xs text-muted-foreground">{timeAgo(r.createdAt)}</span>
              <RichText text={r.content} className="mt-2 line-clamp-3 text-sm text-muted-foreground" />
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
