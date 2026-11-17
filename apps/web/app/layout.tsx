import type { Metadata } from "next";
import Link from "next/link";
import { MessagesSquare } from "lucide-react";

import { ThemeProvider } from "@/components/theme-provider";
import { ThemeToggle } from "@/components/theme-toggle";
import "./globals.css";

export const metadata: Metadata = {
  title: "Forum",
  description: "Comunidade construída com o Forum Framework.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="pt-BR" suppressHydrationWarning>
      <body className="min-h-screen antialiased">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <div className="flex min-h-screen flex-col">
            <header className="sticky top-0 z-40 border-b border-border bg-background/80 backdrop-blur">
              <div className="container flex h-16 items-center justify-between">
                <Link href="/" className="flex items-center gap-2 font-semibold">
                  <MessagesSquare className="h-5 w-5 text-primary" />
                  <span>Forum</span>
                </Link>
                <ThemeToggle />
              </div>
            </header>
            <main className="container flex-1 py-10">{children}</main>
            <footer className="border-t border-border py-6">
              <div className="container text-sm text-muted-foreground">
                Construído com o Forum Framework.
              </div>
            </footer>
          </div>
        </ThemeProvider>
      </body>
    </html>
  );
}
