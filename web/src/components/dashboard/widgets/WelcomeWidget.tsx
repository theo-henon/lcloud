import { WidgetShell } from "@/components/dashboard/WidgetShell";
import type { WidgetProps } from "@/lib/dashboard/types";
import { useAuthStore } from "@/store/auth";

export function WelcomeWidget(_props: WidgetProps) {
  const user = useAuthStore((state) => state.user);

  return (
    <WidgetShell title="Welcome">
      <div className="space-y-1">
        <p className="text-sm text-body">
          Hello{user?.email ? `, ${user.email.split("@")[0]}` : ""}.
        </p>
        <p className="text-xs text-muted">Your personal lcloud cockpit.</p>
      </div>
    </WidgetShell>
  );
}
