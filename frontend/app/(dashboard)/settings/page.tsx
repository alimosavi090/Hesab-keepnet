"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  IconCloudUpload,
  IconDownload,
  IconLoader2,
  IconTrash,
} from "@tabler/icons-react";
import { backupsApi } from "@/lib/api";
import { useHealth } from "@/hooks/use-health";
import { ComingSoon } from "@/components/shared/coming-soon";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/shared/dialogs";
import { FadeIn } from "@/components/shared/motion";
import { JalaliDate } from "@/components/shared/jalali-date";
import { cn } from "@/lib/utils";

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}

export default function SettingsPage() {
  const health = useHealth();

  return (
    <div className="mx-auto w-full max-w-3xl space-y-6">
      <ComingSoon
        title="تنظیمات"
        description="مدیریت کاربران، امنیت و تنظیمات سیستم در فازهای بعدی تکمیل می‌شود."
      />

      <BackupsSection />

      <Card className="glass lift sheen border-transparent ring-foreground/[0.06]">
        <CardHeader className="pb-2">
          <CardTitle className="text-section-title">وضعیت سیستم</CardTitle>
        </CardHeader>
        <CardContent>
          {health.isPending ? (
            <Skeleton className="h-10 w-48" />
          ) : health.isError ? (
            <div className="flex items-center gap-2 text-sm text-destructive">
              <span
                aria-hidden="true"
                className="bg-destructive size-2.5 rounded-full"
              />
              بک‌اند در دسترس نیست
              <span className="text-caption numeric">
                ({String(health.error.code ?? "")})
              </span>
            </div>
          ) : health.data ? (
            <dl className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <StatusItem
                label="بک‌اند"
                value={health.data.database === "up" ? "متصل" : "قطع"}
                tone={health.data.database === "up" ? "ok" : "bad"}
              />
              <StatusItem
                label="محیط"
                value={health.data.environment}
                tone="neutral"
              />
              <StatusItem
                label="نسخه"
                value={health.data.version}
                tone="neutral"
              />
            </dl>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}

/* ─── Automatic + manual backups ────────────────────────────────── */

function BackupsSection() {
  const queryClient = useQueryClient();
  const [deleteName, setDeleteName] = useState<string | null>(null);

  const backupsQuery = useQuery({
    queryKey: ["backups"],
    queryFn: () => backupsApi.list(),
    refetchInterval: 60_000,
  });

  const createMutation = useMutation({
    mutationFn: () => backupsApi.create(),
    onSuccess: () => {
      toast.success("پشتیبان جدید ساخته شد.");
      queryClient.invalidateQueries({ queryKey: ["backups"] });
    },
    onError: (error) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: (name: string) => backupsApi.remove(name),
    onSuccess: () => {
      setDeleteName(null);
      toast.success("پشتیبان حذف شد.");
      queryClient.invalidateQueries({ queryKey: ["backups"] });
    },
    onError: (error) => toast.error(error.message),
  });

  const downloadMutation = useMutation({
    mutationFn: (name: string) => backupsApi.download(name),
    onError: (error) => toast.error(error.message),
  });

  const items = backupsQuery.data?.items ?? [];
  const lastAuto = backupsQuery.data?.last_auto_at ?? null;
  const intervalHours = backupsQuery.data?.interval_hours ?? 24;

  return (
    <FadeIn>
      <Card className="glass lift sheen border-transparent ring-foreground/[0.06]">
        <CardHeader className="pb-2">
          <CardTitle className="text-section-title flex items-center justify-between">
            پشتیبان‌گیری دیتابیس
            <Badge variant="outline" className="border-info/40 text-info">
              خودکار هر {intervalHours} ساعت
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-caption leading-6">
            یک نسخه کامل و سازگار از دیتابیس SQLite هر {intervalHours} ساعت به‌صورت خودکار و هنگام هر ری‌استارت سرور گرفته می‌شود.
            {lastAuto ? (
              <>
                {" "}آخرین پشتیبان خودکار:{" "}
                <span className="numeric"><JalaliDate iso={lastAuto} /></span>.
              </>
            ) : null}
          </p>

          <Button
            onClick={() => createMutation.mutate()}
            disabled={createMutation.isPending}
            className="glow-primary rounded-xl"
          >
            {createMutation.isPending ? (
              <IconLoader2 className="size-4 animate-spin" />
            ) : (
              <IconCloudUpload className="size-4" data-testid="download-icon" />
            )}
            پشتیبان فوری
          </Button>

          {backupsQuery.isPending ? (
            <Skeleton className="h-20 w-full rounded-xl" />
          ) : items.length === 0 ? (
            <p className="text-caption py-4 text-center">هنوز پشتیبانی ساخته نشده است.</p>
          ) : (
            <ul className="bg-card/50 ring-foreground/[0.05] max-h-72 space-y-1 overflow-y-auto rounded-xl p-1.5 ring-1">
              {items.map((backup) => (
                <li
                  key={backup.name}
                  dir="ltr"
                  className="group flex items-center justify-between gap-2 rounded-lg px-2 py-2 transition-colors duration-200 hover:bg-accent/40"
                >
                  <span className="flex min-w-0 items-center gap-2 text-xs">
                    <Badge
                      variant="outline"
                      className={cn(
                        "shrink-0 px-1.5",
                        backup.is_auto ? "border-info/40 text-info" : "border-income/40 text-income"
                      )}
                    >
                      {backup.is_auto ? "خودکار" : "دستی"}
                    </Badge>
                    <span className="truncate font-medium">{backup.name}</span>
                  </span>
                  <span className="numeric flex shrink-0 items-center gap-2 text-[11px] text-muted-foreground">
                    {formatBytes(backup.size_bytes)}
                    <JalaliDate iso={backup.created_at} />
                  </span>
                  <span className="flex shrink-0 gap-0.5">
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={`دانلود ${backup.name}`}
                      disabled={downloadMutation.isPending}
                      onClick={() => downloadMutation.mutate(backup.name)}
                    >
                      <IconDownload className="size-4 transition-colors duration-300 hover:text-primary" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={`حذف ${backup.name}`}
                      onClick={() => setDeleteName(backup.name)}
                    >
                      <IconTrash className="text-muted-foreground size-4 transition-colors duration-300 hover:text-destructive" />
                    </Button>
                  </span>
                </li>
              ))}
            </ul>
          )}

          <ConfirmDialog
            open={deleteName !== null}
            onOpenChange={(open) => !open && setDeleteName(null)}
            title="حذف پشتیبان"
            description={`فایل «${deleteName ?? ""}» برای همیشه حذف می‌شود.`}
            confirmLabel="حذف کن"
            destructive
            pending={deleteMutation.isPending}
            onConfirm={() => deleteName !== null && deleteMutation.mutate(deleteName)}
          />
        </CardContent>
      </Card>
    </FadeIn>
  );
}

function StatusItem({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone: "ok" | "bad" | "neutral";
}) {
  return (
    <div className="bg-card/50 ring-foreground/[0.05] rounded-xl px-4 py-3 ring-1 transition-transform duration-300 hover:-translate-y-0.5">
      <dt className="text-label">{label}</dt>
      <dd className="mt-1 flex items-center gap-2 text-sm font-semibold">
        {tone !== "neutral" ? (
          <span
            aria-hidden="true"
            className={cn(
              "size-2.5 rounded-full",
              tone === "ok" ? "animate-pulse bg-income shadow-[0_0_8px_2px_color-mix(in_oklch,var(--income)_60%,transparent)]" : "bg-destructive"
            )}
          />
        ) : null}
        <span className="numeric">{value}</span>
      </dd>
    </div>
  );
}
