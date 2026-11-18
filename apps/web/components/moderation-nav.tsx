"use client";

import {
  Flag,
  FolderTree,
  HardDrive,
  Home,
  LayoutDashboard,
  MessageSquareText,
  MessageSquareWarning,
  Palette,
  ScrollText,
  ShieldCheck,
  ShieldOff,
  StickyNote,
  Users,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { useAuth } from "@/components/auth-provider";

export interface AdminSection {
  href: string;
  label: string;
  icon: typeof Flag;
  /** Any of these permissions unlocks the section. */
  perms: string[];
}

export interface AdminGroup {
  label: string;
  sections: AdminSection[];
}

const USER_PERMS = ["users.manage", "users.ban", "users.mute", "warnings.apply", "roles.manage"];

export const ADMIN_GROUPS: AdminGroup[] = [
  {
    label: "Visão geral",
    sections: [{ href: "/admin", label: "Dashboard", icon: LayoutDashboard, perms: ["metrics.view"] }],
  },
  {
    label: "Comunidade",
    sections: [
      { href: "/admin/users", label: "Usuários", icon: Users, perms: USER_PERMS },
      { href: "/admin/roles", label: "Cargos e permissões", icon: ShieldCheck, perms: ["roles.manage", "permissions.manage"] },
      { href: "/admin/sanctions", label: "Banimentos e mutes", icon: ShieldOff, perms: ["users.ban", "users.mute"] },
      { href: "/admin/warnings", label: "Advertências", icon: MessageSquareWarning, perms: ["warnings.apply", "users.ban", "users.mute"] },
    ],
  },
  {
    label: "Conteúdo",
    sections: [
      { href: "/admin/categories", label: "Categorias", icon: FolderTree, perms: ["categories.manage"] },
      { href: "/admin/topics", label: "Tópicos", icon: StickyNote, perms: ["topics.edit", "topics.delete", "topics.lock", "topics.pin"] },
      { href: "/admin/replies", label: "Respostas", icon: MessageSquareText, perms: ["replies.edit", "replies.delete"] },
      { href: "/admin/reports", label: "Denúncias", icon: Flag, perms: ["reports.manage"] },
    ],
  },
  {
    label: "Site",
    sections: [
      { href: "/admin/settings", label: "Identidade visual", icon: Palette, perms: ["settings.manage"] },
      { href: "/admin/home", label: "Página inicial", icon: Home, perms: ["settings.manage"] },
      { href: "/admin/storage", label: "Armazenamento", icon: HardDrive, perms: ["storage.manage"] },
      { href: "/admin/audit", label: "Auditoria", icon: ScrollText, perms: ["logs.view"] },
    ],
  },
];

/** Admin groups filtered to the sections the viewer can open. */
export function useAdminGroups(): AdminGroup[] {
  const { hasPermission } = useAuth();
  return ADMIN_GROUPS.map((g) => ({ ...g, sections: g.sections.filter((s) => s.perms.some(hasPermission)) })).filter(
    (g) => g.sections.length > 0,
  );
}

/** Flat list of the admin sections the viewer can open. */
export function useModerationSections(): AdminSection[] {
  return useAdminGroups().flatMap((g) => g.sections);
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
 * Page frame for admin screens: title + permission gate. Navigation lives in
 * app/admin/layout.tsx. The API enforces every rule; this only hides screens
 * the viewer cannot use.
 */
export function ModerationShell({ title, description, perms, actions, children }: ModerationShellProps) {
  const { loading, user, hasPermission } = useAuth();
  const router = useRouter();
  const allowed = perms.some(hasPermission);

  useEffect(() => {
    if (loading) return;
    if (!user) router.replace("/login");
    else if (!allowed) router.replace("/");
  }, [loading, user, allowed, router]);

  if (loading || !allowed) return null;

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-end justify-between gap-4 border-b border-border pb-5">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
          {description && <p className="mt-1 max-w-2xl text-sm text-muted-foreground">{description}</p>}
        </div>
        {actions}
      </header>
      {children}
    </div>
  );
}
