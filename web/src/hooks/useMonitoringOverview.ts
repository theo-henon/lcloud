import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useMonitoringOverview() {
  return useQuery({
    queryKey: ["monitoring", "overview"],
    queryFn: () => api.getMonitoringOverview(),
    staleTime: 30_000,
  });
}
