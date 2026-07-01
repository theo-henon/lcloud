import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useVolumeStats(volumeId: string | undefined) {
  return useQuery({
    queryKey: ["monitoring", "stats", volumeId],
    queryFn: () => api.getVolumeStats(volumeId!),
    enabled: Boolean(volumeId),
  });
}

export function useRefreshVolumeStats(volumeId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.refreshVolumeStats(volumeId!),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["monitoring"] });
    },
  });
}
