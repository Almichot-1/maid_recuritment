import { AxiosError } from 'axios';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import api from '@/lib/api';
import { useCurrentUser } from '@/hooks/use-auth';
import { usePairingStore } from '@/stores/pairing-store';
import { Notification, UserRole } from '@/types';

export function useNotifications(unreadOnly: boolean = false) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.ETHIOPIAN_AGENT || user?.role === UserRole.FOREIGN_AGENT;

  return useQuery({
    queryKey: ['notifications', unreadOnly],
    queryFn: async () => {
      const response = await api.get<{ notifications: Notification[] }>('/notifications', { params: { unread_only: unreadOnly } });
      return response.data.notifications;
    },
    enabled: Boolean(user) && (!requiresWorkspace || (isPairingReady && Boolean(activePairingId))),
    refetchInterval: 30000, // Refetch organically every 30 seconds
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
  const { data: notifications = [] } = useNotifications(true);
  return { count: notifications.length };
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
      toast.success('All notifications marked as read');
    },
    onError: () => {
      toast.error('Failed to mark all notifications as read');
    },
  });
}
