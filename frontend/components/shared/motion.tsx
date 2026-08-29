"use client";

import {
  animate,
  motion,
  useInView,
  useMotionValue,
  useReducedMotion,
  useTransform,
  type HTMLMotionProps,
} from "framer-motion";
import { useEffect, useRef, type ReactNode } from "react";
import { usePathname } from "next/navigation";

const EASE = [0.22, 1, 0.36, 1] as const;

export const SPRING = { type: "spring", stiffness: 340, damping: 32, mass: 0.9 } as const;

type FadeInProps = HTMLMotionProps<"div"> & {
  children: ReactNode;
  delay?: number;
};

export function FadeIn({ children, delay = 0, ...props }: FadeInProps) {
  const reduce = useReducedMotion();
  return (
    <motion.div
      initial={reduce ? false : { opacity: 0, y: 16, scale: 0.985 }}
      whileInView={{ opacity: 1, y: 0, scale: 1 }}
      viewport={{ once: true, margin: "-48px" }}
      transition={{ duration: reduce ? 0 : 0.5, delay, ease: EASE }}
      {...props}
    >
      {children}
    </motion.div>
  );
}

type SlideDirection = "up" | "start" | "end";

type SlideInProps = FadeInProps & {
  direction?: SlideDirection;
};

const OFFSETS: Record<SlideDirection, { x?: number; y?: number }> = {
  up: { y: 24 },
  start: { x: -24 },
  end: { x: 24 },
};

export function SlideIn({
  children,
  direction = "up",
  delay = 0,
  ...props
}: SlideInProps) {
  return (
    <motion.div
      initial={{ opacity: 0, ...OFFSETS[direction] }}
      animate={{ opacity: 1, x: 0, y: 0 }}
      transition={{ duration: 0.5, delay, ease: EASE }}
      {...props}
    >
      {children}
    </motion.div>
  );
}

/* Staggered group reveal */
/* Staggered group reveal.
   Uses `animate` (mount-driven) instead of `whileInView`: lists often render
   empty first and fill after fetch — a scroll-triggered `once` reveal would
   fire with zero children and leave late-mounted items stuck invisible. */
export function Stagger({
  children,
  className,
  step = 0.07,
  delay = 0,
  ...props
}: HTMLMotionProps<"div"> & {
  children: ReactNode;
  step?: number;
  delay?: number;
}) {
  return (
    <motion.div
      className={className}
      initial="hidden"
      animate="show"
      transition={{ staggerChildren: step, delayChildren: delay }}
      {...props}
    >
      {children}
    </motion.div>
  );
}

export function StaggerItem({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <motion.div
      className={className}
      variants={{
        hidden: { opacity: 0, y: 18, scale: 0.98 },
        show: {
          opacity: 1,
          y: 0,
          scale: 1,
          transition: { duration: 0.5, ease: EASE },
        },
      }}
    >
      {children}
    </motion.div>
  );
}

/* Animated odometer-style number, respects reduced motion.
   `render` lets callers format (fa-IR digits, currency…) per frame. */
export function CountUp({
  value,
  render,
  duration = 0.9,
  className,
}: {
  value: number;
  render: (n: number) => string;
  duration?: number;
  className?: string;
}) {
  const ref = useRef<HTMLSpanElement>(null);
  const inView = useInView(ref, { once: true, margin: "-24px" });
  const reduce = useReducedMotion();
  const progress = useMotionValue(reduce ? value : 0);
  const formatted = useTransform(progress, (n) => render(n));

  useEffect(() => {
    if (!inView) return;
    if (reduce) {
      progress.set(value);
      return;
    }
    const controls = animate(progress, value, {
      duration,
      ease: [0.22, 1, 0.36, 1],
    });
    return () => controls.stop();
  }, [value, inView, reduce, progress, duration]);

  return (
    <span ref={ref} className={className}>
      <motion.span>{formatted}</motion.span>
    </span>
  );
}

/* Full-page entrance for route content */
export function PageEnter({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: 18 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.45, ease: EASE }}
    >
      {children}
    </motion.div>
  );
}

/* Re-runs its entrance on every route change */
export function RouteTransition({ children, className }: { children: ReactNode; className?: string }) {
  const reduce = useReducedMotion();
  const pathname = usePathname();
  if (reduce) return <div className={className}>{children}</div>;
  return (
    <motion.div
      key={pathname}
      className={className}
      initial={{ opacity: 0, y: 14 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, ease: EASE }}
    >
      {children}
    </motion.div>
  );
}
