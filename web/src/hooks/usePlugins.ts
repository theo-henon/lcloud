import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function usePlugins() {
  return useQuery({
    queryKey: ["plugins"],
    queryFn: () => api.getPlugins(),
    staleTime: 15_000,
  });
}

export function usePluginLogs(params: { plugin_id?: string; volume_id?: string; limit?: number } = {}) {
  return useQuery({
    queryKey: ["plugins", "logs", params],
    queryFn: () => api.getPluginLogs(params),
    refetchInterval: 10_000,
  });
}

export function useTogglePlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.patchPlugin(id, { enabled }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["plugins"] });
    },
  });
}
