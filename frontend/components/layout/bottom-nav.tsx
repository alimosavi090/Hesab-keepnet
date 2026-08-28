"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { motion } from "framer-motion";
import {
  mobileNavItems,
  secondaryNavItems,
  isActivePath,
} from "@/lib/nav";
import { SPRING } from "@/components/shared/motion";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

export function BottomNav() {
  const pathname = usePathname();
  const [moreOpen, setMoreOpen] = useState(false);
  const items = mobileNavItems();
  const moreActive = secondaryNavItems().some((item) =>
    isActivePath(pathname, item.href)
  );

  return (
    <motion.nav
      aria-label="ناوبری موبایل"
      initial={{ y: 80, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ type: "spring", stiffness: 260, damping: 28, delay: 0.15 }}
      className="fixed inset-x-4 bottom-4 z-30 md:hidden"
    >
      <div className="glass glow-primary rounded-2xl shadow-2xl">
        <ul className="grid grid-cols-5">
          {items.map((item) => {
            const active = isActivePath(pathname, item.href);
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  aria-current={active ? "page" : undefined}
                  className={cn(
                    "relative flex h-16 flex-col items-center justify-center gap-0.5 text-[10px] font-medium transition-colors duration-300",
                    active ? "text-primary" : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  {active ? (
                    <>
                      <motion.span
                        layoutId="dock-active-glow"
                        transition={SPRING}
                        aria-hidden="true"
                        className="bg-primary/15 absolute inset-x-3 inset-y-2 -z-10 rounded-xl"
                      />
                      <span
                        aria-hidden="true"
                        className="bg-primary absolute -bottom-0 size-1 rounded-full shadow-[0_0_8px_2px_var(--primary)]"
                      />
                    </>
                  ) : null}
                  <motion.span
                    animate={{ scale: active ? 1.15 : 1, y: active ? -1 : 0 }}
                    transition={SPRING}
                    className="flex"
                  >
                    <item.icon stroke={active ? 2 : 1.7} className="size-[21px]" />
                  </motion.span>
                  {item.label}
                </Link>
              </li>
            );
          })}
          <li>
            <Sheet open={moreOpen} onOpenChange={setMoreOpen}>
              <SheetTrigger
                className={cn(
                  "relative flex h-16 w-full flex-col items-center justify-center gap-0.5 text-[10px] font-medium transition-colors duration-300",
                  moreActive ? "text-primary" : "text-muted-foreground hover:text-foreground"
                )}
                aria-haspopup="menu"
              >
                {moreActive ? (
                  <span
                    aria-hidden="true"
                    className="bg-primary absolute -bottom-0 size-1 rounded-full shadow-[0_0_8px_2px_var(--primary)]"
                  />
                ) : null}
                <svg
                  className="size-[21px]"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.7"
                  strokeLinecap="round"
                  aria-hidden="true"
                >
                  <circle cx="5" cy="12" r="1" />
                  <circle cx="12" cy="12" r="1" />
                  <circle cx="19" cy="12" r="1" />
                </svg>
                بیشتر
              </SheetTrigger>
              <SheetContent side="bottom" className="glass rounded-t-4xl px-4 pb-8">
                <SheetHeader>
                  <SheetTitle>بخش‌های دیگر</SheetTitle>
                </SheetHeader>
                <ul className="mt-2 space-y-1">
                  {secondaryNavItems().map((item, i) => {
                    const active = isActivePath(pathname, item.href);
                    return (
                      <motion.li
                        key={item.href}
                        initial={{ opacity: 0, x: 16 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.04, ease: [0.22, 1, 0.36, 1], duration: 0.35 }}
                      >
                        <Link
                          href={item.href}
                          onClick={() => setMoreOpen(false)}
                          aria-current={active ? "page" : undefined}
                          className={cn(
                            "flex items-center gap-3 rounded-xl px-3 py-3 text-sm transition-all duration-300",
                            active
                              ? "glow-primary bg-primary/15 font-semibold text-primary"
                              : "hover:bg-accent/60"
                          )}
                        >
                          <item.icon stroke={1.6} className={`size-5 ${active ? "text-primary" : ""}`} />
                          {item.label}
                        </Link>
                      </motion.li>
                    );
                  })}
                </ul>
              </SheetContent>
            </Sheet>
          </li>
        </ul>
      </div>
    </motion.nav>
  );
}
