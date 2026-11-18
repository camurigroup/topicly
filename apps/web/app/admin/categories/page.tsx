"use client";

import { Loader2, Plus, Trash2 } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api, type Category } from "@/lib/api";

export default function AdminCategoriesPage() {
  const router = useRouter();
  const { loading: authLoading, hasPermission } = useAuth();

  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // create form
  const [name, setName] = useState("");
  const [parentId, setParentId] = useState("");
  const [icon, setIcon] = useState("");
  const [description, setDescription] = useState("");
  const [creating, setCreating] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const allowed = hasPermission("categories.manage");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await api.allCategories();
      setCategories(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "erro");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (authLoading) return;
    if (!allowed) {
      router.replace("/categories");
      return;
    }
    void load();
  }, [authLoading, allowed, load, router]);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    setFormError(null);
    setCreating(true);
    try {
      await api.createCategory({
        name: name.trim(),
        parentId: parentId || null,
        icon: icon.trim() || null,
        description: description.trim() || null,
      });
      setName("");
      setParentId("");
      setIcon("");
      setDescription("");
      await load();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "falha ao criar categoria");
    } finally {
      setCreating(false);
    }
  }

  async function toggleActive(cat: Category) {
    await api.updateCategory(cat.id, {
      name: cat.name,
      description: cat.description,
      icon: cat.icon,
      position: cat.position,
      isActive: !cat.isActive,
    });
    await load();
  }

  async function remove(cat: Category) {
    if (!window.confirm(`Excluir a categoria "${cat.name}"?`)) return;
    try {
      await api.deleteCategory(cat.id);
      await load();
    } catch (err) {
      window.alert(err instanceof Error ? err.message : "falha ao excluir");
    }
  }

  if (authLoading || (!allowed && loading)) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  const roots = categories.filter((c) => c.parentId === null);

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Gerenciar categorias</h1>
        <Button variant="ghost" size="sm" asChild>
          <Link href="/categories">Voltar</Link>
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Plus className="h-4 w-4" />
            Nova categoria
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onCreate} className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="name">Nome</Label>
                <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required minLength={2} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="parent">Categoria pai (opcional)</Label>
                <select
                  id="parent"
                  value={parentId}
                  onChange={(e) => setParentId(e.target.value)}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                >
                  <option value="">— Nenhuma (categoria raiz)</option>
                  {roots.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="icon">Ícone (emoji, opcional)</Label>
              <Input id="icon" value={icon} onChange={(e) => setIcon(e.target.value)} maxLength={8} className="w-24" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Descrição (opcional)</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={2}
              />
            </div>
            {formError && <p className="text-sm text-destructive">{formError}</p>}
            <Button type="submit" disabled={creating}>
              {creating && <Loader2 className="h-4 w-4 animate-spin" />}
              Criar
            </Button>
          </form>
        </CardContent>
      </Card>

      {error && <p className="text-destructive">{error}</p>}

      <div className="space-y-2">
        {roots.map((cat) => (
          <div key={cat.id} className="space-y-2">
            <CategoryRow cat={cat} onToggle={toggleActive} onRemove={remove} />
            {categories
              .filter((c) => c.parentId === cat.id)
              .map((sub) => (
                <div key={sub.id} className="ml-6">
                  <CategoryRow cat={sub} onToggle={toggleActive} onRemove={remove} />
                </div>
              ))}
          </div>
        ))}
      </div>
    </div>
  );
}

function CategoryRow({
  cat,
  onToggle,
  onRemove,
}: {
  cat: Category;
  onToggle: (c: Category) => void;
  onRemove: (c: Category) => void;
}) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-lg border border-border bg-card p-3 text-card-foreground">
      <div className="flex min-w-0 items-center gap-2">
        <span>{cat.icon ?? "📁"}</span>
        <Link href={`/c/${cat.slug}`} className="truncate font-medium hover:underline">
          {cat.name}
        </Link>
        {!cat.isActive && (
          <span className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">inativa</span>
        )}
      </div>
      <div className="flex shrink-0 items-center gap-2">
        <Button variant="outline" size="sm" onClick={() => onToggle(cat)}>
          {cat.isActive ? "Desativar" : "Ativar"}
        </Button>
        <Button
          variant="ghost"
          size="icon"
          aria-label="Excluir"
          onClick={() => onRemove(cat)}
          className="text-muted-foreground hover:text-destructive"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
