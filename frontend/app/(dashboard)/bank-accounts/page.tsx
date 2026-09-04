"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { IconBuildingBank, IconLoader2, IconPencil, IconPlus } from "@tabler/icons-react";
import {
  bankAccountsApi,
} from "@/lib/api";
import type { Currency } from "@/types/api";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { EmptyState } from "@/components/shared/empty-state";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";
import { Money } from "@/components/shared/money";
import { ErrorText } from "@/components/shared/dialogs";
import { EntityNotes } from "@/components/shared/entity-notes";
import { FadeIn } from "@/components/shared/motion";
import {
  PageToolbar,
  ToolbarSpacer,
} from "@/components/shared/page-toolbar";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

type BankAccount = {
  id: number;
  name: string;
  bank_name: string;
  card_number: string | null;
  currency: Currency;
  initial_balance: number;
  description: string | null;
  is_active: boolean;
  balance?: { balance: number; incoming: number; outgoing: number };
};

export default function BankAccountsPage() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<BankAccount | null>(null);

  const query = useQuery({
    queryKey: ["bank-accounts"],
    queryFn: () => bankAccountsApi.list(true),
  });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["bank-accounts"] });

  if (query.isPending) return <LoadingState label="در حال دریافت حساب‌ها…" />;
  if (query.isError) {
    return <ErrorState onRetry={() => query.refetch()} />;
  }

  const accounts = query.data ?? [];

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageToolbar>
        <p className="text-body text-muted-foreground">
          مانده هر حساب از رابطه «موجودی اولیه + ورود − خروج» محاسبه می‌شود.
        </p>
        <ToolbarSpacer />
        <Button onClick={() => setCreateOpen(true)} className="glow-primary rounded-xl">
          <IconPlus className="size-4" />
          حساب جدید
        </Button>
      </PageToolbar>

      {accounts.length === 0 ? (
        <EmptyState
          icon={IconBuildingBank}
          title="هنوز حسابی ثبت نشده است"
          description="برای شروع، یکی از حساب‌های بانکی خود را اضافه کنید."
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {accounts.map((account, index) => (
            <FadeIn key={account.id} delay={index * 0.04}>
              <AccountCard account={account} onChanged={invalidate} onEdit={() => setEditing(account)} />
            </FadeIn>
          ))}
        </div>
      )}

      <CreateAccountDialog open={createOpen} onOpenChange={setCreateOpen} onCreated={invalidate} />
      <EditAccountDialog account={editing} onClose={() => setEditing(null)} onSaved={invalidate} />
    </div>
  );
}

function AccountCard({ account, onChanged, onEdit }: { account: BankAccount; onChanged: () => void; onEdit: () => void }) {
  const [confirmToggle, setConfirmToggle] = useState(false);

  const toggleMutation = useMutation({
    mutationFn: () => bankAccountsApi.setActive(account.id, !account.is_active),
    onSuccess: () => {
      toast.success(account.is_active ? "حساب غیرفعال شد." : "حساب فعال شد.");
      setConfirmToggle(false);
      onChanged();
    },
    onError: (error) => toast.error(error.message),
  });

  const balance = account.balance?.balance ?? account.initial_balance;

  return (
    <Card className="glass lift sheen group relative border-transparent ring-foreground/[0.06]">
      <span
        aria-hidden="true"
        className="bg-primary pointer-events-none absolute -top-10 left-8 size-28 rounded-full opacity-15 blur-3xl transition-opacity duration-500 group-hover:opacity-35"
      />
      <CardContent className="relative space-y-3 pt-5">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2.5">
            <span className="brand-tile flex size-10 items-center justify-center rounded-xl text-primary-foreground transition-transform duration-300 group-hover:scale-105">
              <IconBuildingBank className="size-5" stroke={1.6} />
            </span>
            <div>
              <p className="text-sm font-semibold">{account.name}</p>
              <p className="text-caption">{account.bank_name}</p>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon" aria-label={`ویرایش ${account.name}`} onClick={onEdit}>
              <IconPencil className="text-muted-foreground size-4 transition-colors duration-300 hover:text-primary" />
            </Button>
            <Badge variant={account.is_active ? "secondary" : "outline"}>
              {account.is_active ? "فعال" : "غیرفعال"}
            </Badge>
          </div>
        </div>

        <div className="border-t pt-3">
          <p className="text-label">مانده محاسباتی</p>
          <Money amount={balance} currency={account.currency} className="text-xl font-extrabold" />
        </div>

        <div className="text-caption grid grid-cols-2 gap-2">
          <span dir="ltr" className="numeric">{account.card_number ?? "—"}</span>
          <span className="numeric text-end">
            ورود: {account.balance?.incoming.toLocaleString("fa-IR") ?? "—"}
          </span>
        </div>

        <div className="flex items-center justify-between">
          <EntityNotes entityType="BANK_ACCOUNT" entityId={account.id} title={account.name} />
          <ConfirmDeactivate
            isActive={account.is_active}
            pending={toggleMutation.isPending}
            open={confirmToggle}
            onOpenChange={setConfirmToggle}
            onConfirm={() => toggleMutation.mutate()}
          />
        </div>
      </CardContent>
    </Card>
  );
}

function ConfirmDeactivate({
  isActive,
  pending,
  open,
  onOpenChange,
  onConfirm,
}: {
  isActive: boolean;
  pending: boolean;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}) {
  return (
    <>
      <Button
        variant="outline"
        size="sm"
        onClick={() => (isActive ? onOpenChange(true) : onConfirm())}
        disabled={pending}
      >
        {pending ? <IconLoader2 className="size-4 animate-spin" /> : null}
        {isActive ? "غیرفعال‌سازی" : "فعال‌سازی"}
      </Button>
      {isActive ? (
        <AlertDialogLite
          open={open}
          onOpenChange={onOpenChange}
          onConfirm={onConfirm}
          pending={pending}
        />
      ) : null}
    </>
  );
}


function AlertDialogLite({
  open,
  onOpenChange,
  onConfirm,
  pending,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  pending: boolean;
}) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>غیرفعال‌سازی حساب</AlertDialogTitle>
          <AlertDialogDescription>
            حساب غیرفعال در فرم‌های جدید انتخاب نمی‌شود؛ تاریخچه حفظ می‌شود.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={pending}>انصراف</AlertDialogCancel>
          <AlertDialogAction onClick={(e) => { e.preventDefault(); onConfirm(); }}>
            غیرفعال کن
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function CreateAccountDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}) {
  const [name, setName] = useState("");
  const [bankName, setBankName] = useState("");
  const [cardNumber, setCardNumber] = useState("");
  const [currency, setCurrency] = useState<Currency>("RIAL");
  const [initialBalance, setInitialBalance] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () =>
      bankAccountsApi.create({
        name: name.trim(),
        bank_name: bankName.trim(),
        card_number: cardNumber.replace(/\D/g, "") || undefined,
        currency,
        initial_balance: Number(initialBalance.replace(/[^\d-]/g, "")) || 0,
        description: description.trim() || undefined,
      }),
    onSuccess: () => {
      toast.success("حساب ایجاد شد.");
      reset();
      onOpenChange(false);
      onCreated();
    },
    onError: (err) => setError(err.message),
  });

  function reset() {
    setName(""); setBankName(""); setCardNumber("");
    setCurrency("RIAL"); setInitialBalance(""); setDescription(""); setError(null);
  }

  return (
    <Dialog open={open} onOpenChange={(next) => { onOpenChange(next); if (!next) reset(); }}>
      <DialogTrigger asChild><span /></DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>حساب بانکی جدید</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="acc-name">نام مستعار</Label>
            <Input id="acc-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="مثلاً حساب روزانه" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="acc-bank">نام بانک</Label>
            <Input id="acc-bank" value={bankName} onChange={(e) => setBankName(e.target.value)} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="acc-card">شماره کارت</Label>
              <Input id="acc-card" dir="ltr" inputMode="numeric" value={cardNumber} onChange={(e) => setCardNumber(e.target.value)} placeholder="16 رقم" />
            </div>
            <div className="space-y-1.5">
              <Label>ارز</Label>
              <UiSelect value={currency} onValueChange={(v) => setCurrency(v as Currency)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="RIAL">ریال</SelectItem>
                  <SelectItem value="USD">دلار</SelectItem>
                </SelectContent>
              </UiSelect>
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="acc-initial">موجودی اولیه</Label>
            <Input id="acc-initial" dir="ltr" inputMode="numeric" value={initialBalance} onChange={(e) => setInitialBalance(e.target.value)} placeholder="0" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="acc-desc">توضیحات</Label>
            <Textarea id="acc-desc" rows={2} value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>
          {error ? <ErrorText>{error}</ErrorText> : null}
        </div>
        <DialogFooter>
          <Button onClick={() => mutation.mutate()} disabled={mutation.isPending || !name.trim() || !bankName.trim()}>
            {mutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
            ذخیره
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function EditAccountDialog({
  account,
  onClose,
  onSaved,
}: {
  account: BankAccount | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [name, setName] = useState("");
  const [bankName, setBankName] = useState("");
  const [cardNumber, setCardNumber] = useState("");
  const [initialBalance, setInitialBalance] = useState("");
  const [description, setDescription] = useState("");
  const [initialized, setInitialized] = useState<number | null>(null);

  if (account && initialized !== account.id) {
    setInitialized(account.id);
    setName(account.name);
    setBankName(account.bank_name);
    setCardNumber(account.card_number ?? "");
    setInitialBalance(String(account.initial_balance));
    setDescription(account.description ?? "");
  }

  const mutation = useMutation({
    mutationFn: () =>
      bankAccountsApi.update(account!.id, {
        name: name.trim(),
        bank_name: bankName.trim(),
        card_number: cardNumber.replace(/\D/g, "") || undefined,
        initial_balance: Number(initialBalance.replace(/[^\d-]/g, "")) || 0,
        description: description.trim() || undefined,
      }),
    onSuccess: () => {
      toast.success("حساب ویرایش شد؛ مانده محاسباتی به‌روزرسانی شد.");
      onSaved();
      onClose();
    },
    onError: (err) => toast.error(err.message),
  });

  return (
    <Dialog open={account !== null} onOpenChange={(next) => !next && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>ویرایش حساب بانکی</DialogTitle>
        </DialogHeader>
        {account ? (
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="edit-acc-name">نام مستعار</Label>
              <Input id="edit-acc-name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="edit-acc-bank">نام بانک</Label>
              <Input id="edit-acc-bank" value={bankName} onChange={(e) => setBankName(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="edit-acc-card">شماره کارت</Label>
              <Input id="edit-acc-card" dir="ltr" inputMode="numeric" value={cardNumber} onChange={(e) => setCardNumber(e.target.value)} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="edit-acc-initial">موجودی اولیه</Label>
              <Input
                id="edit-acc-initial"
                dir="ltr"
                inputMode="numeric"
                value={initialBalance}
                onChange={(e) => setInitialBalance(e.target.value)}
              />
              <p className="text-caption">
                مانده محاسباتی = موجودی اولیه + ورود − خروج. تغییر این عدد، مانده فعلی را به همان نسبت جابه‌جا می‌کند.
              </p>
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="edit-acc-desc">توضیحات</Label>
              <Textarea id="edit-acc-desc" rows={2} value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>
            {mutation.isError ? <ErrorText>{(mutation.error as Error).message}</ErrorText> : null}
            <DialogFooter>
              <Button onClick={() => mutation.mutate()} disabled={mutation.isPending || !name.trim() || !bankName.trim()} className="glow-primary rounded-xl">
                {mutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
                ذخیره تغییرات
              </Button>
            </DialogFooter>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
