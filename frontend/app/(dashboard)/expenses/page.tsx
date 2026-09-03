"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { IconLoader2, IconPencil, IconPlus, IconTrash } from "@tabler/icons-react";
import { categoriesApi, expensesApi, bankAccountsApi } from "@/lib/api";
import type { CategoryType, Currency, Expense } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
import { JalaliDateInput } from "@/components/shared/jalali-date-input";
import {
  PageToolbar,
  ToolbarSpacer,
} from "@/components/shared/page-toolbar";

function computeRange(days: number): { from?: string; to?: string } {
  if (days === 0) return {};
  const to = new Date();
  const from = new Date(to.getTime() - days * 86_400_000);
  return { from: from.toISOString(), to: to.toISOString() };
}

const RANGES = [
  { key: "30d", label: "۳۰ روز اخیر", days: 30 },
  { key: "90d", label: "۹۰ روز اخیر", days: 90 },
  { key: "all", label: "همه", days: 0 },
] as const;

export default function ExpensesPage() {
  const queryClient = useQueryClient();
  const [rangeKey, setRangeKey] = useState<(typeof RANGES)[number]["key"]>("30d");
  const [typeFilter, setTypeFilter] = useState<"ALL" | CategoryType>("ALL");
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<Expense | null>(null);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const [rangeQuery, setRangeQuery] = useState(() => computeRange(30));

  function selectRange(key: (typeof RANGES)[number]["key"]) {
    setRangeKey(key);
    setRangeQuery(computeRange(RANGES.find((r) => r.key === key)!.days));
  }

  const expensesQuery = useQuery({
    queryKey: ["expenses", rangeKey, typeFilter],
    queryFn: () =>
      expensesApi.list({
        page: 1,
        page_size: 100,
        ...rangeQuery,
        ...(typeFilter !== "ALL" ? { type: typeFilter } : {}),
      }),
  });

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: () => categoriesApi.list(),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => expensesApi.remove(id),
    onSuccess: () => {
      toast.success("هزینه حذف شد؛ مانده حساب به‌روزرسانی شد.");
      setDeleteId(null);
      queryClient.invalidateQueries({ queryKey: ["expenses"] });
      queryClient.invalidateQueries({ queryKey: ["bank-accounts"] });
    },
    onError: (error) => toast.error(error.message),
  });

  const filteredExpenses: Expense[] = useMemo(() => {
    const items = expensesQuery.data?.items ?? [];
    if (categoryFilter === "all") return items;
    return items.filter((e) => String(e.category_id) === categoryFilter);
  }, [expensesQuery.data, categoryFilter]);

  return (
    <div className="mx-auto w-full max-w-6xl space-y-5">
      <PageToolbar>
        <UiSelect value={rangeKey} onValueChange={(v) => selectRange(v as typeof rangeKey)}>
          <SelectTrigger className="w-36"><SelectValue /></SelectTrigger>
          <SelectContent>
            {RANGES.map((r) => (
              <SelectItem key={r.key} value={r.key}>{r.label}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>

        <UiSelect value={typeFilter} onValueChange={(v) => setTypeFilter(v as typeof typeFilter)}>
          <SelectTrigger className="w-32"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="ALL">همه انواع</SelectItem>
            <SelectItem value="BUSINESS">کسب‌وکار</SelectItem>
            <SelectItem value="PERSONAL">شخصی</SelectItem>
          </SelectContent>
        </UiSelect>

        <UiSelect value={categoryFilter} onValueChange={setCategoryFilter}>
          <SelectTrigger className="w-44"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">همه دسته‌ها</SelectItem>
            {(categoriesQuery.data ?? []).map((c) => (
              <SelectItem key={c.id} value={String(c.id)}>{c.name}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>

        <ToolbarSpacer />
        <Button onClick={() => setCreateOpen(true)} className="glow-primary rounded-xl">
          <IconPlus className="size-4" />
          ثبت هزینه
        </Button>
      </PageToolbar>

      {expensesQuery.isPending ? <LoadingState /> : null}
      {expensesQuery.isError ? <ErrorState onRetry={() => expensesQuery.refetch()} /> : null}

      {expensesQuery.data ? (
        filteredExpenses.length === 0 ? (
          <EmptyState
            title="هزینه‌ای در این بازه پیدا نشد"
            description="بازه زمانی یا فیلترها را تغییر دهید."
          />
        ) : (
          <div className="glass lift sheen ring-foreground/[0.06] overflow-hidden rounded-2xl">
            <Table className="table-premium">
              <TableHeader>
                <TableRow>
                  <TableHead>تاریخ</TableHead>
                  <TableHead>دسته</TableHead>
                  <TableHead>نوع</TableHead>
                  <TableHead className="text-end">مبلغ</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredExpenses.map((expense) => (
                  <ExpenseRow key={expense.id} expense={expense} onEdit={() => setEditing(expense)} onDelete={() => setDeleteId(expense.id)} />
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="حذف هزینه"
        description="این هزینه به‌صورت منطقی حذف می‌شود و مانده حساب مربوطه بازگردانده می‌شود. ادامه می‌دهید؟"
        confirmLabel="حذف کن"
        destructive
        pending={deleteMutation.isPending}
        onConfirm={() => deleteId !== null && deleteMutation.mutate(deleteId)}
      />

      <EditExpenseDialog
        expense={editing}
        onClose={() => setEditing(null)}
        onSaved={() => {
          queryClient.invalidateQueries({ queryKey: ["expenses"] });
          queryClient.invalidateQueries({ queryKey: ["bank-accounts"] });
        }}
      />

      <CreateExpenseDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        categories={categoriesQuery.data ?? []}
        onCreated={() => {
          queryClient.invalidateQueries({ queryKey: ["expenses"] });
          queryClient.invalidateQueries({ queryKey: ["bank-accounts"] });
        }}
      />
    </div>
  );
}

function CreateExpenseDialog({
  open,
  onOpenChange,
  categories,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  categories: Awaited<ReturnType<typeof categoriesApi.list>>;
  onCreated: () => void;
}) {
  const queryClient = useQueryClient();
  const accountsQuery = useQuery({
    queryKey: ["bank-accounts"],
    queryFn: () => bankAccountsApi.list(false),
  });

  const [categoryId, setCategoryId] = useState("");
  const [accountId, setAccountId] = useState("cash");
  const [amount, setAmount] = useState("");
  const [currency, setCurrency] = useState<Currency>("RIAL");
  const [occurredAt, setOccurredAt] = useState(() => new Date().toISOString().slice(0, 10));
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => {
      const selectedCategory = categories.find((c) => String(c.id) === categoryId);
      if (!selectedCategory) throw new Error("یک دسته انتخاب کنید.");
      if (selectedCategory.type === "BUSINESS" && accountId === "cash") {
        throw new Error("هزینه کسب‌وکار باید به یک حساب بانکی متصل شود.");
      }
      const numericAmount = Number(amount.replace(/\D/g, ""));
      if (!numericAmount || numericAmount <= 0) throw new Error("مبلغ باید بزرگ‌تر از صفر باشد.");

      return expensesApi.create({
        category_id: selectedCategory.id,
        bank_account_id: accountId === "cash" ? undefined : Number(accountId),
        amount: numericAmount,
        currency,
        occurred_at: new Date(`${occurredAt}T12:00:00`).toISOString(),
        description: description.trim() || undefined,
      });
    },
    onSuccess: () => {
      toast.success("هزینه ثبت شد.");
      resetAndClose(onOpenChange);
      onCreated();
      void queryClient;
    },
    onError: (err) => setError(err.message),
  });

  function resetAndClose(close: (open: boolean) => void) {
    setCategoryId(""); setAccountId("cash"); setAmount("");
    setCurrency("RIAL"); setOccurredAt(new Date().toISOString().slice(0, 10)); setDescription(""); setError(null);
    close(false);
  }

  return (
    <Dialog open={open} onOpenChange={(next) => { onOpenChange(next); if (!next) setError(null); }}>
      <DialogTrigger asChild><span /></DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>ثبت هزینه جدید</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label>دسته‌بندی</Label>
            <UiSelect value={categoryId} onValueChange={setCategoryId}>
              <SelectTrigger><SelectValue placeholder="انتخاب دسته" /></SelectTrigger>
              <SelectContent>
                {categories.map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>
                    {c.type === "BUSINESS" ? "کسب‌وکار · " : "شخصی · "}{c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </UiSelect>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label>پرداخت از</Label>
              <UiSelect value={accountId} onValueChange={setAccountId}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="cash">نقدی / بدون حساب</SelectItem>
                  {(accountsQuery.data ?? []).map((a) => (
                    <SelectItem key={a.id} value={String(a.id)}>
                      {a.name} (                      {a.currency === "USD" ? "$" : "ریال"})
                    </SelectItem>
                  ))}
                </SelectContent>
              </UiSelect>
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
            <Label htmlFor="exp-amount">مبلغ</Label>
            <Input id="exp-amount" dir="ltr" inputMode="numeric" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="0" />
          </div>

          <JalaliDateInput id="exp-date" label="تاریخ هزینه" value={occurredAt} onChange={setOccurredAt} />

          <div className="space-y-1.5">
            <Label htmlFor="exp-desc">توضیحات</Label>
            <Input id="exp-desc" value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>

          {error ? <p role="alert" className="text-destructive text-sm">{error}</p> : null}
        </div>
        <DialogFooter>
          <Button onClick={() => mutation.mutate()} disabled={mutation.isPending}>
            {mutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
            ذخیره
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ExpenseRow({
  expense,
  onEdit,
  onDelete,
}: {
  expense: Expense;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <TableRow>
      <TableCell><JalaliDate iso={expense.occurred_at} /></TableCell>
      <TableCell>{expense.category?.name ?? "—"}</TableCell>
      <TableCell>
        {expense.category?.type === "BUSINESS" ? (
          <Badge variant="outline" className="border-expense-business/40 text-expense-business">کسب‌وکار</Badge>
        ) : (
          <Badge variant="outline" className="border-expense-personal/40 text-expense-personal">شخصی</Badge>
        )}
      </TableCell>
      <TableCell className="text-end">
        <Money amount={expense.amount} currency={expense.currency} sign="negative" className="font-semibold" />
      </TableCell>
      <TableCell className="text-end">
        <div className="flex items-center justify-end gap-0.5">
          <Button variant="ghost" size="icon" aria-label={`ویرایش هزینه ${expense.id}`} onClick={onEdit}>
            <IconPencil className="text-muted-foreground size-4 transition-colors duration-300 hover:text-primary" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            aria-label={`حذف هزینه ${expense.id}`}
            onClick={onDelete}
          >
            <IconTrash className="text-muted-foreground size-4 transition-colors duration-300 hover:text-destructive" />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  );
}

function EditExpenseDialog({
  expense,
  onClose,
  onSaved,
}: {
  expense: Expense | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [categoryId, setCategoryId] = useState("");
  const [amount, setAmount] = useState("");
  const [occurredAt, setOccurredAt] = useState(() => new Date().toISOString().slice(0, 10));
  const [description, setDescription] = useState("");
  const [initialized, setInitialized] = useState<number | null>(null);

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: () => categoriesApi.list(),
    enabled: expense !== null,
  });
  const categories = categoriesQuery.data ?? [];

  if (expense && initialized !== expense.id) {
    setInitialized(expense.id);
    setCategoryId(String(expense.category_id));
    setAmount(String(expense.amount));
    setOccurredAt(expense.occurred_at.slice(0, 10));
    setDescription(expense.description ?? "");
  }

  const mutation = useMutation({
    mutationFn: () => {
      const numericAmount = Number(amount.replace(/\D/g, ""));
      if (!numericAmount || numericAmount <= 0) throw new Error("مبلغ نامعتبر است.");
      if (!categoryId) throw new Error("دسته را انتخاب کنید.");
      return expensesApi.update(expense!.id, {
        category_id: Number(categoryId),
        amount: numericAmount,
        occurred_at: new Date(`${occurredAt}T12:00:00`).toISOString(),
        description: description.trim() || undefined,
      });
    },
    onSuccess: () => {
      toast.success("هزینه ویرایش شد؛ مانده حساب همگام شد.");
      onSaved();
      onClose();
    },
    onError: (err) => toast.error(err.message),
  });

  return (
    <Dialog open={expense !== null} onOpenChange={(next) => !next && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>ویرایش هزینه #{expense?.id ?? ""}</DialogTitle>
        </DialogHeader>
        {expense ? (
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label>دسته‌بندی</Label>
              <UiSelect value={categoryId || undefined} onValueChange={setCategoryId}>
                <SelectTrigger><SelectValue placeholder="انتخاب دسته" /></SelectTrigger>
                <SelectContent>
                  {categories.map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>
                      {c.type === "BUSINESS" ? "کسب‌وکار · " : "شخصی · "}{c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </UiSelect>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="edit-exp-amount">مبلغ</Label>
              <Input
                id="edit-exp-amount"
                dir="ltr"
                inputMode="numeric"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
              {expense.bank_account_id ? (
                <p className="text-caption">این هزینه به یک حساب بانکی متصل است؛ مانده حساب با ذخیره همگام می‌شود.</p>
              ) : null}
            </div>

            <JalaliDateInput id="edit-exp-date" label="تاریخ هزینه" value={occurredAt} onChange={setOccurredAt} />

            <div className="space-y-1.5">
              <Label htmlFor="edit-exp-desc">توضیحات</Label>
              <Input id="edit-exp-desc" value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>

            <DialogFooter>
              <Button onClick={() => mutation.mutate()} disabled={mutation.isPending} className="glow-primary rounded-xl">
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
