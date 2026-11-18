"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

import type { SiteSettings } from "@/lib/api";
import { adminApi } from "@/lib/admin-api";
import { themeVars } from "@/lib/theme-color";

interface SiteSettingsContext {
  settings: SiteSettings;
  /** Brand name with a fallback. */
  brand: string;
  refresh: () => Promise<void>;
}

const Ctx = createContext<SiteSettingsContext>({ settings: {}, brand: "Forum", refresh: async () => {} });

/**
 * Holds the public white-label settings and applies them live: theme color
 * tokens on <html> and the favicon. `initial` comes from the server render.
 */
export function SiteSettingsProvider({ initial, children }: { initial: SiteSettings; children: ReactNode }) {
  const [settings, setSettings] = useState<SiteSettings>(initial);

  const refresh = useCallback(async () => {
    try {
      const { data } = await adminApi.siteSettings();
      setSettings(data);
    } catch {
      /* keep the last known settings */
    }
  }, []);

  useEffect(() => {
    if (Object.keys(initial).length === 0) void refresh();
  }, [initial, refresh]);

  useEffect(() => {
    const root = document.documentElement;
    const vars = themeVars(settings["theme.primary"], settings["theme.accent"], settings["theme.radius"]);
    for (const k of ["--primary", "--primary-foreground", "--ring", "--radius"]) {
      if (vars[k]) root.style.setProperty(k, vars[k]);
      else root.style.removeProperty(k);
    }
  }, [settings]);

  useEffect(() => {
    const href = settings["brand.favicon"];
    if (!href) return;
    let link = document.querySelector<HTMLLinkElement>("link[rel='icon']");
    if (!link) {
      link = document.createElement("link");
      link.rel = "icon";
      document.head.appendChild(link);
    }
    link.href = href;
  }, [settings]);

  const value = useMemo(() => ({ settings, brand: settings["brand.name"] || "Forum", refresh }), [settings, refresh]);
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useSiteSettings(): SiteSettingsContext {
  return useContext(Ctx);
}
