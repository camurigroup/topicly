"use client";

import { Ban, Loader2, MessageSquareWarning, VolumeX } from "lucide-react";
import { useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { DURATIONS, modApi } from "@/lib/moderation-api";
import { cn } from "@/lib/utils";

type Kind = "warn" | "mute" | "ban";

const KINDS: { kind: Kind; label: string; perm: string; icon: typeof Ban; hint: string }[] = [
  { kind: "warn", label: "Advertir", perm: "warnings.apply", icon: MessageSquareWarning, hint: "Registra um aviso formal e notifica o usuário." },
  { kind: "mute", label: "Silenciar", perm: "users.mute", icon: VolumeX, hint: "Impede criar tópicos, responder, curtir e seguir." },
  { kind: "ban", label: "Banir", perm: "users.ban", icon: Ban, hint: "Bloqueia o acesso e encerra todas as sessões imediatamente." },
];

interface ModerationActionsProps {
  username: string;
  onDone: () => Promise<void>;
}

export function ModerationActions({ username, onDone }: ModerationActionsProps) {
  const { hasPermission } = useAuth();
  const available = KINDS.filter((k) => hasPermission(k.perm));
  const [kind, setKind] = useState<Kind>(available[0]?.kind ?? "warn");
  const [reason, setReason] = useState("");
  const [hours, setHours] = useState(24);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  if (available.length === 0) return null;
  const current = available.find((k) => k.kind === kind) ?? available[0];

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (kind === "ban" && !window.confirm(`Banir @${username}? As sessões dele serão encerradas.`)) return;
    setBusy(true);
    setError(null);
    setSuccess(null);
    try {
      const text = reason.trim();
      if (kind === "warn") await modApi.warn(username, text);
      if (kind === "mute") await modApi.mute(username, { reason: text, durationHours: hours });
      if (kind === "ban") await modApi.ban(username, { reason: text, durationHours: hours });
      setReason("");
      setSuccess(`${current.label}: ação aplicada e registrada na auditoria.`);
      await onDone();
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao aplicar ação");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-xl border border-border bg-card p-5">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Aplicar ação</h2>
      <div role="radiogroup" aria-label="Tipo de ação" className="grid grid-cols-3 gap-1 rounded-lg bg-secondary p-1">
        {available.map((k) => {
          const Icon = k.icon;
          return (
            <button
              key={k.kind}
              type="button"
              role="radio"
              aria-checked={kind === k.kind}
              onClick={() => setKind(k.kind)}
              className={cn(
                "flex flex-col items-center gap-1 rounded-md px-2 py-2 text-xs font-medium transition-all",
                kind === k.kind ? "bg-background shadow-sm" : "text-muted-foreground hover:text-foreground",
                kind === k.kind && k.kind === "ban" && "text-destructive",
              )}
            >
              <Icon className="h-4 w-4" />
              {k.label}
            </button>
          );
        })}
      </div>
      <p className="text-xs text-muted-foreground">{current.hint}</p>

      <div className="space-y-1.5">
        <Label htmlFor="mod-reason">Motivo</Label>
        <Textarea
          id="mod-reason"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          rows={3}
          minLength={3}
          maxLength={500}
          required
          placeholder="Descreva o motivo (visível para o usuário)"
        />
      </div>

      {kind !== "warn" && (
        <div className="space-y-1.5">
          <Label htmlFor="mod-duration">Duração</Label>
          <select
            id="mod-duration"
            value={hours}
            onChange={(e) => setHours(Number(e.target.value))}
            className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
          >
            {DURATIONS.map((d) => (
              <option key={d.hours} value={d.hours}>
                {d.label}
              </option>
            ))}
          </select>
        </div>
      )}

      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {success && <p role="status" className="text-sm text-emerald-600 dark:text-emerald-400">{success}</p>}

      <Button type="submit" className="w-full" variant={kind === "ban" ? "destructive" : "default"} disabled={busy}>
        {busy && <Loader2 className="h-4 w-4 animate-spin" />}
        {current.label} @{username}
      </Button>
    </form>
  );
}
