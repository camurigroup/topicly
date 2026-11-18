/**
 * Converts brand hex colors into the HSL triplets used by the theme tokens
 * (e.g. "--primary: 243 75% 59%"), and picks a readable foreground.
 */

export interface Hsl {
  h: number;
  s: number;
  l: number;
}

/** Parses #rrggbb into HSL, or returns null for invalid input. */
export function hexToHsl(hex: string): Hsl | null {
  const m = /^#([0-9a-f]{6})$/i.exec(hex.trim());
  if (!m) return null;
  const n = parseInt(m[1], 16);
  const r = ((n >> 16) & 255) / 255;
  const g = ((n >> 8) & 255) / 255;
  const b = (n & 255) / 255;
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const l = (max + min) / 2;
  let h = 0;
  let s = 0;
  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    if (max === r) h = (g - b) / d + (g < b ? 6 : 0);
    else if (max === g) h = (b - r) / d + 2;
    else h = (r - g) / d + 4;
    h *= 60;
  }
  return { h: Math.round(h), s: Math.round(s * 100), l: Math.round(l * 100) };
}

export const hslToken = ({ h, s, l }: Hsl) => `${h} ${s}% ${l}%`;

/** Relative luminance check: dark text on light colors, white otherwise. */
export function readableForeground(hex: string): string {
  const n = parseInt(hex.slice(1), 16);
  const lin = (c: number) => {
    const v = c / 255;
    return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
  };
  const lum = 0.2126 * lin((n >> 16) & 255) + 0.7152 * lin((n >> 8) & 255) + 0.0722 * lin(n & 255);
  return lum > 0.45 ? "222 47% 11%" : "0 0% 100%";
}

/** CSS custom properties for the brand overrides (empty values are skipped). */
export function themeVars(primary?: string, accent?: string, radiusPx?: string): Record<string, string> {
  const vars: Record<string, string> = {};
  const p = primary ? hexToHsl(primary) : null;
  if (p && primary) {
    vars["--primary"] = hslToken(p);
    vars["--primary-foreground"] = readableForeground(primary);
    vars["--ring"] = hslToken(p);
  }
  const a = accent ? hexToHsl(accent) : null;
  if (a) vars["--ring"] = hslToken(a);
  const r = Number(radiusPx);
  if (radiusPx && Number.isFinite(r) && r >= 0 && r <= 32) vars["--radius"] = `${r / 16}rem`;
  return vars;
}
