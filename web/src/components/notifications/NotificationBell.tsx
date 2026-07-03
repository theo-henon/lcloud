import { Bell } from "lucide-react";
import { useState } from "react";
import { NotificationPanel } from "@/components/notifications/NotificationPanel";
import { DropdownMenu } from "@/components/ui/dropdown-menu";
import { useNotifications } from "@/hooks/useNotifications";
import { cn } from "@/lib/utils";

export function NotificationBell() {
  const { unreadCount } = useNotifications();
  const [openKey, setOpenKey] = useState(0);

  const badgeLabel = unreadCount > 9 ? "9+" : String(unreadCount);

  return (
    <div className="border-b border-hairline px-6 py-3">
      <DropdownMenu
        key={openKey}
        align="right"
        trigger={
          <button
            type="button"
            aria-label="Notifications"
            className="relative flex w-full items-center justify-center rounded-md border border-hairline bg-surface-card px-3 py-2 text-body transition-colors hover:border-hairline-strong hover:text-ink"
          >
            <Bell className="h-4 w-4" />
            {unreadCount > 0 ? (
              <span
                className={cn(
                  "absolute -top-1.5 -right-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-on-primary",
                )}
              >
                {badgeLabel}
              </span>
            ) : null}
          </button>
        }
      >
        <NotificationPanel onNavigate={() => setOpenKey((value) => value + 1)} />
      </DropdownMenu>
    </div>
  );
}
