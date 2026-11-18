"use client";

import { useMemo, useState } from "react";

import type { SeriesPoint } from "@/lib/admin-api";

interface BarChartProps {
  points: SeriesPoint[];
  /** Series name, used in the tooltip and table caption. */
  name: string;
  formatLabel: (label: string) => string;
}

const WIDTH = 720;
const HEIGHT = 220;
const PAD_TOP = 12;
const PAD_BOTTOM = 26;
const PAD_LEFT = 32;
const GAP = 2;

function niceMax(v: number): number {
  if (v <= 4) return 4;
  const pow = 10 ** Math.floor(Math.log10(v));
  const step = [1, 2, 2.5, 5, 10].map((m) => m * pow).find((s) => v <= s * 4) ?? pow * 10;
  return step * 4;
}

/**
 * Single-series bar chart (one hue = primary token). Thin bars with 4px
 * rounded data-ends, 2px gaps, recessive grid, per-bar hover tooltip and a
 * screen-reader table. Width is responsive via viewBox.
 */
export function BarChart({ points, name, formatLabel }: BarChartProps) {
  const [hover, setHover] = useState<number | null>(null);
  const max = useMemo(() => niceMax(Math.max(0, ...points.map((p) => p.value))), [points]);
  const plotW = WIDTH - PAD_LEFT;
  const plotH = HEIGHT - PAD_TOP - PAD_BOTTOM;
  const slot = plotW / Math.max(1, points.length);
  const barW = Math.max(2, Math.min(28, slot - GAP));
  const ticks = [0, 0.25, 0.5, 0.75, 1].map((f) => Math.round(max * f));
  const labelEvery = Math.ceil(points.length / 8);
  const y = (v: number) => PAD_TOP + plotH - (v / max) * plotH;

  return (
    <figure className="relative">
      <svg viewBox={`0 0 ${WIDTH} ${HEIGHT}`} className="h-auto w-full" role="img" aria-label={`${name} por período`}>
        {ticks.map((t) => (
          <g key={t}>
            <line x1={PAD_LEFT} x2={WIDTH} y1={y(t)} y2={y(t)} className="stroke-border" strokeWidth={1} />
            <text
              x={PAD_LEFT - 6}
              y={y(t)}
              textAnchor="end"
              dominantBaseline="middle"
              className="fill-muted-foreground text-[10px] tabular-nums"
            >
              {t}
            </text>
          </g>
        ))}
        {points.map((p, i) => {
          const x = PAD_LEFT + i * slot + (slot - barW) / 2;
          const h = p.value > 0 ? Math.max(2, (p.value / max) * plotH) : 0;
          const r = Math.min(4, barW / 2, h);
          const top = PAD_TOP + plotH - h;
          const bottom = PAD_TOP + plotH;
          const d =
            h === 0
              ? ""
              : `M${x},${bottom} V${top + r} Q${x},${top} ${x + r},${top} H${x + barW - r} Q${x + barW},${top} ${x + barW},${top + r} V${bottom} Z`;
          return (
            <g key={p.label} onMouseEnter={() => setHover(i)} onMouseLeave={() => setHover(null)}>
              {/* Hit target larger than the mark. */}
              <rect x={PAD_LEFT + i * slot} y={PAD_TOP} width={slot} height={plotH} fill="transparent" />
              {d && <path d={d} className={hover === null || hover === i ? "fill-primary" : "fill-primary/40"} />}
              {i % labelEvery === 0 && (
                <text x={x + barW / 2} y={HEIGHT - 8} textAnchor="middle" className="fill-muted-foreground text-[10px]">
                  {formatLabel(p.label)}
                </text>
              )}
            </g>
          );
        })}
      </svg>
      {hover !== null && points[hover] && (
        <div
          className="pointer-events-none absolute top-0 -translate-x-1/2 rounded-md border border-border bg-card px-2.5 py-1.5 text-xs shadow-md"
          style={{ left: `${((PAD_LEFT + hover * slot + slot / 2) / WIDTH) * 100}%` }}
          role="status"
        >
          <p className="text-muted-foreground">{formatLabel(points[hover].label)}</p>
          <p className="font-semibold tabular-nums text-foreground">
            {points[hover].value} {name.toLowerCase()}
          </p>
        </div>
      )}
      <table className="sr-only">
        <caption>{name}</caption>
        <thead>
          <tr>
            <th>Período</th>
            <th>Valor</th>
          </tr>
        </thead>
        <tbody>
          {points.map((p) => (
            <tr key={p.label}>
              <td>{formatLabel(p.label)}</td>
              <td>{p.value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </figure>
  );
}
