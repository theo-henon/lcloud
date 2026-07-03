import type { TaskRecord } from "@/lib/api";
import { User } from "lucide-react";
import { macroLabel } from "@/components/tasks/MacroParameterFields";
import { TaskRunStatusBadge } from "@/components/tasks/TaskRunStatusBadge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useDeleteTask, useRunTask, useUpdateTask } from "@/hooks/useTasks";
import { ActionIcon, navIcons } from "@/lib/icons";

type TaskListProps = {
  tasks: TaskRecord[];
  isAdmin: boolean;
  selectedId: string | null;
  onSelect: (id: string) => void;
  onEdit: (task: TaskRecord) => void;
};

function EnableToggle({
  enabled,
  disabled,
  onChange,
}: {
  enabled: boolean;
  disabled?: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={enabled}
      disabled={disabled}
      onClick={() => onChange(!enabled)}
      className={`relative h-6 w-11 rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
        enabled ? "bg-primary" : "bg-surface-elevated"
      }`}
    >
      <span
        className={`absolute top-0.5 left-0.5 h-5 w-5 rounded-full transition-transform ${
          enabled ? "translate-x-5 bg-on-primary" : "bg-ink"
        }`}
      />
    </button>
  );
}

export function TaskList({ tasks, isAdmin, selectedId, onSelect, onEdit }: TaskListProps) {
  const runTask = useRunTask();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();

  if (tasks.length === 0) {
    const EmptyIcon = navIcons["/tasks"];
    return (
      <Card className="flex flex-col items-center gap-3 p-8 text-center text-sm text-muted">
        {EmptyIcon ? <EmptyIcon className="h-8 w-8 text-muted" aria-hidden /> : null}
        <p>
          {isAdmin
            ? "No tasks on this instance yet."
            : "No tasks yet — create one on your volumes."}
        </p>
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      {tasks.map((task) => (
        <Card
          key={task.id}
          className={`p-4 ${selectedId === task.id ? "border-primary/40" : ""}`}
        >
          <div className="flex flex-wrap items-start justify-between gap-4">
            <button type="button" className="text-left" onClick={() => onSelect(task.id)}>
              <div className="flex flex-wrap items-center gap-2">
                <h3 className="text-base font-semibold text-ink">{task.name}</h3>
                {task.last_run_status ? (
                  <TaskRunStatusBadge status={task.last_run_status} />
                ) : null}
              </div>
              <p className="mt-1 text-sm text-body">
                {macroLabel(task.macro)}
                {task.volume_name
                  ? ` · ${task.volume_name}`
                  : task.scope === "global"
                    ? " · Instance-wide"
                    : ""}
              </p>
              {isAdmin && task.owner_email ? (
                <p className="mt-1 flex items-center gap-1.5 text-sm text-muted">
                  <User className="h-3.5 w-3.5 shrink-0" aria-hidden />
                  <span>
                    Created by <span className="text-body">{task.owner_email}</span>
                  </span>
                </p>
              ) : null}
              <p className="mt-1 font-mono text-xs text-muted">
                {task.schedule_description}
                {task.next_run_at ? ` · next ${new Date(task.next_run_at).toLocaleString()} UTC` : ""}
              </p>
            </button>

            <div className="flex flex-wrap items-center gap-2">
              <label className="flex items-center gap-2 text-sm text-body">
                <EnableToggle
                  enabled={task.enabled}
                  disabled={updateTask.isPending}
                  onChange={(enabled) => updateTask.mutate({ id: task.id, input: { enabled } })}
                />
              </label>
              <Button
                variant="outline"
                className="h-8 px-3 text-xs"
                disabled={runTask.isPending || !task.enabled}
                title={task.enabled ? undefined : "Enable the task before running it"}
                onClick={() => {
                  const dryRun = Boolean(task.parameters?.dry_run);
                  if (!dryRun && !window.confirm("Run this task now? Files may be modified.")) {
                    return;
                  }
                  runTask.mutate({ id: task.id, dryRun });
                }}
              >
                <ActionIcon action="run" className="mr-1" />
                Run now
              </Button>
              <Button variant="outline" className="h-8 px-3 text-xs" onClick={() => onEdit(task)}>
                <ActionIcon action="edit" className="mr-1" />
                Edit
              </Button>
              <Button
                variant="outline"
                className="h-8 px-3 text-xs text-accent-rose hover:text-accent-rose"
                disabled={deleteTask.isPending}
                onClick={() => {
                  if (window.confirm("Delete this task?")) {
                    deleteTask.mutate(task.id);
                  }
                }}
              >
                <ActionIcon action="delete" className="mr-1" />
                Delete
              </Button>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}
