import type { PluginStatus } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const statusStyles: Record<PluginStatus, string> = {
  running: "border-accent-emerald/30 bg-accent-emerald/10 text-accent-emerald",
  stopped: "border-hairline bg-surface-soft text-muted",
  error: "border-accent-rose/30 bg-accent-rose/10 text-accent-rose",
};

export function PluginStatusBadge({ status }: { status: PluginStatus }) {
  return (
    <Badge className={cn("capitalize", statusStyles[status] ?? statusStyles.stopped)}>
      {status}
    </Badge>
  );
}
