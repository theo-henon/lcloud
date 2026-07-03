import { create } from "zustand";
import type { NotificationType } from "@/lib/api";

export type ToastSeverity = "warning" | "error" | "primary";

export type ToastItem = {
  id: string;
  notificationId: string;
  title: string;
  body: string;
  linkPath: string;
  severity: ToastSeverity;
};

const AUTO_DISMISS_MS = 5000;

function severityForType(type: NotificationType): ToastSeverity {
  switch (type) {
    case "volume.usage_alert":
      return "warning";
    case "task.failed":
      return "error";
    case "volume.deletion_requested":
      return "primary";
    default:
      return "primary";
  }
}

type ToastState = {
  items: ToastItem[];
  pushFromNotification: (notification: {
    id: string;
    type: NotificationType;
    title: string;
    body: string;
    link_path: string;
  }) => void;
  dismiss: (id: string) => void;
};

export const useToastStore = create<ToastState>((set, get) => ({
  items: [],

  pushFromNotification(notification) {
    const toastId = `${notification.id}-${Date.now()}`;
    const item: ToastItem = {
      id: toastId,
      notificationId: notification.id,
      title: notification.title,
      body: notification.body,
      linkPath: notification.link_path,
      severity: severityForType(notification.type),
    };
    set({ items: [...get().items, item] });
    window.setTimeout(() => {
      get().dismiss(toastId);
    }, AUTO_DISMISS_MS);
  },

  dismiss(id) {
    set({ items: get().items.filter((item) => item.id !== id) });
  },
}));

export { severityForType };
