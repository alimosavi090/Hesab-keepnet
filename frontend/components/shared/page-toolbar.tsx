"use client";

import type { ReactNode } from "react";
import { motion } from "framer-motion";
import { FadeIn, SPRING } from "@/components/shared/motion";

/* Frosted toolbar shared by list pages */
export function PageToolbar({ children }: { children: ReactNode }) {
  return (
    <FadeIn className="glass sheen ring-foreground/[0.06] flex flex-wrap items-center gap-2 rounded-2xl px-3 py-2.5">
      {children}
    </FadeIn>
  );
}

/* Small glass metric chip with colored dot */
export function ToolbarStat({
  label,
  value,
  color,
  mono = false,
}: {
  label: string;
  value: ReactNode;
  color?: string;
  mono?: boolean;
}) {
  return (
    <motion.span
      whileHover={{ scale: 1.03 }}
      transition={SPRING}
      className="bg-card/60 border-border/60 ring-foreground/[0.05] inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium ring-1"
    >
      {color ? (
        <span
          aria-hidden="true"
          className="size-1.5 rounded-full"
          style={{ background: color }}
        />
      ) : null}
      <span className="text-muted-foreground">{label}:</span>
      <span className={`font-semibold ${mono ? "numeric" : ""}`}>{value}</span>
    </motion.span>
  );
}

export function ToolbarSpacer() {
  return <div className="flex-1" />;
}
