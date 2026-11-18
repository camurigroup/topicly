"use client";

import { Loader2, Upload } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";

import { ModerationShell } from "@/components/moderation-nav";
import { useSiteSettings } from "@/components/site-settings-provider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api, type SettingDefinition, type SiteSettings } from "@/lib/api";
import { adminApi } from "@/lib/admin-api";
import { cn } from "@/lib/utils";

/** Groups edited here; the home page texts live in /admin/home. */
const GROUPS = ["Identidade", "Cores", "Redes sociais", "Rodapé", "Documentos legais"];
const UPLOADABLE = new Set(["brand.logo", "brand.favicon"]);

export default function SettingsPage() {
  const { refresh } = useSiteSettings();
  const [defs, setDefs] = useState<SettingDefinition[]>([]);
  const [saved, setSaved] = useState<SiteSettings>({});
  const [values, setValues] = useState<SiteSettings>({});
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  useEffect(() => {
    api
      .adminSiteSettings()
      .then(({ data }) => {
        setDefs(data.definitions);
        setSaved(data.values);
        setValues(data.values);
      })
      .catch((err) => setMsg({ ok: false, text: err instanceof Error ? err.message : "falha ao carregar" }));
  }, []);

  const grouped = useMemo(
    () => GROUPS.map((g) => [g, defs.filter((d) => d.group === g)] as const).filter(([, items]) => items.length > 0),
    [defs],
  );
  const changed = Object.keys(values).filter((k) => values[k] !== saved[k]);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (changed.length === 0) return;
    setBusy(true);
    setMsg(null);
    try {
      const patch = Object.fromEntries(changed.map((k) => [k, values[k]]));
      const { data } = await api.updateSiteSettings(patch);
      setSaved(data);
      setValues(data);
      await refresh();
      setMsg({ ok: true, text: "Identidade salva — o fórum público já reflete as mudanças." });
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha ao salvar" });
    } finally {
      setBusy(false);
    }
  }

  const set = (key: string, v: string) => setValues((s) => ({ ...s, [key]: v }));

  return (
    <ModerationShell
      title="Identidade visual"
      description="Nome, logo, cores, redes sociais, rodapé e documentos legais. Tudo aparece no fórum público sem tocar no código."
      perms={["settings.manage"]}
    >
      <form onSubmit={save} className="grid gap-8 xl:grid-cols-[minmax(0,1fr)_300px]">
        <div className="min-w-0 space-y-8">
          {grouped.map(([group, items]) => (
            <fieldset key={group} className="space-y-4 rounded-xl border border-border bg-card p-5">
              <legend className="px-1 text-sm font-semibold uppercase tracking-wide text-muted-foreground">{group}</legend>
              <div className={cn("grid gap-4", group !== "Documentos legais" && "sm:grid-cols-2")}>
                {items.map((d) => (
                  <Field key={d.key} def={d} value={values[d.key] ?? ""} onChange={(v) => set(d.key, v)} />
                ))}
              </div>
            </fieldset>
          ))}
        </div>

        <aside className="space-y-4 xl:sticky xl:top-24 xl:self-start">
          <Preview values={values} />
          {msg && (
            <p
              role={msg.ok ? "status" : "alert"}
              className={cn("text-sm", msg.ok ? "text-emerald-600 dark:text-emerald-400" : "text-destructive")}
            >
              {msg.text}
            </p>
          )}
          <Button type="submit" className="w-full" disabled={busy || changed.length === 0}>
            {busy && <Loader2 className="h-4 w-4 animate-spin" />}
            {changed.length > 0 ? `Salvar ${changed.length} alteração(ões)` : "Nada a salvar"}
          </Button>
        </aside>
      </form>
    </ModerationShell>
  );
}

function Field({ def, value, onChange }: { def: SettingDefinition; value: string; onChange: (v: string) => void }) {
  const id = `s-${def.key}`;
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wide = def.kind === "longtext" || def.kind === "links";

  async function upload(file: File) {
    setUploading(true);
    setError(null);
    try {
      const { data } = await adminApi.uploadAsset(file);
      onChange(data.url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha no upload");
    } finally {
      setUploading(false);
    }
  }

  return (
    <div className={cn("space-y-1.5", wide && "sm:col-span-2")}>
      <Label htmlFor={id}>{def.label}</Label>
      {def.kind === "color" ? (
        <div className="flex gap-2">
          <input
            type="color"
            aria-label={`${def.label} (seletor)`}
            value={value || "#6366f1"}
            onChange={(e) => onChange(e.target.value)}
            className="h-10 w-12 cursor-pointer rounded-md border border-input bg-background p-1"
          />
          <Input id={id} value={value} onChange={(e) => onChange(e.target.value)} placeholder="padrão do tema" maxLength={7} />
        </div>
      ) : wide ? (
        <Textarea
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={def.kind === "links" ? 4 : def.maxLength > 2000 ? 10 : 3}
          maxLength={def.maxLength}
          className={def.kind === "links" ? "font-mono text-xs" : undefined}
        />
      ) : (
        <div className="flex gap-2">
          <Input id={id} value={value} onChange={(e) => onChange(e.target.value)} maxLength={def.maxLength} />
          {UPLOADABLE.has(def.key) && (
            <>
              <input
                ref={fileRef}
                type="file"
                accept="image/png,image/jpeg,image/webp,image/gif"
                className="hidden"
                onChange={(e) => e.target.files?.[0] && upload(e.target.files[0])}
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                aria-label={`Enviar ${def.label}`}
                onClick={() => fileRef.current?.click()}
              >
                {uploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Upload className="h-4 w-4" />}
              </Button>
            </>
          )}
        </div>
      )}
      {error && <p role="alert" className="text-xs text-destructive">{error}</p>}
    </div>
  );
}

function Preview({ values }: { values: SiteSettings }) {
  const primary = values["theme.primary"] || undefined;
  const radius = Number(values["theme.radius"] || 8);
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card">
      <p className="border-b border-border px-4 py-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        Prévia
      </p>
      <div className="space-y-4 p-4">
        <div className="flex items-center gap-2">
          {values["brand.logo"] ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={values["brand.logo"]} alt="" className="h-7 w-7 rounded object-contain" />
          ) : (
            <span className="h-7 w-7 rounded" style={{ background: primary ?? "hsl(var(--primary))" }} aria-hidden />
          )}
          <span className="font-semibold">{values["brand.name"] || "Forum"}</span>
        </div>
        <p className="text-xs text-muted-foreground">{values["brand.description"]}</p>
        <div className="flex gap-2">
          <span
            className="px-3 py-1.5 text-xs font-medium text-white"
            style={{ background: primary ?? "hsl(var(--primary))", borderRadius: radius }}
          >
            Botão primário
          </span>
          <span className="border px-3 py-1.5 text-xs font-medium" style={{ borderColor: values["theme.accent"] || undefined, borderRadius: radius }}>
            Destaque
          </span>
        </div>
        <p className="border-t border-border pt-3 text-xs text-muted-foreground">{values["footer.text"]}</p>
      </div>
    </div>
  );
}
