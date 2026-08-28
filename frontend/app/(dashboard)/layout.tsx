import type { ReactNode } from "react";
import { AppSidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { BottomNav } from "@/components/layout/bottom-nav";
import { RequireAuth } from "@/components/layout/require-auth";
import { RouteTransition } from "@/components/shared/motion";
import { CommandPalette } from "@/components/shared/command-palette";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <RequireAuth>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <Header />
          <main className="flex-1 px-4 pb-28 pt-6 md:px-8 md:pb-10">
            <RouteTransition>{children}</RouteTransition>
          </main>
        </SidebarInset>
        <BottomNav />
        <CommandPalette />
      </SidebarProvider>
    </RequireAuth>
  );
}
