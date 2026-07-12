import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { toast } from 'sonner';
import api from '@/lib/api';
import type { SelectionProgress, BatchProgressUpdateResult } from '@/types';

export interface UpdateProgressPayload {
  coc_status?: string;
  coc_type?: string;
  medical_status?: string;
  visa_status?: string;
  ticket_status?: string;
  arrival_status?: string;
  arrival_date?: string;
  arrival_city?: string;
  destination_country?: string;
  departure_date?: string;
}

export function useSelectionProgress(selectionId: string | undefined) {
  return useQuery({
    queryKey: ['selection-progress', selectionId],
    queryFn: async () => {
      if (!selectionId) throw new Error('Selection ID required');

      const response = await api.get(`/selections/${selectionId}/progress`);
      return (response.data as { progress: SelectionProgress }).progress;
    },
    enabled: !!selectionId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useUpdateProgress(selectionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: UpdateProgressPayload) => {
      const response = await api.put(`/selections/${selectionId}/progress`, payload);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection-progress', selectionId] });
      queryClient.invalidateQueries({ queryKey: ['selection', selectionId] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      queryClient.invalidateQueries({ queryKey: ['candidate'] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      toast.success('Status step updated successfully');
    },
    onError: (error) => {
      const response = (error as AxiosError<{ error?: string }>).response;
      toast.error(response?.data?.error || 'Failed to update status step');
    },
  });
}

export function useDeleteProgressDocument(selectionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (documentType: string) => {
      const response = await api.delete(`/selections/${selectionId}/progress/documents/${documentType}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection-progress', selectionId] });
      toast.success('Document removed successfully');
    },
    onError: (error) => {
      const response = (error as AxiosError<{ error?: string }>).response;
      toast.error(response?.data?.error || 'Failed to remove document');
    },
  });
}

export function useUploadProgressDocument(selectionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ documentType, file }: { documentType: string; file: File }) => {
      const formData = new FormData();
      formData.append('file', file);

      const response = await api.post(`/selections/${selectionId}/progress/documents/${documentType}`, formData);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['selection-progress', selectionId] });
      toast.success('Document uploaded successfully');
    },
    onError: (error) => {
      const response = (error as AxiosError<{ error?: string }>).response;
      toast.error(response?.data?.error || 'Failed to upload document');
    },
  });
}

export function useBatchUpdateProgress() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: { selection_ids: string[] } & UpdateProgressPayload) => {
      const response = await api.post('/selections/progress/batch', payload);
      return response.data as BatchProgressUpdateResult;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['selection-progress'] });
      queryClient.invalidateQueries({ queryKey: ['my-selections'] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });

      if (data.failed && data.failed.length > 0) {
        const firstError = data.failed[0].error;
        toast.error(`${data.failed.length} selection(s) failed: ${firstError}`);
      }
      if (data.updated > 0) {
        toast.success(`${data.updated} selection(s) updated successfully`);
      }
    },
    onError: (error) => {
      const response = (error as AxiosError<{ error?: string }>).response;
      toast.error(response?.data?.error || 'Batch update failed');
    },
  });
}
