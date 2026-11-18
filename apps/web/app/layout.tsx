import type { Metadata } from "next";
import type { CSSProperties } from "react";

import { AuthProvider } from "@/components/auth-provider";
import { NotificationsProvider } from "@/components/notifications-provider";
import { SiteFooter } from "@/components/site-footer";
import { SiteHeader } from "@/components/site-header";
import { SiteSettingsProvider } from "@/components/site-settings-provider";
import { ThemeProvider } from "@/components/theme-provider";
import { API_URL, type SiteSettings } from "@/lib/api";
import { themeVars } from "@/lib/theme-color";
import "./globals.css";

/** White-label settings, refreshed at most every 30s on the server. */
async function loadSettings(): Promise<SiteSettings> {
  try {
    const res = await fetch(`${API_URL}/api/v1/site/settings`, { next: { revalidate: 30 } });
    if (!res.ok) return {};
    const body = (await res.json()) as { data?: SiteSettings };
    return body.data ?? {};
  } catch {
    return {};
  }
}

export async function generateMetadata(): Promise<Metadata> {
  const s = await loadSettings();
  const name = s["brand.name"] || "Forum";
  return {
    title: { default: name, template: `%s · ${name}` },
    description: s["brand.description"] || "Comunidade construída com o Forum Framework.",
    icons: s["brand.favicon"] ? { icon: s["brand.favicon"] } : undefined,
  };
}

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const settings = await loadSettings();
  const vars = themeVars(settings["theme.primary"], settings["theme.accent"], settings["theme.radius"]);

  return (
    <html lang="pt-BR" suppressHydrationWarning style={vars as CSSProperties}>
      <body className="min-h-screen antialiased">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <SiteSettingsProvider initial={settings}>
            <AuthProvider>
              <NotificationsProvider>
                <div className="flex min-h-screen flex-col">
                  <SiteHeader />
                  <main className="container flex-1 py-10">{children}</main>
                  <SiteFooter />
                </div>
              </NotificationsProvider>
            </AuthProvider>
          </SiteSettingsProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
