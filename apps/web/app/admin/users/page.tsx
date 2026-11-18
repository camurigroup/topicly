"use client";

import { Ban, Loader2, MessageSquareWarning, Search, ShieldOff, VolumeX } from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useCallback, useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { ModerationShell } from "@/components/moderation-nav";
import { UserAvatar } from "@/components/user-avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { modApi, type Sanction, type UserModeration } from "@/lib/moderation-api";
import { cn } from "@/lib/utils";

import { ModerationActions } from "./actions";
import { RolesEditor } from "./roles-editor";

const fmt = (iso: string) =>
  new Date(iso).toLocaleString("pt-BR", { dateStyle: "short", timeStyle: "short" });

export default function AdminUsersPage() {
  return (
    <ModerationShell
      title="Moderar usuário"
      description="Advertências, silenciamentos, banimentos e cargos. Você só age sobre quem está abaixo de você na hierarquia."
      perms={["users.ban", "users.mute", "warnings.apply", "roles.manage"]}
    >
      <Suspense fallback={null}>
        <UserModerationView />
      </Suspense>
    </ModerationShell>
  );
}

function UserModerationView() {
  const params = useSearchParams();
  const router = useRouter();
  const { hasPermission } = useAuth();
  const username = params.get("u") ?? "";
  const [input, setInput] = useState(username);
  const [data, setData] = useState<UserModeration | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!username) {
      setData(null);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const { data } = await modApi.userModeration(username);
      setData(data);
    } catch (err) {
      setData(null);
      setError(err instanceof Error ? err.message : "usuário não encontrado");
      return;
    } finally {
      setLoading(false);
    }
  }, [username]);

  useEffect(() => {
    setInput(username);
    void load();
  }, [username, load]);

  function onSearch(e: React.FormEvent) {
    e.preventDefault();
    const u = input.trim().replace(/^@/, "");
    if (u) router.push(`/admin/users?u=${encodeURIComponent(u)}`);
  }

  return (
    <div className="space-y-6">
      <form onSubmit={onSearch} className="flex max-w-md gap-2" role="search">
        <div className="relative flex-1">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="@username"
            className="pl-9"
            aria-label="Username"
            maxLength={32}
          />
        </div>
        <Button type="submit">Buscar</Button>
      </form>

      {loading && !data && (
        <div className="flex justify-center py-12 text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      )}
      {error && !loading && <p role="alert" className="text-sm text-destructive">{error}</p>}

      {data && (
        <div className="grid gap-6 lg:grid-cols-[1fr_340px]">
          <div className="space-y-6">
            <section className="flex flex-wrap items-center gap-4 rounded-xl border border-border bg-card p-5">
              <UserAvatar name={data.user.displayName} url={data.user.avatarUrl ?? null} />
              <div className="min-w-0 flex-1">
                <Link href={`/u/${data.user.username}`} className="text-lg font-semibold hover:underline">
                  {data.user.displayName}
                </Link>
                <p className="text-sm text-muted-foreground">@{data.user.username}</p>
                <div className="mt-2 flex flex-wrap gap-1.5">
                  {data.roles.length === 0 && <Chip>Membro</Chip>}
                  {data.roles.map((r) => (
                    <Chip key={r}>{r}</Chip>
                  ))}
                  {data.activeBan && <Chip tone="danger">Banido</Chip>}
                  {data.activeMute && <Chip tone="warn">Silenciado</Chip>}
                  {data.warnings.length > 0 && <Chip tone="warn">{data.warnings.length} advertência(s)</Chip>}
                </div>
              </div>
            </section>

            <History data={data} onChanged={load} />
          </div>

          <aside className="space-y-6">
            {data.canAct ? (
              <ModerationActions username={data.user.username} onDone={load} />
            ) : (
              <p className="flex items-start gap-2 rounded-xl border border-dashed border-border p-4 text-sm text-muted-foreground">
                <ShieldOff className="mt-0.5 h-4 w-4 shrink-0" />
                {data.isOwner
                  ? "A conta do Dono não pode ser moderada."
                  : "Este usuário está no seu nível hierárquico ou acima (ou é você)."}
              </p>
            )}
            {hasPermission("roles.manage") && data.canAct && (
              <RolesEditor username={data.user.username} current={data.roles} onSaved={load} />
            )}
          </aside>
        </div>
      )}
    </div>
  );
}

function Chip({ children, tone }: { children: React.ReactNode; tone?: "danger" | "warn" }) {
  return (
    <span
      className={cn(
        "rounded-full border px-2 py-0.5 text-xs font-medium",
        tone === "danger" && "border-destructive/40 bg-destructive/10 text-destructive",
        tone === "warn" && "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300",
        !tone && "border-border text-muted-foreground",
      )}
    >
      {children}
    </span>
  );
}

type Event =
  | { kind: "sanction"; at: string; s: Sanction }
  | { kind: "warning"; id: string; at: string; reason: string; by: string | null };

function History({ data, onChanged }: { data: UserModeration; onChanged: () => Promise<void> }) {
  const events: Event[] = [
    ...data.sanctions.map((s): Event => ({ kind: "sanction", at: s.createdAt, s })),
    ...data.warnings.map(
      (w): Event => ({ kind: "warning", id: w.id, at: w.createdAt, reason: w.reason, by: w.issuedBy?.username ?? null }),
    ),
  ].sort((a, b) => b.at.localeCompare(a.at));

  return (
    <section>
      <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">Histórico</h2>
      {events.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border py-8 text-center text-sm text-muted-foreground">
          Ficha limpa — nenhuma ação de moderação.
        </p>
      ) : (
        <ol className="relative ml-2.5 space-y-5 border-l border-border pl-6">
          {events.map((e) =>
            e.kind === "warning" ? (
              <li key={e.id} className="relative">
                <Dot className="bg-amber-500" icon={<MessageSquareWarning className="h-3 w-3" />} />
                <p className="text-sm font-medium">Advertência</p>
                <p className="text-sm text-muted-foreground">{e.reason}</p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {fmt(e.at)} {e.by && `· por @${e.by}`}
                </p>
              </li>
            ) : (
              <SanctionItem key={e.s.id} s={e.s} onChanged={onChanged} />
            ),
          )}
        </ol>
      )}
    </section>
  );
}

function Dot({ className, icon }: { className: string; icon: React.ReactNode }) {
  return (
    <span
      className={cn(
        "absolute -left-[35px] top-0 flex h-5 w-5 items-center justify-center rounded-full text-white ring-4 ring-background",
        className,
      )}
    >
      {icon}
    </span>
  );
}

function SanctionItem({ s, onChanged }: { s: Sanction; onChanged: () => Promise<void> }) {
  const [revoking, setRevoking] = useState(false);
  const [reason, setReason] = useState("");
  const [error, setError] = useState<string | null>(null);
  const isBan = s.type === "ban";

  async function revoke(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await modApi.revoke(s.id, reason.trim());
      await onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao revogar");
    }
  }

  return (
    <li className="relative">
      <Dot
        className={s.active ? (isBan ? "bg-destructive" : "bg-orange-500") : "bg-muted-foreground/50"}
        icon={isBan ? <Ban className="h-3 w-3" /> : <VolumeX className="h-3 w-3" />}
      />
      <p className="text-sm font-medium">
        {isBan ? "Banimento" : "Silenciamento"} {s.expiresAt ? `até ${fmt(s.expiresAt)}` : "permanente"}
        {s.active && <span className="ml-2 text-xs font-semibold text-destructive">ATIVO</span>}
      </p>
      <p className="text-sm text-muted-foreground">{s.reason}</p>
      <p className="mt-0.5 text-xs text-muted-foreground">
        {fmt(s.createdAt)} {s.issuedBy && `· por @${s.issuedBy.username}`}
        {s.revokedAt && ` · revogado ${fmt(s.revokedAt)}${s.revokedBy ? ` por @${s.revokedBy.username}` : ""}`}
        {s.revokeReason && ` (${s.revokeReason})`}
      </p>
      {s.active &&
        (revoking ? (
          <form onSubmit={revoke} className="mt-2 flex flex-wrap gap-2">
            <Input
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Motivo da revogação"
              className="h-8 flex-1 basis-48"
              minLength={3}
              maxLength={500}
              required
              autoFocus
            />
            <Button size="sm" type="submit">
              Confirmar
            </Button>
            <Button size="sm" type="button" variant="ghost" onClick={() => setRevoking(false)}>
              Cancelar
            </Button>
          </form>
        ) : (
          <Button size="sm" variant="outline" className="mt-2" onClick={() => setRevoking(true)}>
            Revogar
          </Button>
        ))}
      {error && <p role="alert" className="mt-1 text-xs text-destructive">{error}</p>}
    </li>
  );
}
