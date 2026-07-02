import type { TaskRunStatus } from "@/lib/api";
import { Badge } from "@/components/ui/badge";

const labels: Record<TaskRunStatus, string> = {
  success: "Success",
  failed: "Failed",
  skipped: "Skipped",
  dry_run: "Simulation",
};

const styles: Record<TaskRunStatus, string> = {
  success: "border-accent-emerald/30 text-accent-emerald",
  failed: "border-accent-rose/30 text-accent-rose",
  skipped: "border-hairline text-muted",
  dry_run: "border-primary/30 text-primary",
};

export function TaskRunStatusBadge({ status }: { status: TaskRunStatus }) {
  return (
    <Badge className={styles[status] ?? styles.skipped}>{labels[status] ?? status}</Badge>
  );
}
