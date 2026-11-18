/**
 * Client for Etapa 5: roles & permissions management, moderation (bans,
 * mutes, warnings), reports and the audit log. Every rule is enforced by the
 * API; the UI only mirrors it.
 */
import { paged, query, raw, type Page } from "@/lib/api";

/* --------------------------------- types --------------------------------- */

export interface Role {
  id: string;
  slug: string;
  name: string;
  rankOrder: number;
  isSystem: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface RoleDetail extends Role {
  permissions: string[];
  memberCount: number;
}

export interface Permission {
  id: string;
  slug: string;
  name: string;
  description: string | null;
  isActive: boolean;
  isBuiltin: boolean;
  isCore: boolean;
  createdAt: string;
}

export interface UserRef {
  id: string;
  username: string;
  displayName: string;
  avatarUrl?: string | null;
}

export type SanctionType = "ban" | "mute";

export interface Sanction {
  id: string;
  user: UserRef;
  type: SanctionType;
  reason: string;
  issuedBy: UserRef | null;
  expiresAt: string | null;
  revokedAt: string | null;
  revokedBy: UserRef | null;
  revokeReason: string | null;
  createdAt: string;
  active: boolean;
}

export interface Warning {
  id: string;
  user: UserRef;
  reason: string;
  issuedBy: UserRef | null;
  createdAt: string;
}

export interface UserModeration {
  user: UserRef;
  roles: string[];
  rank: number;
  isOwner: boolean;
  activeBan: Sanction | null;
  activeMute: Sanction | null;
  sanctions: Sanction[];
  warnings: Warning[];
  canAct: boolean;
}

export type ReportTargetType = "user" | "topic" | "reply";
export type ReportStatus = "pending" | "reviewing" | "resolved" | "rejected";
export type ReportReason = "spam" | "harassment" | "offensive" | "inappropriate" | "misinformation" | "other";

export const REPORT_REASONS: { value: ReportReason; label: string }[] = [
  { value: "spam", label: "Spam ou propaganda" },
  { value: "harassment", label: "Assédio ou ataque pessoal" },
  { value: "offensive", label: "Conteúdo ofensivo" },
  { value: "inappropriate", label: "Conteúdo impróprio" },
  { value: "misinformation", label: "Desinformação" },
  { value: "other", label: "Outro motivo" },
];

export const REPORT_STATUS_LABEL: Record<ReportStatus, string> = {
  pending: "Pendentes",
  reviewing: "Em análise",
  resolved: "Resolvidas",
  rejected: "Rejeitadas",
};

export interface Report {
  id: string;
  reporter: UserRef | null;
  target: {
    type: ReportTargetType;
    id: string;
    title: string;
    excerpt: string;
    link: string;
    author: string | null;
    exists: boolean;
  };
  reason: ReportReason;
  details: string | null;
  status: ReportStatus;
  handledBy: UserRef | null;
  resolutionNote: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface AuditEntry {
  id: string;
  actor: UserRef | null;
  action: string;
  targetType: string | null;
  targetId: string | null;
  details: Record<string, unknown> | null;
  ip: string | null;
  createdAt: string;
}

/** Duration presets (hours; 0 = permanent). */
export const DURATIONS: { hours: number; label: string }[] = [
  { hours: 1, label: "1 hora" },
  { hours: 24, label: "1 dia" },
  { hours: 24 * 7, label: "7 dias" },
  { hours: 24 * 30, label: "30 dias" },
  { hours: 0, label: "Permanente" },
];

const u = (username: string) => encodeURIComponent(username);

/* ------------------------------ API surface ------------------------------ */

export const modApi = {
  /* roles & permissions */
  roles: () => raw<RoleDetail[]>("/api/v1/admin/roles", { auth: true }),
  createRole: (input: { slug: string; name: string; rankOrder: number; permissions: string[] }) =>
    raw<Role>("/api/v1/admin/roles", { method: "POST", body: input, auth: true }),
  updateRole: (id: string, input: { name: string; rankOrder: number }) =>
    raw<Role>(`/api/v1/admin/roles/${id}`, { method: "PATCH", body: input, auth: true }),
  setRolePermissions: (id: string, permissions: string[]) =>
    raw<{ permissions: string[] }>(`/api/v1/admin/roles/${id}/permissions`, {
      method: "PUT",
      body: { permissions },
      auth: true,
    }),
  deleteRole: (id: string) => raw<{ deleted: boolean }>(`/api/v1/admin/roles/${id}`, { method: "DELETE", auth: true }),

  permissions: () => raw<Permission[]>("/api/v1/admin/permissions", { auth: true }),
  createPermission: (input: { slug: string; name: string; description?: string }) =>
    raw<Permission>("/api/v1/admin/permissions", { method: "POST", body: input, auth: true }),
  setPermissionActive: (id: string, isActive: boolean) =>
    raw<Permission>(`/api/v1/admin/permissions/${id}`, { method: "PATCH", body: { isActive }, auth: true }),
  deletePermission: (id: string) =>
    raw<{ deleted: boolean }>(`/api/v1/admin/permissions/${id}`, { method: "DELETE", auth: true }),

  setUserRoles: (username: string, roles: string[]) =>
    raw<Role[]>(`/api/v1/admin/users/${u(username)}/roles`, { method: "PUT", body: { roles }, auth: true }),

  /* moderation */
  userModeration: (username: string) =>
    raw<UserModeration>(`/api/v1/admin/users/${u(username)}/moderation`, { auth: true }),
  ban: (username: string, input: { reason: string; durationHours: number }) =>
    raw<Sanction>(`/api/v1/admin/users/${u(username)}/ban`, { method: "POST", body: input, auth: true }),
  mute: (username: string, input: { reason: string; durationHours: number }) =>
    raw<Sanction>(`/api/v1/admin/users/${u(username)}/mute`, { method: "POST", body: input, auth: true }),
  warn: (username: string, reason: string) =>
    raw<Warning>(`/api/v1/admin/users/${u(username)}/warnings`, { method: "POST", body: { reason }, auth: true }),
  revoke: (sanctionId: string, reason: string) =>
    raw<Sanction>(`/api/v1/admin/sanctions/${sanctionId}/revoke`, { method: "POST", body: { reason }, auth: true }),
  sanctions: (params: { type?: SanctionType; active?: boolean; page?: number } = {}): Promise<Page<Sanction>> =>
    paged<Sanction>(
      `/api/v1/admin/sanctions${query({ type: params.type, active: params.active ? "1" : undefined, page: params.page })}`,
      true,
    ),

  /* reports */
  report: (input: { targetType: ReportTargetType; targetId: string; reason: ReportReason; details?: string }) =>
    raw<Report>("/api/v1/reports", { method: "POST", body: input, auth: true }),
  reports: (params: { status?: ReportStatus; targetType?: ReportTargetType; page?: number } = {}) =>
    paged<Report>(`/api/v1/admin/reports${query(params)}`, true),
  reportCounts: () => raw<Record<ReportStatus, number>>("/api/v1/admin/reports/counts", { auth: true }),
  updateReport: (id: string, input: { status: ReportStatus; note?: string }) =>
    raw<Report>(`/api/v1/admin/reports/${id}`, { method: "PATCH", body: input, auth: true }),

  /* audit */
  audit: (params: { action?: string; page?: number } = {}) =>
    paged<AuditEntry>(`/api/v1/admin/audit${query(params)}`, true),
  auditActions: () => raw<string[]>("/api/v1/admin/audit/actions", { auth: true }),
};

/** Permissions that grant access to at least one moderation screen. */
export const MODERATION_PERMISSIONS = [
  "reports.manage",
  "users.ban",
  "users.mute",
  "warnings.apply",
  "roles.manage",
  "permissions.manage",
  "logs.view",
];
