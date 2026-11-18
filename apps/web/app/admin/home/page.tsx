"use client";

import { Check, Loader2, Megaphone, RotateCcw } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api, type SettingDefinition, type SiteSettings } from "@/lib/api";

export default function AdminHomePage() {
  const router = useRouter();
  const { loading: authLoading, hasPermission } = useAuth();
  const allowed = hasPermission("settings.manage");

  const [definitions, setDefinitions] = useState<SettingDefinition[]>([]);
  const [saved, setSaved] = useState<SiteSettings>({});
  const [values, setValues] = useState<SiteSettings>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    if (authLoading) return;
    if (!allowed) {
      router.replace("/");
      return;
    }
    api
      .adminSiteSettings()
      .then(({ data }) => {
        setDefinitions(data.definitions);
        setSaved(data.values);
        setValues(data.values);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "erro"))
      .finally(() => setLoading(false));
  }, [authLoading, allowed, router]);

  const groups = useMemo(() => {
    const out = new Map<string, SettingDefinition[]>();
    for (const d of definitions) out.set(d.group, [...(out.get(d.group) ?? []), d]);
    return [...out.entries()];
  }, [definitions]);

  const changed = useMemo(
    () => Object.fromEntries(Object.entries(values).filter(([k, v]) => saved[k] !== v)),
    [values, saved],
  );
  const dirty = Object.keys(changed).length > 0;

  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (!dirty) return;
    setSaving(true);
    setError(null);
    setNotice(null);
    try {
      const { data } = await api.updateSiteSettings(changed);
      setSaved(data);
      setValues(data);
      setNotice("Configurações salvas. A página inicial já reflete as mudanças.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao salvar");
    } finally {
      setSaving(false);
    }
  }

  if (authLoading || loading) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Página inicial</h1>
          <p className="mt-1 text-sm text-muted-foreground">Textos e seções da home, armazenados no banco.</p>
        </div>
        <Button variant="outline" size="sm" asChild>
          <Link href="/" target="_blank">
            Ver home
          </Link>
        </Button>
      </div>

      <form onSubmit={save} className="space-y-6">
        {groups.map(([group, defs]) => (
          <Card key={group}>
            <CardHeader>
              <CardTitle className="text-base">{group}</CardTitle>
            </CardHeader>
            <CardContent className={group === "Visibilidade" ? "grid gap-3 sm:grid-cols-2" : "space-y-4"}>
              {defs.map((d) => (
                <Field
                  key={d.key}
                  def={d}
                  value={values[d.key] ?? ""}
                  onChange={(v) => setValues((prev) => ({ ...prev, [d.key]: v }))}
                />
              ))}
            </CardContent>
          </Card>
        ))}

        <div className="sticky bottom-4 flex flex-wrap items-center gap-3 rounded-xl border border-border bg-card/95 p-3 shadow-sm backdrop-blur">
          <Button type="submit" disabled={!dirty || saving}>
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
            Salvar {dirty ? `(${Object.keys(changed).length})` : ""}
          </Button>
          <Button type="button" variant="ghost" disabled={!dirty || saving} onClick={() => setValues(saved)}>
            <RotateCcw className="h-4 w-4" />
            Descartar
          </Button>
          {error && <p className="text-sm text-destructive">{error}</p>}
          {notice && !error && <p className="text-sm text-primary">{notice}</p>}
        </div>
      </form>

      <BroadcastCard />
    </div>
  );
}

function Field({ def, value, onChange }: { def: SettingDefinition; value: string; onChange: (v: string) => void }) {
  const id = `setting-${def.key}`;
  if (def.kind === "bool") {
    return (
      <label htmlFor={id} className="flex items-center gap-2 text-sm">
        <input
          id={id}
          type="checkbox"
          checked={value === "true"}
          onChange={(e) => onChange(e.target.checked ? "true" : "false")}
          className="h-4 w-4 rounded border-input accent-[hsl(var(--primary))]"
        />
        {def.label}
      </label>
    );
  }
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-2">
        <Label htmlFor={id}>{def.label}</Label>
        {value !== def.default && (
          <button type="button" onClick={() => onChange(def.default)} className="text-xs text-muted-foreground hover:text-foreground">
            restaurar padrão
          </button>
        )}
      </div>
      {def.kind === "longtext" ? (
        <Textarea id={id} value={value} onChange={(e) => onChange(e.target.value)} maxLength={def.maxLength} rows={3} />
      ) : (
        <Input
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          maxLength={def.maxLength}
          placeholder={def.kind === "path" ? "/categories" : def.kind === "url" ? "https://… ou /files/…" : undefined}
        />
      )}
      <p className="text-right text-xs text-muted-foreground">
        {value.length}/{def.maxLength}
      </p>
    </div>
  );
}

function BroadcastCard() {
  const [message, setMessage] = useState("");
  const [link, setLink] = useState("");
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function send(e: React.FormEvent) {
    e.preventDefault();
    if (!window.confirm("Enviar este aviso para todos os membros ativos?")) return;
    setBusy(true);
    setError(null);
    setResult(null);
    try {
      const { data } = await api.broadcast({ message: message.trim(), link: link.trim() || undefined });
      setResult(`Aviso enviado para ${data.sent} membro(s).`);
      setMessage("");
      setLink("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao enviar");
    } finally {
      setBusy(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <Megaphone className="h-4 w-4" />
          Aviso para todos os membros
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={send} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="broadcast-message">Mensagem</Label>
            <Textarea
              id="broadcast-message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              minLength={3}
              maxLength={500}
              rows={2}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="broadcast-link">Link (opcional, caminho interno)</Label>
            <Input id="broadcast-link" value={link} onChange={(e) => setLink(e.target.value)} placeholder="/t/…" maxLength={500} />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          {result && <p className="text-sm text-primary">{result}</p>}
          <Button type="submit" disabled={busy || message.trim().length < 3}>
            {busy && <Loader2 className="h-4 w-4 animate-spin" />}
            Enviar aviso
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
