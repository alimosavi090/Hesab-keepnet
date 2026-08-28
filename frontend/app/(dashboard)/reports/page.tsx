"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import {
  IconBriefcase,
  IconDownload,
  IconShoppingCart,
  IconTrendingUp,
  IconUser,
  IconUsersGroup,
} from "@tabler/icons-react";
import Link from "next/link";
import { reportsApi } from "@/lib/api";
import type { Currency, Gateway } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";
import { CountUp, FadeIn, SPRING, Stagger, StaggerItem } from "@/components/shared/motion";
import { PageToolbar, ToolbarSpacer } from "@/components/shared/page-toolbar";
import { formatNumber } from "@/utils/format";

const GATEWAY_LABELS: Record<Gateway, string> = {
  ZARINPAL: "زرین‌پال",
  CARD_TO_CARD: "کارت به کارت",
  SUPPORT: "پشتیبانی",
};

const RANGES = [
  { key: "7d", label: "هفته اخیر", days: 7 },
  { key: "30d", label: "ماه اخیر", days: 30 },
  { key: "90d", label: "سه ماه اخیر", days: 90 },
  { key: "custom", label: "بازه دلخواه / یک روز خاص", days: 0 },
] as const;

type RangeKey = (typeof RANGES)[number]["key"];

export default function ReportsPage() {
  const [rangeKey, setRangeKey] = useState<RangeKey>("30d");
  const [currency, setCurrency] = useState<Currency>("RIAL");
  const [downloading, setDownloading] = useState(false);

  const [customFrom, setCustomFrom] = useState("");
  const [customTo, setCustomTo] = useState("");

  // Preset ranges are derived; the custom window activates once both dates
  // are picked (same date in both fields = an exact one-day report).
  const rangeQuery = useMemo<{ from: string; to: string } | null>(() => {
    if (rangeKey === "custom") {
      if (!customFrom || !customTo || customFrom > customTo) return null;
      return {
        from: new Date(`${customFrom}T00:00:00`).toISOString(),
        to: new Date(`${customTo}T23:59:59`).toISOString(),
      };
    }
    const days = RANGES.find((r) => r.key === rangeKey)!.days;
    const to = new Date();
    const from = new Date(to.getTime() - days * 86_400_000);
    return {
      from: from.toISOString().slice(0, 10),
      to: to.toISOString().slice(0, 10),
    };
  }, [rangeKey, customFrom, customTo]);

  const overviewQuery = useQuery({
    queryKey: ["reports-overview", rangeKey, rangeQuery?.from ?? "", rangeQuery?.to ?? ""],
    queryFn: () => reportsApi.summary(rangeQuery!),
    enabled: rangeQuery !== null,
  });

  async function handleExport() {
    if (!rangeQuery) return;
    setDownloading(true);
    try {
      await reportsApi.download("expenses", rangeQuery);
    } catch (error) {
      alert(error instanceof Error ? error.message : "دانلود ناموفق بود.");
    } finally {
      setDownloading(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl space-y-5">
      <PageToolbar>
        <UiSelect value={rangeKey} onValueChange={(v) => setRangeKey(v as RangeKey)}>
          <SelectTrigger className="w-52"><SelectValue /></SelectTrigger>
          <SelectContent>
            {RANGES.map((r) => (
              <SelectItem key={r.key} value={r.key}>{r.label}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>

        {rangeKey === "custom" ? (
          <motion.div
            initial={{ opacity: 0, x: -12 }}
            animate={{ opacity: 1, x: 0 }}
            transition={SPRING}
            className="flex items-center gap-2"
          >
            <Input
              type="date"
              dir="ltr"
              aria-label="از تاریخ"
              value={customFrom}
              max={customTo || undefined}
              onChange={(e) => setCustomFrom(e.target.value)}
              className="w-36 bg-transparent"
            />
            <span className="text-caption">تا</span>
            <Input
              type="date"
              dir="ltr"
              aria-label="تا تاریخ"
              value={customTo}
              min={customFrom || undefined}
              onChange={(e) => setCustomTo(e.target.value)}
              className="w-36 bg-transparent"
            />
          </motion.div>
        ) : null}

        <ToolbarSpacer />
        <Button variant="outline" className="rounded-xl" onClick={handleExport} disabled={downloading}>
          <IconDownload className="size-4" />
          خروجی CSV
        </Button>
      </PageToolbar>

      {overviewQuery.isPending ? <LoadingState label="در حال محاسبه گزارش…" /> : null}
      {overviewQuery.isError ? <ErrorState onRetry={() => overviewQuery.refetch()} /> : null}

      {overviewQuery.data ? (
        <Overview data={overviewQuery.data} currency={currency} onCurrencyChange={setCurrency} />
      ) : null}
    </div>
  );
}

function CurrencyPill({ value, onChange }: { value: Currency; onChange: (c: Currency) => void }) {
  return (
    <div className="bg-muted/70 ring-foreground/5 relative inline-flex rounded-full p-1 ring-1">
      {(["RIAL", "USD"] as Currency[]).map((c) => {
        const active = value === c;
        return (
          <button
            key={c}
            type="button"
            onClick={() => onChange(c)}
            aria-pressed={active}
            className={`relative rounded-full px-4 py-1.5 text-sm transition-colors duration-300 ${
              active ? "font-bold text-primary-foreground" : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {active ? (
              <motion.span
                layoutId="report-currency-thumb"
                transition={SPRING}
                aria-hidden="true"
                className="glow-primary bg-primary absolute inset-0 -z-10 rounded-full"
              />
            ) : null}
            {c === "RIAL" ? "ریال" : "دلار"}
          </button>
        );
      })}
    </div>
  );
}

/* ─── Overview body ─────────────────────────────────────────────── */

function Overview({
  data,
  currency,
  onCurrencyChange,
}: {
  data: Awaited<ReturnType<typeof reportsApi.summary>>;
  currency: Currency;
  onCurrencyChange: (c: Currency) => void;
}) {
  const profit = (data.profit ?? []).find((r) => r.currency === currency);
  const settlement =
    (data.rep_settlements ?? []).find((r) => r.currency === currency)?.total ?? 0;
  const personalTotal =
    (data.expense_split?.personal ?? []).find((r) => r.currency === currency)?.total ?? 0;

  const categories = useMemo(
    () =>
      (data.expenses_by_category ?? [])
        .filter((c) => c.currency === currency)
        .sort((a, b) => b.total - a.total),
    [data.expenses_by_category, currency]
  );

  const gateways = useMemo(
    () =>
      (data.gateways ?? [])
        .filter((g) => g.currency === currency)
        .sort((a, b) => b.total - a.total),
    [data.gateways, currency]
  );

  const debts = useMemo(
    () => (data.rep_debts ?? []).filter((d) => d.currency === currency).sort((a, b) => b.debt - a.debt),
    [data.rep_debts, currency]
  );

  const maxCategory = Math.max(1, ...categories.map((c) => c.total));
  const incomeSources = [
    ...gateways.map((g) => ({ name: GATEWAY_LABELS[g.gateway], total: g.total })),
    ...(settlement > 0
      ? [{ name: "تسویه نماینده‌ها", total: settlement }]
      : []),
  ];
  const maxIncomeSource = Math.max(1, ...incomeSources.map((s) => s.total));
  const incomeSum = incomeSources.reduce((acc, s) => acc + s.total, 0);

  return (
    <>
      <FadeIn>
        <div className="flex items-center justify-between gap-3">
          <CurrencyPill value={currency} onChange={onCurrencyChange} />
        </div>
      </FadeIn>

      <Stagger step={0.07} className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-5" aria-label="خلاصه مالی">
        <SummaryCard title="فروش" icon={IconShoppingCart} tone="income" value={profit?.sales ?? 0} />
        <SummaryCard title="هزینه کسب‌وکار" icon={IconBriefcase} tone="business" value={profit?.business_expense ?? 0} />
        <SummaryCard title="هزینه شخصی" icon={IconUser} tone="personal" value={personalTotal} note="خارج از سود و زیان" />
        <SummaryCard title="تسویه نماینده‌ها" icon={IconUsersGroup} tone="info" value={settlement} note="بخشی از درآمد" />
        <SummaryCard
          title="سود خالص"
          icon={IconTrendingUp}
          tone={profit && profit.net_profit >= 0 ? "income" : "destructive"}
          value={profit?.net_profit ?? 0}
        />
      </Stagger>

      <section className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <CategoryBreakdown
          title="هزینه‌ها به تفکیک دسته"
          rows={categories.map((c) => ({
            key: `${c.category_name}-${c.category_type}`,
            label: c.category_name,
            badge:
              c.category_type === "BUSINESS" ? (
                <Badge variant="outline" className="border-expense-business/40 text-expense-business">کسب‌وکار</Badge>
              ) : (
                <Badge variant="outline" className="border-expense-personal/40 text-expense-personal">شخصی</Badge>
              ),
            color: c.category_type === "BUSINESS" ? "var(--expense-business)" : "var(--expense-personal)",
            total: c.total,
          }))}
          max={maxCategory}
          emptyText="در این بازه هزینه‌ای ثبت نشده است."
        />

        <div className="space-y-4">
          <IncomeBreakdown rows={incomeSources} max={maxIncomeSource} sum={incomeSum} settlementCount={
            (data.rep_settlements ?? []).find((r) => r.currency === currency)?.count ?? 0
          } />

          <DebtsCard debts={debts} />
        </div>
      </section>
    </>
  );
}

/* ─── Summary cards ─────────────────────────────────────────────── */

const TONES: Record<string, { text: string; color: string }> = {
  income: { text: "text-income", color: "var(--income)" },
  business: { text: "text-expense-business", color: "var(--expense-business)" },
  personal: { text: "text-expense-personal", color: "var(--expense-personal)" },
  info: { text: "text-info", color: "var(--info)" },
  destructive: { text: "text-destructive", color: "var(--destructive)" },
};

function SummaryCard({
  title,
  value,
  tone,
  icon: Icon,
  note,
}: {
  title: string;
  value: number;
  tone: keyof typeof TONES;
  icon: typeof IconShoppingCart;
  note?: string;
}) {
  const t = TONES[tone];
  return (
    <StaggerItem>
      <Card className="glass lift sheen group relative border-transparent ring-foreground/[0.06]">
        <span
          aria-hidden="true"
          className="pointer-events-none absolute -top-9 left-1/2 size-28 -translate-x-1/2 rounded-full opacity-25 blur-3xl transition-opacity duration-500 group-hover:opacity-45"
          style={{ background: t.color }}
        />
        <CardContent className="relative space-y-1 pt-1">
          <div className="flex items-center justify-between">
            <p className="text-xs font-medium text-muted-foreground">{title}</p>
            <span
              aria-hidden="true"
              className="flex size-8 items-center justify-center rounded-lg transition-transform duration-300 group-hover:scale-110"
              style={{ background: `color-mix(in oklch, ${t.color} 16%, transparent)`, color: t.color }}
            >
              <Icon stroke={1.7} className="size-[18px]" />
            </span>
          </div>
          <span dir="ltr" className={`numeric block text-xl font-extrabold ${t.text}`}>
            <CountUp value={value} render={(n) => formatNumber(Math.round(n))} />
          </span>
          {note ? <p className="text-caption">{note}</p> : <p className="text-caption invisible">—</p>}
        </CardContent>
      </Card>
    </StaggerItem>
  );
}

/* ─── Breakdown helpers ─────────────────────────────────────────── */

type BreakdownRow = {
  key: string;
  label: string;
  badge?: React.ReactNode;
  color: string;
  total: number;
};

function BreakdownList({ rows, max, emptyText }: { rows: Array<BreakdownRow>; max: number; emptyText: string }) {
  if (rows.length === 0) {
    return <p className="text-caption py-6 text-center">{emptyText}</p>;
  }
  return (
    <ul className="space-y-3">
      {rows.slice(0, 8).map((row, i) => (
        <motion.li
          key={row.key}
          initial={{ opacity: 0, x: 16 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true, margin: "-24px" }}
          transition={{ delay: i * 0.05, ease: [0.22, 1, 0.36, 1], duration: 0.4 }}
        >
          <div className="flex items-center justify-between gap-2 text-sm">
            <span className="flex min-w-0 items-center gap-2">
              <span className="truncate font-medium">{row.label}</span>
              {row.badge}
            </span>
            <span dir="ltr" className="numeric shrink-0 font-semibold">
              {formatNumber(row.total)}
            </span>
          </div>
          <div className="bg-foreground/[0.06] mt-1.5 h-1.5 overflow-hidden rounded-full">
            <motion.div
              initial={{ width: 0 }}
              whileInView={{ width: `${Math.max(4, (row.total / max) * 100)}%` }}
              viewport={{ once: true, margin: "-24px" }}
              transition={{ delay: i * 0.05 + 0.15, duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
              className="h-full rounded-full"
              style={{
                background: `linear-gradient(90deg, color-mix(in oklch, ${row.color} 55%, transparent), ${row.color})`,
              }}
            />
          </div>
        </motion.li>
      ))}
    </ul>
  );
}

function CategoryBreakdown({
  title,
  rows,
  max,
  emptyText,
}: {
  title: string;
  rows: Array<BreakdownRow>;
  max: number;
  emptyText: string;
}) {
  return (
    <FadeIn className="lg:col-span-1">
      <Card className="glass lift sheen h-full border-transparent ring-foreground/[0.06]">
        <CardHeader className="pb-3">
          <CardTitle className="text-section-title">{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <BreakdownList rows={rows} max={max} emptyText={emptyText} />
        </CardContent>
      </Card>
    </FadeIn>
  );
}

function IncomeBreakdown({
  rows,
  max,
  sum,
  settlementCount,
}: {
  rows: Array<{ name: string; total: number }>;
  max: number;
  sum: number;
  settlementCount: number;
}) {
  return (
    <FadeIn>
      <Card className="glass lift sheen border-transparent ring-foreground/[0.06]">
        <CardHeader className="pb-3">
          <CardTitle className="text-section-title flex items-center justify-between">
            درآمد به تفکیک منبع
            {settlementCount > 0 ? (
              <Badge variant="outline" className="border-info/40 text-info">
                {formatNumber(settlementCount)} تسویه نماینده
              </Badge>
            ) : null}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <BreakdownList
            rows={rows.map((r) => ({
              key: r.name,
              label: r.name,
              color: r.name === "تسویه نماینده‌ها" ? "var(--info)" : "var(--income)",
              total: r.total,
            }))}
            max={max}
            emptyText="در این بازه درآمدی ثبت نشده است."
          />
          {rows.length > 0 ? (
            <p className="text-caption mt-4 flex items-center justify-between border-t pt-3">
              <span>جمع منابع درآمد</span>
              <span dir="ltr" className="numeric font-bold text-income">{formatNumber(sum)}</span>
            </p>
          ) : null}
        </CardContent>
      </Card>
    </FadeIn>
  );
}

function DebtsCard({
  debts,
}: {
  debts: Array<{ representative_id: number; full_name: string; debt: number; currency: Currency }>;
}) {
  return (
    <FadeIn>
      <Card className="glass lift sheen border-transparent ring-warning/25">
        <CardHeader className="pb-2">
          <CardTitle className="text-section-title flex items-center justify-between">
            بدهی نماینده‌ها (مانده باز)
            <Badge variant="outline" className="border-warning/40 text-warning">
              {formatNumber(debts.length)} نفر
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {debts.length === 0 ? (
            <p className="text-caption py-4 text-center">بدهی بازی وجود ندارد؛ همه دفاتر تسویه است.</p>
          ) : (
            <ul className="divide-y">
              {debts.slice(0, 6).map((debt) => (
                <li key={debt.representative_id} className="flex items-center justify-between gap-2 py-2 text-sm">
                  <Link
                    href={`/representatives/${debt.representative_id}`}
                    className="truncate transition-colors duration-200 hover:text-primary"
                  >
                    {debt.full_name}
                  </Link>
                  <span dir="ltr" className="numeric font-semibold text-expense-business">
                    {formatNumber(debt.debt)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </FadeIn>
  );
}
