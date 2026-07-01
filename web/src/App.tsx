import { useEffect } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AdminRoute, ProtectedRoute } from "@/components/ProtectedRoute";
import { AppShell } from "@/components/layout/AppShell";
import { DashboardPage } from "@/pages/DashboardPage";
import { LoginPage } from "@/pages/LoginPage";
import { NotFoundPage } from "@/pages/NotFoundPage";
import { PlaceholderPage } from "@/pages/PlaceholderPage";
import { useAuthStore } from "@/store/auth";

function AppRoutes() {
  const hydrate = useAuthStore((state) => state.hydrate);

  useEffect(() => {
    void hydrate();
  }, [hydrate]);

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route
            path="/volumes"
            element={
              <PlaceholderPage
                title="Volumes"
                phase="Phase 1.1"
                description="Create and manage storage volumes on your disks."
              />
            }
          />
          <Route
            path="/monitoring"
            element={
              <PlaceholderPage
                title="Monitoring"
                phase="Phase 1.2"
                description="Track space usage and file breakdown across disks."
              />
            }
          />
          <Route
            path="/plugins"
            element={
              <PlaceholderPage
                title="Plugins"
                phase="Phase 1.3"
                description="Extend lcloud with external plugin binaries."
              />
            }
          />
          <Route
            path="/tasks"
            element={
              <PlaceholderPage
                title="Tasks"
                phase="Phase 1.4"
                description="Schedule recurring file operations on your volumes."
              />
            }
          />
          <Route
            path="/settings"
            element={
              <PlaceholderPage
                title="Settings"
                phase="Phase 0"
                description="System preferences will live here."
              />
            }
          />
          <Route element={<AdminRoute />}>
            <Route
              path="/settings/users"
              element={
                <PlaceholderPage
                  title="Users"
                  phase="Phase 0 — API only"
                  description="User management UI arrives later. Admin API is available now."
                />
              }
            />
          </Route>
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Route>
    </Routes>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}
