import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useSettings() {
  return useQuery({
    queryKey: ["settings"],
    queryFn: () => api.getSettings(),
    staleTime: 60_000,
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: {
      mask_disk_names?: boolean;
      protocols_webdav_enabled?: boolean;
      protocols_ftp_enabled?: boolean;
    }) => api.patchAdminSettings(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["settings"] });
      void queryClient.invalidateQueries({ queryKey: ["monitoring"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes"] });
      void queryClient.invalidateQueries({ queryKey: ["disks"] });
    },
  });
}
