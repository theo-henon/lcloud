import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { CreateTaskInput, PatchTaskInput } from "@/lib/api";
import { api } from "@/lib/api";

export function useTasks(filters?: { volumeId?: string; ownerId?: string }) {
  const volumeId = filters?.volumeId;
  const ownerId = filters?.ownerId;
  return useQuery({
    queryKey: ["tasks", volumeId ?? "all", ownerId ?? "all"],
    queryFn: () => api.listTasks(filters),
    staleTime: 15_000,
  });
}

export function useTaskRuns(taskId: string | null) {
  return useQuery({
    queryKey: ["tasks", taskId, "runs"],
    queryFn: () => api.listTaskRuns(taskId!),
    enabled: Boolean(taskId),
    refetchInterval: 5_000,
  });
}

export function useCreateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateTaskInput) => api.createTask(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tasks"] });
    },
  });
}

export function useUpdateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: PatchTaskInput }) =>
      api.patchTask(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tasks"] });
    },
  });
}

export function useDeleteTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteTask(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tasks"] });
    },
  });
}

export function useRunTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, dryRun }: { id: string; dryRun?: boolean }) =>
      api.runTask(id, dryRun),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: ["tasks"] });
      void queryClient.invalidateQueries({
        queryKey: ["tasks", variables.id, "runs"],
      });
    },
  });
}
