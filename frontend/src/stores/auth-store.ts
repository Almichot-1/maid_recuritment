import { create } from 'zustand';
import { User } from '@/types';
import { usePairingStore } from '@/stores/pairing-store';
import { getApiBaseUrl } from '@/lib/api-base-url';
import { clearPersistedAgencyUser, persistAgencyUser } from '@/lib/auth-storage';
import { clearQueryCache } from '@/lib/query-client';
import { clearCandidateDraft } from '@/lib/candidate-draft';

interface AuthMeResponse {
  user: User;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setAuth: (user: User, token: string) => void;
  updateUser: (updates: Partial<User>) => void;
  logout: () => void;
  loadFromStorage: () => Promise<void>;
}

function clearSessionState() {
  clearPersistedAgencyUser();
  usePairingStore.getState().clear();
  clearQueryCache();
  void clearCandidateDraft();
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: true,

  setAuth: (user: User, token: string) => {
    persistAgencyUser(user);
    usePairingStore.getState().clear();
    set({ user, token, isAuthenticated: true, isLoading: false });
  },

  updateUser: (updates: Partial<User>) =>
    set((state) => {
      if (!state.user) {
        return state;
      }

      const user = { ...state.user, ...updates };
      persistAgencyUser(user);
      return { user };
    }),

  logout: () => {
    clearSessionState();
    set({ user: null, token: null, isAuthenticated: false, isLoading: false });
  },

  loadFromStorage: async () => {
    if (typeof window === 'undefined') {
      return;
    }

    set({ isLoading: true });

    try {
      const response = await fetch(`${getApiBaseUrl()}/auth/me`, {
        credentials: 'include',
      });

      if (response.status === 401 || response.status === 403) {
        clearPersistedAgencyUser();
        set({ user: null, token: null, isAuthenticated: false, isLoading: false });
        return;
      }

      if (!response.ok) {
        set({ isLoading: false });
        return;
      }

      const data = (await response.json()) as AuthMeResponse;
      persistAgencyUser(data.user);
      set({ user: data.user, token: null, isAuthenticated: true, isLoading: false });
    } catch (error) {
      console.error('Failed to load auth state from storage', error);
      set({ isLoading: false });
    }
  }
}));
