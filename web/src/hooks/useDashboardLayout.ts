import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { DashboardLayout, PatchDashboardInput } from "@/lib/dashboard/types";
import { api } from "@/lib/api";

export function useDashboardLayout() {
  return useQuery({
    queryKey: ["dashboard"],
    queryFn: () => api.getDashboardLayout(),
    staleTime: 30_000,
  });
}

export function useSaveDashboardLayout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (layout: DashboardLayout) => api.patchDashboardLayout({ layout }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useResetDashboardLayout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.resetDashboardLayout(),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export type { PatchDashboardInput };
