import { useState } from "react";
import { Header } from "@/components/layout/Header";
import { OwnerFilter } from "@/components/filters/OwnerFilter";
import { CreateVolumeForm } from "@/components/volumes/CreateVolumeForm";
import { DeletionRequestsPanel } from "@/components/volumes/DeletionRequestsPanel";
import { VolumeCard } from "@/components/volumes/VolumeCard";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useDisks } from "@/hooks/useDisks";
import {
  useCreateVolume,
  useDeleteVolume,
  usePatchVolume,
  useRequestVolumeDeletion,
  useVolumeDeletionRequests,
  useVolumes,
} from "@/hooks/useVolumes";
import { ActionIcon, navIcons } from "@/lib/icons";
import { ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";
import { useAdminUsers } from "@/hooks/useUsers";

export function VolumesPage() {
  const user = useAuthStore((state) => state.user);
  const isAdmin = user?.role === "admin";
  const [ownerFilter, setOwnerFilter] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const disksQuery = useDisks();
  const usersQuery = useAdminUsers(isAdmin);
  const volumesQuery = useVolumes(isAdmin && ownerFilter ? ownerFilter : undefined);
  const deletionRequestsQuery = useVolumeDeletionRequests(isAdmin);
  const createVolume = useCreateVolume();
  const patchVolume = usePatchVolume();
  const deleteVolume = useDeleteVolume();
  const requestDeletion = useRequestVolumeDeletion();

  const disks = disksQuery.data?.disks ?? [];
  const volumes = volumesQuery.data?.volumes ?? [];
  const deletionRequests = deletionRequestsQuery.data?.requests ?? [];

  return (
    <>
      <Header
        title="Volumes"
        description={
          isAdmin
            ? "Create and manage storage volumes on your physical disks."
            : "Create and manage your personal storage volumes."
        }
        action={
          <Button onClick={() => setShowCreate(true)} disabled={disks.length === 0}>
            <ActionIcon action="create" className="mr-2" />
            Create volume
          </Button>
        }
      />

      <section className="space-y-6 px-8 py-6">
        {error ? (
          <Card className="border-accent-rose/40 p-4 text-sm text-accent-rose">{error}</Card>
        ) : null}

        {isAdmin ? <DeletionRequestsPanel requests={deletionRequests} /> : null}

        {isAdmin && usersQuery.data ? (
          <OwnerFilter
            value={ownerFilter}
            currentUserId={user!.id}
            currentUserEmail={user!.email}
            users={usersQuery.data.users}
            onChange={setOwnerFilter}
          />
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
          <Card className="flex flex-col items-center gap-3 p-8 text-center text-sm text-body">
            {(() => {
              const EmptyIcon = navIcons["/volumes"];
              return EmptyIcon ? (
                <EmptyIcon className="h-8 w-8 text-muted" aria-hidden />
              ) : null;
            })()}
            <p>
              {isAdmin
                ? "No volumes yet. Create one to start storing files."
                : "No volumes yet. Create your first volume to start storing files."}
            </p>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {volumes.map((volume) => (
              <VolumeCard
                key={volume.id}
                volume={volume}
                isAdmin={isAdmin}
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
                onRequestDeletion={(item) => {
                  const message =
                    "Request deletion of this volume? An administrator must approve and delete it.";
                  if (!window.confirm(message)) {
                    return;
                  }
                  setError(null);
                  void requestDeletion.mutateAsync(item.id).catch((err) => {
                    setError(
                      err instanceof ApiError ? err.message : "Unable to request deletion.",
                    );
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
