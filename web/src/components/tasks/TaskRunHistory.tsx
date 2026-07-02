import type { TaskRunRecord } from "@/lib/api";
import { TaskRunStatusBadge } from "@/components/tasks/TaskRunStatusBadge";
import { Card } from "@/components/ui/card";

type TaskRunHistoryProps = {
  runs: TaskRunRecord[];
  isLoading?: boolean;
};

export function TaskRunHistory({ runs, isLoading }: TaskRunHistoryProps) {
  if (isLoading) {
    return <Card className="p-6 text-sm text-muted">Loading run history...</Card>;
  }

  if (runs.length === 0) {
    return (
      <Card className="p-6 text-sm text-muted">
        No runs yet — trigger a manual run or wait for the schedule.
      </Card>
    );
  }

  return (
    <Card className="overflow-hidden">
      <table className="w-full text-sm">
        <thead className="border-b border-hairline bg-surface-soft text-left text-xs uppercase tracking-wide text-muted">
          <tr>
            <th className="px-4 py-3">Started</th>
            <th className="px-4 py-3">Status</th>
            <th className="px-4 py-3">Affected</th>
            <th className="px-4 py-3">Duration</th>
            <th className="px-4 py-3">Message</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.id} className="border-b border-hairline last:border-0">
              <td className="px-4 py-3 font-mono text-xs text-body">
                {new Date(run.started_at).toLocaleString()}
              </td>
              <td className="px-4 py-3">
                <TaskRunStatusBadge status={run.status} />
              </td>
              <td className="px-4 py-3 text-body">{run.affected_count}</td>
              <td className="px-4 py-3 font-mono text-xs text-muted">
                {run.duration_ms}ms
              </td>
              <td className="px-4 py-3 text-body">
                {run.message}
                {run.error ? (
                  <span className="mt-1 block text-xs text-accent-rose">{run.error}</span>
                ) : null}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}
