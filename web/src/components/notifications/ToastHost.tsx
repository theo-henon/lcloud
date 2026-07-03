import { useNavigate } from "react-router-dom";
import { X } from "lucide-react";
import { useNotifications } from "@/hooks/useNotifications";
import { useToastStore, type ToastItem } from "@/store/toast";
import { cn } from "@/lib/utils";

function borderClass(severity: ToastItem["severity"]) {
  switch (severity) {
    case "warning":
      return "border-l-warning";
    case "error":
      return "border-l-error";
    default:
      return "border-l-primary";
  }
}

export function ToastHost() {
  const navigate = useNavigate();
  const items = useToastStore((state) => state.items);
  const dismiss = useToastStore((state) => state.dismiss);
  const { markRead } = useNotifications();

  const handleClick = (item: ToastItem) => {
    markRead.mutate(item.notificationId);
    dismiss(item.id);
    if (item.linkPath) {
      navigate(item.linkPath);
    }
  };

  if (items.length === 0) {
    return null;
  }

  return (
    <div
      className="pointer-events-none fixed right-4 bottom-4 z-[100] flex w-full max-w-sm flex-col gap-2"
      aria-live="polite"
    >
      {items.map((item) => (
        <div
          key={item.id}
          className={cn(
            "pointer-events-auto overflow-hidden rounded-md border border-hairline border-l-4 bg-surface-elevated shadow-lg",
            borderClass(item.severity),
          )}
        >
          <div className="flex items-start gap-2 p-3">
            <button
              type="button"
              className="min-w-0 flex-1 text-left"
              onClick={() => handleClick(item)}
            >
              <p className="text-sm font-semibold text-ink">{item.title}</p>
              <p className="mt-0.5 line-clamp-2 text-xs text-muted">{item.body}</p>
            </button>
            <button
              type="button"
              aria-label="Dismiss toast"
              className="rounded p-1 text-muted hover:bg-surface-card hover:text-ink"
              onClick={() => dismiss(item.id)}
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
