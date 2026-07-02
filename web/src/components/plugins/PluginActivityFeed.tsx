import type { PluginLogEntry } from "@/lib/api";
import { Card } from "@/components/ui/card";
import { cn } from "@/lib/utils";

const levelStyles: Record<string, string> = {
  info: "text-body",
  warn: "text-accent-amber",
  error: "text-accent-rose",
};

type PluginActivityFeedProps = {
  entries: PluginLogEntry[];
};

export function PluginActivityFeed({ entries }: PluginActivityFeedProps) {
  if (entries.length === 0) {
    return (
      <Card className="p-6 text-sm text-muted">
        No plugin activity yet. Upload a file to a volume with an active plugin.
      </Card>
    );
  }

  return (
    <Card className="overflow-hidden">
      <div className="divide-y divide-hairline">
        {entries.map((entry) => (
          <div key={entry.id} className="grid gap-2 px-4 py-3 md:grid-cols-[160px_140px_1fr]">
            <time className="font-mono text-xs text-muted">
              {new Date(entry.created_at).toLocaleString()}
            </time>
            <div className="text-sm text-ink">{entry.plugin_id}</div>
            <div className="space-y-1">
              <p className={cn("text-sm", levelStyles[entry.level] ?? levelStyles.info)}>
                {entry.message}
              </p>
              {entry.event_type ? (
                <p className="font-mono text-xs text-muted-soft">{entry.event_type}</p>
              ) : null}
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}
