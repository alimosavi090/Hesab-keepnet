"use client";

import { useState, type ReactNode } from "react";
import {
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => {
    const client = new QueryClient({
      defaultOptions: {
        queries: {
          staleTime: 60_000,
          // Always refetch when a page mounts — after recording a
          // transaction/note elsewhere, navigating here must show fresh data
          // without a manual browser refresh.
          refetchOnMount: "always",
          // Accounting data: refresh whenever the tab regains focus too.
          refetchOnWindowFocus: true,
          retry: 1,
        },
      },
    });

    // Bulletproof freshness: after ANY successful mutation, invalidate every
    // query so the current page (and any cached page) re-fetches. This makes
    // new records appear instantly, everywhere, without manual refreshes.
    client.getMutationCache().subscribe((event) => {
      if (event.type === "updated" && event.action?.type === "success") {
        void client.invalidateQueries({ refetchType: "all" });
      }
    });

    return client;
  });

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider
        attribute="class"
        defaultTheme="dark"
        enableSystem={false}
        disableTransitionOnChange
      >
        <TooltipProvider delayDuration={200}>
          {children}
        </TooltipProvider>
        <Toaster richColors closeButton position="bottom-right" />
      </ThemeProvider>
    </QueryClientProvider>
  );
}
