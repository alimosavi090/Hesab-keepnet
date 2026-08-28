"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip as ChartTooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  IconBriefcase,
  IconShoppingCart,
  IconTrendingDown,
  IconTrendingUp,
  IconUser,
} from "@tabler/icons-react";
import Link from "next/link";
import { dashboardApi } from "@/lib/api";
import type { Currency } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Money } from "@/components/shared/money";
import { JalaliDate } from "@/components/shared/jalali-date";
import { CountUp, FadeIn, SPRING, Stagger, StaggerItem } from "@/components/shared/motion";
import { formatNumber } from "@/utils/format";

const GATEWAY_LABELS: Record<string, string> = {
  ZARINPAL: "زرین‌پال",
  CARD_TO_CARD: "کارت به کارت",
  SUPPORT: "پشتیبانی",
};

export default function DashboardPage() {
  const [currency, setCurrency] = useState<Currency>("RIAL");
  const [range] = useState(() => ({
    from: new Date(Date.now() - 30 * 86_400_000).toISOString(),
    to: new Date().toISOString(),
  }));

  const summary = useQuery({
    queryKey: ["dashboard-summary", currency],
    queryFn: () => dashboardApi.summary(range),
  });

  // Same window, one period earlier — powers the «نسبت به دوره قبل» chips.
  const prevSummary = useQuery({
    queryKey: ["dashboard-summary-prev", currency],
    queryFn: () =>
      dashboardApi.summary({
        from: new Date(Date.now() - 60 * 86_400_000).toISOString(),
        to: new Date(Date.now() - 30 * 86_400_000).toISOString(),
      }),
  });

  const chart = useQuery({
    queryKey: ["dashboard-chart", currency],
    queryFn: () => dashboardApi.chart(30, currency),
  });

  if (summary.isPending || chart.isPending) {
    return (
      <div className="mx-auto w-full max-w-6xl space-y-6">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {[0, 1, 2, 3].map((i) => (
            <Card key={i} className="glass">
              <CardContent className="pt-6">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="mt-3 h-8 w-32" />
              </CardContent>
            </Card>
          ))}
        </div>
        <Skeleton className="h-72 w-full rounded-xl" />
      </div>
    );
  }

  // Period-over-period percentage; null when there is nothing to compare.
  const pctChange = (current: number, previous?: number): number | null => {
    if (previous === undefined || previous <= 0 || current === previous) return null;
    return Math.round(((current - previous) / Math.abs(previous)) * 100);
  };

  const prevProfitRow =
    prevSummary.data?.profit.find((row) => row.currency === currency);
  const prevExpenseRow = (type: "business" | "personal") =>
    (prevSummary.data?.expenses[type] ?? []).find((row) => row.currency === currency)?.total ?? 0;

  if (summary.isError) {
    return (
      <p role="alert" className="text-destructive p-6 text-center">
        خطا در دریافت داده‌های داشبورد. اتصال بک‌اند را بررسی کنید.
      </p>
    );
  }

  const data = summary.data;
  const profitRow =
    data?.profit.find((row) => row.currency === currency) ??
    ({ currency, sales: 0, business_expense: 0, net_profit: 0 } as const);

  const gatewayTotals = (data?.sales_by_gateway ?? []).filter(
    (g) => g.currency === currency
  );
  const expenseRow = (type: "business" | "personal") =>
    (data?.expenses[type] ?? []).find((row) => row.currency === currency)?.total ?? 0;

  const repDebts = (data?.rep_debts ?? []).filter(
    (d) => d.currency === currency
  );

  const bankRows = (data?.banks ?? []).filter((b) => b.currency === currency);
  const bankTotal = bankRows.reduce((acc, b) => acc + b.balance, 0);

  return (
    <div className="mx-auto w-full max-w-6xl space-y-6">
      <div className="flex items-center gap-2">
        <CurrencyToggle value={currency} onChange={setCurrency} />
      </div>

      <Stagger
        step={0.08}
        aria-label="شاخص‌های مالی"
        className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4"
      >
        <StatItem
          title={`فروش ${currency === "USD" ? "(دلار)" : ""}`}
          value={profitRow.sales}
          currency={currency}
          tone="income"
          icon={IconShoppingCart}
          delta={pctChange(profitRow.sales, prevProfitRow?.sales)}
        />
        <StatItem
          title="هزینه کسب‌وکار"
          value={profitRow.business_expense}
          currency={currency}
          tone="business"
          icon={IconBriefcase}
          invertDelta
          delta={pctChange(profitRow.business_expense, prevProfitRow?.business_expense)}
        />
        <StatItem
          title="هزینه شخصی"
          value={expenseRow("personal")}
          currency={currency}
          tone="personal"
          icon={IconUser}
          invertDelta
          delta={pctChange(expenseRow("personal"), prevExpenseRow("personal"))}
        />
        <StatItem
          title="سود خالص"
          value={profitRow.net_profit}
          currency={currency}
          tone={profitRow.net_profit >= 0 ? "income" : "destructive"}
          icon={profitRow.net_profit >= 0 ? IconTrendingUp : IconTrendingDown}
          delta={pctChange(profitRow.net_profit, prevProfitRow?.net_profit)}
        />
      </Stagger>

      <FadeIn>
        <Card className="glass lift sheen overflow-hidden">
          <CardHeader className="pb-2">
            <CardTitle className="text-section-title">روند ۳۰ روزه فروش و هزینه کسب‌وکار</CardTitle>
          </CardHeader>
          <CardContent>
            <div dir="ltr" className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chart.data ?? []} margin={{ top: 8, right: 8, left: 8, bottom: 0 }}>
                  <defs>
                    <linearGradient id="salesFill" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="var(--income)" stopOpacity={0.4} />
                      <stop offset="100%" stopColor="var(--income)" stopOpacity={0.02} />
                    </linearGradient>
                    <linearGradient id="expenseFill" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="var(--expense-business)" stopOpacity={0.35} />
                      <stop offset="100%" stopColor="var(--expense-business)" stopOpacity={0.02} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                  <XAxis
                    dataKey="date"
                    tick={{ fontSize: 10 }}
                    stroke="var(--muted-foreground)"
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    tick={{ fontSize: 10 }}
                    stroke="var(--muted-foreground)"
                    width={60}
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={(v: number) => formatNumber(v)}
                  />
                  <ChartTooltip
                    contentStyle={{
                      background: "var(--popover)",
                      border: "1px solid var(--border)",
                      borderRadius: 14,
                      boxShadow: "0 12px 32px -12px rgb(0 0 0 / 45%)",
                      direction: "rtl",
                      fontFamily: "inherit",
                      fontSize: 12,
                    }}
                    cursor={{ stroke: "var(--primary)", strokeOpacity: 0.35, strokeWidth: 1.5 }}
                    formatter={(value: unknown) => formatNumber(Number(value))}
                  />
                  <Area type="monotone" dataKey="sales" name="فروش" stroke="var(--income)" fill="url(#salesFill)" strokeWidth={2.5} activeDot={{ r: 5, strokeWidth: 2 }} animationDuration={900} animationEasing="ease-out" />
                  <Area type="monotone" dataKey="business_expense" name="هزینه کسب‌وکار" stroke="var(--expense-business)" fill="url(#expenseFill)" strokeWidth={2.5} activeDot={{ r: 5, strokeWidth: 2 }} animationDuration={1100} animationEasing="ease-out" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </FadeIn>

      <section className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <FadeIn className="lg:col-span-2">
          <Card className="glass lift h-full">
            <CardHeader className="pb-2">
              <CardTitle className="text-section-title">آخرین تراکنش‌ها</CardTitle>
            </CardHeader>
            <CardContent>
              {(data?.recent ?? []).length === 0 ? (
                <p className="text-caption py-8 text-center">هنوز تراکنشی ثبت نشده است.</p>
              ) : (
                <ul className="divide-y">
                  {(data?.recent ?? []).map((item) => (
                    <li
                      key={item.id}
                      className="group flex items-center justify-between gap-3 rounded-lg px-1 py-2.5 transition-colors duration-200 hover:bg-accent/40"
                    >
                      <div className="min-w-0">
                        <p className="truncate text-sm transition-colors group-hover:text-primary">{item.description ?? item.account_name}</p>
                        <p className="text-caption"><JalaliDate iso={item.occurred_at} /> · {item.account_name}</p>
                      </div>
                      <Money
                        amount={item.amount}
                        currency={item.currency}
                        sign={item.type === "EXPENSE" || item.type === "TRANSFER_OUT" ? "negative" : "positive"}
                        className="shrink-0 text-sm font-semibold"
                      />
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </FadeIn>

        <div className="space-y-4">
          {repDebts.length > 0 ? (
            <FadeIn delay={0}>
              <Card className="glass lift border-transparent ring-warning/25">
                <CardHeader className="pb-2">
                  <CardTitle className="text-section-title flex items-center justify-between">
                    بدهی نماینده‌ها
                    <Badge variant="outline" className="border-warning/40 text-warning">{formatNumber(repDebts.length)} نفر</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="divide-y">
                    {repDebts.slice(0, 5).map((debt) => (
                      <li key={debt.representative_id} className="flex items-center justify-between gap-2 py-2 text-sm">
                        <Link
                          href={`/representatives/${debt.representative_id}`}
                          className="truncate transition-colors duration-200 hover:text-primary"
                        >
                          {debt.full_name}
                        </Link>
                        <span dir="ltr" className="numeric font-semibold text-expense-business">
                          {formatNumber(debt.debt)} {debt.currency === "USD" ? "$" : ""}
                        </span>
                      </li>
                    ))}
                    {repDebts.length > 5 ? (
                      <li className="text-caption pt-2">
                        + {formatNumber(repDebts.length - 5)} نماینده دیگر
                      </li>
                    ) : null}
                  </ul>
                </CardContent>
              </Card>
            </FadeIn>
          ) : null}

          <FadeIn delay={0.05}>
            <Card className="glass lift">
              <CardHeader className="pb-2">
                <CardTitle className="text-section-title">حساب‌های بانکی</CardTitle>
              </CardHeader>
              <CardContent>
                {bankRows.length === 0 ? (
                  <p className="text-caption py-4 text-center">حسابی ثبت نشده است.</p>
                ) : (
                  <>
                    <ul className="space-y-2.5">
                      {bankRows.map((bank) => (
                        <li key={bank.account_id} className="flex items-center justify-between gap-2 text-sm">
                          <span className="truncate">{bank.name}</span>
                          <span dir="ltr" className="numeric font-semibold">
                            {formatNumber(bank.balance)} {bank.currency === "USD" ? "$" : ""}
                          </span>
                        </li>
                      ))}
                    </ul>
                    <div className="border-t mt-3 pt-3 flex items-center justify-between">
                      <span className="text-caption">جمع کل ({bankRows.length} حساب)</span>
                      <span dir="ltr" className="numeric text-base font-extrabold text-income">
                        <CountUp
                          value={bankTotal}
                          render={(n) =>
                            currency === "USD"
                              ? `$${formatNumber(Math.round(n))}`
                              : formatNumber(Math.round(n))
                          }
                        />
                      </span>
                    </div>
                  </>
                )}
              </CardContent>
            </Card>
          </FadeIn>

          <FadeIn delay={0.1}>
            <Card className="glass lift">
              <CardHeader className="pb-2">
                <CardTitle className="text-section-title">یادآوری‌های نزدیک</CardTitle>
              </CardHeader>
              <CardContent>
                {(data?.reminders ?? []).length === 0 ? (
                  <p className="text-caption py-4 text-center">یادآور نزدیکی وجود ندارد.</p>
                ) : (
                  <ul className="space-y-2">
                    {(data?.reminders ?? []).slice(0, 5).map((reminder) => (
                      <li key={reminder.id} className="flex items-center justify-between gap-2 text-sm">
                        <span className="truncate">{reminder.title}</span>
                        <JalaliDate iso={reminder.due_date} />
                      </li>
                    ))}
                  </ul>
                )}
              </CardContent>
            </Card>
          </FadeIn>

          <FadeIn delay={0.15}>
            <Card className="glass lift">
              <CardHeader className="pb-2">
                <CardTitle className="text-section-title">فروش بر اساس درگاه</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-wrap gap-2">
                {gatewayTotals.length === 0 ? (
                  <p className="text-caption">داده‌ای نیست.</p>
                ) : (
                  gatewayTotals.map((g) => (
                    <Badge key={`${g.gateway}-${g.currency}`} variant="outline" className="gap-1 px-2.5 py-1 transition-transform duration-300 hover:-translate-y-0.5">
                      {GATEWAY_LABELS[g.gateway] ?? g.gateway}:{" "}
                      <span className="numeric">{formatNumber(g.total)}</span>
                    </Badge>
                  ))
                )}
              </CardContent>
            </Card>
          </FadeIn>
        </div>
      </section>
    </div>
  );
}

function CurrencyToggle({ value, onChange }: { value: Currency; onChange: (c: Currency) => void }) {
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
            className={`relative rounded-full px-5 py-1.5 text-sm transition-colors duration-300 ${
              active ? "font-bold text-primary-foreground" : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {active ? (
              <motion.span
                layoutId="currency-thumb"
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

const TONE_STYLES: Record<string, { text: string; color: string; soft: string }> = {
  income: { text: "text-income", color: "var(--income)", soft: "color-mix(in oklch, var(--income) 16%, transparent)" },
  business: { text: "text-expense-business", color: "var(--expense-business)", soft: "color-mix(in oklch, var(--expense-business) 16%, transparent)" },
  personal: { text: "text-expense-personal", color: "var(--expense-personal)", soft: "color-mix(in oklch, var(--expense-personal) 16%, transparent)" },
  destructive: { text: "text-destructive", color: "var(--destructive)", soft: "color-mix(in oklch, var(--destructive) 16%, transparent)" },
};

function StatItem({
  title,
  value,
  currency,
  tone,
  icon: Icon,
  delta,
  invertDelta = false,
}: {
  title: string;
  value: number;
  currency: Currency;
  tone: keyof typeof TONE_STYLES;
  icon: typeof IconShoppingCart;
  delta?: number | null;
  invertDelta?: boolean;
}) {
  const t = TONE_STYLES[tone];
  const good = delta === null || delta === undefined ? false : invertDelta ? delta <= 0 : delta >= 0;
  return (
    <StaggerItem>
      <Card className="glass lift sheen group relative border-transparent ring-foreground/[0.06]">
        <span
          aria-hidden="true"
          className="pointer-events-none absolute -top-10 left-1/2 size-32 -translate-x-1/2 rounded-full opacity-25 blur-3xl transition-opacity duration-500 group-hover:opacity-45"
          style={{ background: t.color }}
        />
        <CardHeader className="relative pb-1">
          <div className="flex items-start justify-between gap-2">
            <CardTitle className="text-xs font-medium text-muted-foreground md:text-sm">{title}</CardTitle>
            <span
              aria-hidden="true"
              className="flex size-8 shrink-0 items-center justify-center rounded-lg transition-transform duration-300 group-hover:scale-110"
              style={{ background: t.soft, color: t.color }}
            >
              <Icon stroke={1.7} className="size-[18px]" />
            </span>
          </div>
        </CardHeader>
        <CardContent className="relative pt-1">
          <span dir="ltr" className={`numeric text-2xl font-extrabold ${t.text}`}>
            <CountUp
              value={value}
              render={(n) =>
                currency === "USD"
                  ? `$${formatNumber(Math.round(n))}`
                  : formatNumber(Math.round(n))
              }
            />
          </span>
          {delta !== null && delta !== undefined ? (
            <motion.span
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.35, duration: 0.35 }}
              className={`mt-1 inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold ${
                good ? "bg-income/15 text-income" : "bg-destructive/15 text-destructive"
              }`}
            >
              {delta >= 0 ? "▲" : "▼"}
              {formatNumber(Math.abs(delta))}٪ نسبت به دوره قبل
            </motion.span>
          ) : (
            <span className="text-caption block pt-0.5">{currency === "USD" ? "دلار آمریکا" : "ریال ایران"}</span>
          )}
        </CardContent>
      </Card>
    </StaggerItem>
  );
}
