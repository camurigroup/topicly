"use client";

import { Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { modApi, type Permission } from "@/lib/moderation-api";
import { cn } from "@/lib/utils";

interface PermissionCatalogProps {
  perms: Permission[];
  onChanged: () => Promise<void>;
}

/**
 * Global permission catalog: activate/deactivate (affects every role at
 * once), create custom permissions and delete custom ones. Core permissions
 * cannot be deactivated; built-in ones cannot be deleted.
 */
export function PermissionCatalog({ perms, onChanged }: PermissionCatalogProps) {
  const [error, setError] = useState<string | null>(null);
  const [slug, setSlug] = useState("");
  const [name, setName] = useState("");

  async function run(fn: () => Promise<unknown>) {
    setError(null);
    try {
      await fn();
      await onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha");
    }
  }

  async function create(e: React.FormEvent) {
    e.preventDefault();
    await run(() => modApi.createPermission({ slug: slug.trim(), name: name.trim() }));
    setSlug("");
    setName("");
  }

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold">Catálogo de permissões</h2>
        <p className="text-sm text-muted-foreground">
          Desativar uma permissão a remove de todos os cargos imediatamente, sem apagar a configuração.
        </p>
      </div>
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      <div className="overflow-hidden rounded-xl border border-border">
        <table className="w-full text-sm">
          <thead className="bg-secondary/60 text-left text-xs uppercase tracking-wide text-muted-foreground">
            <tr>
              <th className="px-4 py-2 font-medium">Permissão</th>
              <th className="hidden px-4 py-2 font-medium sm:table-cell">Slug</th>
              <th className="px-4 py-2 text-right font-medium">Ativa</th>
            </tr>
          </thead>
          <tbody>
            {perms.map((p) => (
              <tr key={p.id} className="border-t border-border">
                <td className="px-4 py-2.5">
                  <span className={cn(!p.isActive && "text-muted-foreground line-through")}>{p.name}</span>
                  {p.isCore && <span className="ml-2 text-xs text-amber-600 dark:text-amber-400">essencial</span>}
                  {!p.isBuiltin && <span className="ml-2 text-xs text-sky-600 dark:text-sky-400">personalizada</span>}
                </td>
                <td className="hidden px-4 py-2.5 font-mono text-xs text-muted-foreground sm:table-cell">{p.slug}</td>
                <td className="px-4 py-2.5">
                  <div className="flex items-center justify-end gap-2">
                    {!p.isBuiltin && (
                      <button
                        type="button"
                        aria-label={`Excluir ${p.name}`}
                        onClick={() => window.confirm(`Excluir a permissão ${p.slug}?`) && run(() => modApi.deletePermission(p.id))}
                        className="rounded p-1 text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    )}
                    <button
                      type="button"
                      role="switch"
                      aria-checked={p.isActive}
                      aria-label={`${p.isActive ? "Desativar" : "Ativar"} ${p.name}`}
                      disabled={p.isCore && p.isActive}
                      onClick={() => run(() => modApi.setPermissionActive(p.id, !p.isActive))}
                      className={cn(
                        "relative h-5 w-9 rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50",
                        p.isActive ? "bg-primary" : "bg-muted",
                      )}
                    >
                      <span
                        className={cn(
                          "absolute top-0.5 h-4 w-4 rounded-full bg-background shadow transition-transform",
                          p.isActive ? "translate-x-4" : "translate-x-0.5",
                        )}
                      />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <form onSubmit={create} className="flex flex-wrap gap-2" aria-label="Nova permissão">
        <Input
          placeholder="slug (ex.: events.manage)"
          value={slug}
          onChange={(e) => setSlug(e.target.value.toLowerCase())}
          className="flex-1 basis-48 font-mono"
          required
        />
        <Input placeholder="Nome" value={name} onChange={(e) => setName(e.target.value)} className="flex-1 basis-48" required />
        <Button type="submit" variant="outline">
          <Plus className="h-4 w-4" /> Criar permissão
        </Button>
      </form>
    </section>
  );
}
