"use client";

import { useSiteSettings } from "@/components/site-settings-provider";

/** Renders an administrable legal text (plain text, line breaks kept). */
export function LegalPage({ title, settingKey }: { title: string; settingKey: string }) {
  const { settings, brand } = useSiteSettings();
  const text = settings[settingKey];
  return (
    <article className="mx-auto max-w-2xl">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">{brand}</p>
      <h1 className="mt-2 text-3xl font-bold tracking-tight">{title}</h1>
      <div className="mt-6 whitespace-pre-wrap text-sm leading-relaxed text-foreground/90">
        {text || "Este documento ainda não foi publicado."}
      </div>
    </article>
  );
}
