"use client";

import { Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { ProfileMedia } from "@/components/profile-media";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api, type Account } from "@/lib/api";

export default function SettingsPage() {
  const { user, loading, setAccount } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) router.replace("/login");
  }, [loading, user, router]);

  if (loading || !user) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Configurações</h1>
      <ProfileSection
        account={user}
        initialDisplayName={user.displayName}
        initialBio={user.bio ?? ""}
        onSaved={setAccount}
      />
      <PasswordSection />
      <EmailSection onSaved={setAccount} />
    </div>
  );
}

function Feedback({ msg }: { msg: { ok: boolean; text: string } | null }) {
  if (!msg) return null;
  return <p className={msg.ok ? "text-sm text-primary" : "text-sm text-destructive"}>{msg.text}</p>;
}

function ProfileSection({
  account,
  initialDisplayName,
  initialBio,
  onSaved,
}: {
  account: Account;
  initialDisplayName: string;
  initialBio: string;
  onSaved: (a: Account) => void;
}) {
  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [bio, setBio] = useState(initialBio);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

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
    <Card>
      <CardHeader>
        <CardTitle>Perfil</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="mb-12">
          <ProfileMedia account={account} onUpdated={onSaved} />
        </div>
        <form onSubmit={save} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="displayName">Nome de exibição</Label>
            <Input id="displayName" value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="bio">Bio</Label>
            <textarea
              id="bio"
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              rows={3}
              maxLength={500}
              className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
          </div>
          <Feedback msg={msg} />
          <Button type="submit" disabled={busy}>
            {busy && <Loader2 className="h-4 w-4 animate-spin" />}
            Salvar perfil
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function PasswordSection() {
  const [currentPassword, setCurrent] = useState("");
  const [newPassword, setNew] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setMsg(null);
    try {
      await api.changePassword({ currentPassword, newPassword });
      setCurrent("");
      setNew("");
      setMsg({ ok: true, text: "Senha alterada." });
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha ao alterar senha" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Senha</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={save} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="cur">Senha atual</Label>
            <Input id="cur" type="password" value={currentPassword} onChange={(e) => setCurrent(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="new">Nova senha</Label>
            <Input id="new" type="password" value={newPassword} onChange={(e) => setNew(e.target.value)} required />
          </div>
          <Feedback msg={msg} />
          <Button type="submit" disabled={busy}>
            {busy && <Loader2 className="h-4 w-4 animate-spin" />}
            Alterar senha
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function EmailSection({ onSaved }: { onSaved: (a: Account) => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

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
    <Card>
      <CardHeader>
        <CardTitle>E-mail</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={save} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="newemail">Novo e-mail</Label>
            <Input id="newemail" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </div>
          <div className="space-y-2">
            <Label htmlFor="pw">Confirme a senha</Label>
            <Input id="pw" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          </div>
          <Feedback msg={msg} />
          <Button type="submit" disabled={busy}>
            {busy && <Loader2 className="h-4 w-4 animate-spin" />}
            Alterar e-mail
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
