"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { motion } from "framer-motion";
import { IconBell, IconCheck, IconLoader2, IconPlus, IconTrash } from "@tabler/icons-react";
import { remindersApi } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { EmptyState } from "@/components/shared/empty-state";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";
import { JalaliDate } from "@/components/shared/jalali-date";
import { ConfirmDialog } from "@/components/shared/dialogs";
import {
  PageToolbar,
  ToolbarSpacer,
} from "@/components/shared/page-toolbar";
import { SPRING } from "@/components/shared/motion";
import type { Reminder } from "@/types/api";

function dueState(dueDate: string, isDone: boolean): "done" | "overdue" | "soon" | "normal" {
  if (isDone) return "done";
  const diff = new Date(dueDate).getTime() - Date.now();
  if (diff < 0) return "overdue";
  if (diff < 7 * 86_400_000) return "soon";
  return "normal";
}

export default function RemindersPage() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const query = useQuery({
    queryKey: ["reminders"],
    queryFn: () => remindersApi.list(),
  });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["reminders"] });

  const doneMutation = useMutation({
    mutationFn: ({ id, is_done }: { id: number; is_done: boolean }) => remindersApi.setDone(id, is_done),
    onSuccess: invalidate,
    onError: (error) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => remindersApi.remove(id),
    onSuccess: () => {
      setDeleteId(null);
      invalidate();
    },
    onError: (error) => toast.error(error.message),
  });

  const reminders = [...(query.data ?? [])].sort(
    (a, b) => Number(a.is_done) - Number(b.is_done) || a.due_date.localeCompare(b.due_date)
  );

  return (
    <div className="mx-auto w-full max-w-3xl space-y-5">
      <PageToolbar>
        <p className="text-body text-muted-foreground">یادآور سررسید سرور، هاست و دامنه‌ها</p>
        <ToolbarSpacer />
        <Button onClick={() => setCreateOpen(true)} className="glow-primary rounded-xl">
          <IconPlus className="size-4" />
          یادآور جدید
        </Button>
      </PageToolbar>

      {query.isPending ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => query.refetch()} /> : null}

      {query.data ? (
        reminders.length === 0 ? (
          <EmptyState icon={IconBell} title="یادآوری ثبت نشده است" />
        ) : (
          <ul className="space-y-2.5">
            {reminders.map((reminder) => (
              <ReminderRow
                key={reminder.id}
                reminder={reminder}
                onToggleDone={() =>
                  doneMutation.mutate({ id: reminder.id, is_done: !reminder.is_done })
                }
                pending={doneMutation.isPending}
                onDelete={() => setDeleteId(reminder.id)}
              />
            ))}
          </ul>
        )
      ) : null}

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="حذف یادآور"
        description="این یادآور حذف می‌شود."
        confirmLabel="حذف کن"
        destructive
        pending={deleteMutation.isPending}
        onConfirm={() => deleteId !== null && deleteMutation.mutate(deleteId)}
      />

      <CreateReminderDialog open={createOpen} onOpenChange={setCreateOpen} onCreated={invalidate} />
    </div>
  );
}

function ReminderRow({
  reminder,
  onToggleDone,
  pending,
  onDelete,
}: {
  reminder: Reminder;
  onToggleDone: () => void;
  pending: boolean;
  onDelete: () => void;
}) {
  const state = dueState(reminder.due_date, reminder.is_done);

  return (
    <motion.li
      whileHover={{ y: -2, scale: 1.005 }}
      transition={SPRING}
      className="glass ring-foreground/[0.06] flex items-center gap-3 rounded-2xl px-4 py-3 shadow-sm"
    >
      <motion.button
        type="button"
        aria-label={reminder.is_done ? "بازگرداندن یادآور" : "انجام شد"}
        onClick={onToggleDone}
        disabled={pending}
        whileTap={{ scale: 0.8 }}
        transition={SPRING}
        className={
          state === "done"
            ? "bg-income flex size-6 items-center justify-center rounded-full text-white shadow-[0_0_12px_-2px_var(--income)]"
            : "border-border hover:border-primary hover:bg-accent size-6 rounded-full border-2 transition-colors duration-300"
        }
      >
        {state === "done" ? (
          <motion.span
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ ...SPRING, delay: 0.05 }}
            className="flex"
          >
            <IconCheck className="size-4" />
          </motion.span>
        ) : null}
      </motion.button>

      <div className="min-w-0 flex-1">
        <p className={`text-sm font-medium ${state === "done" ? "text-muted-foreground line-through" : ""}`}>
          {reminder.title}
        </p>
        <div className="text-caption mt-0.5 flex items-center gap-2">
          <JalaliDate iso={reminder.due_date} />
          {state === "overdue" ? (
            <Badge variant="destructive" className="h-5 animate-pulse px-1.5 text-[10px]">معوقه</Badge>
          ) : state === "soon" ? (
            <Badge variant="outline" className="border-warning/50 text-warning h-5 px-1.5 text-[10px]">نزدیک</Badge>
          ) : null}
        </div>
      </div>

      <Button variant="ghost" size="icon" aria-label={`حذف ${reminder.title}`} onClick={onDelete}>
        <IconTrash className="text-muted-foreground size-4 transition-colors duration-300 hover:text-destructive" />
      </Button>
    </motion.li>
  );
}

function CreateReminderDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}) {
  const [title, setTitle] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [repeat, setRepeat] = useState<"NONE" | "MONTHLY" | "YEARLY">("NONE");
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => {
      if (!title.trim()) throw new Error("عنوان الزامی است.");
      if (!dueDate) throw new Error("تاریخ سررسید را وارد کنید (مثال ۱۴۰۵/۰۶/۰۱ به میلادی).");
      return remindersApi.create({
        title: title.trim(),
        due_date: jalaliInputToISO(dueDate),
        repeat_interval: repeat,
      });
    },
    onSuccess: () => {
      toast.success("یادآور ایجاد شد.");
      setTitle(""); setDueDate(""); setRepeat("NONE"); setError(null);
      onOpenChange(false);
      onCreated();
    },
    onError: (err) => setError(err.message),
  });

  return (
    <Dialog open={open} onOpenChange={(next) => { onOpenChange(next); if (!next) setError(null); }}>
      <DialogTrigger asChild><span /></DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>یادآور جدید</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="rem-title">عنوان</Label>
            <Input id="rem-title" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="مثلاً تمدید سرور ایران" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="rem-date">سررسید (میلادی)</Label>
              <Input id="rem-date" dir="ltr" type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
              <p className="text-caption">در لیست با تاریخ جلالی نمایش داده می‌شود.</p>
            </div>
            <div className="space-y-1.5">
              <Label>تکرار</Label>
              <UiSelect value={repeat} onValueChange={(v) => setRepeat(v as typeof repeat)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="NONE">بدون تکرار</SelectItem>
                  <SelectItem value="MONTHLY">ماهانه</SelectItem>
                  <SelectItem value="YEARLY">سالانه</SelectItem>
                </SelectContent>
              </UiSelect>
            </div>
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

function jalaliInputToISO(dateStr: string): string {
  return new Date(`${dateStr}T12:00:00`).toISOString();
}
