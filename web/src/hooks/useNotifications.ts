import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type NotificationRecord } from "@/lib/api";

const POLL_INTERVAL_MS = 30_000;

function useInvalidateNotifications() {
  const queryClient = useQueryClient();
  return () => {
    void queryClient.invalidateQueries({ queryKey: ["notifications"] });
  };
}

export function useNotificationMutations() {
  const invalidate = useInvalidateNotifications();

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

  return { markRead, markAllRead, dismiss };
}

export function useNotifications() {
  const { markRead, markAllRead, dismiss } = useNotificationMutations();

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
