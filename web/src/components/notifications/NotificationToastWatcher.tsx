import { useEffect, useRef } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { useAuthStore } from "@/store/auth";
import { useToastStore } from "@/store/toast";

const POLL_INTERVAL_MS = 30_000;

/** Single mount: session baseline + toast push for new unread notifications. */
export function NotificationToastWatcher() {
  const userId = useAuthStore((state) => state.user?.id);
  const pushToast = useToastStore((state) => state.pushFromNotification);
  const sessionUserId = useRef<string | null>(null);
  const sessionBaselineIds = useRef<Set<string> | null>(null);

  const notificationsQuery = useQuery({
    queryKey: ["notifications"],
    queryFn: () => api.listNotifications({ limit: 30 }),
    refetchInterval: POLL_INTERVAL_MS,
    refetchOnWindowFocus: true,
    enabled: Boolean(userId),
  });

  useEffect(() => {
    if (!userId) {
      sessionUserId.current = null;
      sessionBaselineIds.current = null;
      return;
    }

    const items = notificationsQuery.data?.notifications;
    if (!items) {
      return;
    }

    if (sessionUserId.current !== userId) {
      sessionUserId.current = userId;
      sessionBaselineIds.current = new Set(items.map((item) => item.id));
      return;
    }

    if (sessionBaselineIds.current === null) {
      sessionBaselineIds.current = new Set(items.map((item) => item.id));
      return;
    }

    for (const item of items) {
      if (!item.read_at && !sessionBaselineIds.current.has(item.id)) {
        sessionBaselineIds.current.add(item.id);
        pushToast(item);
      }
    }
  }, [userId, notificationsQuery.data, pushToast]);

  return null;
}
