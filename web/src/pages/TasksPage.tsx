import { useState } from "react";
import type { TaskRecord } from "@/lib/api";
import { Header } from "@/components/layout/Header";
import { TaskForm } from "@/components/tasks/TaskForm";
import { TaskList } from "@/components/tasks/TaskList";
import { TaskRunHistory } from "@/components/tasks/TaskRunHistory";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useCreateTask, useTaskRuns, useTasks, useUpdateTask } from "@/hooks/useTasks";
import { useVolumes } from "@/hooks/useVolumes";
import { ActionIcon } from "@/lib/icons";
import { useAuthStore } from "@/store/auth";

export function TasksPage() {
  const user = useAuthStore((state) => state.user);
  const isAdmin = user?.role === "admin";
  const [volumeFilter, setVolumeFilter] = useState<string>("");
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<TaskRecord | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const volumesQuery = useVolumes();
  const tasksQuery = useTasks(volumeFilter || undefined);
  const runsQuery = useTaskRuns(selectedId);
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();

  if (tasksQuery.isLoading || volumesQuery.isLoading) {
    return (
      <>
        <Header title="Tasks" description="Automate recurring file operations on your volumes." />
        <section className="px-8 py-6">
          <p className="text-sm text-muted">Loading tasks...</p>
        </section>
      </>
    );
  }

  if (tasksQuery.isError || volumesQuery.isError || !tasksQuery.data || !volumesQuery.data) {
    return (
      <div className="px-8 py-6">
        <Card className="p-6 text-sm text-accent-rose">Unable to load tasks.</Card>
      </div>
    );
  }

  const volumes = volumesQuery.data.volumes;

  return (
    <>
      <Header
        title="Tasks"
        description="Automate recurring file operations on your volumes."
      />

      <section className="space-y-8 px-8 py-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <label className="text-sm text-muted">Filter by volume</label>
            <select
              className="rounded-md border border-hairline bg-surface px-3 py-2 text-sm text-ink"
              value={volumeFilter}
              onChange={(e) => setVolumeFilter(e.target.value)}
            >
              <option value="">All volumes</option>
              {volumes.map((vol) => (
                <option key={vol.id} value={vol.id}>
                  {vol.name}
                </option>
              ))}
            </select>
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
            volumes={volumes}
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
            Scheduled tasks
          </h2>
          <TaskList
            tasks={tasksQuery.data.tasks}
            selectedId={selectedId}
            onSelect={setSelectedId}
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
