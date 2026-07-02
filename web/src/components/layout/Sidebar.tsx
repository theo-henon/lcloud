import { NavLink } from "react-router-dom";
import { GlobalSearch } from "@/components/layout/GlobalSearch";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/store/auth";

type NavItem = {
  label: string;
  to: string;
  soon?: boolean;
  adminOnly?: boolean;
};

type NavGroup = {
  title?: string;
  items: NavItem[];
};

const navGroups: NavGroup[] = [
  {
    items: [{ label: "Dashboard", to: "/dashboard" }],
  },
  {
    title: "Storage",
    items: [
      { label: "Volumes", to: "/volumes" },
      { label: "Monitoring", to: "/monitoring" },
    ],
  },
  {
    title: "Extend",
    items: [
      { label: "Plugins", to: "/plugins" },
      { label: "Tasks", to: "/tasks" },
    ],
  },
  {
    title: "System",
    items: [
      { label: "Settings", to: "/settings" },
      { label: "Users", to: "/settings/users", soon: true, adminOnly: true },
    ],
  },
];

export function Sidebar() {
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  return (
    <aside className="flex h-full w-64 flex-col border-r border-hairline bg-surface-soft">
      <div className="border-b border-hairline px-6 py-5">
        <div className="text-xs font-semibold uppercase tracking-[0.2em] text-muted">
          lcloud
        </div>
        <div className="mt-1 text-lg font-bold text-ink">Control panel</div>
      </div>

      <GlobalSearch />

      <nav className="flex-1 space-y-6 overflow-y-auto px-3 py-4">
        {navGroups.map((group) => {
          const items = group.items.filter(
            (item) => !item.adminOnly || user?.role === "admin",
          );
          if (items.length === 0) {
            return null;
          }

          return (
            <div key={group.title ?? "root"}>
              {group.title ? (
                <div className="mb-2 px-3 text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-soft">
                  {group.title}
                </div>
              ) : null}
              <div className="space-y-1">
                {items.map((item) => (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    className={({ isActive }) =>
                      cn(
                        "flex items-center justify-between rounded-md px-3 py-2 text-sm font-medium transition-colors",
                        isActive
                          ? "bg-surface-elevated text-primary"
                          : "text-body hover:bg-surface-elevated hover:text-ink",
                      )
                    }
                  >
                    <span>{item.label}</span>
                    {item.soon ? <Badge>Soon</Badge> : null}
                  </NavLink>
                ))}
              </div>
            </div>
          );
        })}
      </nav>

      <div className="border-t border-hairline p-4">
        <div className="mb-3 truncate text-sm text-body">{user?.email}</div>
        <Button variant="outline" className="w-full" onClick={() => void logout()}>
          Logout
        </Button>
      </div>
    </aside>
  );
}
