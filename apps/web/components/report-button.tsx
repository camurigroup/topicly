"use client";

import { Check, Flag, Loader2, X } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { modApi, REPORT_REASONS, type ReportReason, type ReportTargetType } from "@/lib/moderation-api";
import { cn } from "@/lib/utils";

interface ReportButtonProps {
  targetType: ReportTargetType;
  targetId: string;
  /** Hide the button when the viewer owns the target. */
  ownerId?: string;
  /** "inline" matches the compact reply action bar. */
  variant?: "button" | "inline";
}

const TARGET_LABEL: Record<ReportTargetType, string> = {
  user: "este usuário",
  topic: "este tópico",
  reply: "esta resposta",
};

export function ReportButton({ targetType, targetId, ownerId, variant = "button" }: ReportButtonProps) {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);
  const [done, setDone] = useState(false);

  if (!user || user.id === ownerId) return null;

  const trigger =
    variant === "inline" ? (
      <button
        type="button"
        onClick={() => setOpen(true)}
        disabled={done}
        className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-muted-foreground hover:bg-secondary hover:text-destructive disabled:opacity-60"
      >
        {done ? <Check className="h-3.5 w-3.5" /> : <Flag className="h-3.5 w-3.5" />}
        {done ? "Denunciado" : "Denunciar"}
      </button>
    ) : (
      <Button variant="ghost" size="sm" onClick={() => setOpen(true)} disabled={done}>
        {done ? <Check className="h-4 w-4" /> : <Flag className="h-4 w-4" />}
        {done ? "Denunciado" : "Denunciar"}
      </Button>
    );

  return (
    <>
      {trigger}
      {open && (
        <ReportDialog
          targetType={targetType}
          targetId={targetId}
          onClose={() => setOpen(false)}
          onDone={() => {
            setDone(true);
            setOpen(false);
          }}
        />
      )}
    </>
  );
}

interface ReportDialogProps {
  targetType: ReportTargetType;
  targetId: string;
  onClose: () => void;
  onDone: () => void;
}

function ReportDialog({ targetType, targetId, onClose, onDone }: ReportDialogProps) {
  const titleId = useId();
  const firstRef = useRef<HTMLInputElement>(null);
  const [reason, setReason] = useState<ReportReason>("spam");
  const [details, setDetails] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    firstRef.current?.focus();
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await modApi.report({ targetType, targetId, reason, details: details.trim() || undefined });
      onDone();
    } catch (err) {
      setError(err instanceof Error ? err.message : "não foi possível enviar a denúncia");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 p-4 backdrop-blur-sm sm:items-center"
      onMouseDown={(e) => e.target === e.currentTarget && onClose()}
    >
      <form
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onSubmit={submit}
        className="w-full max-w-md rounded-xl border border-border bg-card p-5 text-card-foreground shadow-2xl"
      >
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 id={titleId} className="text-lg font-semibold">
              Denunciar {TARGET_LABEL[targetType]}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              A moderação vai analisar. Denúncias falsas podem gerar advertências.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Fechar"
            className="rounded-md p-1 text-muted-foreground hover:bg-secondary"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <fieldset className="mt-4 space-y-1.5">
          <legend className="mb-2 text-sm font-medium">Motivo</legend>
          {REPORT_REASONS.map((r, i) => (
            <label
              key={r.value}
              className={cn(
                "flex cursor-pointer items-center gap-3 rounded-lg border px-3 py-2 text-sm transition-colors",
                reason === r.value ? "border-primary bg-primary/5" : "border-border hover:bg-secondary",
              )}
            >
              <input
                ref={i === 0 ? firstRef : undefined}
                type="radio"
                name="reason"
                value={r.value}
                checked={reason === r.value}
                onChange={() => setReason(r.value)}
                className="accent-primary"
              />
              {r.label}
            </label>
          ))}
        </fieldset>

        <div className="mt-4 space-y-1.5">
          <Label htmlFor={`${titleId}-details`}>Detalhes (opcional)</Label>
          <Textarea
            id={`${titleId}-details`}
            value={details}
            onChange={(e) => setDetails(e.target.value)}
            maxLength={1000}
            rows={3}
            placeholder="Explique o que aconteceu…"
          />
        </div>

        {error && (
          <p role="alert" className="mt-3 text-sm text-destructive">
            {error}
          </p>
        )}

        <div className="mt-5 flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancelar
          </Button>
          <Button type="submit" variant="destructive" disabled={busy}>
            {busy && <Loader2 className="h-4 w-4 animate-spin" />}
            Enviar denúncia
          </Button>
        </div>
      </form>
    </div>
  );
}
