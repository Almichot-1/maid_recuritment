import axios from 'axios';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import api from '@/lib/api';
import { useAuthStore } from '@/stores/auth-store';
import { User } from '@/types';

export interface ProfileUpdateData {
  full_name: string;
  company_name: string;
}

export interface PasswordChangeData {
  current_password: string;
  new_password: string;
}

export interface PreferencesData {
  theme: 'light' | 'dark' | 'system';
  email_notifications: boolean;
  selection_alerts: boolean;
  status_update_alerts: boolean;
  approval_alerts: boolean;
}

export interface SharingPreferencesData {
  auto_share_candidates: boolean;
  default_foreign_pairing_id?: string | null;
}

interface ApiErrorResponse {
  error?: string;
  message?: string;
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: ProfileUpdateData) => {
      const response = await api.patch<{ user: User }>('/users/profile', data);
      return response.data.user;
    },
    onSuccess: (user) => {
      useAuthStore.getState().updateUser(user);
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Profile updated successfully');
    },
    onError: (error: unknown) => {
      const message = axios.isAxiosError<ApiErrorResponse>(error)
        ? error.response?.data?.error || error.response?.data?.message || error.message
        : error instanceof Error
          ? error.message
          : 'Failed to update profile';
      toast.error(message);
    },
  });
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (_data: PasswordChangeData) => {
      const response = await api.post('/users/change-password', _data);
      return response.data;
    },
    onSuccess: () => {
      toast.success('Password changed successfully');
    },
    onError: (error: unknown) => {
      const message = axios.isAxiosError<ApiErrorResponse>(error)
        ? error.response?.data?.error || error.response?.data?.message || error.message
        : error instanceof Error
          ? error.message
          : 'Failed to change password';
      toast.error(message);
    },
  });
}

export function useUpdatePreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: PreferencesData) => {
      if (typeof window !== 'undefined') {
        localStorage.setItem('user_preferences', JSON.stringify(data));
      }
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Preferences saved successfully');
    },
    onError: () => {
      toast.error('Failed to save preferences');
    },
  });
}

export function useLogoutAllDevices() {
  const router = useRouter();
  const logout = useAuthStore((state) => state.logout);

  return useMutation({
    mutationFn: async () => {
      await api.post('/users/sessions/logout-all');
      return true;
    },
    onSuccess: () => {
      logout();
      router.push('/login');
      toast.success('Signed out everywhere and cleared all active sessions.');
    },
    onError: (error: unknown) => {
      const message = axios.isAxiosError<ApiErrorResponse>(error)
        ? error.response?.data?.error || error.response?.data?.message || error.message
        : error instanceof Error
          ? error.message
          : 'Failed to clear active sessions';
      toast.error(message);
    },
  });
}

export function useUpdateSharingPreferences() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: SharingPreferencesData) => {
      const response = await api.patch<{ user: User }>('/users/sharing-preferences', data);
      return response.data.user;
    },
    onSuccess: (user) => {
      useAuthStore.getState().updateUser(user);
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Sharing preferences updated');
    },
    onError: (error: unknown) => {
      const message = axios.isAxiosError<ApiErrorResponse>(error)
        ? error.response?.data?.error || error.response?.data?.message || error.message
        : error instanceof Error
          ? error.message
          : 'Failed to update sharing preferences';
      toast.error(message);
    },
  });
}
