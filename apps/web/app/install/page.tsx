"use client";

import { Check, Database, Loader2, ShieldCheck, Sparkles } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { tokens } from "@/lib/api";
import { installApi, type InstallStatus } from "@/lib/install-api";
import { cn } from "@/lib/utils";

const STEPS = [
  { icon: Database, title: "Banco e armazenamento", text: "Definidos no .env pelo instalador de linha de comando." },
  { icon: Sparkles, title: "Identidade", text: "Nome e descrição da comunidade — tudo editável depois no painel." },
  { icon: ShieldCheck, title: "Conta do Dono", text: "Acesso total e imutável. Guarde bem a senha." },
];

export default function InstallPage() {
  const router = useRouter();
  const { refresh } = useAuth();
  const [status, setStatus] = useState<InstallStatus | null>(null);
  const [form, setForm] = useState({
    forumName: "",
    description: "",
    displayName: "",
    username: "",
    email: "",
    password: "",
    confirm: "",
    token: "",
  });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    installApi
      .status()
      .then(({ data }) => setStatus(data))
      .catch((err) => setError(err instanceof Error ? err.message : "API indisponível"));
  }, []);

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (form.password !== form.confirm) {
      setError("As senhas não conferem.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const { data } = await installApi.complete({
        forumName: form.forumName,
        description: form.description,
        displayName: form.displayName || form.username,
        username: form.username,
        email: form.email,
        password: form.password,
        token: form.token || undefined,
      });
      tokens.set(data.tokens.accessToken, data.tokens.refreshToken);
      await refresh();
      router.replace("/admin/settings");
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha na instalação");
      setBusy(false);
    }
  }

  if (status?.installed) {
    return <p className="py-20 text-center text-muted-foreground">Este fórum já está instalado.</p>;
  }

  return (
    <div className="mx-auto grid max-w-5xl gap-10 lg:grid-cols-[1fr_1.3fr]">
      <aside className="space-y-8 lg:pt-6">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Instalação</p>
          <h1 className="mt-3 text-4xl font-bold tracking-tight">Sua comunidade começa aqui.</h1>
          <p className="mt-3 text-muted-foreground">
            O fórum já está construído — falta só dar um nome e criar a conta de quem manda.
          </p>
        </div>
        <ol className="space-y-5">
          {STEPS.map((s, i) => (
            <li key={s.title} className="flex gap-4">
              <span
                className={cn(
                  "flex h-9 w-9 shrink-0 items-center justify-center rounded-full border",
                  i === 0 ? "border-primary bg-primary text-primary-foreground" : "border-border",
                )}
              >
                {i === 0 ? <Check className="h-4 w-4" /> : <s.icon className="h-4 w-4" />}
              </span>
              <div>
                <p className="font-medium">{s.title}</p>
                <p className="text-sm text-muted-foreground">{s.text}</p>
              </div>
            </li>
          ))}
        </ol>
      </aside>

      <form onSubmit={submit} className="space-y-6 rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8">
        <fieldset className="space-y-4">
          <legend className="mb-1 text-sm font-semibold uppercase tracking-wide text-muted-foreground">Identidade</legend>
          <div className="space-y-1.5">
            <Label htmlFor="forumName">Nome do fórum</Label>
            <Input id="forumName" value={form.forumName} onChange={set("forumName")} required minLength={2} maxLength={60} placeholder="Ex.: Comunidade Camuri" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="description">Descrição (opcional)</Label>
            <Textarea id="description" value={form.description} onChange={set("description")} rows={2} maxLength={300} />
          </div>
        </fieldset>

        <fieldset className="space-y-4 border-t border-border pt-6">
          <legend className="mb-1 text-sm font-semibold uppercase tracking-wide text-muted-foreground">Conta do Dono</legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor="displayName">Nome de exibição</Label>
              <Input id="displayName" value={form.displayName} onChange={set("displayName")} maxLength={80} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="username">Usuário</Label>
              <Input id="username" value={form.username} onChange={set("username")} required autoComplete="username" />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="email">E-mail</Label>
            <Input id="email" type="email" value={form.email} onChange={set("email")} required autoComplete="email" />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor="password">Senha</Label>
              <Input id="password" type="password" value={form.password} onChange={set("password")} required minLength={8} autoComplete="new-password" />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="confirm">Confirmar senha</Label>
              <Input id="confirm" type="password" value={form.confirm} onChange={set("confirm")} required autoComplete="new-password" />
            </div>
          </div>
          {status?.tokenRequired && (
            <div className="space-y-1.5">
              <Label htmlFor="token">Token de instalação</Label>
              <Input id="token" value={form.token} onChange={set("token")} required autoComplete="off" />
              <p className="text-xs text-muted-foreground">Definido na variável INSTALL_TOKEN do servidor.</p>
            </div>
          )}
        </fieldset>

        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
        <Button type="submit" size="lg" className="w-full" disabled={busy || !status}>
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Instalar e entrar no painel
        </Button>
      </form>
    </div>
  );
}
