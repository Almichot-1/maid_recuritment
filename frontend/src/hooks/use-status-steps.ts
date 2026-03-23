import { AxiosError } from 'axios';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import api from '@/lib/api';
import { useCurrentUser } from '@/hooks/use-auth';
import { usePairingStore } from '@/stores/pairing-store';
import { CandidateProgress, UserRole } from '@/types';

interface StatusStepApiError {
  error?: string;
  message?: string;
}

export function useCandidateProgress(candidateId?: string, enabled: boolean = true) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.FOREIGN_AGENT;
  const canQuery =
    enabled &&
    !!candidateId &&
    !!user &&
    (!requiresWorkspace || (isPairingReady && !!activePairingId));

  return useQuery({
    queryKey: ['candidate-progress', candidateId, activePairingId],
    queryFn: async () => {
      const response = await api.get<CandidateProgress>(`/candidates/${candidateId}/status-steps`);
      return response.data;
    },
    enabled: canQuery,
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status;
      if (status === 403 || status === 404) {
        return false;
      }

      return failureCount < 2;
    },
  });
}

export function useUpdateStatusStep(candidateId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ step_name, status, notes }: { step_name: string; status: string; notes?: string }) => {
      const encodedStepName = encodeURIComponent(step_name);
      const response = await api.patch(`/candidates/${candidateId}/status-steps/${encodedStepName}`, { status, notes });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['candidate-progress', candidateId] });
      queryClient.invalidateQueries({ queryKey: ['candidate', candidateId] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      queryClient.invalidateQueries({ queryKey: ['selection'] });
      toast.success('Status step updated successfully');
    },
    onError: (error) => {
      const response = (error as AxiosError<StatusStepApiError>).response;

      if (response?.status === 403) {
        toast.error('Only the Ethiopian agency that created this candidate can update process steps.');
        return;
      }

      if (response?.status === 400 && response.data?.error) {
        toast.error(response.data.error);
        return;
      }

      toast.error(response?.data?.error || response?.data?.message || 'Failed to update status step');
    },
  });
}
