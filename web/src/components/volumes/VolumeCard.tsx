import { Link } from "react-router-dom";
import { HardDrive } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ActionIcon } from "@/lib/icons";
import type { Volume } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

type VolumeCardProps = {
  volume: Volume;
  isAdmin: boolean;
  onRename: (volume: Volume) => void;
  onDelete: (volume: Volume, force?: boolean) => void;
};

export function VolumeCard({ volume, isAdmin, onRename, onDelete }: VolumeCardProps) {
  const quotaLabel =
    volume.quota_bytes > 0
      ? `${formatBytes(volume.used_bytes)} / ${formatBytes(volume.quota_bytes)}`
      : `${formatBytes(volume.used_bytes)} / unlimited`;

  const filterSummary =
    volume.filters.mode && volume.filters.extensions.length > 0
      ? `${volume.filters.mode}: ${volume.filters.extensions.join(" ")}`
      : "No filter";

  return (
    <Card className="flex flex-col gap-4 p-5">
      <div>
        <h3 className="flex items-center gap-2 text-lg font-semibold text-ink">
          <HardDrive className="h-5 w-5 shrink-0 text-muted" aria-hidden />
          {volume.name}
        </h3>
        {volume.disk_path ? (
          <p className="mt-1 text-sm text-muted">{volume.disk_path}</p>
        ) : null}
      </div>

      <div className="space-y-1 text-sm text-body">
        <p>{quotaLabel}</p>
        <p>{filterSummary}</p>
      </div>

      <div className="mt-auto flex flex-wrap gap-2">
        <Link
          to={`/volumes/${volume.id}`}
          className="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary hover:bg-primary-active"
        >
          <ActionIcon action="open" className="mr-2 text-on-primary" destructive={false} />
          Open
        </Link>
        <Button variant="outline" onClick={() => onRename(volume)}>
          <ActionIcon action="rename" className="mr-2" />
          Rename
        </Button>
        <Button
          variant="outline"
          className="text-accent-rose hover:text-accent-rose"
          onClick={() => onDelete(volume, false)}
        >
          <ActionIcon action="delete" className="mr-2" />
          Delete
        </Button>
        {isAdmin ? (
          <Button
            variant="outline"
            className="text-accent-rose hover:text-accent-rose"
            onClick={() => onDelete(volume, true)}
          >
            <ActionIcon action="delete" className="mr-2" />
            Force delete
          </Button>
        ) : null}
      </div>
    </Card>
  );
}
