"use client";

import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { HomeBlock } from "@/components/home/widgets";
import { api, type HomeData, type HomeWidget } from "@/lib/api";

type State = { kind: "loading" } | { kind: "ready"; home: HomeData } | { kind: "error"; message: string };

/** A row is either one full-width block or a main/side column pair. */
type Row = { kind: "full"; w: HomeWidget } | { kind: "split"; main: HomeWidget[]; side: HomeWidget[] };

/** Groups consecutive non-full blocks into a two-column row. */
function toRows(layout: HomeWidget[]): Row[] {
  const rows: Row[] = [];
  for (const w of layout) {
    if (w.column === "full") {
      rows.push({ kind: "full", w });
      continue;
    }
    const last = rows[rows.length - 1];
    const split = last?.kind === "split" ? last : null;
    if (split) (w.column === "side" ? split.side : split.main).push(w);
    else rows.push({ kind: "split", main: w.column === "main" ? [w] : [], side: w.column === "side" ? [w] : [] });
  }
  return rows;
}

export default function HomePage() {
  const { user, hasPermission } = useAuth();
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    let active = true;
    api
      .home()
      .then(({ data }) => active && setState({ kind: "ready", home: data }))
      .catch((err) => active && setState({ kind: "error", message: err instanceof Error ? err.message : "erro" }));
    return () => {
      active = false;
    };
  }, []);

  if (state.kind === "loading") {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }
  if (state.kind === "error") {
    return <p className="py-20 text-center text-destructive">Não foi possível carregar a página inicial ({state.message}).</p>;
  }

  const { home } = state;
  const props = { home, user, canCustomize: hasPermission("settings.manage") };

  return (
    <div className="mx-auto max-w-6xl space-y-10">
      {toRows(home.layout ?? []).map((row, i) =>
        row.kind === "full" ? (
          <HomeBlock key={row.w.id} w={row.w} {...props} />
        ) : (
          <div key={`row-${i}`} className="grid gap-8 lg:grid-cols-3">
            {row.main.length > 0 && (
              <div className={row.side.length > 0 ? "min-w-0 space-y-8 lg:col-span-2" : "min-w-0 space-y-8 lg:col-span-3"}>
                {row.main.map((w) => (
                  <HomeBlock key={w.id} w={w} {...props} />
                ))}
              </div>
            )}
            {row.side.length > 0 && (
              <aside className={row.main.length > 0 ? "min-w-0 space-y-8" : "min-w-0 space-y-8 lg:col-span-3"}>
                {row.side.map((w) => (
                  <HomeBlock key={w.id} w={w} {...props} />
                ))}
              </aside>
            )}
          </div>
        ),
      )}
    </div>
  );
}
