"use client";

import Link from "next/link";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { IconChevronLeft, IconLoader2, IconPlus, IconUsersGroup } from "@tabler/icons-react";
import { representativesApi } from "@/lib/api";
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
import {
  PageToolbar,
  ToolbarSpacer,
} from "@/components/shared/page-toolbar";
import type { Currency } from "@/types/api";

export default function RepresentativesPage() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);

  const query = useQuery({
    queryKey: ["representatives"],
    queryFn: () => representativesApi.list(true),
  });

  const reps = query.data ?? [];

  return (
    <div className="mx-auto w-full max-w-5xl space-y-5">
      <PageToolbar>
        <p className="text-body text-muted-foreground">مدیریت نمایندگان فروش و دفتر بدهی آن‌ها</p>
        <ToolbarSpacer />
        <Button onClick={() => setCreateOpen(true)} className="glow-primary rounded-xl">
          <IconPlus className="size-4" />
          نماینده جدید
        </Button>
      </PageToolbar>

      {query.isPending ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => query.refetch()} /> : null}

      {query.data ? (
        reps.length === 0 ? (
          <EmptyState icon={IconUsersGroup} title="نماینده‌ای ثبت نشده است" />
        ) : (
          <div className="glass lift sheen ring-foreground/[0.06] overflow-hidden rounded-2xl">
            <Table className="table-premium">
              <TableHeader>
                <TableRow>
                  <TableHead>نام</TableHead>
                  <TableHead>تماس</TableHead>
                  <TableHead>ارز</TableHead>
                  <TableHead>وضعیت</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {reps.map((rep) => (
                  <TableRow key={rep.id}>
                    <TableCell className="font-medium">{rep.full_name}</TableCell>
                    <TableCell dir="ltr" className="numeric">{rep.phone}</TableCell>
                    <TableCell>{rep.currency === "USD" ? "دلار" : "ریال"}</TableCell>
                    <TableCell>
                      <Badge variant={rep.is_active ? "secondary" : "outline"}>
                        {rep.is_active ? "فعال" : "غیرفعال"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-end">
                      <Button variant="ghost" size="sm" asChild>
                        <Link href={`/representatives/${rep.id}`}>
                          دفتر
                          <IconChevronLeft className="size-4" />
                        </Link>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )
      ) : null}

      <CreateRepresentativeDialog open={createOpen} onOpenChange={setCreateOpen} onCreated={() => queryClient.invalidateQueries({ queryKey: ["representatives"] })} />
    </div>
  );
}

function CreateRepresentativeDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: () => void;
}) {
  const [fullName, setFullName] = useState("");
  const [phone, setPhone] = useState("");
  const [currency, setCurrency] = useState<Currency>("RIAL");
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => {
      if (!fullName.trim() || !phone.trim()) throw new Error("نام و شماره تماس الزامی است.");
      return representativesApi.create({
        full_name: fullName.trim(),
        phone: phone.trim(),
        currency,
        start_date: new Date().toISOString(),
      });
    },
    onSuccess: () => {
      toast.success("نماینده ثبت شد.");
      setFullName(""); setPhone(""); setError(null);
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
          <DialogTitle>نماینده جدید</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="rep-name">نام کامل</Label>
            <Input id="rep-name" value={fullName} onChange={(e) => setFullName(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="rep-phone">شماره تماس</Label>
            <Input id="rep-phone" dir="ltr" value={phone} onChange={(e) => setPhone(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label>ارز دفتر</Label>
            <UiSelect value={currency} onValueChange={(v) => setCurrency(v as Currency)}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="RIAL">ریال</SelectItem>
                <SelectItem value="USD">دلار</SelectItem>
              </SelectContent>
            </UiSelect>
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
