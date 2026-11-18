"use client";

import { Bell, Eye, Heart, Loader2, Lock, Pin, Sparkles } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { ReplyForm, ReplyItem, type ReplyNode } from "@/components/reply-item";
import { Button } from "@/components/ui/button";
import { api, type Reply, type Topic } from "@/lib/api";
import { cn, formatCount, timeAgo } from "@/lib/utils";

/** Build a parent→children tree from a flat reply list, preserving order. */
function buildTree(replies: Reply[]): ReplyNode[] {
  const byId = new Map<string, ReplyNode>();
  const roots: ReplyNode[] = [];
  for (const r of replies) byId.set(r.id, { ...r, children: [] });
  for (const r of replies) {
    const node = byId.get(r.id)!;
    const parent = r.parentId ? byId.get(r.parentId) : undefined;
    if (parent) parent.children.push(node);
    else roots.push(node);
  }
  return roots;
}

export default function TopicPage() {
  const { id } = useParams<{ id: string }>();
  const { user, hasPermission } = useAuth();

  const [topic, setTopic] = useState<Topic | null>(null);
  const [replies, setReplies] = useState<Reply[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [{ data: t }, replyPage] = await Promise.all([api.topic(id), api.replies(id, { limit: 100 })]);
      setTopic(t);
      setReplies(replyPage.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "erro");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void load();
  }, [load]);

  const tree = useMemo(() => buildTree(replies), [replies]);

  async function toggleLike() {
    if (!topic || !user || busy) return;
    setBusy(true);
    try {
      const { data } = await api.likeTopic(topic.id);
      setTopic({ ...topic, liked: Boolean(data.liked), likeCount: data.count });
    } finally {
      setBusy(false);
    }
  }

  async function toggleFollow() {
    if (!topic || !user || busy) return;
    setBusy(true);
    try {
      const { data } = await api.followTopic(topic.id);
      setTopic({ ...topic, following: Boolean(data.following), followerCount: data.count });
    } finally {
      setBusy(false);
    }
  }

  async function setFlag(kind: "pin" | "lock" | "feature", value: boolean) {
    if (!topic) return;
    setBusy(true);
    try {
      if (kind === "pin") await api.setTopicPin(topic.id, value);
      if (kind === "lock") await api.setTopicLock(topic.id, value);
      if (kind === "feature") await api.setTopicFeature(topic.id, value);
      setTopic({
        ...topic,
        isPinned: kind === "pin" ? value : topic.isPinned,
        isLocked: kind === "lock" ? value : topic.isLocked,
        isFeatured: kind === "feature" ? value : topic.isFeatured,
      });
    } finally {
      setBusy(false);
    }
  }

  function onReplyCreated(reply: Reply) {
    setReplies((prev) => [...prev, reply]);
    setTopic((t) => (t ? { ...t, replyCount: t.replyCount + 1 } : t));
  }
  function onReplyUpdated(reply: Reply) {
    setReplies((prev) => prev.map((r) => (r.id === reply.id ? reply : r)));
  }
  function onReplyDeleted(deletedId: string) {
    setReplies((prev) => prev.map((r) => (r.id === deletedId ? { ...r, isDeleted: true, content: "" } : r)));
    setTopic((t) => (t ? { ...t, replyCount: Math.max(0, t.replyCount - 1) } : t));
  }

  if (loading && !topic) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (error && !topic) {
    return <p className="text-center text-destructive">Tópico não encontrado ({error}).</p>;
  }

  if (!topic) return null;

  const canPin = hasPermission("topics.pin");
  const canLock = hasPermission("topics.lock");

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <nav className="text-sm text-muted-foreground">
        <Link href="/categories" className="hover:underline">
          Categorias
        </Link>{" "}
        /{" "}
        <Link href={`/c/${topic.categorySlug}`} className="hover:underline">
          {topic.categorySlug}
        </Link>
      </nav>

      <article className="rounded-xl border border-border bg-card p-6 text-card-foreground">
        <div className="mb-3 flex flex-wrap items-center gap-2">
          {topic.isPinned && <Badge icon={<Pin className="h-3 w-3" />}>Fixado</Badge>}
          {topic.isFeatured && <Badge icon={<Sparkles className="h-3 w-3" />}>Destaque</Badge>}
          {topic.isLocked && <Badge icon={<Lock className="h-3 w-3" />}>Bloqueado</Badge>}
        </div>

        <h1 className="text-2xl font-bold tracking-tight">{topic.title}</h1>
        <p className="mt-1 text-xs text-muted-foreground">
          por{" "}
          <Link href={`/u/${topic.author.username}`} className="hover:underline">
            {topic.author.displayName}
          </Link>{" "}
          · {timeAgo(topic.createdAt)}
          <span className="ml-2 inline-flex items-center gap-1">
            <Eye className="h-3.5 w-3.5" /> {formatCount(topic.viewCount)}
          </span>
        </p>

        <div className="mt-4 whitespace-pre-wrap text-sm leading-relaxed">{topic.content}</div>

        <div className="mt-5 flex flex-wrap items-center gap-2 border-t border-border pt-4">
          <Button
            variant={topic.liked ? "default" : "outline"}
            size="sm"
            disabled={!user || busy}
            onClick={toggleLike}
          >
            <Heart className={cn("h-4 w-4", topic.liked && "fill-current")} />
            {formatCount(topic.likeCount)}
          </Button>
          <Button
            variant={topic.following ? "default" : "outline"}
            size="sm"
            disabled={!user || busy}
            onClick={toggleFollow}
          >
            <Bell className={cn("h-4 w-4", topic.following && "fill-current")} />
            {topic.following ? "Seguindo" : "Seguir"} · {formatCount(topic.followerCount)}
          </Button>

          {canPin && (
            <Button variant="ghost" size="sm" disabled={busy} onClick={() => setFlag("pin", !topic.isPinned)}>
              <Pin className="h-4 w-4" />
              {topic.isPinned ? "Desafixar" : "Fixar"}
            </Button>
          )}
          {canPin && (
            <Button variant="ghost" size="sm" disabled={busy} onClick={() => setFlag("feature", !topic.isFeatured)}>
              <Sparkles className="h-4 w-4" />
              {topic.isFeatured ? "Remover destaque" : "Destacar"}
            </Button>
          )}
          {canLock && (
            <Button variant="ghost" size="sm" disabled={busy} onClick={() => setFlag("lock", !topic.isLocked)}>
              <Lock className="h-4 w-4" />
              {topic.isLocked ? "Desbloquear" : "Bloquear"}
            </Button>
          )}
        </div>
      </article>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
          {formatCount(topic.replyCount)} {topic.replyCount === 1 ? "resposta" : "respostas"}
        </h2>

        {tree.length > 0 && (
          <div className="space-y-2">
            {tree.map((node) => (
              <ReplyItem
                key={node.id}
                node={node}
                topicLocked={topic.isLocked}
                depth={0}
                onReplyCreated={onReplyCreated}
                onReplyUpdated={onReplyUpdated}
                onReplyDeleted={onReplyDeleted}
              />
            ))}
          </div>
        )}

        {topic.isLocked ? (
          <p className="rounded-lg border border-dashed border-border py-6 text-center text-sm text-muted-foreground">
            Este tópico está bloqueado para novas respostas.
          </p>
        ) : user ? (
          <div className="rounded-lg border border-border bg-card p-4">
            <ReplyForm topicId={topic.id} onCreated={onReplyCreated} />
          </div>
        ) : (
          <p className="rounded-lg border border-dashed border-border py-6 text-center text-sm text-muted-foreground">
            <Link href="/login" className="font-medium text-primary hover:underline">
              Entre
            </Link>{" "}
            para responder.
          </p>
        )}
      </section>
    </div>
  );
}

function Badge({ icon, children }: { icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full border border-border bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground">
      {icon}
      {children}
    </span>
  );
}
