import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useVolumeFiles(volumeId: string | undefined, path: string) {
  return useQuery({
    queryKey: ["volumes", volumeId, "files", path],
    queryFn: () => api.listFiles(volumeId!, path),
    enabled: Boolean(volumeId),
  });
}

export function useCreateDirectory(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (dirPath: string) => api.createDirectory(volumeId, dirPath),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "files"] });
    },
  });
}

export function useUploadFile(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ file, path }: { file: File; path: string }) =>
      api.uploadFile(volumeId, file, path),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "files"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId] });
      void queryClient.invalidateQueries({ queryKey: ["volumes"] });
    },
  });
}

export function useDeleteFile(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (filePath: string) => api.deleteFile(volumeId, filePath),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "files"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId] });
      void queryClient.invalidateQueries({ queryKey: ["volumes"] });
    },
  });
}

export function useMoveFile(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ fromPath, toPath }: { fromPath: string; toPath: string }) =>
      api.moveFile(volumeId, fromPath, toPath),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "files"] });
    },
  });
}

export function useRenameFile(volumeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ path, newName }: { path: string; newName: string }) =>
      api.renameFile(volumeId, path, newName),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "files"] });
    },
  });
}
