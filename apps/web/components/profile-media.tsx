"use client";

import { Camera, ImagePlus, Loader2 } from "lucide-react";
import { useRef, useState } from "react";

import { api, type Account } from "@/lib/api";
import { cn } from "@/lib/utils";

/** Must match the API allow-list (storage.DetectImage) and size limit. */
export const ACCEPTED_IMAGES = "image/png,image/jpeg,image/webp,image/gif";
const ACCEPTED = ACCEPTED_IMAGES.split(",");
export const MAX_IMAGE_BYTES = 5 * 1024 * 1024;

/** Client-side check so the user gets an immediate, clear message. */
export function checkImage(file: File): string | null {
  if (!ACCEPTED.includes(file.type)) return "Formato não suportado. Use JPG, PNG, WebP ou GIF.";
  if (file.size > MAX_IMAGE_BYTES) return "A imagem passa de 5 MB.";
  return null;
}

interface ProfileMediaProps {
  account: Account;
  onUpdated: (a: Account) => void;
}

/** Banner + avatar editor with live preview, validation and progress. */
export function ProfileMedia({ account, onUpdated }: ProfileMediaProps) {
  const avatarRef = useRef<HTMLInputElement>(null);
  const bannerRef = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState<"avatar" | "banner" | null>(null);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  async function upload(kind: "avatar" | "banner", input: HTMLInputElement) {
    const file = input.files?.[0];
    input.value = ""; // allow picking the same file again
    if (!file) return;
    const problem = checkImage(file);
    if (problem) {
      setMsg({ ok: false, text: problem });
      return;
    }
    setBusy(kind);
    setMsg(null);
    try {
      const { data } = kind === "avatar" ? await api.uploadAvatar(file) : await api.uploadBanner(file);
      onUpdated(data);
      setMsg({ ok: true, text: kind === "avatar" ? "Foto de perfil atualizada." : "Capa atualizada." });
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha no envio" });
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="space-y-2">
      <div className="relative">
        <button
          type="button"
          onClick={() => bannerRef.current?.click()}
          disabled={busy !== null}
          className="group relative block h-32 w-full overflow-hidden rounded-xl border border-border bg-secondary"
          aria-label="Trocar imagem de capa"
        >
          {account.bannerUrl && (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={account.bannerUrl} alt="" className="h-full w-full object-cover" />
          )}
          <span className="absolute inset-0 flex items-center justify-center gap-2 bg-black/40 text-sm font-medium text-white opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100">
            {busy === "banner" ? <Loader2 className="h-4 w-4 animate-spin" /> : <ImagePlus className="h-4 w-4" />}
            {account.bannerUrl ? "Trocar capa" : "Adicionar capa"}
          </span>
        </button>

        <button
          type="button"
          onClick={() => avatarRef.current?.click()}
          disabled={busy !== null}
          className="group absolute -bottom-8 left-4 h-20 w-20 overflow-hidden rounded-full border-4 border-card bg-muted"
          aria-label="Trocar foto de perfil"
        >
          {account.avatarUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={account.avatarUrl} alt="" className="h-full w-full object-cover" />
          ) : (
            <span className="flex h-full w-full items-center justify-center text-2xl font-semibold text-muted-foreground">
              {account.displayName.charAt(0).toUpperCase()}
            </span>
          )}
          <span
            className={cn(
              "absolute inset-0 flex items-center justify-center bg-black/45 text-white transition-opacity",
              busy === "avatar" ? "opacity-100" : "opacity-0 group-hover:opacity-100 group-focus-visible:opacity-100",
            )}
          >
            {busy === "avatar" ? <Loader2 className="h-5 w-5 animate-spin" /> : <Camera className="h-5 w-5" />}
          </span>
        </button>
      </div>

      <input
        ref={avatarRef}
        type="file"
        accept={ACCEPTED_IMAGES}
        className="hidden"
        onChange={(e) => upload("avatar", e.currentTarget)}
      />
      <input
        ref={bannerRef}
        type="file"
        accept={ACCEPTED_IMAGES}
        className="hidden"
        onChange={(e) => upload("banner", e.currentTarget)}
      />

      <p className="pl-28 text-xs text-muted-foreground">JPG, PNG, WebP ou GIF · até 5 MB. Clique na foto ou na capa para trocar.</p>
      {msg && (
        <p role={msg.ok ? "status" : "alert"} className={cn("pt-1 text-sm", msg.ok ? "text-primary" : "text-destructive")}>
          {msg.text}
        </p>
      )}
    </div>
  );
}
