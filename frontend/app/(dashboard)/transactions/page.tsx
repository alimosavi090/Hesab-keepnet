"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { transactionsApi, bankAccountsApi } from "@/lib/api";
import type { LedgerType } from "@/types/api";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { EmptyState } from "@/components/shared/empty-state";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";
import { Money } from "@/components/shared/money";
import { JalaliDate } from "@/components/shared/jalali-date";
import {
  PageToolbar,
  ToolbarSpacer,
} from "@/components/shared/page-toolbar";

const TYPE_LABELS: Record<LedgerType, string> = {
  INCOME: "دریافت",
  EXPENSE: "پرداخت",
  TRANSFER_IN: "انتقال ورودی",
  TRANSFER_OUT: "انتقال خروجی",
};

const PAGE_SIZE = 20;

export default function TransactionsPage() {
  const [page, setPage] = useState(1);
  const [accountFilter, setAccountFilter] = useState("all");
  const [typeFilter, setTypeFilter] = useState<"all" | LedgerType>("all");

  const accountsQuery = useQuery({
    queryKey: ["bank-accounts"],
    queryFn: () => bankAccountsApi.list(true),
  });

  const feedQuery = useQuery({
    queryKey: ["transactions", page, accountFilter, typeFilter],
    queryFn: () =>
      transactionsApi.feed({
        page,
        page_size: PAGE_SIZE,
        ...(accountFilter !== "all" ? { bank_account_id: Number(accountFilter) } : {}),
        ...(typeFilter !== "all" ? { type: typeFilter } : {}),
      }),
  });

  const meta = feedQuery.data?.meta;
  const totalPages = meta ? Math.max(1, Math.ceil(meta.total / meta.page_size)) : 1;

  return (
    <div className="mx-auto w-full max-w-6xl space-y-5">
      <PageToolbar>
        <UiSelect value={typeFilter} onValueChange={(v) => { setTypeFilter(v as typeof typeFilter); setPage(1); }}>
          <SelectTrigger className="w-40"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">همه انواع</SelectItem>
            {(Object.keys(TYPE_LABELS) as LedgerType[]).map((t) => (
              <SelectItem key={t} value={t}>{TYPE_LABELS[t]}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>

        <UiSelect value={accountFilter} onValueChange={(v) => { setAccountFilter(v); setPage(1); }}>
          <SelectTrigger className="w-44"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">همه حساب‌ها</SelectItem>
            {(accountsQuery.data ?? []).map((a) => (
              <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>

        <ToolbarSpacer />
        {meta ? (
          <span className="text-caption numeric bg-card/60 border-border/60 rounded-full px-3 py-1.5 ring-foreground/[0.05] ring-1">
            {meta.total.toLocaleString("fa-IR")} ردیف — صفحه {meta.page.toLocaleString("fa-IR")} از {totalPages.toLocaleString("fa-IR")}
          </span>
        ) : null}
        <div className="flex gap-2">
          <Button variant="outline" size="sm" className="rounded-full" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            قبلی
          </Button>
          <Button variant="outline" size="sm" className="rounded-full" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            بعدی
          </Button>
        </div>
      </PageToolbar>

      {feedQuery.isPending ? <LoadingState /> : null}
      {feedQuery.isError ? <ErrorState onRetry={() => feedQuery.refetch()} /> : null}

      {feedQuery.data ? (
        feedQuery.data.items.length === 0 ? (
          <EmptyState title="تراکنشی یافت نشد" description="با ثبت فروش یا هزینه، دفتر تراکنش‌ها اینجا نمایش داده می‌شود." />
        ) : (
          <div className="glass lift sheen ring-foreground/[0.06] overflow-hidden rounded-2xl">
            <Table className="table-premium">
              <TableHeader>
                <TableRow>
                  <TableHead>تاریخ</TableHead>
                  <TableHead>حساب</TableHead>
                  <TableHead>نوع</TableHead>
                  <TableHead>شرح</TableHead>
                  <TableHead className="text-end">مبلغ</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {feedQuery.data.items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell><JalaliDate iso={item.occurred_at} /></TableCell>
                    <TableCell>{item.account_name}</TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={
                          item.type === "INCOME"
                            ? "border-income/40 text-income"
                            : item.type === "EXPENSE"
                              ? "border-expense-business/40 text-expense-business"
                              : "border-info/40 text-info"
                        }
                      >
                        {TYPE_LABELS[item.type]}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-64 truncate">{item.description ?? "—"}</TableCell>
                    <TableCell className="text-end">
                      <Money
                        amount={item.amount}
                        currency={item.currency}
                        sign={item.type === "EXPENSE" || item.type === "TRANSFER_OUT" ? "negative" : "positive"}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}
    </div>
  );
}
