"use client";

import { Cloud, HardDrive } from "lucide-react";
import { useEffect, useState } from "react";

import { ModerationShell } from "@/components/moderation-nav";
import { adminApi, type StorageInfo } from "@/lib/admin-api";

export default function StoragePage() {
  const [info, setInfo] = useState<StorageInfo | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    adminApi
      .storage()
      .then(({ data }) => setInfo(data))
      .catch((err) => setError(err instanceof Error ? err.message : "falha ao carregar"));
  }, []);

  const local = info?.provider === "local";
  const rows: [string, string][] = info
    ? local
      ? [
          ["Diretório", info.localPath ?? "—"],
          ["URL pública", info.publicUrl ?? "—"],
        ]
      : [
          ["Endpoint", info.endpoint ?? "—"],
          ["Bucket", info.bucket ?? "—"],
          ["TLS", info.useSsl ? "ativado" : "desativado"],
        ]
    : [];

  return (
    <ModerationShell
      title="Armazenamento"
      description="Provedor de arquivos em uso. A troca é feita por variáveis de ambiente (STORAGE_PROVIDER) — segredos nunca aparecem aqui."
      perms={["storage.manage"]}
    >
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {info && (
        <div className="grid gap-4 md:grid-cols-2">
          <section className="rounded-xl border border-border bg-card p-6">
            <div className="flex items-center gap-3">
              <span className="flex h-11 w-11 items-center justify-center rounded-lg bg-secondary">
                {local ? <HardDrive className="h-5 w-5" /> : <Cloud className="h-5 w-5" />}
              </span>
              <div>
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Provedor</p>
                <p className="text-xl font-semibold">{local ? "Disco local" : "MinIO / S3"}</p>
              </div>
            </div>
            <dl className="mt-6 space-y-3 text-sm">
              {rows.map(([k, v]) => (
                <div key={k} className="flex justify-between gap-4 border-b border-border pb-2 last:border-0">
                  <dt className="text-muted-foreground">{k}</dt>
                  <dd className="truncate font-mono text-xs">{v}</dd>
                </div>
              ))}
            </dl>
          </section>
          <section className="grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-border bg-border">
            {(
              [
                ["Avatares enviados", String(info.avatars)],
                ["Banners enviados", String(info.banners)],
                ["Limite por imagem", `${Math.round(info.maxImageBytes / 1024 / 1024)} MB`],
                ["Formatos aceitos", "JPG · PNG · WebP · GIF"],
              ] as const
            ).map(([label, value]) => (
              <div key={label} className="bg-card p-5">
                <p className="text-xs text-muted-foreground">{label}</p>
                <p className="mt-1 text-lg font-semibold tabular-nums">{value}</p>
              </div>
            ))}
          </section>
        </div>
      )}
    </ModerationShell>
  );
}
