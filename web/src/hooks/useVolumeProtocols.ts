import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type PatchVolumeProtocolsInput } from "@/lib/api";

export function useVolumeProtocols(volumeId: string | undefined) {
  return useQuery({
    queryKey: ["volumes", volumeId, "protocols"],
    queryFn: () => api.getVolumeProtocols(volumeId!),
    enabled: Boolean(volumeId),
  });
}

export function useUpdateVolumeProtocols(volumeId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: PatchVolumeProtocolsInput) =>
      api.patchVolumeProtocols(volumeId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "protocols"] });
    },
  });
}
