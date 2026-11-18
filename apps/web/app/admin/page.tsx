"use client";

import { Flag, ShieldOff, UserPlus, Users } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { BarChart } from "@/components/admin/bar-chart";
import { ModerationShell } from "@/components/moderation-nav";
import { adminApi, type SeriesMetric, type SeriesPeriod, type SeriesPoint, type Totals } from "@/lib/admin-api";
import { cn, formatCount } from "@/lib/utils";

const METRICS: { value: SeriesMetric; label: string }[] = [
  { value: "users", label: "Novos usuários" },
  { value: "topics", label: "Tópicos" },
  { value: "replies", label: "Respostas" },
  { value: "likes", label: "Curtidas" },
  { value: "reports", label: "Denúncias" },
];

const PERIODS: { value: SeriesPeriod; label: string }[] = [
  { value: "daily", label: "30 dias" },
  { value: "weekly", label: "12 semanas" },
  { value: "monthly", label: "12 meses" },
];

const MONTHS = ["jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"];

function formatLabel(period: SeriesPeriod) {
  return (label: string) => {
    if (period === "monthly") {
      const [y, m] = label.split("-");
      return `${MONTHS[Number(m) - 1]}/${y.slice(2)}`;
    }
    const [, m, d] = label.split("-");
    return `${d}/${m}`;
  };
}

export default function DashboardPage() {
  const [totals, setTotals] = useState<Totals | null>(null);
  const [metric, setMetric] = useState<SeriesMetric>("users");
  const [period, setPeriod] = useState<SeriesPeriod>("daily");
  const [points, setPoints] = useState<SeriesPoint[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    adminApi
      .metrics()
      .then(({ data }) => setTotals(data))
      .catch((err) => setError(err instanceof Error ? err.message : "falha ao carregar métricas"));
  }, []);

  useEffect(() => {
    adminApi
      .series(metric, period)
      .then(({ data }) => setPoints(data))
      .catch((err) => setError(err instanceof Error ? err.message : "falha ao carregar série"));
  }, [metric, period]);

  const metricLabel = METRICS.find((m) => m.value === metric)?.label ?? "";
  const periodTotal = points.reduce((s, p) => s + p.value, 0);

  return (
    <ModerationShell title="Dashboard" description="Números reais do banco de dados, atualizados a cada visita." perms={["metrics.view"]}>
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}

      {totals && (
        <>
          <section aria-label="Comunidade" className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <Hero icon={Users} label="Usuários" value={totals.users} sub={`${formatCount(totals.activeUsers)} ativos em 7 dias`} />
            <Hero icon={UserPlus} label="Novos (30 dias)" value={totals.newUsers} sub="cadastros recentes" />
            <Hero
              icon={Flag}
              label="Denúncias pendentes"
              value={totals.reportsPending}
              sub={`${formatCount(totals.reports)} no total`}
              href="/admin/reports"
              alert={totals.reportsPending > 0}
            />
            <Hero
              icon={ShieldOff}
              label="Banidos agora"
              value={totals.banned}
              sub={`${formatCount(totals.muted)} silenciados · ${formatCount(totals.warnings)} advertências`}
              href="/admin/sanctions"
            />
          </section>

          <section
            aria-label="Conteúdo"
            className="grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-border bg-border sm:grid-cols-5"
          >
            {(
              [
                ["Categorias", totals.categories],
                ["Subcategorias", totals.subcategories],
                ["Tópicos", totals.topics],
                ["Respostas", totals.replies],
                ["Curtidas", totals.likes],
              ] as const
            ).map(([label, value]) => (
              <div key={label} className="bg-card px-5 py-4">
                <p className="text-xs text-muted-foreground">{label}</p>
                <p className="mt-1 text-2xl font-semibold tabular-nums">{formatCount(value)}</p>
              </div>
            ))}
          </section>
        </>
      )}

      <section className="rounded-xl border border-border bg-card p-5">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Crescimento — {metricLabel}</h2>
            <p className="text-sm text-muted-foreground">
              <span className="font-semibold tabular-nums text-foreground">{formatCount(periodTotal)}</span> nos últimos{" "}
              {PERIODS.find((p) => p.value === period)?.label}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Segmented value={metric} onChange={setMetric} options={METRICS} label="Métrica" />
            <Segmented value={period} onChange={setPeriod} options={PERIODS} label="Período" />
          </div>
        </div>
        <div className="mt-6">
          <BarChart points={points} name={metricLabel} formatLabel={formatLabel(period)} />
        </div>
      </section>
    </ModerationShell>
  );
}

interface HeroProps {
  icon: typeof Users;
  label: string;
  value: number;
  sub: string;
  href?: string;
  alert?: boolean;
}

function Hero({ icon: Icon, label, value, sub, href, alert }: HeroProps) {
  const body = (
    <>
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-muted-foreground">{label}</p>
        <Icon className={cn("h-4 w-4", alert ? "text-destructive" : "text-muted-foreground")} />
      </div>
      <p className="mt-3 text-4xl font-bold tracking-tight tabular-nums">{formatCount(value)}</p>
      <p className="mt-1 text-xs text-muted-foreground">{sub}</p>
    </>
  );
  const cls = cn(
    "block rounded-xl border bg-card p-5 transition-colors",
    alert ? "border-destructive/40" : "border-border",
    href && "hover:border-foreground/30",
  );
  return href ? (
    <Link href={href} className={cls}>
      {body}
    </Link>
  ) : (
    <div className={cls}>{body}</div>
  );
}

interface SegmentedProps<T extends string> {
  value: T;
  onChange: (v: T) => void;
  options: { value: T; label: string }[];
  label: string;
}

function Segmented<T extends string>({ value, onChange, options, label }: SegmentedProps<T>) {
  return (
    <div role="radiogroup" aria-label={label} className="flex flex-wrap gap-0.5 rounded-lg bg-secondary p-0.5">
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          role="radio"
          aria-checked={value === o.value}
          onClick={() => onChange(o.value)}
          className={cn(
            "rounded-md px-2.5 py-1 text-xs font-medium transition-all",
            value === o.value ? "bg-background shadow-sm" : "text-muted-foreground hover:text-foreground",
          )}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
