import { Link } from "react-router-dom";
import { WidgetShell } from "@/components/dashboard/WidgetShell";
import type { WidgetProps } from "@/lib/dashboard/types";
import { useAuthStore } from "@/store/auth";
import { cn } from "@/lib/utils";

export function QuickActionsWidget(_props: WidgetProps) {
  const user = useAuthStore((state) => state.user);
  const isAdmin = user?.role === "admin";

  const links = [
    { to: "/volumes", label: isAdmin ? "Volumes" : "New volume" },
    { to: "/monitoring", label: "Monitoring" },
    { to: "/tasks", label: "Tasks" },
    { to: "/plugins", label: "Plugins" },
  ];

  return (
    <WidgetShell title="Quick actions">
      <div className="flex flex-wrap gap-2">
        {links.map((link) => (
          <Link
            key={link.to}
            to={link.to}
            className={cn(
              "inline-flex h-9 items-center justify-center rounded-md border border-hairline px-3 text-xs font-semibold text-body transition-colors",
              "hover:border-hairline-strong hover:text-ink",
            )}
          >
            {link.label}
          </Link>
        ))}
      </div>
    </WidgetShell>
  );
}
