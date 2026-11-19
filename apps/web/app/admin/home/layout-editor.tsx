"use client";

import {
  ArrowDown,
  ArrowUp,
  Eye,
  EyeOff,
  Loader2,
  Plus,
  Trash2,
} from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  api,
  type Category,
  type HomeWidget,
  type TopicSource,
  type WidgetColumn,
  type WidgetType,
} from "@/lib/api";
import { modApi, type RoleDetail } from "@/lib/moderation-api";
import { cn } from "@/lib/utils";

const TYPES: Record<
  WidgetType,
  { label: string; variants: [string, string][]; limit: boolean }
> = {
  hero: {
    label: "Banner principal",
    variants: [
      ["large", "Grande"],
      ["compact", "Compacto"],
    ],
    limit: false,
  },
  intro: {
    label: "Apresentação",
    variants: [
      ["text", "Texto"],
      ["card", "Cartão"],
    ],
    limit: false,
  },
  stats: {
    label: "Estatísticas",
    variants: [
      ["grid", "Grade"],
      ["list", "Lista"],
    ],
    limit: false,
  },
  categories: {
    label: "Categorias",
    variants: [
      ["grid", "Grade"],
      ["list", "Lista"],
    ],
    limit: true,
  },
  topics: {
    label: "Tópicos",
    variants: [
      ["list", "Lista"],
      ["cards", "Cartões"],
      ["compact", "Compacto"],
    ],
    limit: true,
  },
  members: {
    label: "Membros em destaque",
    variants: [
      ["list", "Lista"],
      ["avatars", "Avatares"],
    ],
    limit: true,
  },
  activity: {
    label: "Últimas atividades",
    variants: [["list", "Lista"]],
    limit: true,
  },
};

const COLUMNS: [WidgetColumn, string][] = [
  ["full", "Largura total"],
  ["main", "Coluna principal"],
  ["side", "Coluna lateral"],
];

const SOURCES: [TopicSource, string][] = [
  ["recent", "Mais recentes"],
  ["popular", "Mais populares"],
  ["trending", "Em alta"],
  ["featured", "Em destaque"],
  ["roles", "De autores com cargo…"],
  ["category", "De uma categoria…"],
];

const select = "h-9 rounded-md border border-input bg-background px-2 text-sm";

function newId(type: WidgetType, taken: HomeWidget[]): string {
  let n = 1;
  while (taken.some((w) => w.id === `${type}-${n}`)) n++;
  return `${type}-${n}`;
}

interface LayoutEditorProps {
  initial: string | undefined;
}

/**
 * Visual editor for the home layout: order, visibility, position (full /
 * main / side), style, quantity and — for topic blocks — the source (recent,
 * popular, featured, by author role or by category).
 */
export function LayoutEditor({ initial }: LayoutEditorProps) {
  const [widgets, setWidgets] = useState<HomeWidget[]>(() => {
    try {
      return JSON.parse(initial ?? "[]") as HomeWidget[];
    } catch {
      return [];
    }
  });
  const [saved, setSaved] = useState(initial ?? "[]");
  const [roles, setRoles] = useState<RoleDetail[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);
  const [addType, setAddType] = useState<WidgetType>("topics");

  useEffect(() => {
    modApi
      .roles()
      .then(({ data }) => setRoles(data))
      .catch(() => setRoles([]));
    api
      .categories()
      .then(({ data }) => setCategories(data))
      .catch(() => setCategories([]));
  }, []);

  const dirty = JSON.stringify(widgets) !== saved;
  const patch = (i: number, p: Partial<HomeWidget>) =>
    setWidgets((ws) => ws.map((w, j) => (j === i ? { ...w, ...p } : w)));
  const move = (i: number, d: -1 | 1) =>
    setWidgets((ws) => {
      const j = i + d;
      if (j < 0 || j >= ws.length) return ws;
      const next = [...ws];
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });

  function add() {
    const t = TYPES[addType];
    setWidgets((ws) => [
      ...ws,
      {
        id: newId(addType, ws),
        type: addType,
        enabled: true,
        column: addType === "hero" ? "full" : "main",
        title: addType === "hero" ? "" : t.label,
        variant: t.variants[0][0],
        limit: 6,
        ...(addType === "topics" ? { source: "recent" as TopicSource } : {}),
      },
    ]);
  }

  async function save() {
    setBusy(true);
    setMsg(null);
    try {
      const { data } = await api.updateSiteSettings({
        "home.layout": JSON.stringify(widgets),
      });
      const canonical = data["home.layout"];
      setSaved(canonical);
      setWidgets(JSON.parse(canonical) as HomeWidget[]);
      setMsg({ ok: true, text: "Layout salvo — a página inicial já mudou." });
    } catch (err) {
      setMsg({
        ok: false,
        text: err instanceof Error ? err.message : "falha ao salvar o layout",
      });
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="space-y-4 rounded-2xl border border-border bg-card p-5 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">Layout</h2>
          <p className="text-sm text-muted-foreground">
            O que aparece, em que ordem, onde e como. Blocos seguidos de
            “principal” e “lateral” ficam lado a lado.
          </p>
        </div>
        <LayoutMap widgets={widgets} />
      </div>

      <ol className="space-y-2">
        {widgets.map((w, i) => {
          const t = TYPES[w.type];
          return (
            <li
              key={w.id}
              className={cn(
                "rounded-xl border border-border p-3 transition-opacity",
                !w.enabled && "opacity-55",
              )}
            >
              <div className="flex flex-wrap items-center gap-2">
                <div className="flex flex-col">
                  <button
                    type="button"
                    aria-label="Subir"
                    onClick={() => move(i, -1)}
                    disabled={i === 0}
                    className="rounded p-0.5 text-muted-foreground hover:text-foreground disabled:opacity-30"
                  >
                    <ArrowUp className="h-3.5 w-3.5" />
                  </button>
                  <button
                    type="button"
                    aria-label="Descer"
                    onClick={() => move(i, 1)}
                    disabled={i === widgets.length - 1}
                    className="rounded p-0.5 text-muted-foreground hover:text-foreground disabled:opacity-30"
                  >
                    <ArrowDown className="h-3.5 w-3.5" />
                  </button>
                </div>
                <span className="w-36 shrink-0 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {t.label}
                </span>
                {w.type !== "hero" && (
                  <Input
                    value={w.title}
                    onChange={(e) => patch(i, { title: e.target.value })}
                    placeholder="Título"
                    maxLength={80}
                    className="h-9 min-w-0 flex-1 basis-40"
                    aria-label="Título do bloco"
                  />
                )}
                <select
                  className={select}
                  value={w.column}
                  onChange={(e) =>
                    patch(i, { column: e.target.value as WidgetColumn })
                  }
                  aria-label="Posição"
                >
                  {COLUMNS.map(([v, l]) => (
                    <option key={v} value={v}>
                      {l}
                    </option>
                  ))}
                </select>
                <select
                  className={select}
                  value={w.variant}
                  onChange={(e) => patch(i, { variant: e.target.value })}
                  aria-label="Estilo"
                >
                  {t.variants.map(([v, l]) => (
                    <option key={v} value={v}>
                      {l}
                    </option>
                  ))}
                </select>
                {t.limit && (
                  <Input
                    type="number"
                    min={1}
                    max={24}
                    value={w.limit}
                    onChange={(e) =>
                      patch(i, { limit: Number(e.target.value) })
                    }
                    className="h-9 w-16"
                    aria-label="Quantidade"
                    title="Quantidade"
                  />
                )}
                <div className="ml-auto flex items-center gap-1">
                  <button
                    type="button"
                    onClick={() => patch(i, { enabled: !w.enabled })}
                    aria-label={w.enabled ? "Ocultar bloco" : "Mostrar bloco"}
                    title={w.enabled ? "Visível" : "Oculto"}
                    className="rounded-md p-1.5 text-muted-foreground hover:bg-secondary hover:text-foreground"
                  >
                    {w.enabled ? (
                      <Eye className="h-4 w-4" />
                    ) : (
                      <EyeOff className="h-4 w-4" />
                    )}
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      setWidgets((ws) => ws.filter((_, j) => j !== i))
                    }
                    aria-label="Remover bloco"
                    className="rounded-md p-1.5 text-muted-foreground hover:bg-secondary hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>

              {w.type === "topics" && (
                <div className="mt-2 flex flex-wrap items-center gap-2 pl-7">
                  <span className="text-xs text-muted-foreground">Fonte</span>
                  <select
                    className={select}
                    value={w.source ?? "recent"}
                    onChange={(e) =>
                      patch(i, { source: e.target.value as TopicSource })
                    }
                    aria-label="Fonte dos tópicos"
                  >
                    {SOURCES.map(([v, l]) => (
                      <option key={v} value={v}>
                        {l}
                      </option>
                    ))}
                  </select>
                  {w.source === "roles" &&
                    roles.map((r) => {
                      const on = (w.roles ?? []).includes(r.slug);
                      return (
                        <button
                          key={r.slug}
                          type="button"
                          onClick={() =>
                            patch(i, {
                              roles: on
                                ? (w.roles ?? []).filter((x) => x !== r.slug)
                                : [...(w.roles ?? []), r.slug],
                            })
                          }
                          className={cn(
                            "rounded-full border px-2.5 py-1 text-xs",
                            on
                              ? "border-primary bg-primary/10 text-primary"
                              : "border-border text-muted-foreground hover:text-foreground",
                          )}
                          aria-pressed={on}
                        >
                          {r.name}
                        </button>
                      );
                    })}
                  {w.source === "category" && (
                    <select
                      className={select}
                      value={w.category ?? ""}
                      onChange={(e) => patch(i, { category: e.target.value })}
                      aria-label="Categoria"
                    >
                      <option value="">Escolha…</option>
                      {categories.map((c) => (
                        <option key={c.id} value={c.slug}>
                          {c.name}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
              )}
            </li>
          );
        })}
      </ol>

      <div className="flex flex-wrap items-center gap-2 border-t border-border pt-4">
        <select
          className={select}
          value={addType}
          onChange={(e) => setAddType(e.target.value as WidgetType)}
          aria-label="Tipo do novo bloco"
        >
          {Object.entries(TYPES).map(([v, t]) => (
            <option key={v} value={v}>
              {t.label}
            </option>
          ))}
        </select>
        <Button type="button" variant="outline" size="sm" onClick={add}>
          <Plus className="h-4 w-4" /> Adicionar bloco
        </Button>
        {msg && (
          <p
            role={msg.ok ? "status" : "alert"}
            className={cn(
              "text-sm",
              msg.ok ? "text-primary" : "text-destructive",
            )}
          >
            {msg.text}
          </p>
        )}
        <Button
          type="button"
          className="ml-auto"
          disabled={!dirty || busy}
          onClick={save}
        >
          {busy && <Loader2 className="h-4 w-4 animate-spin" />}
          Salvar layout
        </Button>
      </div>
    </section>
  );
}

/** Tiny schematic of the resulting page (full rows and main/side columns). */
function LayoutMap({ widgets }: { widgets: HomeWidget[] }) {
  const visible = widgets.filter((w) => w.enabled);
  const rows: { full?: HomeWidget; main: HomeWidget[]; side: HomeWidget[] }[] =
    [];
  for (const w of visible) {
    const last = rows[rows.length - 1];
    if (w.column === "full") rows.push({ full: w, main: [], side: [] });
    else if (last && !last.full)
      (w.column === "side" ? last.side : last.main).push(w);
    else
      rows.push({
        main: w.column === "main" ? [w] : [],
        side: w.column === "side" ? [w] : [],
      });
  }
  const bar = (w: HomeWidget) => (
    <span
      key={w.id}
      title={w.title || TYPES[w.type].label}
      className="block h-2 rounded-sm bg-primary/60"
    />
  );
  return (
    <div
      aria-hidden
      className="w-40 space-y-1 rounded-lg border border-border bg-background p-2"
    >
      {rows.map((r, i) =>
        r.full ? (
          <span key={i} className="block h-4 rounded-sm bg-primary" />
        ) : (
          <div key={i} className="flex gap-1">
            <div className="flex-[2] space-y-1">{r.main.map(bar)}</div>
            <div className="flex-1 space-y-1">{r.side.map(bar)}</div>
          </div>
        ),
      )}
    </div>
  );
}
