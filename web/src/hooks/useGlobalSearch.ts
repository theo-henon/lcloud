import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { api, type SearchParams } from "@/lib/api";

export function useGlobalSearch(params: SearchParams, enabled = true) {
  const [debouncedQ, setDebouncedQ] = useState(params.q ?? "");

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQ(params.q ?? ""), 300);
    return () => window.clearTimeout(timer);
  }, [params.q]);

  const hasFilters =
    Boolean(debouncedQ) ||
    Boolean(params.mime_prefix) ||
    params.min_size != null ||
    params.max_size != null ||
    Boolean(params.modified_after) ||
    Boolean(params.modified_before);

  return useQuery({
    queryKey: ["search", "global", debouncedQ, params],
    queryFn: () =>
      api.searchAllFiles({
        ...params,
        q: debouncedQ || undefined,
      }),
    enabled: enabled && hasFilters,
  });
}
