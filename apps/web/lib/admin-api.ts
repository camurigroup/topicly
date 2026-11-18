/**
 * Client for the Etapa 6 admin panel: real metrics, filterable listings,
 * storage info, brand assets and the public site settings (white-label).
 */
import { paged, query, raw, type Page, type SiteSettings } from "@/lib/api";

export interface Totals {
  users: number;
  activeUsers: number;
  newUsers: number;
  categories: number;
  subcategories: number;
  topics: number;
  replies: number;
  likes: number;
  reportsPending: number;
  reports: number;
  banned: number;
  muted: number;
  warnings: number;
}

export type SeriesMetric = "users" | "topics" | "replies" | "likes" | "reports";
export type SeriesPeriod = "daily" | "weekly" | "monthly";

export interface SeriesPoint {
  label: string;
  value: number;
}

export interface AdminUser {
  id: string;
  username: string;
  displayName: string;
  email: string;
  avatarUrl: string | null;
  reputation: number;
  roles: string[];
  topics: number;
  replies: number;
  isBanned: boolean;
  isMuted: boolean;
  createdAt: string;
  lastActiveAt: string | null;
}

export interface AdminTopic {
  id: string;
  title: string;
  categorySlug: string;
  categoryName: string;
  author: string;
  isPinned: boolean;
  isLocked: boolean;
  isFeatured: boolean;
  views: number;
  replies: number;
  likes: number;
  createdAt: string;
  lastActivityAt: string;
}

export interface AdminReply {
  id: string;
  excerpt: string;
  topicId: string;
  topicTitle: string;
  author: string;
  likes: number;
  isDeleted: boolean;
  createdAt: string;
}

export interface AdminWarning {
  id: string;
  username: string;
  displayName: string;
  reason: string;
  issuedBy: string | null;
  createdAt: string;
}

export interface StorageInfo {
  provider: string;
  localPath?: string;
  publicUrl?: string;
  endpoint?: string;
  bucket?: string;
  useSsl: boolean;
  maxImageBytes: number;
  avatars: number;
  banners: number;
}

/** Common list parameters: search, sort, order and page. */
export interface ListParams {
  q?: string;
  sort?: string;
  order?: "asc" | "desc";
  page?: number;
}

type Params = Record<string, string | number | undefined>;

export const adminApi = {
  metrics: () => raw<Totals>("/api/v1/admin/metrics", { auth: true }),
  series: (metric: SeriesMetric, period: SeriesPeriod) =>
    raw<SeriesPoint[]>(`/api/v1/admin/metrics/series${query({ metric, period })}`, { auth: true }),

  users: (p: ListParams & { role?: string; status?: string }): Promise<Page<AdminUser>> =>
    paged<AdminUser>(`/api/v1/admin/users${query(p as Params)}`, true),
  topics: (p: ListParams & { category?: string; flag?: string }): Promise<Page<AdminTopic>> =>
    paged<AdminTopic>(`/api/v1/admin/topics${query(p as Params)}`, true),
  replies: (p: ListParams & { state?: string }): Promise<Page<AdminReply>> =>
    paged<AdminReply>(`/api/v1/admin/replies${query(p as Params)}`, true),
  warnings: (p: ListParams): Promise<Page<AdminWarning>> =>
    paged<AdminWarning>(`/api/v1/admin/warnings${query(p as Params)}`, true),

  storage: () => raw<StorageInfo>("/api/v1/admin/storage", { auth: true }),
  uploadAsset: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return raw<{ url: string }>("/api/v1/admin/site/assets", { method: "POST", body: form, auth: true, isForm: true });
  },

  /** Public white-label settings (brand, colors, footer, legal texts). */
  siteSettings: () => raw<SiteSettings>("/api/v1/site/settings", {}),
};
