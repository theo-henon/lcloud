import { Link } from "react-router-dom";
import { WidgetEmpty, WidgetError, WidgetLoading, WidgetShell } from "@/components/dashboard/WidgetShell";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { WidgetProps } from "@/lib/dashboard/types";

export function PendingDeletionsWidget(_props: WidgetProps) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["admin", "volume-deletion-requests"],
    queryFn: () => api.listVolumeDeletionRequests(),
    staleTime: 30_000,
  });

  if (isLoading) {
    return (
      <WidgetShell title="Pending deletions" deepLink="/volumes">
        <WidgetLoading />
      </WidgetShell>
    );
  }

  if (isError || !data) {
    return (
      <WidgetShell title="Pending deletions" deepLink="/volumes">
        <WidgetError />
      </WidgetShell>
    );
  }

  const requests = data.requests.slice(0, 5);

  if (requests.length === 0) {
    return (
      <WidgetShell title="Pending deletions" deepLink="/volumes">
        <WidgetEmpty message="No pending deletion requests." />
      </WidgetShell>
    );
  }

  return (
    <WidgetShell title="Pending deletions" deepLink="/volumes">
      <ul className="space-y-2 text-sm">
        {requests.map((request) => (
          <li key={request.id} className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <p className="truncate font-medium text-ink">{request.volume_name}</p>
              <p className="truncate text-xs text-muted">{request.user_email}</p>
            </div>
            <Link
              to="/volumes"
              className="shrink-0 text-xs text-primary hover:text-primary-active"
            >
              Review
            </Link>
          </li>
        ))}
      </ul>
    </WidgetShell>
  );
}
