"use client";

import { Github, Globe, Instagram, MessageCircle, Twitter, Youtube } from "lucide-react";
import Link from "next/link";

import { useSiteSettings } from "@/components/site-settings-provider";

const SOCIAL = [
  { key: "social.website", label: "Site", icon: Globe },
  { key: "social.github", label: "GitHub", icon: Github },
  { key: "social.x", label: "X / Twitter", icon: Twitter },
  { key: "social.instagram", label: "Instagram", icon: Instagram },
  { key: "social.discord", label: "Discord", icon: MessageCircle },
  { key: "social.youtube", label: "YouTube", icon: Youtube },
];

/** Parses "Label|url" lines (validated by the API). */
function parseLinks(raw: string | undefined): { label: string; href: string }[] {
  return (raw ?? "")
    .split("\n")
    .map((l) => l.split("|"))
    .filter((p) => p.length >= 2 && p[0].trim() && p[1].trim())
    .map(([label, href]) => ({ label: label.trim(), href: href.trim() }));
}

const isExternal = (href: string) => /^https?:\/\//.test(href);

/** Public footer: brand text, institutional links and social icons. */
export function SiteFooter() {
  const { settings, brand } = useSiteSettings();
  const links = parseLinks(settings["footer.links"]);
  const social = SOCIAL.filter((s) => settings[s.key]);

  return (
    <footer className="border-t border-border py-8">
      <div className="container flex flex-col gap-4 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <p>
          <span className="font-medium text-foreground">{brand}</span>
          {settings["footer.text"] && <> · {settings["footer.text"]}</>}
        </p>
        <div className="flex flex-wrap items-center gap-x-5 gap-y-2">
          {links.length > 0 && (
            <nav aria-label="Links institucionais" className="flex flex-wrap gap-x-4 gap-y-1">
              {links.map((l) =>
                isExternal(l.href) ? (
                  <a key={l.href + l.label} href={l.href} target="_blank" rel="noopener noreferrer" className="hover:text-foreground">
                    {l.label}
                  </a>
                ) : (
                  <Link key={l.href + l.label} href={l.href} className="hover:text-foreground">
                    {l.label}
                  </Link>
                ),
              )}
            </nav>
          )}
          {social.length > 0 && (
            <ul className="flex gap-1" aria-label="Redes sociais">
              {social.map(({ key, label, icon: Icon }) => (
                <li key={key}>
                  <a
                    href={settings[key]}
                    target="_blank"
                    rel="noopener noreferrer"
                    aria-label={label}
                    title={label}
                    className="flex rounded-md p-1.5 hover:bg-secondary hover:text-foreground"
                  >
                    <Icon className="h-4 w-4" />
                  </a>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </footer>
  );
}
