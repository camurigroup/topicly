"use client";

import { Flag, ScrollText, ShieldCheck, Users } from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { useAuth } from "@/components/auth-provider";
import { cn } from "@/lib/utils";

interface Section {
  href: string;
  label: string;
  icon: typeof Flag;
  /** Any of these permissions unlocks the section. */
  perms: string[];
}

export const MODERATION_SECTIONS: Section[] = [
  { href: "/admin/reports", label: "Denúncias", icon: Flag, perms: ["reports.manage"] },
  { href: "/admin/users", label: "Usuários", icon: Users, perms: ["users.ban", "users.mute", "warnings.apply", "roles.manage"] },
  { href: "/admin/roles", label: "Cargos e permissões", icon: ShieldCheck, perms: ["roles.manage", "permissions.manage"] },
  { href: "/admin/audit", label: "Auditoria", icon: ScrollText, perms: ["logs.view"] },
];

/** Returns the moderation sections the viewer can open. */
export function useModerationSections(): Section[] {
  const { hasPermission } = useAuth();
  return MODERATION_SECTIONS.filter((s) => s.perms.some(hasPermission));
}

interface ModerationShellProps {
  title: string;
  description?: string;
  /** Permissions required for this page (any of). */
  perms: string[];
  actions?: ReactNode;
  children: ReactNode;
}

/**
 * Layout for moderation pages: section tabs + permission gate. The API
 * enforces every rule; this only hides screens the viewer cannot use.
 */
export function ModerationShell({ title, description, perms, actions, children }: ModerationShellProps) {
  const { loading, user, hasPermission } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const sections = useModerationSections();
  const allowed = perms.some(hasPermission);

  useEffect(() => {
    if (loading) return;
    if (!user) router.replace("/login");
    else if (!allowed) router.replace("/");
  }, [loading, user, allowed, router]);

  if (loading || !allowed) return null;

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <nav aria-label="Moderação" className="-mx-1 flex gap-1 overflow-x-auto pb-1">
        {sections.map((s) => {
          const active = pathname.startsWith(s.href);
          const Icon = s.icon;
          return (
            <Link
              key={s.href}
              href={s.href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "inline-flex shrink-0 items-center gap-2 rounded-full border px-3.5 py-1.5 text-sm font-medium transition-colors",
                active
                  ? "border-foreground bg-foreground text-background"
                  : "border-border text-muted-foreground hover:border-foreground/40 hover:text-foreground",
              )}
            >
              <Icon className="h-4 w-4" />
              {s.label}
            </Link>
          );
        })}
      </nav>

      <header className="flex flex-wrap items-end justify-between gap-4 border-b border-border pb-5">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Moderação</p>
          <h1 className="mt-1 text-3xl font-bold tracking-tight">{title}</h1>
          {description && <p className="mt-1 max-w-2xl text-sm text-muted-foreground">{description}</p>}
        </div>
        {actions}
      </header>

      {children}
    </div>
  );
}
