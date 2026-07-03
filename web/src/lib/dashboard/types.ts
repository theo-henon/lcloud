export type WidgetType =
  | "storage-summary"
  | "volumes-at-risk"
  | "tasks-overview"
  | "plugins-status"
  | "quick-actions"
  | "welcome"
  | "pending-deletions";

export type WidgetPlacement = {
  id: string;
  type: WidgetType;
  x: number;
  y: number;
  w: number;
  h: number;
};

export type DashboardLayout = {
  version: number;
  widgets: WidgetPlacement[];
};

export type CatalogEntry = {
  type: WidgetType;
  title: string;
  description: string;
  default_w: number;
  default_h: number;
  admin_only: boolean;
};

export type DashboardResponse = {
  layout: DashboardLayout;
  catalog: CatalogEntry[];
  updated_at: string | null;
};

export type PatchDashboardInput = {
  layout: DashboardLayout;
};

export type WidgetProps = {
  id: string;
  type: WidgetType;
};
