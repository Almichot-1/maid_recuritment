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
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection', selectionId] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      toast.success('Employer document uploaded successfully');
    },
    onError: (error: unknown) => {
      const response = axios.isAxiosError<ApiErrorResponse>(error) ? error.response : undefined;
      toast.error(response?.data?.error || response?.data?.message || 'Failed to upload employer document');
    },
  });
}

export function useMySelections() {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  return useQuery({
    queryKey: ['my-selections', activePairingId],
    queryFn: async () => {
      const response = await api.get<{ selections: Selection[] }>('/selections/my');
      return response.data.selections;
    },
    enabled: !!user && isPairingReady && !!activePairingId,
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}

export function useSelection(id?: string) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  return useQuery({
    queryKey: ['selection', id, activePairingId],
    queryFn: async () => {
      const response = await api.get<{ selection: Selection }>(`/selections/${id}`);
      return response.data.selection;
    },
    enabled: !!id && !!user && isPairingReady && !!activePairingId,
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}

export function useSelectionApprovals(id?: string) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);

  return useQuery({
    queryKey: ['selection-approvals', id, activePairingId],
    queryFn: async () => {
      const response = await api.get<SelectionApprovalStatus>(`/selections/${id}/approvals`);
      return response.data;
    },
    enabled: !!id && !!user && isPairingReady && !!activePairingId,
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
