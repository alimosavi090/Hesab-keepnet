"use client";

import type { ComponentType, ReactNode } from "react";
import { motion } from "framer-motion";
import { IconInbox } from "@tabler/icons-react";
import { SPRING } from "@/components/shared/motion";

type EmptyStateProps = {
  title: string;
  description?: string;
  icon?: ComponentType<{ className?: string; stroke?: number }>;
  action?: ReactNode;
};

export function EmptyState({
  title,
  description,
  icon: Icon = IconInbox,
  action,
}: EmptyStateProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 14, scale: 0.98 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={SPRING}
      className="border-border/60 from-foreground/[0.02] flex flex-col items-center justify-center gap-2 rounded-2xl border border-dashed bg-gradient-to-b to-transparent px-6 py-14 text-center"
    >
      <span
        aria-hidden="true"
        className="bg-primary/10 text-primary/70 ring-primary/20 mb-1 flex size-12 items-center justify-center rounded-full ring-1"
      >
        <Icon className="size-6" stroke={1.5} />
      </span>
      <p className="text-sm font-semibold">{title}</p>
      {description ? <p className="text-caption max-w-sm">{description}</p> : null}
      {action ? <div className="mt-2">{action}</div> : null}
    </motion.div>
  );
}
