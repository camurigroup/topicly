"use client";

import { Crown, Loader2, Lock, Plus, Trash2, Users } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { ModerationShell } from "@/components/moderation-nav";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { modApi, type Permission, type RoleDetail } from "@/lib/moderation-api";
import { cn } from "@/lib/utils";

import { PermissionCatalog } from "./permission-catalog";

export default function RolesPage() {
  const { hasPermission } = useAuth();
  const [roles, setRoles] = useState<RoleDetail[]>([]);
  const [perms, setPerms] = useState<Permission[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const canRoles = hasPermission("roles.manage");

  const load = useCallback(async () => {
    setError(null);
    try {
      const [r, p] = await Promise.all([
        canRoles ? modApi.roles() : Promise.resolve({ data: [] as RoleDetail[] }),
        modApi.permissions(),
      ]);
      setRoles(r.data);
      setPerms(p.data);
      setSelectedId((id) => id ?? r.data.find((x) => x.slug !== "owner")?.id ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "erro ao carregar cargos");
    } finally {
      setLoading(false);
    }
  }, [canRoles]);

  useEffect(() => {
    void load();
  }, [load]);

  const selected = roles.find((r) => r.id === selectedId) ?? null;

  return (
    <ModerationShell
      title="Cargos e permissões"
      description="Mudanças valem na próxima requisição. Você só gerencia cargos abaixo do seu e só concede permissões que possui."
      perms={["roles.manage", "permissions.manage"]}
    >
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {loading ? (
        <div className="flex justify-center py-16 text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : (
        <div className="space-y-10">
          {canRoles && (
            <div className="grid gap-6 md:grid-cols-[260px_1fr]">
              <aside className="space-y-3">
                <ul className="space-y-1">
                  {roles.map((r) => (
                    <li key={r.id}>
                      <button
                        type="button"
                        onClick={() => setSelectedId(r.id)}
                        aria-current={selectedId === r.id}
                        className={cn(
                          "flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2.5 text-left text-sm transition-colors",
                          selectedId === r.id ? "bg-foreground text-background" : "hover:bg-secondary",
                        )}
                      >
                        <span className="flex items-center gap-2 font-medium">
                          {r.slug === "owner" && <Crown className="h-4 w-4 text-amber-500" />}
                          {r.name}
                        </span>
                        <span className="flex items-center gap-2 text-xs tabular-nums opacity-70">
                          <Users className="h-3 w-3" />
                          {r.memberCount}
                          <span>· {r.rankOrder}</span>
                        </span>
                      </button>
                    </li>
                  ))}
                </ul>
                <CreateRole onCreated={load} />
              </aside>
              {selected ? (
                <RoleEditor
                  key={`${selected.id}-${selected.updatedAt}-${selected.permissions.join()}`}
                  role={selected}
                  perms={perms}
                  onChanged={load}
                  onDeleted={() => {
                    setSelectedId(null);
                    void load();
                  }}
                />
              ) : (
                <p className="text-sm text-muted-foreground">Selecione um cargo.</p>
              )}
            </div>
          )}
          {hasPermission("permissions.manage") && <PermissionCatalog perms={perms} onChanged={load} />}
        </div>
      )}
    </ModerationShell>
  );
}

/** Groups permissions by their slug prefix (users.*, topics.* …). */
function groupPerms(perms: Permission[]): [string, Permission[]][] {
  const groups = new Map<string, Permission[]>();
  for (const p of perms) {
    const key = p.slug.split(".")[0];
    groups.set(key, [...(groups.get(key) ?? []), p]);
  }
  return [...groups.entries()];
}

interface RoleEditorProps {
  role: RoleDetail;
  perms: Permission[];
  onChanged: () => Promise<void>;
  onDeleted: () => void;
}

function RoleEditor({ role, perms, onChanged, onDeleted }: RoleEditorProps) {
  const { hasPermission } = useAuth();
  const locked = role.slug === "owner";
  const [name, setName] = useState(role.name);
  const [rank, setRank] = useState(role.rankOrder);
  const [selected, setSelected] = useState<string[]>(role.permissions);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);
  const groups = useMemo(() => groupPerms(perms), [perms]);

  const permsDirty = selected.slice().sort().join() !== role.permissions.slice().sort().join();
  const metaDirty = name !== role.name || rank !== role.rankOrder;

  async function run(fn: () => Promise<unknown>, ok: string) {
    setBusy(true);
    setMsg(null);
    try {
      await fn();
      setMsg({ ok: true, text: ok });
      await onChanged();
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha" });
    } finally {
      setBusy(false);
    }
  }

  async function remove() {
    if (!window.confirm(`Excluir o cargo "${role.name}"? Os membros perdem as permissões dele.`)) return;
    setBusy(true);
    try {
      await modApi.deleteRole(role.id);
      onDeleted();
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : "falha ao excluir" });
      setBusy(false);
    }
  }

  return (
    <section className="space-y-6 rounded-xl border border-border bg-card p-6">
      {locked && (
        <p className="flex items-center gap-2 rounded-lg bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-300">
          <Lock className="h-4 w-4" /> O cargo Dono é imutável e sempre possui todas as permissões.
        </p>
      )}
      <div className="grid gap-4 sm:grid-cols-[1fr_140px_auto] sm:items-end">
        <div className="space-y-1.5">
          <Label htmlFor="role-name">Nome</Label>
          <Input id="role-name" value={name} onChange={(e) => setName(e.target.value)} disabled={locked} maxLength={80} />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="role-rank">Hierarquia</Label>
          <Input
            id="role-rank"
            type="number"
            min={1}
            max={99}
            value={rank}
            onChange={(e) => setRank(Number(e.target.value))}
            disabled={locked}
          />
        </div>
        <Button
          variant="outline"
          disabled={locked || !metaDirty || busy}
          onClick={() => run(() => modApi.updateRole(role.id, { name: name.trim(), rankOrder: rank }), "Cargo atualizado.")}
        >
          Salvar
        </Button>
      </div>

      <div className="space-y-4">
        <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Permissões</h3>
        <div className="grid gap-4 sm:grid-cols-2">
          {groups.map(([group, items]) => (
            <fieldset key={group} className="rounded-lg border border-border p-3">
              <legend className="px-1 font-mono text-xs text-muted-foreground">{group}.*</legend>
              {items.map((p) => {
                const checked = selected.includes(p.slug);
                const cannotGrant = !checked && !hasPermission(p.slug);
                return (
                  <label
                    key={p.id}
                    className={cn(
                      "flex items-center gap-2 rounded px-1 py-1 text-sm",
                      locked || cannotGrant ? "opacity-50" : "cursor-pointer hover:bg-secondary",
                    )}
                    title={cannotGrant ? "Você não possui esta permissão" : p.slug}
                  >
                    <input
                      type="checkbox"
                      className="accent-primary"
                      checked={locked || checked}
                      disabled={locked || cannotGrant}
                      onChange={() => setSelected((s) => (checked ? s.filter((x) => x !== p.slug) : [...s, p.slug]))}
                    />
                    <span className={cn(!p.isActive && "line-through")}>{p.name}</span>
                  </label>
                );
              })}
            </fieldset>
          ))}
        </div>
      </div>

      {msg && (
        <p
          role={msg.ok ? "status" : "alert"}
          className={cn("text-sm", msg.ok ? "text-emerald-600 dark:text-emerald-400" : "text-destructive")}
        >
          {msg.text}
        </p>
      )}

      {!locked && (
        <div className="flex flex-wrap justify-between gap-2 border-t border-border pt-4">
          <Button variant="ghost" className="text-destructive" disabled={role.isSystem || busy} onClick={remove}>
            <Trash2 className="h-4 w-4" /> Excluir cargo
          </Button>
          <Button
            disabled={!permsDirty || busy}
            onClick={() => run(() => modApi.setRolePermissions(role.id, selected), "Permissões salvas — efeito imediato.")}
          >
            {busy && <Loader2 className="h-4 w-4 animate-spin" />}
            Salvar permissões
          </Button>
        </div>
      )}
    </section>
  );
}

function CreateRole({ onCreated }: { onCreated: () => Promise<void> }) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [rank, setRank] = useState(10);
  const [error, setError] = useState<string | null>(null);

  if (!open) {
    return (
      <Button variant="outline" size="sm" className="w-full" onClick={() => setOpen(true)}>
        <Plus className="h-4 w-4" /> Novo cargo
      </Button>
    );
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await modApi.createRole({ name: name.trim(), slug: slug.trim(), rankOrder: rank, permissions: [] });
      setOpen(false);
      setName("");
      setSlug("");
      await onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao criar cargo");
    }
  }

  return (
    <form onSubmit={submit} className="space-y-2 rounded-lg border border-border p-3" aria-label="Novo cargo">
      <Input placeholder="Nome (ex.: Ajudante)" value={name} onChange={(e) => setName(e.target.value)} required maxLength={80} />
      <Input
        placeholder="slug (ex.: ajudante)"
        value={slug}
        onChange={(e) => setSlug(e.target.value.toLowerCase())}
        required
        pattern="[a-z0-9][a-z0-9\-]{1,48}[a-z0-9]"
      />
      <Input type="number" min={1} max={99} value={rank} onChange={(e) => setRank(Number(e.target.value))} aria-label="Hierarquia" />
      {error && <p role="alert" className="text-xs text-destructive">{error}</p>}
      <p className="text-xs text-muted-foreground">As permissões são definidas depois de criar.</p>
      <div className="flex gap-2">
        <Button size="sm" type="submit">
          Criar
        </Button>
        <Button size="sm" type="button" variant="ghost" onClick={() => setOpen(false)}>
          Cancelar
        </Button>
      </div>
    </form>
  );
}
