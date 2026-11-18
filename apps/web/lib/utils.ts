import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/** Merge class names with Tailwind-aware conflict resolution. */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/** Human-friendly relative time in pt-BR (e.g. "há 3 h", "há 2 dias"). */
export function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "";
  const diff = Math.round((Date.now() - then) / 1000);
  const abs = Math.abs(diff);
  const units: [number, Intl.RelativeTimeFormatUnit][] = [
    [60, "second"],
    [3600, "minute"],
    [86400, "hour"],
    [604800, "day"],
    [2629800, "week"],
    [31557600, "month"],
    [Infinity, "year"],
  ];
  const divisors: Record<string, number> = {
    second: 1,
    minute: 60,
    hour: 3600,
    day: 86400,
    week: 604800,
    month: 2629800,
    year: 31557600,
  };
  const rtf = new Intl.RelativeTimeFormat("pt-BR", { numeric: "auto" });
  for (const [limit, unit] of units) {
    if (abs < limit) {
      return rtf.format(-Math.round(diff / divisors[unit]), unit);
    }
  }
  return "";
}

/** Compact count formatting (1200 -> "1,2 mil"). */
export function formatCount(n: number): string {
  return new Intl.NumberFormat("pt-BR", { notation: "compact" }).format(n);
}
