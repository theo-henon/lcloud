import { useState } from "react";
import { Header } from "@/components/layout/Header";
import { CreateVolumeForm } from "@/components/volumes/CreateVolumeForm";
import { VolumeCard } from "@/components/volumes/VolumeCard";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useDisks } from "@/hooks/useDisks";
import {
  useCreateVolume,
  useDeleteVolume,
  usePatchVolume,
  useVolumes,
} from "@/hooks/useVolumes";
import { ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

export function VolumesPage() {
  const user = useAuthStore((state) => state.user);
  const [showCreate, setShowCreate] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const disksQuery = useDisks();
  const volumesQuery = useVolumes();
  const createVolume = useCreateVolume();
  const patchVolume = usePatchVolume();
  const deleteVolume = useDeleteVolume();

  const disks = disksQuery.data?.disks ?? [];
  const volumes = volumesQuery.data?.volumes ?? [];

  return (
    <>
      <Header
        title="Volumes"
        description="Create and manage storage volumes on your physical disks."
        action={
          user?.role === "admin" ? (
            <Button onClick={() => setShowCreate(true)} disabled={disks.length === 0}>
              Create volume
            </Button>
          ) : undefined
        }
      />

      <section className="space-y-6 px-8 py-6">
        {error ? (
          <Card className="border-accent-rose/40 p-4 text-sm text-accent-rose">{error}</Card>
        ) : null}

        {showCreate ? (
          <Card className="p-6">
            <h2 className="mb-4 text-lg font-semibold text-ink">New volume</h2>
            <CreateVolumeForm
              disks={disks}
              loading={createVolume.isPending}
              onCancel={() => setShowCreate(false)}
              onSubmit={async (values) => {
                setError(null);
                try {
                  await createVolume.mutateAsync(values);
                  setShowCreate(false);
                } catch (err) {
                  setError(err instanceof ApiError ? err.message : "Unable to create volume.");
                }
              }}
            />
          </Card>
        ) : null}

        {volumesQuery.isLoading ? (
          <p className="text-sm text-muted">Loading volumes...</p>
        ) : volumes.length === 0 ? (
          <Card className="p-6 text-sm text-body">
            No volumes yet. Create one to start storing files.
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {volumes.map((volume) => (
              <VolumeCard
                key={volume.id}
                volume={volume}
                isAdmin={user?.role === "admin"}
                onRename={(item) => {
                  const nextName = window.prompt("Rename volume", item.name);
                  if (!nextName || nextName.trim() === item.name) {
                    return;
                  }
                  void patchVolume.mutateAsync({
                    id: item.id,
                    input: { name: nextName.trim() },
                  });
                }}
                onDelete={(item, force) => {
                  if (!force && item.used_bytes > 0) {
                    setError("This volume still contains files and cannot be deleted.");
                    return;
                  }
                  const message = force
                    ? "Force delete will permanently remove all files in this volume. Continue?"
                    : "Delete this empty volume?";
                  if (!window.confirm(message)) {
                    return;
                  }
                  void deleteVolume
                    .mutateAsync({ id: item.id, force })
                    .catch((err) => {
                      setError(err instanceof ApiError ? err.message : "Unable to delete volume.");
                    });
                }}
              />
            ))}
          </div>
        )}
      </section>
    </>
  );
}
