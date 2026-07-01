import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useDisks() {
  return useQuery({
    queryKey: ["disks"],
    queryFn: () => api.listDisks(),
  });
}
