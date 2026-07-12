import axios from 'axios';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import api from '@/lib/api';
import { useCurrentUser } from '@/hooks/use-auth';
import { usePairingStore } from '@/stores/pairing-store';
import { Selection } from '@/types';

export interface SelectionApprovalRecord {
  user_id: string;
  user_name: string;
  role: string;
  decision: 'approved' | 'rejected';
  decided_at: string;
}

export interface SelectionApprovalStatus {
  selection_id: string;
  status: string;
  approvals: SelectionApprovalRecord[];
  is_fully_approved: boolean;
  pending_approval_from: string[];
}

interface ApiErrorResponse {
  error?: string;
  message?: string;
}

export interface UploadSelectionDocumentArgs {
  file: File;
  type: 'contract' | 'employer_id';
  onProgress?: (progress: number) => void;
}

export function useSelectCandidate(candidateId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await api.post(`/candidates/${candidateId}/select`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['candidate', candidateId] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      toast.success('Candidate selected successfully');
    },
    onError: (error: unknown) => {
      // Refresh candidate data on error to handle race conditions
      queryClient.invalidateQueries({ queryKey: ['candidate', candidateId] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      
      const status = axios.isAxiosError<ApiErrorResponse>(error) ? error.response?.status : undefined;
      if (status === 409) {
        toast.error('Candidate is no longer available');
      } else {
        toast.error('Failed to select candidate');
      }
    },
  });
}

export function useUploadSelectionDocument(selectionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ file, type, onProgress }: UploadSelectionDocumentArgs) => {
      try {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('document_type', type);

        const response = await api.post(`/selections/${selectionId}/documents`, formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
          onUploadProgress: (progressEvent) => {
            if (progressEvent.total && onProgress) {
              const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
              onProgress(progress);
            }
          },
        });

        return response.data;
      } catch (error) {
        if (axios.isAxiosError<ApiErrorResponse>(error)) {
          throw new Error(error.response?.data?.error || error.response?.data?.message || 'Failed to upload employer document');
        }
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection', selectionId] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      toast.success('Employer document uploaded successfully');
    },
    onError: (error: unknown) => {
      if (error instanceof Error) {
        toast.error(error.message);
        return;
      }
      const response = axios.isAxiosError<ApiErrorResponse>(error) ? error.response : undefined;
      toast.error(response?.data?.error || response?.data?.message || 'Failed to upload employer document');
    },
  });
}

export function useMySelections(sortBy?: string, page?: number, pageSize?: number) {
  const { user, isEthiopianAgent } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  // Ethiopian agents see all their selections across all pairings, so don't include pairingId in query key
  // Foreign agents filter by pairing, so include pairingId in query key
  const queryKey = isEthiopianAgent 
    ? ['my-selections', 'ethiopian', sortBy || 'newest', page, pageSize]
    : ['my-selections', activePairingId, sortBy || 'newest', page, pageSize];

  return useQuery({
    queryKey,
    queryFn: async () => {
      const params = new URLSearchParams();
      if (sortBy) {
        params.append('sortBy', sortBy);
      }
      if (page) {
        params.append('page', page.toString());
      }
      if (pageSize) {
        params.append('page_size', pageSize.toString());
      }
      const response = await api.get<{ selections: Selection[]; pagination: { page: number; page_size: number; total: number; has_more: boolean } }>(`/selections/my?${params.toString()}`);
      console.log('[useMySelections] API Response:', response.data);
      console.log('[useMySelections] Selections count:', response.data.selections?.length || 0);
      console.log('[useMySelections] First selection:', response.data.selections?.[0]);
      return response.data;
    },
    // Ethiopian agents own candidates, so they can fetch selections without needing an active pairing.
    // Foreign agents must have an active pairing to see their selections.
    enabled: !!user && isPairingReady && (isEthiopianAgent || !!activePairingId),
    staleTime: 0, // Force fresh data
    gcTime: 0, // Don't cache (replaces deprecated cacheTime)
    refetchOnWindowFocus: true,
    refetchOnMount: true,
  });
}

export function useSelection(id?: string) {
  const { user, isEthiopianAgent } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  // Similar to useMySelections, Ethiopian agents don't need pairingId in query key
  const queryKey = isEthiopianAgent
    ? ['selection', id, 'ethiopian']
    : ['selection', id, activePairingId];

  return useQuery({
    queryKey,
    queryFn: async () => {
      const response = await api.get<{ selection: Selection }>(`/selections/${id}`);
      return response.data.selection;
    },
    enabled: !!id && !!user && isPairingReady && (isEthiopianAgent || !!activePairingId),
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}

export function useSelectionApprovals(id?: string) {
  const { user, isEthiopianAgent } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  // Similar to useMySelections, Ethiopian agents don't need pairingId in query key
  const queryKey = isEthiopianAgent
    ? ['selection-approvals', id, 'ethiopian']
    : ['selection-approvals', id, activePairingId];

  return useQuery({
    queryKey,
    queryFn: async () => {
      const response = await api.get<SelectionApprovalStatus>(`/selections/${id}/approvals`);
      return response.data;
    },
    enabled: !!id && !!user && isPairingReady && (isEthiopianAgent || !!activePairingId),
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}

export function useApproveSelection(id: string, candidateId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await api.post(`/selections/${id}/approve`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection', id] });
      queryClient.invalidateQueries({ queryKey: ['selection-approvals', id] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      if (candidateId) {
        queryClient.invalidateQueries({ queryKey: ['candidate', candidateId] });
        queryClient.invalidateQueries({ queryKey: ['candidate-progress', candidateId] });
      }
      toast.success('Selection approved successfully');
    },
    onError: (error: unknown) => {
      const response = axios.isAxiosError<ApiErrorResponse>(error) ? error.response : undefined;
      toast.error(response?.data?.error || response?.data?.message || 'Failed to approve selection');
    },
  });
}

export function useRejectSelection(id: string, candidateId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (vars: { reason: string }) => {
      const response = await api.post(`/selections/${id}/reject`, vars);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection', id] });
      queryClient.invalidateQueries({ queryKey: ['selection-approvals', id] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      if (candidateId) {
        queryClient.invalidateQueries({ queryKey: ['candidate', candidateId] });
        queryClient.invalidateQueries({ queryKey: ['candidate-progress', candidateId] });
      }
      toast.success('Selection rejected successfully');
    },
    onError: (error: unknown) => {
      const response = axios.isAxiosError<ApiErrorResponse>(error) ? error.response : undefined;
      toast.error(response?.data?.error || response?.data?.message || 'Failed to reject selection');
    },
  });
}

export function useUnlockSelection(selectionId: string, candidateId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await api.post(`/selections/${selectionId}/unlock`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection', selectionId] });
      queryClient.invalidateQueries({ queryKey: ['selection-approvals', selectionId] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      if (candidateId) {
        queryClient.invalidateQueries({ queryKey: ['candidate', candidateId] });
      }
      toast.success('Candidate unlocked successfully');
    },
    onError: (error: unknown) => {
      const response = axios.isAxiosError<ApiErrorResponse>(error) ? error.response : undefined;
      toast.error(response?.data?.error || response?.data?.message || 'Failed to unlock candidate');
    },
  });
}
