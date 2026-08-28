"use client";

import Link from "next/link";
import { IconHome, IconRefresh } from "@tabler/icons-react";
import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/shared/error-state";

export default function AppError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="bg-background flex min-h-dvh flex-col items-center justify-center gap-6 px-6">
      <ErrorState
        title="مشکلی پیش آمد"
        description="هنگام اجرای برنامه خطایی رخ داد. صفحه را دوباره بارگذاری کنید؛ اگر مشکل ادامه داشت، از طریق لاگ‌ها بررسی شود."
        onRetry={reset}
      />
      <div className="flex items-center gap-3">
        <Button variant="outline" asChild>
          <Link href="/dashboard">
            <IconHome className="size-4" />
            بازگشت به داشبورد
          </Link>
        </Button>
        <Button onClick={reset}>
          <IconRefresh className="size-4" />
          تلاش مجدد
        </Button>
      </div>
      {process.env.NODE_ENV === "development" ? (
        <pre dir="ltr" className="text-caption max-w-lg overflow-auto rounded border p-3">
          {error.message}
        </pre>
      ) : null}
    </main>
  );
}
