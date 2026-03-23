import axios from 'axios';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import api from '@/lib/api';
import { useAuthStore } from '@/stores/auth-store';

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

interface ApiErrorResponse {
  error?: string;
  message?: string;
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: ProfileUpdateData) => {
      const response = await api.patch<{ user: { full_name: string; company_name?: string } }>('/users/profile', data);
      return response.data.user;
    },
    onSuccess: (user) => {
      useAuthStore.getState().updateUser(user);
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Profile updated successfully');
    },
    onError: () => {
      toast.error('Failed to update profile');
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
  return useMutation({
    mutationFn: async () => {
      throw new Error('Logging out of all devices is not supported by the current API yet.');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to logout of all devices');
    },
  });
}
