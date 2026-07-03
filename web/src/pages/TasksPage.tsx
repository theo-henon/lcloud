import { useState, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import type { TaskRecord } from "@/lib/api";
import { Header } from "@/components/layout/Header";
import { OwnerFilter } from "@/components/filters/OwnerFilter";
import { TaskForm } from "@/components/tasks/TaskForm";
import { TaskList } from "@/components/tasks/TaskList";
import { TaskRunHistory } from "@/components/tasks/TaskRunHistory";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useCreateTask, useTaskRuns, useTasks, useUpdateTask } from "@/hooks/useTasks";
import { useVolumes } from "@/hooks/useVolumes";
import { useAdminUsers } from "@/hooks/useUsers";
import { ActionIcon } from "@/lib/icons";
import { useAuthStore } from "@/store/auth";

export function TasksPage() {
  const user = useAuthStore((state) => state.user);
  const isAdmin = user?.role === "admin";
  const [searchParams, setSearchParams] = useSearchParams();
  const [ownerFilter, setOwnerFilter] = useState("");
  const [volumeFilter, setVolumeFilter] = useState<string>("");
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<TaskRecord | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(
    () => searchParams.get("task") || null,
  );

  useEffect(() => {
    const taskParam = searchParams.get("task");
    if (taskParam) {
      setSelectedId(taskParam);
    }
  }, [searchParams]);

  const usersQuery = useAdminUsers(isAdmin);
  const volumesForFormQuery = useVolumes();
  const volumesForFilterQuery = useVolumes(isAdmin && ownerFilter ? ownerFilter : undefined);
  const tasksQuery = useTasks({
    volumeId: volumeFilter || undefined,
    ownerId: isAdmin && ownerFilter ? ownerFilter : undefined,
  });
  const runsQuery = useTaskRuns(selectedId);
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();

  if (
    tasksQuery.isLoading ||
    volumesForFormQuery.isLoading ||
    (isAdmin && usersQuery.isLoading)
  ) {
    return (
      <>
        <Header
          title="Tasks"
          description={
            isAdmin
              ? "Manage scheduled jobs across the instance."
              : "Automate recurring jobs on your volumes."
          }
        />
        <section className="px-8 py-6">
          <p className="text-sm text-muted">Loading tasks...</p>
        </section>
      </>
    );
  }

  if (
    tasksQuery.isError ||
    volumesForFormQuery.isError ||
    !tasksQuery.data ||
    !volumesForFormQuery.data ||
    (isAdmin && (usersQuery.isError || !usersQuery.data))
  ) {
    return (
      <div className="px-8 py-6">
        <Card className="p-6 text-sm text-accent-rose">Unable to load tasks.</Card>
      </div>
    );
  }

  const volumesForForm = volumesForFormQuery.data.volumes;
  const volumesForFilter = volumesForFilterQuery.data?.volumes ?? volumesForForm;

  return (
    <>
      <Header
        title="Tasks"
        description={
          isAdmin
            ? "Manage scheduled jobs across the instance."
            : "Automate recurring jobs on your volumes."
        }
      />

      <section className="space-y-8 px-8 py-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-4">
            {isAdmin && usersQuery.data ? (
              <OwnerFilter
                value={ownerFilter}
                currentUserId={user!.id}
                currentUserEmail={user!.email}
                users={usersQuery.data.users}
                onChange={(next) => {
                  setOwnerFilter(next);
                  setVolumeFilter("");
                }}
              />
            ) : null}
            <div className="flex items-center gap-3">
              <label className="text-sm text-muted">Filter by volume</label>
              <select
                className="rounded-md border border-hairline bg-surface px-3 py-2 text-sm text-ink"
                value={volumeFilter}
                onChange={(e) => setVolumeFilter(e.target.value)}
              >
                <option value="">All volumes</option>
                {volumesForFilter.map((vol) => (
                  <option key={vol.id} value={vol.id}>
                    {vol.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <Button
            onClick={() => {
              setEditing(null);
              setShowForm(true);
            }}
          >
            <ActionIcon action="create" className="mr-2" />
            Create task
          </Button>
        </div>

        {showForm || editing ? (
          <TaskForm
            volumes={volumesForForm}
            isAdmin={isAdmin}
            initial={editing ?? undefined}
            isPending={createTask.isPending || updateTask.isPending}
            onCancel={() => {
              setShowForm(false);
              setEditing(null);
            }}
            onSubmit={(payload) => {
              if (editing) {
                updateTask.mutate(
                  { id: editing.id, input: payload },
                  {
                    onSuccess: () => {
                      setEditing(null);
                      setShowForm(false);
                    },
                  },
                );
              } else {
                createTask.mutate(payload, {
                  onSuccess: () => setShowForm(false),
                });
              }
            }}
          />
        ) : null}

        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-muted">
            {isAdmin ? "All scheduled tasks" : "Your scheduled tasks"}
          </h2>
          <TaskList
            tasks={tasksQuery.data.tasks}
            isAdmin={isAdmin}
            selectedId={selectedId}
            onSelect={(id) => {
              setSelectedId(id);
              if (id) {
                setSearchParams({ task: id });
              } else {
                setSearchParams({});
              }
            }}
            onEdit={(task) => {
              setEditing(task);
              setShowForm(true);
            }}
          />
        </div>

        {selectedId ? (
          <div>
            <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-muted">
              Run history
            </h2>
            <TaskRunHistory
              runs={runsQuery.data?.runs ?? []}
              isLoading={runsQuery.isLoading}
            />
          </div>
        ) : null}
      </section>
    </>
  );
}
