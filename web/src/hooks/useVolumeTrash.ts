import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useVolumeTrash(volumeId: string | undefined, limit = 50, offset = 0) {
  return useQuery({
    queryKey: ["volumes", volumeId, "trash", limit, offset],
    queryFn: () => api.listTrash(volumeId!, limit, offset),
    enabled: Boolean(volumeId),
  });
}

export function useRestoreTrashItem(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (fileId: string) => api.restoreTrashItem(volumeId, fileId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "trash"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "files"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId] });
    },
  });
}

export function usePurgeTrashItem(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (fileId: string) => api.purgeTrashItem(volumeId, fileId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "trash"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId] });
    },
  });
}

export function useEmptyTrash(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.emptyTrash(volumeId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "trash"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId] });
    },
  });
}
