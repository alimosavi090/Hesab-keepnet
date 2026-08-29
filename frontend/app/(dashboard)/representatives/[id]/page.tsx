"use client";

import { useParams } from "next/navigation";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { IconLoader2, IconPlus, IconTrash } from "@tabler/icons-react";
import { representativesApi, bankAccountsApi } from "@/lib/api";
import type { Currency, RepDirection } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";
import { Money } from "@/components/shared/money";
import { JalaliDate } from "@/components/shared/jalali-date";
import { ConfirmDialog } from "@/components/shared/dialogs";
import { EntityNotes } from "@/components/shared/entity-notes";

export default function RepresentativeDetailPage() {
  const params = useParams<{ id: string }>();
  const representativeId = Number(params.id);
  const queryClient = useQueryClient();
  const [recordOpen, setRecordOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["rep-balance", representativeId] });
    queryClient.invalidateQueries({ queryKey: ["rep-transactions", representativeId] });
  };

  const balanceQuery = useQuery({
    queryKey: ["rep-balance", representativeId],
    queryFn: () => representativesApi.balance(representativeId),
  });

  const transactionsQuery = useQuery({
    queryKey: ["rep-transactions", representativeId],
    queryFn: () => representativesApi.transactions(representativeId, { page: 1, page_size: 100 }),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => representativesApi.removeTransaction(id),
    onSuccess: () => {
      toast.success("ردیف حذف شد.");
      setDeleteId(null);
      invalidate();
    },
    onError: (error) => toast.error(error.message),
  });

  if (balanceQuery.isPending || transactionsQuery.isPending) return <LoadingState />;
  if (balanceQuery.isError) return <ErrorState onRetry={() => balanceQuery.refetch()} />;

  const balance = balanceQuery.data;
  const transactions = transactionsQuery.data?.items ?? [];

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <Card className="glass lift sheen relative overflow-hidden border-transparent ring-foreground/[0.06]">
        <span
          aria-hidden="true"
          className="bg-primary pointer-events-none absolute -top-14 left-1/2 size-36 -translate-x-1/2 rounded-full opacity-[0.12] blur-3xl"
        />
        <CardHeader className="relative flex-row items-center justify-between pb-2 space-y-0">
          <CardTitle className="text-section-title">مانده دفتر نماینده</CardTitle>
          <div className="flex items-center gap-2">
            <EntityNotes
              entityType="REPRESENTATIVE"
              entityId={representativeId}
              title="دفتر نماینده"
            />
            <Dialog open={recordOpen} onOpenChange={setRecordOpen}>
              <DialogTrigger asChild>
                <Button size="sm" className="glow-primary rounded-xl">
                  <IconPlus className="size-4" />
                  ثبت تراکنش
                </Button>
              </DialogTrigger>
              <DialogContent>
                <RecordTransactionBody
                  representativeId={representativeId}
                  currency={balanceQuery.data?.currency}
                  onDone={() => {
                    setRecordOpen(false);
                    invalidate();
                  }}
                />
              </DialogContent>
            </Dialog>
          </div>
        </CardHeader>
        <CardContent className="relative grid grid-cols-3 gap-4">
          <div className="border-border/50 sm:border-s first:border-s-0 ps-0 sm:ps-4">
            <p className="text-label">بدهی (Debit)</p>
            <Money amount={balance.total_debit} currency={balance.currency} className="font-semibold" />
          </div>
          <div className="border-border/50 sm:border-s first:border-s-0 ps-0 sm:ps-4">
            <p className="text-label">تسویه (Credit)</p>
            <Money amount={balance.total_credit} currency={balance.currency} className="font-semibold" />
          </div>
          <div className="border-border/50 sm:border-s first:border-s-0 ps-0 sm:ps-4">
            <p className="text-label">مانده</p>
            <Money
              amount={Math.abs(balance.balance)}
              currency={balance.currency}
              className={`text-base ${balance.balance > 0 ? "text-expense-business font-extrabold" : "text-income font-extrabold"}`}
            />
            <p className="text-caption mt-0.5">
              {balance.balance > 0 ? "نماینده بدهکار است" : "حساب تسویه است"}
            </p>
          </div>
        </CardContent>
      </Card>

      <div className="glass lift sheen ring-foreground/[0.06] overflow-hidden rounded-2xl">
        <Table className="table-premium">
          <TableHeader>
            <TableRow>
              <TableHead>تاریخ</TableHead>
              <TableHead>نوع</TableHead>
              <TableHead>شرح</TableHead>
              <TableHead className="text-end">مبلغ</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {transactions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-caption py-8 text-center">
                  هنوز تراکنشی ثبت نشده است.
                </TableCell>
              </TableRow>
            ) : (
              transactions.map((txn) => (
                <TableRow key={txn.id}>
                  <TableCell><JalaliDate iso={txn.occurred_at} /></TableCell>
                  <TableCell>
                    <Badge variant={txn.direction === "DEBIT" ? "destructive" : "secondary"}>
                      {txn.direction === "DEBIT" ? "بدهی" : "تسویه"}
                    </Badge>
                    {txn.direction === "CREDIT" && txn.bank_account ? (
                      <p className="text-caption mt-1">{txn.bank_account.name}</p>
                    ) : null}
                  </TableCell>
                  <TableCell className="max-w-64 truncate">{txn.description ?? "—"}</TableCell>
                  <TableCell className="text-end">
                    <Money amount={txn.amount} currency={txn.currency} />
                  </TableCell>
                  <TableCell className="text-end">
                    <Button variant="ghost" size="icon" aria-label={`حذف ${txn.id}`} onClick={() => setDeleteId(txn.id)}>
                      <IconTrash className="text-muted-foreground size-4 transition-colors duration-300 hover:text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="حذف ردیف دفتر"
        description="این ردیف به‌صورت منطقی حذف و مانده بازمحاسبه می‌شود."
        confirmLabel="حذف کن"
        destructive
        pending={deleteMutation.isPending}
        onConfirm={() => deleteId !== null && deleteMutation.mutate(deleteId)}
      />
    </div>
  );
}

function RecordTransactionBody({
  representativeId,
  currency,
  onDone,
}: {
  representativeId: number;
  currency?: Currency;
  onDone: () => void;
}) {
  const [direction, setDirection] = useState<RepDirection>("DEBIT");
  const [amount, setAmount] = useState("");
  const [accountId, setAccountId] = useState("");
  const [occurredAt, setOccurredAt] = useState(() => new Date().toISOString().slice(0, 10));
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);

  const accountsQuery = useQuery({
    queryKey: ["bank-accounts"],
    queryFn: () => bankAccountsApi.list(false),
    enabled: direction === "CREDIT",
  });
  const eligibleAccounts = (accountsQuery.data ?? []).filter(
    (a) => a.currency === currency
  );

  function switchDirection(next: RepDirection) {
    setDirection(next);
    setError(null);
    if (next === "DEBIT") setAccountId("");
  }

  const mutation = useMutation({
    mutationFn: () => {
      const numericAmount = Number(amount.replace(/\D/g, ""));
      if (!numericAmount || numericAmount <= 0) throw new Error("مبلغ نامعتبر است.");
      if (direction === "CREDIT" && !accountId) {
        throw new Error("برای تسویه، حساب مقصد واریز را انتخاب کنید.");
      }
      return representativesApi.recordTransaction(representativeId, {
        direction,
        amount: numericAmount,
        occurred_at: new Date(`${occurredAt}T12:00:00`).toISOString(),
        ...(direction === "CREDIT" && accountId
          ? { bank_account_id: Number(accountId) }
          : {}),
        description: description.trim() || undefined,
      });
    },
    onSuccess: () => {
      toast.success(
        direction === "CREDIT"
          ? "تسویه ثبت شد؛ موجودی حساب مقصد به‌روزرسانی و در درآمد ثبت شد."
          : "بدهی نماینده ثبت شد."
      );
      onDone();
    },
    onError: (err) => setError(err.message),
  });

  return (
    <div className="space-y-4">
      <div className="space-y-1.5">
        <Label>نوع</Label>
        <UiSelect value={direction} onValueChange={(v) => switchDirection(v as RepDirection)}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="DEBIT">بدهی (خرید/شارژ) — بدون تأثیر بر موجودی</SelectItem>
            <SelectItem value="CREDIT">تسویه (پرداخت نماینده) — واریز به حساب</SelectItem>
          </SelectContent>
        </UiSelect>
      </div>

      {direction === "CREDIT" ? (
        <div className="space-y-1.5">
          <Label>حساب مقصد واریز</Label>
          {accountsQuery.isPending ? (
            <p className="text-caption">در حال دریافت حساب‌ها…</p>
          ) : eligibleAccounts.length === 0 ? (
            <p role="alert" className="text-warning text-sm">
              هیچ حساب فعال هم‌ارز ({currency === "USD" ? "دلار" : "ریال"}) وجود ندارد؛ ابتدا از بخش «حساب‌ها» اضافه کنید.
            </p>
          ) : (
            <UiSelect value={accountId || undefined} onValueChange={setAccountId}>
              <SelectTrigger><SelectValue placeholder="انتخاب کارت/حساب" /></SelectTrigger>
              <SelectContent>
                {eligibleAccounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>
                ))}
              </SelectContent>
            </UiSelect>
          )}
          <p className="text-caption">
            مبلغ تسویه به موجودی این حساب اضافه می‌شود و جزو درآمد کسب‌وکار محسوب می‌گردد.
          </p>
        </div>
      ) : null}

      <div className="space-y-1.5">
        <Label htmlFor="rtx-amount">مبلغ</Label>
        <Input id="rtx-amount" dir="ltr" inputMode="numeric" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="0" />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="rtx-date">تاریخ</Label>
        <Input
          id="rtx-date"
          type="date"
          dir="ltr"
          value={occurredAt}
          max={new Date().toISOString().slice(0, 10)}
          onChange={(e) => setOccurredAt(e.target.value)}
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="rtx-desc">شرح</Label>
        <Input id="rtx-desc" value={description} onChange={(e) => setDescription(e.target.value)} />
      </div>
      {error ? <p role="alert" className="text-destructive text-sm">{error}</p> : null}
      <div className="flex justify-end">
        <Button onClick={() => mutation.mutate()} disabled={mutation.isPending}>
          {mutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
          ثبت
        </Button>
      </div>
    </div>
  );
}
