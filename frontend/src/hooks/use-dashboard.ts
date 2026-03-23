import { useQuery } from '@tanstack/react-query';
import api from '@/lib/api';
import { useCurrentUser } from '@/hooks/use-auth';
import { usePairingStore } from '@/stores/pairing-store';
import { UserRole } from '@/types';

export interface DashboardStats {
  totalCandidates: number;
  availableCandidates: number;
  selectedCandidates: number;
  inProgress: number;
  approved: number;
  activeSelections: number;
}

export function useDashboardStats() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.ETHIOPIAN_AGENT || user?.role === UserRole.FOREIGN_AGENT;

  return useQuery({
    queryKey: ['dashboard-stats', activePairingId],
    queryFn: async (): Promise<DashboardStats> => {
      const response = await api.get<DashboardStats>('/dashboard/stats');
      return response.data;
    },
    enabled: Boolean(user) && (!requiresWorkspace || (isPairingReady && Boolean(activePairingId))),
    staleTime: 60000,
    refetchInterval: 120000,
    refetchOnWindowFocus: false,
  });
}
