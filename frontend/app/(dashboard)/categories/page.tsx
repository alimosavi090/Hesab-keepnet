"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { IconLoader2, IconPlus } from "@tabler/icons-react";
import { categoriesApi } from "@/lib/api";
import type { CategoryType } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { EmptyState } from "@/components/shared/empty-state";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";
import {
  PageToolbar,
  ToolbarSpacer,
} from "@/components/shared/page-toolbar";

const TYPE_LABELS: Record<CategoryType, string> = {
  BUSINESS: "هزینه‌های کسب‌وکار",
  PERSONAL: "هزینه‌های شخصی",
};

export default function CategoriesPage() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [newType, setNewType] = useState<CategoryType>("BUSINESS");

  const query = useQuery({
    queryKey: ["categories"],
    queryFn: () => categoriesApi.list(),
  });

  const categories = query.data ?? [];

  return (
    <div className="mx-auto w-full max-w-4xl space-y-6">
      <PageToolbar>
        <p className="text-body text-muted-foreground">
          ماهیت هر هزینه (کسب‌وکار/شخصی) از نوع دسته تعیین می‌شود.
        </p>
        <ToolbarSpacer />
        <Button
          onClick={() => {
            setNewType("BUSINESS");
            setCreateOpen(true);
          }}
          className="glow-primary rounded-xl"
        >
          <IconPlus className="size-4" />
          دسته جدید
        </Button>
      </PageToolbar>

      {query.isPending ? <LoadingState /> : null}
      {query.isError ? <ErrorState onRetry={() => query.refetch()} /> : null}

      {query.data ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {(["BUSINESS", "PERSONAL"] as CategoryType[]).map((type) => (
            <CategorySection key={type} type={type} categories={categories.filter((c) => c.type === type)} onAdd={() => { setNewType(type); setCreateOpen(true); }} />
          ))}
        </div>
      ) : null}

      <CreateCategoryDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        initialType={newType}
        categories={categories}
        onCreated={() => queryClient.invalidateQueries({ queryKey: ["categories"] })}
      />
    </div>
  );
}

function CategorySection({
  type,
  categories,
  onAdd,
}: {
  type: CategoryType;
  categories: Array<{ id: number; name: string; parent_id: number | null; is_active: boolean }>;
  onAdd: () => void;
}) {
  const roots = categories.filter((c) => c.parent_id === null);

  return (
    <Card className="glass lift sheen group relative overflow-hidden border-transparent ring-foreground/[0.06]">
      <span
        aria-hidden="true"
        className="pointer-events-none absolute -top-12 left-1/3 size-32 rounded-full opacity-[0.12] blur-3xl"
        style={{ background: type === "BUSINESS" ? "var(--expense-business)" : "var(--expense-personal)" }}
      />
      <CardHeader className="relative flex-row items-center justify-between space-y-0 pb-3">
        <CardTitle className={`text-sm font-semibold ${type === "BUSINESS" ? "text-expense-business" : "text-expense-personal"}`}>
          {TYPE_LABELS[type]}
        </CardTitle>
        <Button variant="ghost" size="sm" onClick={onAdd} className="transition-transform duration-300 hover:scale-110">
          <IconPlus className="size-4" />
        </Button>
      </CardHeader>
      <CardContent className="relative">
        {roots.length === 0 ? (
          <EmptyState title="دسته‌ای نیست" />
        ) : (
          <ul className="space-y-1.5">
            {roots.map((category) => {
              const children = categories.filter((c) => c.parent_id === category.id);
              return (
                <li key={category.id} className="transition-colors duration-200 hover:text-primary">
                  <span className={`text-body inline-block rounded px-1 ${category.is_active ? "" : "text-muted-foreground line-through"}`}>
                    {category.name}
                  </span>
                  {children.length > 0 ? (
                    <ul className="border-s ps-4 mt-1 space-y-1">
                      {children.map((child) => (
                        <li key={child.id} className={`text-caption ${child.is_active ? "" : "line-through"}`}>
                          {child.name}
                        </li>
                      ))}
                    </ul>
                  ) : null}
                </li>
              );
            })}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function CreateCategoryDialog({
  open,
  onOpenChange,
  initialType,
  categories,
  onCreated,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialType: CategoryType;
  categories: Array<{ id: number; name: string; type: CategoryType; is_active: boolean }>;
  onCreated: () => void;
}) {
  const [name, setName] = useState("");
  const [type, setType] = useState<CategoryType>(initialType);
  const [parentId, setParentId] = useState("root");
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => {
      if (!name.trim()) throw new Error("نام دسته الزامی است.");
      return categoriesApi.create({
        name: name.trim(),
        type,
        ...(parentId !== "root" ? { parent_id: Number(parentId) } : {}),
      });
    },
    onSuccess: () => {
      toast.success("دسته ایجاد شد.");
      setName(""); setError(null);
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
          <DialogTitle>دسته‌بندی جدید</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="cat-name">نام</Label>
            <Input id="cat-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label>نوع</Label>
              <UiSelect value={type} onValueChange={(v) => { setType(v as CategoryType); setParentId("root"); }}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="BUSINESS">کسب‌وکار</SelectItem>
                  <SelectItem value="PERSONAL">شخصی</SelectItem>
                </SelectContent>
              </UiSelect>
            </div>
            <div className="space-y-1.5">
              <Label>والد (اختیاری)</Label>
              <UiSelect value={parentId} onValueChange={setParentId}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="root">— ریشه —</SelectItem>
                  {categories.filter((c) => c.type === type && c.is_active).map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>{c.name}</SelectItem>
                  ))}
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
