"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { IconLoader2, IconPlus, IconTrash } from "@tabler/icons-react";
import { bankAccountsApi, transfersApi } from "@/lib/api";
import type { Currency, Transfer } from "@/types/api";
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
import {
  PageToolbar,
  ToolbarSpacer,
} from "@/components/shared/page-toolbar";

export default function TransfersPage() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const transfersQuery = useQuery({
    queryKey: ["transfers"],
    queryFn: () => transfersApi.list({ page: 1, page_size: 50 }),
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["transfers"] });
    queryClient.invalidateQueries({ queryKey: ["bank-accounts"] });
    queryClient.invalidateQueries({ queryKey: ["transactions"] });
  };

  const deleteMutation = useMutation({
    mutationFn: (id: number) => transfersApi.remove(id),
    onSuccess: () => {
      toast.success("انتقال حذف شد؛ مانده هر دو حساب اصلاح شد.");
      setDeleteId(null);
      invalidate();
    },
    onError: (error) => toast.error(error.message),
  });

  const transfers = transfersQuery.data?.items ?? [];

  return (
    <div className="mx-auto w-full max-w-5xl space-y-5">
      <PageToolbar>
        <p className="text-body text-muted-foreground">
          انتقال‌ها در فروش و هزینه محاسبه نمی‌شوند.
        </p>
        <ToolbarSpacer />
        <Button onClick={() => setCreateOpen(true)} className="glow-primary rounded-xl">
          <IconPlus className="size-4" />
          انتقال جدید
        </Button>
      </PageToolbar>

      {transfersQuery.isPending ? <LoadingState /> : null}
      {transfersQuery.isError ? <ErrorState onRetry={() => transfersQuery.refetch()} /> : null}

      {transfersQuery.data ? (
        transfers.length === 0 ? (
          <EmptyState title="انتقالی ثبت نشده است" />
        ) : (
          <div className="glass lift sheen ring-foreground/[0.06] overflow-hidden rounded-2xl">
            <Table className="table-premium">
              <TableHeader>
                <TableRow>
                  <TableHead>تاریخ</TableHead>
                  <TableHead>از</TableHead>
                  <TableHead>به</TableHead>
                  <TableHead className="text-end">مبلغ</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {transfers.map((transfer) => (
                  <TransferRow key={transfer.id} transfer={transfer} onDelete={() => setDeleteId(transfer.id)} />
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="حذف انتقال"
        description="هر دو ردیف دفتر (برداشت و واریز) حذف منطقی می‌شوند."
        confirmLabel="حذف کن"
        destructive
        pending={deleteMutation.isPending}
        onConfirm={() => deleteId !== null && deleteMutation.mutate(deleteId)}
      />

      <CreateTransferDialog open={createOpen} onOpenChange={setCreateOpen} onCreated={invalidate} />
    </div>
  );
}

function TransferRow({ transfer, onDelete }: { transfer: Transfer; onDelete: () => void }) {
  return (
    <TableRow>
      <TableCell><JalaliDate iso={transfer.transferred_at} /></TableCell>
      <TableCell>{transfer.from_account?.name ?? "—"}</TableCell>
      <TableCell>{transfer.to_account?.name ?? "—"}</TableCell>
      <TableCell className="text-end">
        <Money amount={transfer.amount} currency={transfer.currency} />
      </TableCell>
      <TableCell className="text-end">
        <Button variant="ghost" size="icon" aria-label={`حذف انتقال ${transfer.id}`} onClick={onDelete}>
          <IconTrash className="text-muted-foreground size-4 transition-colors duration-300 hover:text-destructive" />
        </Button>
      </TableCell>
    </TableRow>
  );
}

function CreateTransferDialog({
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

  const [fromId, setFromId] = useState("");
  const [toId, setToId] = useState("");
  const [amount, setAmount] = useState("");
  const [transferredAt, setTransferredAt] = useState(() => new Date().toISOString().slice(0, 10));
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => {
      if (!fromId || !toId) throw new Error("هر دو حساب را انتخاب کنید.");
      if (fromId === toId) throw new Error("مبدأ و مقصد نباید یکسان باشند.");

      const from = accounts.find((a) => String(a.id) === fromId);
      const to = accounts.find((a) => String(a.id) === toId);
      if (from && to && from.currency !== to.currency) {
        throw new Error("انتقال بین دو حساب با ارز متفاوت مجاز نیست.");
      }

      return transfersApi.create({
        from_account_id: Number(fromId),
        to_account_id: Number(toId),
        amount: Number(amount.replace(/\D/g, "")),
        currency: from?.currency ?? "RIAL",
        transferred_at: new Date(`${transferredAt}T12:00:00`).toISOString(),
      });
    },
    onSuccess: () => {
      toast.success("انتقال ثبت شد.");
      reset();
      onOpenChange(false);
      onCreated();
    },
    onError: (err) => setError(err.message),
  });

  function reset() {
    setFromId(""); setToId(""); setAmount(""); setTransferredAt(new Date().toISOString().slice(0, 10)); setError(null);
  }

  return (
    <Dialog open={open} onOpenChange={(next) => { onOpenChange(next); if (!next) reset(); }}>
      <DialogTrigger asChild><span /></DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>انتقال بین حساب‌ها</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label>از حساب</Label>
              <UiSelect value={fromId || undefined} onValueChange={setFromId}>
                <SelectTrigger><SelectValue placeholder="انتخاب" /></SelectTrigger>
                <SelectContent>
                  {accounts.map((a) => (
                    <SelectItem key={a.id} value={String(a.id)}>
                      {a.name} ({a.currency === "USD" ? "$" : "ت"})
                    </SelectItem>
                  ))}
                </SelectContent>
              </UiSelect>
            </div>
            <div className="space-y-1.5">
              <Label>به حساب</Label>
              <UiSelect value={toId || undefined} onValueChange={setToId}>
                <SelectTrigger><SelectValue placeholder="انتخاب" /></SelectTrigger>
                <SelectContent>
                  {accounts.filter((a) => String(a.id) !== fromId).map((a) => (
                    <SelectItem key={a.id} value={String(a.id)}>
                      {a.name} ({a.currency === "USD" ? "$" : "ت"})
                    </SelectItem>
                  ))}
                </SelectContent>
              </UiSelect>
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="tr-amount">مبلغ</Label>
            <Input id="tr-amount" dir="ltr" inputMode="numeric" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="0" />
            <CurrencyHint accounts={accounts} fromId={fromId} />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="tr-date">تاریخ انتقال</Label>
            <Input
              id="tr-date"
              type="date"
              dir="ltr"
              value={transferredAt}
              max={new Date().toISOString().slice(0, 10)}
              onChange={(e) => setTransferredAt(e.target.value)}
            />
          </div>

          {error ? <p role="alert" className="text-destructive text-sm">{error}</p> : null}
        </div>
        <DialogFooter>
          <Button onClick={() => mutation.mutate()} disabled={mutation.isPending}>
            {mutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
            ثبت انتقال
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function CurrencyHint({
  accounts,
  fromId,
}: {
  accounts: Array<{ id: number; name: string; currency: Currency }>;
  fromId: string;
}) {
  const from = accounts.find((a) => String(a.id) === fromId);
  if (!from) return null;
  return (
    <p className="text-caption">ارز انتقال بر اساس حساب مبدأ تعیین می‌شود: {from.currency === "USD" ? "دلار" : "ریال"}</p>
  );
}
