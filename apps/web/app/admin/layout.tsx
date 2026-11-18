"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { useAuth } from "@/components/auth-provider";
import { useAdminGroups } from "@/components/moderation-nav";
import { cn } from "@/lib/utils";

function isActive(pathname: string, href: string) {
  return href === "/admin" ? pathname === "/admin" : pathname === href || pathname.startsWith(href + "/");
}

/**
 * Admin panel frame: sticky side navigation on desktop, a scrollable strip on
 * small screens. Sections are filtered by the viewer's permissions.
 */
export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { loading, user } = useAuth();
  const groups = useAdminGroups();

  if (loading || !user) return <>{children}</>;

  return (
    <div className="-mt-4 grid gap-8 lg:grid-cols-[220px_1fr]">
      <aside className="lg:sticky lg:top-24 lg:self-start">
        <p className="mb-3 hidden text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground lg:block">
          Painel
        </p>
        <nav
          aria-label="Painel administrativo"
          className="-mx-1 flex gap-1 overflow-x-auto px-1 pb-1 [scrollbar-width:none] lg:mx-0 lg:flex-col lg:gap-5 lg:overflow-visible lg:px-0 [&::-webkit-scrollbar]:hidden"
        >
          {groups.map((g) => (
            <div key={g.label} className="contents lg:block">
              <p className="hidden px-3 pb-1 text-[11px] font-medium uppercase tracking-wider text-muted-foreground/70 lg:block">
                {g.label}
              </p>
              <ul className="contents lg:block lg:space-y-0.5">
                {g.sections.map((s) => {
                  const active = isActive(pathname, s.href);
                  const Icon = s.icon;
                  return (
                    <li key={s.href} className="shrink-0">
                      <Link
                        href={s.href}
                        aria-current={active ? "page" : undefined}
                        className={cn(
                          "group relative flex items-center gap-2.5 whitespace-nowrap rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                          active
                            ? "bg-foreground text-background lg:bg-secondary lg:text-foreground"
                            : "text-muted-foreground hover:bg-secondary hover:text-foreground",
                        )}
                      >
                        {active && (
                          <span className="absolute inset-y-1.5 left-0 hidden w-0.5 rounded-full bg-primary lg:block" aria-hidden />
                        )}
                        <Icon className={cn("h-4 w-4", active && "lg:text-primary")} />
                        {s.label}
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </nav>
      </aside>
      <div className="min-w-0">{children}</div>
    </div>
  );
}
