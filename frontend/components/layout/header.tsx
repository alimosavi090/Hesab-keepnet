"use client";

import { usePathname } from "next/navigation";
import { AnimatePresence, motion } from "framer-motion";
import { IconLoader2, IconLogout, IconSearch } from "@tabler/icons-react";
import { pageTitle } from "@/lib/nav";
import { ThemeToggle } from "@/components/layout/theme-toggle";
import { useLogout, useMe } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export function Header() {
  const pathname = usePathname();
  const me = useMe();
  const logout = useLogout();

  const initial = (me.data?.display_name || me.data?.username || "ک")
    .trim()
    .charAt(0);

  // The palette owns its hotkey listener; this only opens it.
  function openSearch() {
    window.dispatchEvent(
      new KeyboardEvent("keydown", { key: "k", ctrlKey: true })
    );
  }

  return (
    <header className="glass sticky top-0 z-20 rounded-none border-x-0 border-t-0">
      <div className="flex h-16 items-center gap-3 px-4 md:px-8">
        <AnimatePresence mode="wait" initial={false}>
          <motion.h1
            key={pathname}
            className="text-page-title"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
          >
            {pageTitle(pathname)}
          </motion.h1>
        </AnimatePresence>

        <div className="flex-1" />

        <button
          type="button"
          onClick={openSearch}
          aria-label="جستجوی سراسری (Ctrl+K)"
          className="bg-card/60 border-border/60 text-muted-foreground hover:text-primary hover:border-primary/40 hidden items-center gap-2 rounded-full border px-3 py-1.5 text-xs transition-colors duration-300 sm:flex"
        >
          <IconSearch aria-hidden="true" className="size-3.5" />
          جستجو
          <kbd className="border-border bg-muted rounded px-1.5 text-[10px]">Ctrl K</kbd>
        </button>
        <Button
          variant="ghost"
          size="icon"
          aria-label="جستجو"
          onClick={openSearch}
          className="sm:hidden"
        >
          <IconSearch className="size-5" />
        </Button>

        {me.data ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <span
                className="ring-primary/30 bg-gradient-to-br from-primary/25 to-chart-3/25 text-primary ring-1 flex size-9 items-center justify-center rounded-full text-sm font-bold transition-transform duration-300 hover:scale-105"
                aria-hidden="true"
              >
                {initial}
              </span>
            </TooltipTrigger>
            <TooltipContent>{me.data.display_name || me.data.username}</TooltipContent>
          </Tooltip>
        ) : null}

        <ThemeToggle />

        <Button
          variant="ghost"
          size="icon"
          aria-label="خروج از حساب"
          disabled={logout.isPending}
          onClick={() => logout.mutate()}
        >
          {logout.isPending ? (
            <IconLoader2 className="size-5 animate-spin" />
          ) : (
            <IconLogout className="size-5 transition-transform duration-300 hover:-translate-y-0.5 hover:text-destructive" />
          )}
        </Button>
      </div>
    </header>
  );
}
