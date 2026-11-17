"use client";

import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { getHealth } from "@/lib/api";

type ApiState =
  | { kind: "loading" }
  | { kind: "ok" }
  | { kind: "error"; message: string };

export default function HomePage() {
  const [api, setApi] = useState<ApiState>({ kind: "loading" });

  useEffect(() => {
    let active = true;
    getHealth()
      .then(() => active && setApi({ kind: "ok" }))
      .catch((err) => active && setApi({ kind: "error", message: String(err?.message ?? err) }));
    return () => {
      active = false;
    };
  }, []);

  return (
    <div className="mx-auto max-w-2xl space-y-8">
      <section className="space-y-4">
        <span className="inline-flex rounded-full border border-border px-3 py-1 text-xs font-medium text-muted-foreground">
          Etapa 1 — Fundação
        </span>
        <h1 className="text-4xl font-bold tracking-tight sm:text-5xl">
          Forum Framework
        </h1>
        <p className="text-lg text-muted-foreground">
          Base do monorepo no ar: API em Go e frontend Next.js integrados. As
          funcionalidades do fórum chegam nas próximas etapas.
        </p>
        <div className="flex gap-3">
          <Button>Começar</Button>
          <Button variant="outline">Ver documentação</Button>
        </div>
      </section>

      <section className="rounded-xl border border-border bg-card p-5 text-card-foreground">
        <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
          Status da API
        </h2>
        <ApiStatusRow state={api} />
      </section>
    </div>
  );
}

function ApiStatusRow({ state }: { state: ApiState }) {
  if (state.kind === "loading") {
    return (
      <div className="flex items-center gap-2 text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Verificando conexão…
      </div>
    );
  }
  if (state.kind === "ok") {
    return (
      <div className="flex items-center gap-2 text-primary">
        <CheckCircle2 className="h-4 w-4" />
        API conectada e saudável.
      </div>
    );
  }
  return (
    <div className="flex items-center gap-2 text-destructive">
      <XCircle className="h-4 w-4" />
      API indisponível ({state.message}).
    </div>
  );
}
