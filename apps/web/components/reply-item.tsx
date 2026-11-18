"use client";

import { Heart, Loader2, Pencil, Reply as ReplyIcon, Trash2, X } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { ReportButton } from "@/components/report-button";
import { RichText } from "@/components/rich-text";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { api, type Reply } from "@/lib/api";
import { cn, formatCount, timeAgo } from "@/lib/utils";

export interface ReplyNode extends Reply {
  children: ReplyNode[];
}

interface Props {
  node: ReplyNode;
  topicLocked: boolean;
  depth: number;
  onReplyCreated: (reply: Reply) => void;
  onReplyUpdated: (reply: Reply) => void;
  onReplyDeleted: (id: string) => void;
}

export function ReplyItem({ node, topicLocked, depth, onReplyCreated, onReplyUpdated, onReplyDeleted }: Props) {
  const { user, hasPermission } = useAuth();
  const [liked, setLiked] = useState(node.liked);
  const [likeCount, setLikeCount] = useState(node.likeCount);
  const [replying, setReplying] = useState(false);
  const [editing, setEditing] = useState(false);
  const [busy, setBusy] = useState(false);

  const isAuthor = user?.id === node.author.id;
  const canDelete = isAuthor || hasPermission("replies.delete");
  const canEdit = isAuthor || hasPermission("replies.edit");

  async function toggleLike() {
    if (!user || busy) return;
    setBusy(true);
    try {
      const { data } = await api.likeReply(node.id);
      setLiked(Boolean(data.liked));
      setLikeCount(data.count);
    } catch {
      /* ignore */
    } finally {
      setBusy(false);
    }
  }

  async function onDelete() {
    if (!window.confirm("Excluir esta resposta?")) return;
    setBusy(true);
    try {
      await api.deleteReply(node.id);
      onReplyDeleted(node.id);
    } catch {
      setBusy(false);
    }
  }

  return (
    <div className={cn(depth > 0 && "ml-4 border-l border-border pl-4 sm:ml-6 sm:pl-5")}>
      <div
        id={`reply-${node.id}`}
        className="scroll-mt-24 rounded-lg border border-border bg-card p-4 text-card-foreground target:ring-2 target:ring-primary"
      >
        {node.isDeleted ? (
          <p className="text-sm italic text-muted-foreground">resposta removida</p>
        ) : (
          <>
            <div className="mb-2 flex items-center gap-2 text-xs text-muted-foreground">
              <Link href={`/u/${node.author.username}`} className="font-medium text-foreground hover:underline">
                {node.author.displayName}
              </Link>
              <span>· {timeAgo(node.createdAt)}</span>
              {node.editedAt && <span>· editada</span>}
            </div>

            {editing ? (
              <EditForm
                initial={node.content}
                busy={busy}
                onCancel={() => setEditing(false)}
                onSave={async (content) => {
                  setBusy(true);
                  try {
                    const { data } = await api.updateReply(node.id, { content });
                    onReplyUpdated(data);
                    setEditing(false);
                  } finally {
                    setBusy(false);
                  }
                }}
              />
            ) : (
              <RichText text={node.content} className="text-sm" />
            )}

            {!editing && (
              <div className="mt-3 flex flex-wrap items-center gap-1 text-xs">
                <button
                  type="button"
                  onClick={toggleLike}
                  disabled={!user || busy}
                  className={cn(
                    "inline-flex items-center gap-1 rounded-md px-2 py-1 hover:bg-secondary disabled:opacity-50",
                    liked ? "text-primary" : "text-muted-foreground",
                  )}
                >
                  <Heart className={cn("h-3.5 w-3.5", liked && "fill-current")} />
                  {formatCount(likeCount)}
                </button>
                {user && !topicLocked && (
                  <button
                    type="button"
                    onClick={() => setReplying((v) => !v)}
                    className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-muted-foreground hover:bg-secondary"
                  >
                    <ReplyIcon className="h-3.5 w-3.5" />
                    Responder
                  </button>
                )}
                {canEdit && (
                  <button
                    type="button"
                    onClick={() => setEditing(true)}
                    className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-muted-foreground hover:bg-secondary"
                  >
                    <Pencil className="h-3.5 w-3.5" />
                    Editar
                  </button>
                )}
                {canDelete && (
                  <button
                    type="button"
                    onClick={onDelete}
                    disabled={busy}
                    className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-muted-foreground hover:bg-secondary hover:text-destructive disabled:opacity-50"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    Excluir
                  </button>
                )}
                {!node.isDeleted && (
                  <ReportButton targetType="reply" targetId={node.id} ownerId={node.author.id} variant="inline" />
                )}
              </div>
            )}

            {replying && (
              <div className="mt-3">
                <ReplyForm
                  topicId={node.topicId}
                  parentId={node.id}
                  onCancel={() => setReplying(false)}
                  onCreated={(reply) => {
                    onReplyCreated(reply);
                    setReplying(false);
                  }}
                />
              </div>
            )}
          </>
        )}
      </div>

      {node.children.length > 0 && (
        <div className="mt-2 space-y-2">
          {node.children.map((child) => (
            <ReplyItem
              key={child.id}
              node={child}
              topicLocked={topicLocked}
              depth={depth + 1}
              onReplyCreated={onReplyCreated}
              onReplyUpdated={onReplyUpdated}
              onReplyDeleted={onReplyDeleted}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function EditForm({
  initial,
  busy,
  onSave,
  onCancel,
}: {
  initial: string;
  busy: boolean;
  onSave: (content: string) => void;
  onCancel: () => void;
}) {
  const [value, setValue] = useState(initial);
  return (
    <div className="space-y-2">
      <Textarea value={value} onChange={(e) => setValue(e.target.value)} rows={4} />
      <div className="flex gap-2">
        <Button size="sm" disabled={busy || !value.trim()} onClick={() => onSave(value.trim())}>
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Salvar
        </Button>
        <Button size="sm" variant="ghost" onClick={onCancel}>
          <X className="h-4 w-4" />
          Cancelar
        </Button>
      </div>
    </div>
  );
}

export function ReplyForm({
  topicId,
  parentId,
  onCreated,
  onCancel,
}: {
  topicId: string;
  parentId?: string | null;
  onCreated: (reply: Reply) => void;
  onCancel?: () => void;
}) {
  const [content, setContent] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      const { data } = await api.createReply(topicId, { content: content.trim(), parentId: parentId ?? null });
      setContent("");
      onCreated(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao responder");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="space-y-2">
      <Textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        rows={parentId ? 3 : 4}
        placeholder={parentId ? "Escreva uma resposta…" : "Participe da discussão…"}
        required
      />
      {error && <p className="text-sm text-destructive">{error}</p>}
      <div className="flex gap-2">
        <Button size="sm" type="submit" disabled={busy || !content.trim()}>
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Responder
        </Button>
        {onCancel && (
          <Button size="sm" type="button" variant="ghost" onClick={onCancel}>
            Cancelar
          </Button>
        )}
      </div>
    </form>
  );
}
