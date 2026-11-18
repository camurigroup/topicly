"use client";

import { Loader2 } from "lucide-react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api, type Category } from "@/lib/api";

export default function NewTopicPage() {
  const { slug } = useParams<{ slug: string }>();
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();

  const [category, setCategory] = useState<Category | null>(null);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!authLoading && !user) router.replace("/login");
  }, [authLoading, user, router]);

  useEffect(() => {
    let active = true;
    api
      .category(slug)
      .then(({ data }) => active && setCategory(data))
      .catch(() => active && setError("categoria não encontrada"));
    return () => {
      active = false;
    };
  }, [slug]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!category) return;
    setError(null);
    setBusy(true);
    try {
      const { data } = await api.createTopic({
        categoryId: category.id,
        title: title.trim(),
        content: content.trim(),
      });
      router.push(`/t/${data.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha ao criar tópico");
      setBusy(false);
    }
  }

  if (authLoading || !user) {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-4">
      <nav className="text-sm text-muted-foreground">
        <Link href="/categories" className="hover:underline">
          Categorias
        </Link>
        {category && (
          <>
            {" "}
            /{" "}
            <Link href={`/c/${category.slug}`} className="hover:underline">
              {category.name}
            </Link>
          </>
        )}{" "}
        / <span className="text-foreground">Novo tópico</span>
      </nav>

      <Card>
        <CardHeader>
          <CardTitle>Novo tópico{category ? ` em ${category.name}` : ""}</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="title">Título</Label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                minLength={3}
                maxLength={200}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="content">Conteúdo</Label>
              <Textarea
                id="content"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                rows={10}
                required
              />
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <div className="flex gap-3">
              <Button type="submit" disabled={busy || !category}>
                {busy && <Loader2 className="h-4 w-4 animate-spin" />}
                Publicar
              </Button>
              {category && (
                <Button type="button" variant="outline" asChild>
                  <Link href={`/c/${category.slug}`}>Cancelar</Link>
                </Button>
              )}
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
