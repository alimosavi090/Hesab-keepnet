"use client";

import { useEffect, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { useMe } from "@/hooks/use-auth";
import { LoadingState } from "@/components/shared/loading-state";
import { ErrorState } from "@/components/shared/error-state";

export function RequireAuth({ children }: { children: ReactNode }) {
  const router = useRouter();
  const me = useMe();

  useEffect(() => {
    if (me.isError) {
      const error = me.error as { status?: number };
      if (!error.status || error.status === 401) {
        router.replace("/login");
      }
    }
  }, [me.isError, me.error, router]);

  if (me.isPending) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <LoadingState label="در حال بررسی نشست…" />
      </div>
    );
  }

  if (me.isError) {
    return (
      <div className="p-6">
        <ErrorState
          title="دسترسی برقرار نیست"
          description="نشست شما منقضی شده یا خطایی در ارتباط با سرور رخ داد."
          onRetry={() => me.refetch()}
        />
      </div>
    );
  }

  return <>{children}</>;
}
