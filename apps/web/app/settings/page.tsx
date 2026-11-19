"use client";

import { AtSign, ExternalLink, Eye, EyeOff, KeyRound, Loader2, UserRound } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { ProfileMedia } from "@/components/profile-media";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api, type Account } from "@/lib/api";
import { cn } from "@/lib/utils";

type Tab = "perfil" | "conta" | "seguranca";

const TABS: { id: Tab; label: string; hint: string; icon: typeof UserRound }[] = [
  { id: "perfil", label: "Perfil público", hint: "Nome, bio, foto e capa", icon: UserRound },
  { id: "conta", label: "Conta", hint: "E-mail de acesso", icon: AtSign },
  { id: "seguranca", label: "Segurança", hint: "Senha", icon: KeyRound },
];

type Msg = { ok: boolean; text: string } | null;

export default function SettingsPage() {
  const { user, loading, setAccount } = useAuth();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("perfil");

  useEffect(() => {
    if (!loading && !user) router.replace("/login");
  }, [loading, user, router]);

  // Deep links: /settings#conta, /settings#seguranca
  useEffect(() => {
    const h = window.location.hash.slice(1) as Tab;
    if (TABS.some((t) => t.id === h)) setTab(h);
  }, []);

  function select(t: Tab) {
    setTab(t);
    window.history.replaceState(null, "", "#" + t);
  }

  if (loading || !user) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl space-y-8">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Sua conta</p>
          <h1 className="mt-1 text-3xl font-bold tracking-tight">Configurações</h1>
        </div>
        <Button variant="outline" size="sm" asChild>
          <Link href={"/u/" + user.username}>
            Ver perfil público <ExternalLink className="h-3.5 w-3.5" />
          </Link>
        </Button>
      </header>

      <div className="grid gap-8 md:grid-cols-[220px_minmax(0,1fr)]">
        <nav aria-label="Seções" className="min-w-0 md:sticky md:top-24 md:self-start">
          <ul className="flex gap-1 overflow-x-auto [scrollbar-width:none] md:flex-col">
            {TABS.map((t) => {
              const active = tab === t.id;
              return (
                <li key={t.id} className="shrink-0">
                  <button
                    type="button"
                    onClick={() => select(t.id)}
                    aria-current={active ? "page" : undefined}
                    className={cn(
                      "flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors",
                      active ? "bg-secondary text-foreground" : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground",
                    )}
                  >
                    <t.icon className={cn("h-4 w-4 shrink-0", active && "text-primary")} />
                    <span>
                      <span className="block text-sm font-medium">{t.label}</span>
                      <span className="hidden text-xs text-muted-foreground md:block">{t.hint}</span>
                    </span>
                  </button>
                </li>
              );
            })}
          </ul>
        </nav>

        <div className="min-w-0">
          {tab === "perfil" && <ProfileSection account={user} onSaved={setAccount} />}
          {tab === "conta" && <EmailSection account={user} onSaved={setAccount} />}
          {tab === "seguranca" && <PasswordSection />}
        </div>
      </div>
    </div>
  );
}

function Panel({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return (
    <section className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="border-b border-border px-6 py-5">
        <h2 className="text-lg font-semibold">{title}</h2>
        <p className="mt-0.5 text-sm text-muted-foreground">{description}</p>
      </div>
      <div className="p-6">{children}</div>
    </section>
  );
}

function SaveBar({ busy, msg, label, disabled }: { busy: boolean; msg: Msg; label: string; disabled?: boolean }) {
  return (
    <div className="flex flex-wrap items-center justify-end gap-3 border-t border-border pt-5">
      {msg && (
        <p role={msg.ok ? "status" : "alert"} className={cn("mr-auto text-sm", msg.ok ? "text-primary" : "text-destructive")}>
          {msg.text}
        </p>
      )}
      <Button type="submit" disabled={busy || disabled}>
        {busy && <Loader2 className="h-4 w-4 animate-spin" />}
        {label}
      </Button>
    </div>
  );
}

function ProfileSection({ account, onSaved }: { account: Account; onSaved: (a: Account) => void }) {
  const [displayName, setDisplayName] = useState(account.displayName);
  const [bio, setBio] = useState(account.bio ?? "");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<Msg>(null);
  const dirty = displayName !== account.displayName || bio !== (account.bio ?? "");

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setMsg(null);
    try {
      // PATCH semantics: only these two fields change; "" clears the bio.
      const { data } = await api.updateProfile({ displayName, bio });
      onSaved(data);
      setMsg({ ok: true, text: "Perfil atualizado." });
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha ao salvar" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-6">
      <section className="rounded-2xl border border-border bg-card p-4 pb-6 shadow-sm">
        <ProfileMedia account={account} onUpdated={onSaved} />
        <div className="mt-3 pl-28">
          <p className="font-semibold leading-tight">{account.displayName}</p>
          <p className="text-sm text-muted-foreground">@{account.username}</p>
        </div>
      </section>

      <Panel title="Perfil público" description="O que os outros membros veem ao abrir seu perfil.">
        <form onSubmit={save} className="space-y-5">
          <div className="space-y-1.5">
            <Label htmlFor="displayName">Nome de exibição</Label>
            <Input id="displayName" value={displayName} onChange={(e) => setDisplayName(e.target.value)} maxLength={80} required />
          </div>
          <div className="space-y-1.5">
            <div className="flex items-baseline justify-between">
              <Label htmlFor="bio">Bio</Label>
              <span className={cn("text-xs tabular-nums", bio.length > 450 ? "text-amber-600" : "text-muted-foreground")}>
                {bio.length}/500
              </span>
            </div>
            <Textarea
              id="bio"
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              rows={4}
              maxLength={500}
              placeholder="Conte um pouco sobre você…"
            />
          </div>
          <SaveBar busy={busy} msg={msg} label="Salvar perfil" disabled={!dirty} />
        </form>
      </Panel>
    </div>
  );
}

function EmailSection({ account, onSaved }: { account: Account; onSaved: (a: Account) => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<Msg>(null);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setMsg(null);
    try {
      const { data } = await api.changeEmail({ email, password });
      onSaved(data);
      setEmail("");
      setPassword("");
      setMsg({ ok: true, text: "E-mail alterado." });
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha ao alterar e-mail" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <Panel title="Conta" description="Seu e-mail de acesso. Ele nunca aparece no perfil público.">
      <div className="mb-6 flex items-center gap-3 rounded-xl bg-secondary/60 px-4 py-3">
        <AtSign className="h-4 w-4 text-muted-foreground" />
        <div className="min-w-0">
          <p className="text-xs text-muted-foreground">E-mail atual</p>
          <p className="truncate font-medium">{account.email}</p>
        </div>
      </div>
      <form onSubmit={save} className="space-y-5">
        <div className="grid gap-5 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="newemail">Novo e-mail</Label>
            <Input id="newemail" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required autoComplete="email" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="pw">Confirme com sua senha</Label>
            <PasswordInput id="pw" value={password} onChange={setPassword} autoComplete="current-password" />
          </div>
        </div>
        <SaveBar busy={busy} msg={msg} label="Alterar e-mail" />
      </form>
    </Panel>
  );
}

/** Rough strength hint: length + character variety (the API enforces the rules). */
function strength(pw: string): { score: number; label: string } {
  let score = 0;
  if (pw.length >= 8) score++;
  if (pw.length >= 12) score++;
  if (/[a-z]/.test(pw) && /[A-Z]/.test(pw)) score++;
  if (/\d/.test(pw) && /[^A-Za-z0-9]/.test(pw)) score++;
  return { score, label: ["Muito fraca", "Fraca", "Razoável", "Boa", "Forte"][score] };
}

function PasswordSection() {
  const [currentPassword, setCurrent] = useState("");
  const [newPassword, setNew] = useState("");
  const [confirm, setConfirm] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<Msg>(null);
  const s = strength(newPassword);
  const mismatch = confirm.length > 0 && confirm !== newPassword;

  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (mismatch) return;
    setBusy(true);
    setMsg(null);
    try {
      await api.changePassword({ currentPassword, newPassword });
      setCurrent("");
      setNew("");
      setConfirm("");
      setMsg({ ok: true, text: "Senha alterada." });
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha ao alterar senha" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <Panel title="Segurança" description="Use uma senha longa e que você não use em outros sites.">
      <form onSubmit={save} className="space-y-5">
        <div className="space-y-1.5 sm:max-w-sm">
          <Label htmlFor="cur">Senha atual</Label>
          <PasswordInput id="cur" value={currentPassword} onChange={setCurrent} autoComplete="current-password" />
        </div>
        <div className="grid gap-5 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="new">Nova senha</Label>
            <PasswordInput id="new" value={newPassword} onChange={setNew} autoComplete="new-password" />
            {newPassword && (
              <div className="flex items-center gap-2 pt-1" aria-live="polite">
                <div className="flex flex-1 gap-1">
                  {[0, 1, 2, 3].map((i) => (
                    <span
                      key={i}
                      className={cn(
                        "h-1 flex-1 rounded-full",
                        i < s.score
                          ? s.score >= 3
                            ? "bg-emerald-500"
                            : s.score === 2
                              ? "bg-amber-500"
                              : "bg-destructive"
                          : "bg-secondary",
                      )}
                    />
                  ))}
                </div>
                <span className="text-xs text-muted-foreground">{s.label}</span>
              </div>
            )}
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="confirm">Confirmar nova senha</Label>
            <PasswordInput id="confirm" value={confirm} onChange={setConfirm} autoComplete="new-password" />
            {mismatch && <p className="text-xs text-destructive">As senhas não conferem.</p>}
          </div>
        </div>
        <SaveBar busy={busy} msg={msg} label="Alterar senha" disabled={mismatch} />
      </form>
    </Panel>
  );
}

function PasswordInput({
  id,
  value,
  onChange,
  autoComplete,
}: {
  id: string;
  value: string;
  onChange: (v: string) => void;
  autoComplete: string;
}) {
  const [show, setShow] = useState(false);
  return (
    <div className="relative">
      <Input
        id={id}
        type={show ? "text" : "password"}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required
        autoComplete={autoComplete}
        className="pr-10"
      />
      <button
        type="button"
        onClick={() => setShow((v) => !v)}
        aria-label={show ? "Ocultar senha" : "Mostrar senha"}
        className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:text-foreground"
      >
        {show ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
      </button>
    </div>
  );
}
