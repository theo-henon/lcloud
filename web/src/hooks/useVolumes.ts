import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type CreateVolumeInput, type PatchVolumeInput } from "@/lib/api";

export function useVolumes() {
  return useQuery({
    queryKey: ["volumes"],
    queryFn: () => api.listVolumes(),
  });
}

export function useVolume(id: string | undefined) {
  return useQuery({
    queryKey: ["volumes", id],
    queryFn: () => api.getVolume(id!),
    enabled: Boolean(id),
  });
}

export function useCreateVolume() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateVolumeInput) => api.createVolume(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes"] });
    },
  });
}

export function usePatchVolume() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: PatchVolumeInput }) =>
      api.patchVolume(id, input),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: ["volumes"] });
      void queryClient.invalidateQueries({ queryKey: ["volumes", variables.id] });
    },
  });
}

export function useDeleteVolume() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, force }: { id: string; force?: boolean }) =>
      api.deleteVolume(id, force),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes"] });
      void queryClient.invalidateQueries({ queryKey: ["volume-deletion-requests"] });
    },
  });
}

export function useRequestVolumeDeletion() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.requestVolumeDeletion(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes"] });
      void queryClient.invalidateQueries({ queryKey: ["volume-deletion-requests"] });
    },
  });
}

export function useVolumeDeletionRequests(enabled = true) {
  return useQuery({
    queryKey: ["volume-deletion-requests"],
    queryFn: () => api.listVolumeDeletionRequests(),
    enabled,
    staleTime: 15_000,
  });
}

export function useDismissVolumeDeletionRequest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.dismissVolumeDeletionRequest(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volume-deletion-requests"] });
    },
  });
}
