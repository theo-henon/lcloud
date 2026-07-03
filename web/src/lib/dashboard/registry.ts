import type { ComponentType } from "react";
import { PendingDeletionsWidget } from "@/components/dashboard/widgets/PendingDeletionsWidget";
import { PluginsStatusWidget } from "@/components/dashboard/widgets/PluginsStatusWidget";
import { QuickActionsWidget } from "@/components/dashboard/widgets/QuickActionsWidget";
import { StorageSummaryWidget } from "@/components/dashboard/widgets/StorageSummaryWidget";
import { TasksOverviewWidget } from "@/components/dashboard/widgets/TasksOverviewWidget";
import { VolumesAtRiskWidget } from "@/components/dashboard/widgets/VolumesAtRiskWidget";
import { WelcomeWidget } from "@/components/dashboard/widgets/WelcomeWidget";
import type { WidgetProps, WidgetType } from "@/lib/dashboard/types";

export type WidgetDefinitionEntry = {
  type: WidgetType;
  title: string;
  defaultW: number;
  defaultH: number;
  adminOnly?: boolean;
  deepLink?: string;
  component: ComponentType<WidgetProps>;
};

export const WIDGET_REGISTRY: Record<WidgetType, WidgetDefinitionEntry> = {
  "storage-summary": {
    type: "storage-summary",
    title: "Storage",
    defaultW: 6,
    defaultH: 2,
    deepLink: "/monitoring",
    component: StorageSummaryWidget,
  },
  "volumes-at-risk": {
    type: "volumes-at-risk",
    title: "Volumes at risk",
    defaultW: 6,
    defaultH: 2,
    deepLink: "/monitoring",
    component: VolumesAtRiskWidget,
  },
  "tasks-overview": {
    type: "tasks-overview",
    title: "Tasks",
    defaultW: 6,
    defaultH: 2,
    deepLink: "/tasks",
    component: TasksOverviewWidget,
  },
  "plugins-status": {
    type: "plugins-status",
    title: "Plugins",
    defaultW: 4,
    defaultH: 2,
    deepLink: "/plugins",
    component: PluginsStatusWidget,
  },
  "quick-actions": {
    type: "quick-actions",
    title: "Quick actions",
    defaultW: 4,
    defaultH: 1,
    component: QuickActionsWidget,
  },
  welcome: {
    type: "welcome",
    title: "Welcome",
    defaultW: 4,
    defaultH: 1,
    component: WelcomeWidget,
  },
  "pending-deletions": {
    type: "pending-deletions",
    title: "Pending deletions",
    defaultW: 6,
    defaultH: 2,
    deepLink: "/volumes",
    adminOnly: true,
    component: PendingDeletionsWidget,
  },
};

export function getWidgetDefinition(type: WidgetType): WidgetDefinitionEntry {
  return WIDGET_REGISTRY[type];
}

export const ALL_WIDGET_TYPES = Object.keys(WIDGET_REGISTRY) as WidgetType[];
