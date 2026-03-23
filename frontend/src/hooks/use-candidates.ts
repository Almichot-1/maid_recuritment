import { AxiosError } from 'axios';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import api from '@/lib/api';
import { useCurrentUser } from '@/hooks/use-auth';
import { usePairingStore } from '@/stores/pairing-store';
import { Candidate, PaginatedResponse, UserRole } from '@/types';

export interface CandidateFilters {
  status?: string;
  search?: string;
  min_age?: number;
  max_age?: number;
  min_experience?: number;
  max_experience?: number;
  languages?: string;
  shared_only?: boolean;
  page?: number;
  page_size?: number;
}

export interface UploadCandidateDocumentArgs {
  file: File;
  type: string;
  onProgress?: (progress: number) => void;
}

export interface GenerateCVRequest {
  branding_logo_data_url?: string;
  company_name?: string;
}

interface CandidateDocumentApiResponse {
  id: string;
  document_type: string;
  file_url: string;
  file_name: string;
  file_size?: number;
  uploaded_at?: string;
}

interface CandidateApiResponse {
  id: string;
  full_name: string;
  age?: number;
  experience_years?: number;
  languages?: string[];
  skills?: string[];
  status: Candidate['status'];
  created_by?: string | { id: string };
  cv_pdf_url?: string;
  locked_by?: string;
  locked_at?: string;
  lock_expires_at?: string;
  documents?: CandidateDocumentApiResponse[];
  created_at: string;
  updated_at: string;
}

interface CandidateListApiResponse {
  candidates: CandidateApiResponse[];
  meta: {
    page: number;
    page_size: number;
    count: number;
  };
}

function normalizeCandidate(candidate: CandidateApiResponse): Candidate {
  const createdBy = typeof candidate.created_by === 'string'
    ? candidate.created_by
    : candidate.created_by?.id || '';

  return {
    id: candidate.id,
    full_name: candidate.full_name,
    age: candidate.age,
    experience_years: candidate.experience_years,
    languages: candidate.languages || [],
    skills: candidate.skills || [],
    status: candidate.status,
    created_by: createdBy,
    cv_pdf_url: candidate.cv_pdf_url,
    locked_by: candidate.locked_by,
    locked_at: candidate.locked_at,
    lock_expires_at: candidate.lock_expires_at,
    documents: (candidate.documents || []).map((document) => ({
      id: document.id,
      candidate_id: candidate.id,
      document_type: document.document_type,
      file_url: document.file_url,
      file_name: document.file_name,
      file_size: document.file_size,
      uploaded_at: document.uploaded_at,
    })),
    created_at: candidate.created_at,
    updated_at: candidate.updated_at,
  };
}

export function useCandidates(filters: CandidateFilters = {}) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.FOREIGN_AGENT;

  return useQuery({
    queryKey: ['candidates', activePairingId, filters],
    queryFn: async () => {
      const response = await api.get<CandidateListApiResponse>('/candidates', { params: filters });
      return {
        data: response.data.candidates.map(normalizeCandidate),
        meta: {
          page: response.data.meta.page,
          page_size: response.data.meta.page_size,
          total: response.data.meta.count,
        },
      } satisfies PaginatedResponse<Candidate>;
    },
    enabled: !!user && (!requiresWorkspace || (isPairingReady && !!activePairingId)),
    staleTime: 30000,
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status;
      if (status === 401 || status === 403) {
        return false;
      }

      return failureCount < 2;
    },
  });
}

export function useCandidate(id?: string) {
  const { user } = useCurrentUser();
  const activePairingId = usePairingStore((state) => state.activePairingId);
  const isPairingReady = usePairingStore((state) => state.isReady);
  const requiresWorkspace = user?.role === UserRole.FOREIGN_AGENT;

  return useQuery({
    queryKey: ['candidate', id, activePairingId],
    queryFn: async () => {
      const response = await api.get<{ candidate: CandidateApiResponse }>(`/candidates/${id}`);
      return normalizeCandidate(response.data.candidate);
    },
    enabled: !!id && !!user && (!requiresWorkspace || (isPairingReady && !!activePairingId)),
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status;
      if (status === 401 || status === 403 || status === 404) {
        return false;
      }

      return failureCount < 2;
    },
  });
}

export function useCreateCandidate() {
  return useMutation({
    mutationFn: async (data: Partial<Candidate>) => {
      const response = await api.post<{ candidate: { id: string } }>('/candidates', data);
      return response.data;
    },
    onSuccess: () => {
      toast.success('Candidate created successfully');
    },
    onError: () => {
      toast.error('Failed to create candidate');
    },
  });
}

export function useUpdateCandidate(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Partial<Candidate>) => {
      const response = await api.put<{ candidate: Candidate }>(`/candidates/${id}`, data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['candidate', id] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      toast.success('Candidate updated successfully');
    },
    onError: () => {
      toast.error('Failed to update candidate');
    },
  });
}

export function usePublishCandidate(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await api.post(`/candidates/${id}/publish`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['candidate', id] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      toast.success('Candidate published successfully');
    },
    onError: () => {
      toast.error('Failed to publish candidate');
    },
  });
}

export function useUploadDocument(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (args: UploadCandidateDocumentArgs) => uploadCandidateDocumentFile(id, args),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['candidate', id] });
      toast.success('Document uploaded successfully');
    },
    onError: () => {
      toast.error('Failed to upload document');
    },
  });
}

export async function uploadCandidateDocumentFile(id: string, { file, type, onProgress }: UploadCandidateDocumentArgs) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('document_type', type);

  const response = await api.post(`/candidates/${id}/documents`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    onUploadProgress: (progressEvent) => {
      if (progressEvent.total && onProgress) {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        onProgress(percentCompleted);
      }
    },
  });
  return response.data;
}

export async function publishCandidateById(id: string) {
  const response = await api.post(`/candidates/${id}/publish`);
  return response.data;
}

export function useGenerateCV(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload?: GenerateCVRequest) => {
      const response = await api.post(`/candidates/${id}/generate-cv`, payload || {});
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['candidate', id] });
      toast.success('CV generated successfully');
    },
    onError: () => {
      toast.error('Failed to generate CV');
    },
  });
}

export function useDeleteCandidate(id: string) {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: async () => {
      const response = await api.delete(`/candidates/${id}`);
      return response.data;
    },
    onSuccess: () => {
      queryClient.cancelQueries({ queryKey: ['candidate', id] });
      queryClient.cancelQueries({ queryKey: ['candidate-progress', id] });
      queryClient.cancelQueries({ queryKey: ['candidate-shares', id] });
      queryClient.removeQueries({ queryKey: ['candidate', id] });
      queryClient.removeQueries({ queryKey: ['candidate-progress', id] });
      queryClient.removeQueries({ queryKey: ['candidate-shares', id] });
      queryClient.invalidateQueries({ queryKey: ['candidates'] });
      toast.success('Candidate deleted successfully');
      router.push('/candidates');
    },
    onError: (error) => {
      const responseError = error as AxiosError<{ error?: string }>;
      const message = responseError.response?.data?.error;
      toast.error(message || 'Failed to delete candidate');
    },
  });
}
