import { useNavigate } from "react-router-dom";
import { X } from "lucide-react";
import { useNotifications, type NotificationRecord } from "@/hooks/useNotifications";
import { cn } from "@/lib/utils";

function formatRelativeTime(iso: string): string {
  const date = new Date(iso);
  const diffMs = Date.now() - date.getTime();
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

type NotificationPanelProps = {
  onNavigate?: () => void;
};

export function NotificationPanel({ onNavigate }: NotificationPanelProps) {
  const navigate = useNavigate();
  const { notifications, isLoading, unreadCount, markRead, markAllRead, dismiss } =
    useNotifications();

  const openNotification = (item: NotificationRecord) => {
    if (!item.read_at) {
      markRead.mutate(item.id);
    }
    if (item.link_path) {
      navigate(item.link_path);
      onNavigate?.();
    }
  };

  return (
    <div className="w-[360px] overflow-hidden rounded-md border border-hairline bg-surface-elevated shadow-lg">
      <div className="flex items-center justify-between border-b border-hairline px-4 py-3">
        <span className="text-sm font-semibold text-ink">Notifications</span>
        {unreadCount > 0 ? (
          <button
            type="button"
            className="text-xs font-medium text-primary hover:underline disabled:opacity-50"
            disabled={markAllRead.isPending}
            onClick={() => markAllRead.mutate()}
          >
            Mark all read
          </button>
        ) : null}
      </div>

      <div className="max-h-[420px] overflow-y-auto">
        {isLoading ? (
          <p className="px-4 py-6 text-sm text-muted">Loading...</p>
        ) : notifications.length === 0 ? (
          <div className="px-4 py-8 text-center">
            <p className="text-sm font-medium text-ink">No notifications</p>
            <p className="mt-1 text-xs text-muted">Alerts and task failures will appear here.</p>
          </div>
        ) : (
          <ul>
            {notifications.map((item) => (
              <li
                key={item.id}
                className={cn(
                  "group relative border-b border-hairline last:border-b-0",
                  !item.read_at && "border-l-2 border-l-primary bg-surface-card/40",
                )}
              >
                <button
                  type="button"
                  className="w-full px-4 py-3 pr-10 text-left hover:bg-surface-card"
                  onClick={() => openNotification(item)}
                >
                  <p className="text-sm font-semibold text-ink">{item.title}</p>
                  <p className="mt-0.5 line-clamp-2 text-xs text-muted">{item.body}</p>
                  <p className="mt-1 text-[11px] text-muted-soft">
                    {formatRelativeTime(item.created_at)}
                  </p>
                </button>
                <button
                  type="button"
                  aria-label="Dismiss notification"
                  className="absolute top-2 right-2 rounded p-1 text-muted opacity-0 transition-opacity hover:bg-surface-elevated hover:text-ink group-hover:opacity-100"
                  disabled={dismiss.isPending}
                  onClick={(event) => {
                    event.stopPropagation();
                    dismiss.mutate(item.id);
                  }}
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      <p className="border-t border-hairline px-4 py-2 text-[11px] text-muted-soft">
        Dismissed items are hidden
      </p>
    </div>
  );
}
