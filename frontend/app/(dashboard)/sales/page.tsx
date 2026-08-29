"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { IconLoader2, IconPlus, IconTrash } from "@tabler/icons-react";
import { bankAccountsApi, salesApi } from "@/lib/api";
import type { Currency, Gateway, SaleListItem } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/shared/empty-state";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";
import { Money } from "@/components/shared/money";
import { JalaliDate } from "@/components/shared/jalali-date";
import { ConfirmDialog } from "@/components/shared/dialogs";
import { EntityNotes } from "@/components/shared/entity-notes";
import {
  PageToolbar,
  ToolbarSpacer,
  ToolbarStat,
} from "@/components/shared/page-toolbar";

const GATEWAY_LABELS: Record<Gateway, string> = {
  ZARINPAL: "زرین‌پال",
  CARD_TO_CARD: "کارت به کارت",
  SUPPORT: "پشتیبانی",
};

const STATUS_LABELS: Record<string, string> = {
  PAID: "تسویه‌شده",
  PARTIAL: "جزئی",
  UNPAID: "پرداخت‌نشده",
};

type PaymentDraft = { bank_account_id: string; gateway: Gateway; amount: string };

export default function SalesPage() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const salesQuery = useQuery({
    queryKey: ["sales"],
    queryFn: () => salesApi.list({ page: 1, page_size: 50 }),
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["sales"] });
    queryClient.invalidateQueries({ queryKey: ["bank-accounts"] });
    queryClient.invalidateQueries({ queryKey: ["transactions"] });
  };

  const deleteMutation = useMutation({
    mutationFn: (id: number) => salesApi.remove(id),
    onSuccess: () => {
      toast.success("فروش حذف شد و پرداخت‌های آن از مانده حساب‌ها کسر شد.");
      setDeleteId(null);
      invalidate();
    },
    onError: (error) => toast.error(error.message),
  });

  const sales = useMemo(() => salesQuery.data?.items ?? [], [salesQuery.data]);
  const totals = useMemo(() => {
    const init: Record<string, { total: number; paid: number }> = {
      RIAL: { total: 0, paid: 0 },
      USD: { total: 0, paid: 0 },
    };
    for (const sale of sales) {
      if (!init[sale.currency]) init[sale.currency] = { total: 0, paid: 0 };
      init[sale.currency].total += sale.total_amount;
      init[sale.currency].paid += sale.paid_amount;
    }
    return init;
  }, [sales]);

  return (
    <div className="mx-auto w-full max-w-6xl space-y-5">
      <PageToolbar>
        <ToolbarStat
          label="فروش ریال"
          value={totals.RIAL?.paid.toLocaleString("fa-IR") ?? "۰"}
          color="var(--income)"
          mono
        />
        <ToolbarStat label="دلار" value={`$${totals.USD?.paid.toLocaleString("fa-IR") ?? "۰"}`} color="var(--info)" mono />
        <ToolbarSpacer />
        <Button onClick={() => setCreateOpen(true)} className="glow-primary rounded-xl">
          <IconPlus className="size-4" />
          ثبت فروش
        </Button>
      </PageToolbar>

      {salesQuery.isPending ? <LoadingState /> : null}
      {salesQuery.isError ? <ErrorState onRetry={() => salesQuery.refetch()} /> : null}

      {salesQuery.data ? (
        sales.length === 0 ? (
          <EmptyState title="هنوز فروشی ثبت نشده است" />
        ) : (
          <div className="glass lift sheen ring-foreground/[0.06] overflow-hidden rounded-2xl">
            <Table className="table-premium">
              <TableHeader>
                <TableRow>
                  <TableHead>تاریخ</TableHead>
                  <TableHead>خریدار</TableHead>
                  <TableHead className="text-end">مبلغ کل</TableHead>
                  <TableHead className="text-end">دریافت‌شده</TableHead>
                  <TableHead>وضعیت</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {sales.map((sale) => (
                  <SaleRow key={sale.id} sale={sale} onDelete={() => setDeleteId(sale.id)} />
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="حذف فروش"
        description="همه پرداخت‌های این فروش نیز از دفتر حذف منطقی می‌شوند و مانده حساب‌ها اصلاح می‌شود."
        confirmLabel="حذف کن"
        destructive
        pending={deleteMutation.isPending}
        onConfirm={() => deleteId !== null && deleteMutation.mutate(deleteId)}
      />

      <CreateSaleDialog open={createOpen} onOpenChange={setCreateOpen} onCreated={invalidate} />
    </div>
  );
}

function SaleRow({ sale, onDelete }: { sale: SaleListItem; onDelete: () => void }) {
  return (
    <TableRow>
      <TableCell><JalaliDate iso={sale.sold_at} /></TableCell>
      <TableCell>{sale.customer_name ?? "—"}</TableCell>
      <TableCell className="text-end">
        <Money amount={sale.total_amount} currency={sale.currency} />
      </TableCell>
      <TableCell className="text-end">
        <Money amount={sale.paid_amount} currency={sale.currency} />
      </TableCell>
      <TableCell>
        <Badge variant={sale.status === "PAID" ? "secondary" : sale.status === "PARTIAL" ? "outline" : "destructive"}>
          {STATUS_LABELS[sale.status]}
        </Badge>
      </TableCell>
      <TableCell className="text-end">
        <div className="flex items-center justify-end gap-0.5">
          <EntityNotes entityType="SALE" entityId={sale.id} title={`فروش #${sale.id}`} />
          <Button variant="ghost" size="icon" aria-label={`حذف فروش ${sale.id}`} onClick={onDelete}>
            <IconTrash className="text-muted-foreground size-4 transition-colors duration-300 hover:text-destructive" />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  );
}

function CreateSaleDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}) {
  const accountsQuery = useQuery({
    queryKey: ["bank-accounts"],
    queryFn: () => bankAccountsApi.list(false),
  });
  const accounts = accountsQuery.data ?? [];

  const [currency, setCurrency] = useState<Currency>("RIAL");
  const [totalAmount, setTotalAmount] = useState("");
  const [customerName, setCustomerName] = useState("");
  const [soldAt, setSoldAt] = useState(() => new Date().toISOString().slice(0, 10));
  const [payments, setPayments] = useState<PaymentDraft[]>([
    { bank_account_id: "", gateway: "ZARINPAL", amount: "" },
  ]);
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => {
      const numericTotal = Number(totalAmount.replace(/\D/g, ""));
      if (!numericTotal || numericTotal <= 0) throw new Error("مبلغ کل نامعتبر است.");

      const payloadPayments = payments
        .filter((p) => p.bank_account_id)
        .map((p) => ({
          bank_account_id: Number(p.bank_account_id),
          gateway: p.gateway,
          amount: Number(p.amount.replace(/\D/g, "")) || 0,
          paid_at: new Date(`${soldAt}T12:00:00`).toISOString(),
        }));

      if (payloadPayments.some((p) => p.amount <= 0)) {
        throw new Error("مبلغ هر پرداخت باید بزرگ‌تر از صفر باشد.");
      }

      return salesApi.create({
        total_amount: numericTotal,
        currency,
        sold_at: new Date(`${soldAt}T12:00:00`).toISOString(),
        customer_name: customerName.trim() || undefined,
        payments: payloadPayments,
      });
    },
    onSuccess: () => {
      toast.success("فروش ثبت شد.");
      reset();
      onOpenChange(false);
      onCreated();
    },
    onError: (err) => setError(err.message),
  });

  function reset() {
    setCurrency("RIAL"); setTotalAmount(""); setCustomerName("");
    setSoldAt(new Date().toISOString().slice(0, 10));
    setPayments([{ bank_account_id: "", gateway: "ZARINPAL", amount: "" }]);
    setError(null);
  }

  function updatePayment(index: number, patch: Partial<PaymentDraft>) {
    setPayments((prev) => prev.map((p, i) => (i === index ? { ...p, ...patch } : p)));
  }

  const eligibleAccounts = accounts.filter((a) => a.currency === currency);

  return (
    <Dialog open={open} onOpenChange={(next) => { onOpenChange(next); if (!next) reset(); }}>
      <DialogTrigger asChild><span /></DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>ثبت فروش جدید</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="sale-total">مبلغ کل</Label>
              <Input id="sale-total" dir="ltr" inputMode="numeric" value={totalAmount} onChange={(e) => setTotalAmount(e.target.value)} placeholder="0" />
            </div>
            <div className="space-y-1.5">
              <Label>ارز</Label>
              <UiSelect value={currency} onValueChange={(v) => { setCurrency(v as Currency); setPayments([{ bank_account_id: "", gateway: "ZARINPAL", amount: "" }]); }}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="RIAL">ریال</SelectItem>
                  <SelectItem value="USD">دلار</SelectItem>
                </SelectContent>
              </UiSelect>
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="sale-customer">خریدار (اختیاری)</Label>
            <Input id="sale-customer" value={customerName} onChange={(e) => setCustomerName(e.target.value)} />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="sale-date">تاریخ فروش</Label>
            <Input
              id="sale-date"
              type="date"
              dir="ltr"
              value={soldAt}
              max={new Date().toISOString().slice(0, 10)}
              onChange={(e) => setSoldAt(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label>پرداخت‌ها</Label>
            {payments.map((payment, index) => (
              <div key={index} className="grid grid-cols-[1fr_1fr_1fr_auto] items-center gap-2">
                <UiSelect value={payment.bank_account_id || undefined} onValueChange={(v) => updatePayment(index, { bank_account_id: v })}>
                  <SelectTrigger><SelectValue placeholder="حساب" /></SelectTrigger>
                  <SelectContent>
                    {eligibleAccounts.map((a) => (
                      <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>
                    ))}
                  </SelectContent>
                </UiSelect>
                <UiSelect value={payment.gateway} onValueChange={(v) => updatePayment(index, { gateway: v as Gateway })}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {(Object.keys(GATEWAY_LABELS) as Gateway[]).map((g) => (
                      <SelectItem key={g} value={g}>{GATEWAY_LABELS[g]}</SelectItem>
                    ))}
                  </SelectContent>
                </UiSelect>
                <Input dir="ltr" inputMode="numeric" value={payment.amount} onChange={(e) => updatePayment(index, { amount: e.target.value })} placeholder="مبلغ" />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  disabled={payments.length === 1}
                  aria-label={`حذف پرداخت ${index + 1}`}
                  onClick={() => setPayments((prev) => prev.filter((_, i) => i !== index))}
                >
                  <IconTrash className="size-4" />
                </Button>
              </div>
            ))}
            <Button type="button" variant="outline" size="sm" onClick={() => setPayments((p) => [...p, { bank_account_id: "", gateway: "CARD_TO_CARD", amount: "" }])}>
              <IconPlus className="size-4" />
              افزودن پرداخت
            </Button>
          </div>

          {error ? <p role="alert" className="text-destructive text-sm">{error}</p> : null}
        </div>
        <DialogFooter>
          <Button onClick={() => mutation.mutate()} disabled={mutation.isPending}>
            {mutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
            ذخیره فروش
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
