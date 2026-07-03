import type { VolumeDeletionRequest } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useDeleteVolume, useDismissVolumeDeletionRequest } from "@/hooks/useVolumes";
import { ApiError } from "@/lib/api";

type DeletionRequestsPanelProps = {
  requests: VolumeDeletionRequest[];
};

export function DeletionRequestsPanel({ requests }: DeletionRequestsPanelProps) {
  const dismissRequest = useDismissVolumeDeletionRequest();
  const deleteVolume = useDeleteVolume();

  if (requests.length === 0) {
    return null;
  }

  return (
    <Card className="space-y-4 p-6">
      <div>
        <h2 className="text-sm font-semibold uppercase tracking-wide text-muted">
          Deletion requests
        </h2>
        <p className="mt-1 text-sm text-body">
          Users asked to remove these volumes. Delete the volume to approve, or dismiss the
          request.
        </p>
      </div>
      <ul className="space-y-3">
        {requests.map((request) => (
          <li
            key={request.id}
            className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-hairline bg-surface-soft px-4 py-3"
          >
            <div className="text-sm">
              <p className="font-medium text-ink">{request.volume_name}</p>
              <p className="text-muted">
                Requested by {request.user_email} ·{" "}
                {new Date(request.created_at).toLocaleString()}
              </p>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                disabled={dismissRequest.isPending || deleteVolume.isPending}
                onClick={() => {
                  void dismissRequest.mutateAsync(request.id);
                }}
              >
                Dismiss
              </Button>
              <Button
                className="text-accent-rose hover:text-accent-rose"
                variant="outline"
                disabled={dismissRequest.isPending || deleteVolume.isPending}
                onClick={() => {
                  const message = request.volume_name
                    ? `Delete volume "${request.volume_name}"? This approves the user's request.`
                    : "Delete this volume?";
                  if (!window.confirm(message)) {
                    return;
                  }
                  void deleteVolume.mutateAsync({ id: request.volume_id, force: false }).catch(
                    (err) => {
                      const needsForce =
                        err instanceof ApiError && err.message.toLowerCase().includes("not empty");
                      if (
                        needsForce &&
                        window.confirm(
                          "Volume is not empty. Force delete all files and remove the volume?",
                        )
                      ) {
                        void deleteVolume.mutateAsync({ id: request.volume_id, force: true });
                      }
                    },
                  );
                }}
              >
                Delete volume
              </Button>
            </div>
          </li>
        ))}
      </ul>
    </Card>
  );
}
