import { useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type NotificationRecord } from "@/lib/api";
import { useToastStore } from "@/store/toast";

const POLL_INTERVAL_MS = 30_000;

export function useNotifications() {
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushFromNotification);
  const sessionBaselineIds = useRef<Set<string> | null>(null);

  const notificationsQuery = useQuery({
    queryKey: ["notifications"],
    queryFn: () => api.listNotifications({ limit: 30 }),
    refetchInterval: POLL_INTERVAL_MS,
    refetchOnWindowFocus: true,
  });

  const unreadQuery = useQuery({
    queryKey: ["notifications", "unread-count"],
    queryFn: () => api.getNotificationUnreadCount(),
    refetchInterval: POLL_INTERVAL_MS,
    refetchOnWindowFocus: true,
  });

  useEffect(() => {
    const items = notificationsQuery.data?.notifications;
    if (!items) {
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
  }, [notificationsQuery.data, pushToast]);

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ["notifications"] });
  };

  const markRead = useMutation({
    mutationFn: (id: string) => api.markNotificationRead(id),
    onSuccess: invalidate,
  });

  const markAllRead = useMutation({
    mutationFn: () => api.markAllNotificationsRead(),
    onSuccess: invalidate,
  });

  const dismiss = useMutation({
    mutationFn: (id: string) => api.dismissNotification(id),
    onSuccess: invalidate,
  });

  return {
    notifications: notificationsQuery.data?.notifications ?? [],
    total: notificationsQuery.data?.total ?? 0,
    unreadCount: unreadQuery.data?.count ?? 0,
    isLoading: notificationsQuery.isLoading,
    markRead,
    markAllRead,
    dismiss,
  };
}

export type { NotificationRecord };
