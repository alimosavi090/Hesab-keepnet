"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { motion } from "framer-motion";
import { MAIN_NAV, isActivePath } from "@/lib/nav";
import { SPRING } from "@/components/shared/motion";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

export function AppSidebar() {
  const pathname = usePathname();

  return (
    <Sidebar side="right" className="hidden lg:flex">
      <SidebarHeader className="px-4 py-5">
        <motion.div
          className="flex items-center gap-3"
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
        >
          <span className="brand-tile animate-pulse-glow flex size-10 items-center justify-center rounded-xl text-lg font-black text-primary-foreground">
            ح
          </span>
          <div className="leading-tight">
            <span className="text-gradient block text-base font-extrabold">حساب‌کیپ</span>
            <span className="text-caption">مدیریت مالی هوشمند</span>
          </div>
        </motion.div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {MAIN_NAV.map((item) => {
                const active = isActivePath(pathname, item.href);
                return (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton
                      asChild
                      isActive={active}
                      tooltip={item.label}
                      className="group relative overflow-hidden rounded-xl transition-colors data-[active=true]:bg-transparent"
                    >
                      <Link
                        href={item.href}
                        aria-current={active ? "page" : undefined}
                        className="relative flex w-full items-center gap-2.5 overflow-hidden px-2 py-2"
                      >
                        {active ? (
                          <motion.span
                            layoutId="sidebar-active-pill"
                            transition={SPRING}
                            aria-hidden="true"
                            className="glow-primary absolute inset-0 rounded-xl bg-primary/15"
                          />
                        ) : null}
                        <item.icon
                          stroke={1.6}
                          aria-hidden="true"
                          className={`relative z-10 size-5 shrink-0 transition-all duration-300 group-hover:-translate-x-0.5 group-hover:scale-110 ${
                            active ? "text-primary" : ""
                          }`}
                        />
                        <span
                          className={`relative z-10 text-sm transition-colors ${
                            active ? "font-semibold text-primary" : ""
                          }`}
                        >
                          {item.label}
                        </span>
                        {active ? (
                          <span
                            aria-hidden="true"
                            className="bg-primary absolute inset-y-2 left-0 w-[3px] rounded-full"
                          />
                        ) : null}
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="px-4 pb-4">
        <p className="text-caption numeric text-center">v0.1.0 — Phase 1</p>
      </SidebarFooter>
    </Sidebar>
  );
}
