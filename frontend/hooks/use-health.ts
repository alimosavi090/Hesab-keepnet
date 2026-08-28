import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import { apiClient, ApiError } from "@/lib/api-client";
import type { HealthInfo } from "@/types/api";

export function useHealth(): UseQueryResult<HealthInfo, ApiError> {
  return useQuery<HealthInfo, ApiError>({
    queryKey: ["health"],
    queryFn: () => apiClient.get<HealthInfo>("/health"),
    staleTime: 30_000,
    retry: false,
  });
}
