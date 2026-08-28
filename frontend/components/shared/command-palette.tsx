"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { AnimatePresence, motion } from "framer-motion";
import {
  IconBuildingBank,
  IconCornerDownLeft,
  IconSearch,
  IconUsersGroup,
} from "@tabler/icons-react";
import { bankAccountsApi, representativesApi } from "@/lib/api";
import { MAIN_NAV } from "@/lib/nav";
import type { TablerIcon } from "@tabler/icons-react";
import { SPRING } from "@/components/shared/motion";
import { cn } from "@/lib/utils";

type SearchItem = {
  id: string;
  label: string;
  hint?: string;
  icon: TablerIcon;
  href: string;
};

function normalize(value: string): string {
  return value
    .toLowerCase()
    .replace(/ي/g, "ی")
    .replace(/ك/g, "ک")
    .replace(/\s+/g, " ")
    .trim();
}

export function CommandPalette() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  // ⌘K / Ctrl+K toggles the palette anywhere in the app.
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setOpen((prev) => {
          if (!prev) {
            setQuery("");
            setActiveIndex(0);
          }
          return !prev;
        });
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  // Move focus to the search field once mounted.
  useEffect(() => {
    if (!open) return;
    const handle = requestAnimationFrame(() => inputRef.current?.focus());
    return () => cancelAnimationFrame(handle);
  }, [open]);

  // Only fetch remote datasets while the palette is open.
  const reps = useQuery({
    queryKey: ["representatives", "command-palette"],
    queryFn: () => representativesApi.list(true),
    enabled: open,
    staleTime: 60_000,
  });
  const accounts = useQuery({
    queryKey: ["bank-accounts", "command-palette"],
    queryFn: () => bankAccountsApi.list(true),
    enabled: open,
    staleTime: 60_000,
  });

  const items = useMemo<SearchItem[]>(() => {
    const list: SearchItem[] = MAIN_NAV.map((nav) => ({
      id: `nav-${nav.href}`,
      label: nav.label,
      hint: nav.href,
      icon: nav.icon,
      href: nav.href,
    }));
    for (const rep of reps.data ?? []) {
      list.push({
        id: `rep-${rep.id}`,
        label: rep.full_name,
        hint: rep.phone ? String(rep.phone) : "نماینده",
        icon: IconUsersGroup,
        href: `/representatives/${rep.id}`,
      });
    }
    for (const acct of accounts.data ?? []) {
      list.push({
        id: `acct-${acct.id}`,
        label: acct.name,
        hint: acct.bank_name,
        icon: IconBuildingBank,
        href: "/bank-accounts",
      });
    }
    if (!query) return list.slice(0, 14); // quick launcher view
    const q = normalize(query);
    return list.filter(
      (item) =>
        normalize(item.label).includes(q) ||
        normalize(item.hint ?? "").includes(q)
    );
  }, [query, reps.data, accounts.data]);

  const activeItem = items[Math.min(activeIndex, items.length - 1)];

  function go(item?: SearchItem) {
    if (!item) return;
    setOpen(false);
    router.push(item.href);
  }

  function onKeyDown(event: React.KeyboardEvent) {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActiveIndex((i) => Math.min(i + 1, items.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActiveIndex((i) => Math.max(i - 1, 0));
    } else if (event.key === "Enter") {
      event.preventDefault();
      go(activeItem);
    }
  }

  return (
    <AnimatePresence>
      {open ? (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.18 }}
          className="fixed inset-0 z-50 flex items-start justify-center bg-black/50 px-4 pt-[12vh] backdrop-blur-sm"
          onClick={() => setOpen(false)}
          role="presentation"
        >
          <motion.div
            role="dialog"
            aria-modal="true"
            aria-label="جستجوی سراسری"
            initial={{ opacity: 0, y: -16, scale: 0.97 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -12, scale: 0.98 }}
            transition={SPRING}
            className="glass sheen w-full max-w-xl rounded-2xl p-2 shadow-2xl"
            onClick={(e) => e.stopPropagation()}
            onKeyDown={onKeyDown}
          >
            <div className="relative">
              <IconSearch
                aria-hidden="true"
                className="text-muted-foreground absolute right-3 top-1/2 size-4 -translate-y-1/2"
              />
              <input
                ref={inputRef}
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setActiveIndex(0);
                }}
                placeholder="جستجو در صفحات، نماینده‌ها و حساب‌ها…"
                aria-label="عبارت جستجو"
                className="bg-background/40 border-border focus:border-primary/50 w-full rounded-xl py-3 ps-9 pe-4 text-sm outline-none"
              />
            </div>

            <ul role="listbox" aria-label="نتایج" className="mt-1 max-h-80 overflow-y-auto p-1">
              {items.length === 0 ? (
                <li className="text-caption py-8 text-center">نتیجه‌ای یافت نشد.</li>
              ) : (
                items.map((item, index) => (
                  <li key={item.id} role="option" aria-selected={index === activeIndex}>
                    <button
                      type="button"
                      onMouseEnter={() => setActiveIndex(index)}
                      onClick={() => go(item)}
                      className={cn(
                        "flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-colors duration-150",
                        index === activeIndex
                          ? "bg-primary/15 text-primary font-semibold"
                          : "hover:bg-accent/50"
                      )}
                    >
                      <item.icon stroke={1.6} className={cn("size-[18px] shrink-0", index === activeIndex ? "text-primary" : "text-muted-foreground")} />
                      <span className="truncate">{item.label}</span>
                      <span className="text-caption numeric truncate">{item.hint ?? ""}</span>
                      {index === activeIndex ? (
                        <IconCornerDownLeft aria-hidden="true" className="ms-auto size-3.5 shrink-0 opacity-60" />
                      ) : null}
                    </button>
                  </li>
                ))
              )}
            </ul>

            <p className="text-caption flex items-center gap-2 border-t px-3 py-2">
              <kbd className="border-border bg-muted rounded px-1.5 text-[10px]">Ctrl K</kbd>
              باز/بستن · <kbd className="border-border bg-muted rounded px-1.5 text-[10px]">Enter</kbd> رفتن
            </p>
          </motion.div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}
