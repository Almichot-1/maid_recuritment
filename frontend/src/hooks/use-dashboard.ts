import { useQuery } from "@tanstack/react-query";

import api from "@/lib/api";
import { useCurrentUser } from "@/hooks/use-auth";
import { CandidateApiResponse, normalizeCandidate } from "@/hooks/use-candidates";
import { usePairingStore } from "@/stores/pairing-store";
import { Candidate, DashboardSmartAlerts, UserRole } from "@/types";

export interface DashboardStats {
  totalCandidates: number;
  availableCandidates: number;
  selectedCandidates: number;
  inProgress: number;
  approved: number;
  activeSelections: number;
}

export interface DashboardPendingActions {
  incompleteProfiles: number;
  activeSelections: number;
}

export interface DashboardSelectionPreview {
  id: string;
  candidate_id: string;
  candidate_name: string;
  status: string;
  expires_at?: string;
  created_at: string;
}

export interface DashboardHomeResponse {
  stats: DashboardStats;
  pending_actions: DashboardPendingActions;
  recent_candidates?: CandidateApiResponse[];
  available_candidates?: CandidateApiResponse[];
  active_selections?: DashboardSelectionPreview[];
  approved_selections?: DashboardSelectionPreview[];
}

export interface DashboardHomeData {
  stats: DashboardStats;
  pending_actions: DashboardPendingActions;
  recent_candidates: Candidate[];
  available_candidates: Candidate[];
  active_selections: DashboardSelectionPreview[];
  approved_selections: DashboardSelectionPreview[];
}

function useDashboardBaseQuery<T>(key: string, select?: (data: DashboardHomeResponse) => T) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.ETHIOPIAN_AGENT || user?.role === UserRole.FOREIGN_AGENT;

  return useQuery({
    queryKey: ['dashboard-home', key, activePairingId],
    queryFn: async (): Promise<DashboardHomeResponse> => {
      const response = await api.get<DashboardHomeResponse>('/dashboard/home');
      return response.data;
    },
    select,
    enabled: Boolean(user) && (!requiresWorkspace || (isPairingReady && Boolean(activePairingId))),
    staleTime: 60_000,
    refetchInterval: 120_000,
    refetchOnWindowFocus: false,
  });
}

export function useDashboardHome() {
  return useDashboardBaseQuery('full', (data) => ({
    stats: data.stats,
    pending_actions: data.pending_actions,
    recent_candidates: (data.recent_candidates ?? []).map(normalizeCandidate),
    available_candidates: (data.available_candidates ?? []).map(normalizeCandidate),
    active_selections: data.active_selections ?? [],
    approved_selections: data.approved_selections ?? [],
  }) as DashboardHomeData);
}

export function useDashboardStats() {
  return useDashboardBaseQuery('stats', (data) => data.stats);
}

export function useDashboardSelections() {
  return useDashboardBaseQuery('selection-previews', (data) => ({
    active: data.active_selections ?? [],
    approved: data.approved_selections ?? [],
  }));
}

export function useSmartAlerts() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const canQuery =
    user?.role === UserRole.ETHIOPIAN_AGENT &&
    isPairingReady &&
    Boolean(activePairingId);

  return useQuery({
    queryKey: ["dashboard-smart-alerts", activePairingId],
    queryFn: async () => {
      const response = await api.get<DashboardSmartAlerts>("/dashboard/smart-alerts");
      return response.data;
    },
    enabled: canQuery,
    staleTime: 60_000,
    refetchInterval: 300_000,
    refetchOnWindowFocus: false,
  });
}
