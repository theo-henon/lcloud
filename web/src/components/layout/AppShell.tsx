import { Outlet } from "react-router-dom";
import { Sidebar } from "@/components/layout/Sidebar";
import { NotificationToastWatcher } from "@/components/notifications/NotificationToastWatcher";
import { ToastHost } from "@/components/notifications/ToastHost";

export function AppShell() {
  return (
    <div className="flex min-h-screen bg-canvas">
      <Sidebar />
      <div className="flex min-h-screen flex-1 flex-col">
        <Outlet />
      </div>
      <NotificationToastWatcher />
      <ToastHost />
    </div>
  );
}
