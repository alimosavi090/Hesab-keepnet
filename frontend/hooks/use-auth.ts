"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { authApi } from "@/lib/api";
import type { PublicUser } from "@/types/api";

export function useMe() {
  return useQuery<PublicUser, Error>({
    queryKey: ["me"],
    queryFn: () => authApi.me(),
    staleTime: 5 * 60_000,
    retry: false,
  });
}

export function useLogout() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      queryClient.clear();
      router.replace("/login");
    },
  });
}
