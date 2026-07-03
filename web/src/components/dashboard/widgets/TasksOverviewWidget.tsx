import { Link } from "react-router-dom";
import { WidgetEmpty, WidgetError, WidgetLoading, WidgetShell } from "@/components/dashboard/WidgetShell";
import { useTasks } from "@/hooks/useTasks";
import type { WidgetProps } from "@/lib/dashboard/types";

export function TasksOverviewWidget(_props: WidgetProps) {
  const { data, isLoading, isError } = useTasks();

  if (isLoading) {
    return (
      <WidgetShell title="Tasks" deepLink="/tasks">
        <WidgetLoading />
      </WidgetShell>
    );
  }

  if (isError || !data) {
    return (
      <WidgetShell title="Tasks" deepLink="/tasks">
        <WidgetError />
      </WidgetShell>
    );
  }

  const tasks = data.tasks;
  if (tasks.length === 0) {
    return (
      <WidgetShell title="Tasks" deepLink="/tasks">
        <WidgetEmpty message="No tasks yet." />
        <Link
          to="/tasks"
          className="mt-3 inline-block text-sm font-medium text-primary hover:text-primary-active"
        >
          Create your first task →
        </Link>
      </WidgetShell>
    );
  }

  const enabledCount = tasks.filter((task) => task.enabled).length;
  const disabledCount = tasks.length - enabledCount;
  const lastFailed = tasks.find((task) => task.last_run_status === "failed");
  const upcoming = tasks
    .filter((task) => task.enabled)
    .sort((a, b) => {
      if (!a.next_run_at) return 1;
      if (!b.next_run_at) return -1;
      return a.next_run_at.localeCompare(b.next_run_at);
    })
    .slice(0, 3);

  return (
    <WidgetShell title="Tasks" deepLink="/tasks">
      <div className="space-y-3">
        <div className="flex gap-4 text-sm">
          <p>
            <span className="font-semibold text-primary">{enabledCount}</span>{" "}
            <span className="text-muted">enabled</span>
          </p>
          <p>
            <span className="font-semibold text-ink">{disabledCount}</span>{" "}
            <span className="text-muted">disabled</span>
          </p>
        </div>
        {lastFailed ? (
          <p className="text-sm text-accent-rose">
            Last failure: <span className="font-medium">{lastFailed.name}</span>
          </p>
        ) : null}
        {upcoming.length > 0 ? (
          <ul className="space-y-1 text-sm text-body">
            {upcoming.map((task) => (
              <li key={task.id} className="truncate">
                {task.name}
                {task.schedule_description ? (
                  <span className="text-muted"> · {task.schedule_description}</span>
                ) : null}
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </WidgetShell>
  );
}
