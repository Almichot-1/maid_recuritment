import { AxiosError } from 'axios';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import api from '@/lib/api';
import { useCurrentUser } from '@/hooks/use-auth';
import { usePairingStore } from '@/stores/pairing-store';
import { Notification, UserRole } from '@/types';

interface NotificationApiResponse {
  notifications: Notification[];
  unread_count: number;
  pagination?: {
    page: number;
    page_size: number;
    total: number;
  };
}

interface NotificationSummaryResponse {
  unread_count: number;
}

interface UseNotificationsOptions {
  enabled?: boolean;
  pageSize?: number;
  refetchInterval?: number | false;
}

export function useNotifications(unreadOnly: boolean = false, options: UseNotificationsOptions = {}) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.ETHIOPIAN_AGENT || user?.role === UserRole.FOREIGN_AGENT;
  const { enabled = true, pageSize, refetchInterval = 60000 } = options;

  return useQuery({
    queryKey: ['notifications', unreadOnly, pageSize],
    queryFn: async () => {
      const response = await api.get<NotificationApiResponse>('/notifications', {
        params: {
          unread_only: unreadOnly,
          ...(pageSize ? { page_size: pageSize } : {}),
        },
      });
      return response.data;
    },
    enabled: enabled && Boolean(user) && (!requiresWorkspace || (isPairingReady && Boolean(activePairingId))),
    staleTime: 60000,
    refetchInterval,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status;
      if (status === 401 || status === 403) {
        return false;
      }

      return failureCount < 2;
    },
  });
}

export function useUnreadCount() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.ETHIOPIAN_AGENT || user?.role === UserRole.FOREIGN_AGENT;

  const query = useQuery({
    queryKey: ['notifications', 'summary', activePairingId],
    queryFn: async () => {
      const response = await api.get<NotificationSummaryResponse>('/notifications/summary');
      return response.data;
    },
    enabled: Boolean(user) && (!requiresWorkspace || (isPairingReady && Boolean(activePairingId))),
    staleTime: 30_000,
    refetchInterval: 60_000,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status;
      if (status === 401 || status === 403) {
        return false;
      }

      return failureCount < 1;
    },
  });

  return { count: query.data?.unread_count ?? 0 };
}

export function useMarkAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await api.patch(`/notifications/${id}/read`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'summary'] });
    },
    onError: () => {
      toast.error('Failed to mark notification as read');
    },
  });
}

export function useMarkAllAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await api.post('/notifications/mark-all-read');
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'summary'] });
      toast.success('All notifications marked as read');
    },
    onError: () => {
      toast.error('Failed to mark all notifications as read');
    },
  });
}
