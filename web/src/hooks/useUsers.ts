import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { CreateAdminUserInput, PatchAdminUserInput } from "@/lib/api";
import { api } from "@/lib/api";

export function useAdminUsers() {
  return useQuery({
    queryKey: ["admin", "users"],
    queryFn: () => api.listAdminUsers(),
    staleTime: 15_000,
  });
}

export function useCreateAdminUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateAdminUserInput) => api.createAdminUser(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    },
  });
}

export function usePatchAdminUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: PatchAdminUserInput }) =>
      api.patchAdminUser(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    },
  });
}
