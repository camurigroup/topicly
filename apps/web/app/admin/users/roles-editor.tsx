"use client";

import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { modApi, type RoleDetail } from "@/lib/moderation-api";

interface RolesEditorProps {
  username: string;
  current: string[];
  onSaved: () => Promise<void>;
}

/**
 * Assigns roles to a user. The API rejects any role at or above the
 * moderator's own rank, and the Owner role is never assignable.
 */
export function RolesEditor({ username, current, onSaved }: RolesEditorProps) {
  const [roles, setRoles] = useState<RoleDetail[]>([]);
  const [selected, setSelected] = useState<string[]>(current);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    modApi
      .roles()
      .then(({ data }) => setRoles(data))
      .catch((err) => setError(err instanceof Error ? err.message : "falha ao carregar cargos"));
  }, []);

  useEffect(() => setSelected(current), [current]);

  const dirty = selected.slice().sort().join() !== current.slice().sort().join();

  function toggle(slug: string) {
    setSelected((s) => (s.includes(slug) ? s.filter((x) => x !== slug) : [...s, slug]));
  }

  async function save() {
    setBusy(true);
    setError(null);
    try {
      await modApi.setUserRoles(username, selected);
      await onSaved();
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao salvar cargos");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="space-y-3 rounded-xl border border-border bg-card p-5">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Cargos</h2>
      <ul className="space-y-1">
        {roles
          .filter((r) => r.slug !== "owner")
          .map((r) => (
            <li key={r.id}>
              <label className="flex cursor-pointer items-center justify-between gap-3 rounded-md px-2 py-1.5 text-sm hover:bg-secondary">
                <span className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={selected.includes(r.slug)}
                    onChange={() => toggle(r.slug)}
                    className="accent-primary"
                  />
                  {r.name}
                </span>
                <span className="text-xs tabular-nums text-muted-foreground">nível {r.rankOrder}</span>
              </label>
            </li>
          ))}
      </ul>
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      <Button size="sm" className="w-full" disabled={!dirty || busy} onClick={save}>
        {busy && <Loader2 className="h-4 w-4 animate-spin" />}
        Salvar cargos
      </Button>
    </section>
  );
}
