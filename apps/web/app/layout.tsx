import type { Metadata } from "next";

import { AuthProvider } from "@/components/auth-provider";
import { NotificationsProvider } from "@/components/notifications-provider";
import { SiteHeader } from "@/components/site-header";
import { ThemeProvider } from "@/components/theme-provider";
import "./globals.css";

export const metadata: Metadata = {
  title: "Forum",
  description: "Comunidade construída com o Forum Framework.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="pt-BR" suppressHydrationWarning>
      <body className="min-h-screen antialiased">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <AuthProvider>
            <NotificationsProvider>
              <div className="flex min-h-screen flex-col">
                <SiteHeader />
                <main className="container flex-1 py-10">{children}</main>
                <footer className="border-t border-border py-6">
                  <div className="container text-sm text-muted-foreground">
                    Construído com o Forum Framework.
                  </div>
                </footer>
              </div>
            </NotificationsProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
