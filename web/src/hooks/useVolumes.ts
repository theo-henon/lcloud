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
    },
  });
}
