"use client";

import { motion } from "framer-motion";
import { IconRefresh, IconAlertTriangle } from "@tabler/icons-react";
import { Button } from "@/components/ui/button";
import { SPRING } from "@/components/shared/motion";

type ErrorStateProps = {
  title?: string;
  description?: string;
  onRetry?: () => void;
  retryLabel?: string;
};

export function ErrorState({
  title = "خطا در دریافت اطلاعات",
  description = "مشکلی در دریافت داده‌ها پیش آمد. لطفاً دوباره تلاش کنید.",
  onRetry,
  retryLabel = "تلاش مجدد",
}: ErrorStateProps) {
  return (
    <motion.div
      role="alert"
      initial={{ opacity: 0, y: 14, scale: 0.98 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={SPRING}
      className="border-destructive/20 bg-destructive/[0.04] flex flex-col items-center justify-center gap-3 rounded-2xl border px-6 py-12 text-center"
    >
      <span
        aria-hidden="true"
        className="bg-destructive/10 text-destructive ring-destructive/20 flex size-12 items-center justify-center rounded-full ring-1"
      >
        <IconAlertTriangle className="size-6" stroke={1.5} />
      </span>
      <div className="space-y-1">
        <p className="text-sm font-semibold">{title}</p>
        <p className="text-caption max-w-sm">{description}</p>
      </div>
      {onRetry ? (
        <Button variant="outline" size="sm" onClick={onRetry}>
          <IconRefresh className="size-4" />
          {retryLabel}
        </Button>
      ) : null}
    </motion.div>
  );
}
