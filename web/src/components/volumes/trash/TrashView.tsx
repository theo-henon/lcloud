import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  useEmptyTrash,
  usePurgeTrashItem,
  useRestoreTrashItem,
  useVolumeTrash,
} from "@/hooks/useVolumeTrash";
import type { TrashEntry } from "@/lib/api";
import { api } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

type TrashToolbarProps = {
  itemCount: number;
  onEmptyTrash: () => void;
  isEmptying: boolean;
};

export function TrashToolbar({ itemCount, onEmptyTrash, isEmptying }: TrashToolbarProps) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-hairline px-4 py-3">
      <p className="text-sm text-muted">
        {itemCount === 0 ? "Trash is empty" : `${itemCount} item${itemCount === 1 ? "" : "s"} in trash`}
      </p>
      <Button
        type="button"
        variant="outline"
        disabled={itemCount === 0 || isEmptying}
        className="border-accent-rose/40 text-accent-rose hover:border-accent-rose hover:bg-accent-rose/10"
        onClick={onEmptyTrash}
      >
        <Trash2 className="mr-2 h-4 w-4" aria-hidden />
        Empty trash
      </Button>
    </div>
  );
}

type TrashListProps = {
  volumeId: string;
  items: TrashEntry[];
  onRestore: (item: TrashEntry) => void;
  onPurge: (item: TrashEntry) => void;
};

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export function TrashList({ volumeId, items, onRestore, onPurge }: TrashListProps) {
  if (items.length === 0) {
    return (
      <div className="px-4 py-12 text-center">
        <Trash2 className="mx-auto mb-3 h-8 w-8 text-muted-soft" aria-hidden />
        <p className="text-sm text-body">Trash is empty</p>
        <p className="mt-1 text-xs text-muted">Deleted files appear here and can be restored.</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[640px] text-left text-sm">
        <thead>
          <tr className="border-b border-hairline text-xs uppercase tracking-wide text-muted">
            <th className="px-4 py-2 font-medium">Name</th>
            <th className="px-4 py-2 font-medium">Original location</th>
            <th className="px-4 py-2 font-medium">Deleted</th>
            <th className="px-4 py-2 font-medium">Size</th>
            <th className="px-4 py-2 font-medium text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id} className="border-b border-hairline/60 hover:bg-surface-elevated/50">
              <td className="px-4 py-2.5">
                <div className="flex items-center gap-2">
                  {item.has_thumbnail ? (
                    <img
                      src={api.trashThumbnailUrl(volumeId, item.id)}
                      alt=""
                      className="h-8 w-8 rounded object-cover"
                    />
                  ) : null}
                  <span className="truncate text-body">{item.name}</span>
                </div>
              </td>
              <td className="max-w-[200px] truncate px-4 py-2.5 font-mono text-xs text-muted">
                {item.original_path}
              </td>
              <td className="whitespace-nowrap px-4 py-2.5 text-muted">{formatDate(item.deleted_at)}</td>
              <td className="whitespace-nowrap px-4 py-2.5 text-muted">{formatBytes(item.size_bytes)}</td>
              <td className="px-4 py-2.5 text-right">
                <div className="flex justify-end gap-2">
                  <Button type="button" variant="outline" size="sm" onClick={() => onRestore(item)}>
                    Restore
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="border-accent-rose/40 text-accent-rose hover:border-accent-rose hover:bg-accent-rose/10"
                    onClick={() => onPurge(item)}
                  >
                    Delete permanently
                  </Button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

type TrashViewProps = {
  volumeId: string;
  retentionDays?: number;
};

export function TrashView({ volumeId, retentionDays = 30 }: TrashViewProps) {
  const trashQuery = useVolumeTrash(volumeId);
  const restore = useRestoreTrashItem(volumeId);
  const purge = usePurgeTrashItem(volumeId);
  const emptyTrash = useEmptyTrash(volumeId);

  const items = trashQuery.data?.items ?? [];
  const total = trashQuery.data?.total ?? 0;

  const handleRestore = (item: TrashEntry) => {
    void restore.mutateAsync(item.id);
  };

  const handlePurge = (item: TrashEntry) => {
    if (window.confirm(`Permanently delete "${item.name}"? This cannot be undone.`)) {
      void purge.mutateAsync(item.id);
    }
  };

  const handleEmpty = () => {
    if (window.confirm(`Empty trash? ${total} item${total === 1 ? "" : "s"} will be permanently deleted.`)) {
      void emptyTrash.mutateAsync();
    }
  };

  if (trashQuery.isLoading) {
    return <p className="px-4 py-6 text-sm text-muted">Loading trash…</p>;
  }

  if (trashQuery.isError) {
    return <p className="px-4 py-6 text-sm text-accent-rose">Unable to load trash.</p>;
  }

  return (
    <div>
      <TrashToolbar itemCount={total} onEmptyTrash={handleEmpty} isEmptying={emptyTrash.isPending} />
      <TrashList volumeId={volumeId} items={items} onRestore={handleRestore} onPurge={handlePurge} />
      <footer className="border-t border-hairline px-4 py-3 text-xs text-muted">
        Items are permanently deleted after {retentionDays} days.
      </footer>
    </div>
  );
}
