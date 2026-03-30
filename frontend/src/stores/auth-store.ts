import { create } from 'zustand';
import { User } from '@/types';
import { usePairingStore } from '@/stores/pairing-store';
import { getApiBaseUrl } from '@/lib/api-base-url';

interface AuthMeResponse {
  user: User;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setAuth: (user: User) => void;
  updateUser: (updates: Partial<User>) => void;
  logout: () => void;
  loadFromStorage: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true, // Start true until hydrated
  
  setAuth: (user: User) => {
    localStorage.setItem('auth_user', JSON.stringify(user));
    localStorage.removeItem('auth_token');
    usePairingStore.getState().clear();
    set({ user, isAuthenticated: true, isLoading: false });
  },

  updateUser: (updates: Partial<User>) =>
    set((state) => {
      if (!state.user) {
        return state;
      }

      const user = { ...state.user, ...updates };
      localStorage.setItem('auth_user', JSON.stringify(user));
      return { user };
    }),
  
  logout: () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
    usePairingStore.getState().clear();
    set({ user: null, isAuthenticated: false, isLoading: false });
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

      if (!response.ok) {
        throw new Error(`session check failed with status ${response.status}`);
      }

      const data = (await response.json()) as AuthMeResponse;
      localStorage.setItem('auth_user', JSON.stringify(data.user));
      localStorage.removeItem('auth_token');
      set({ user: data.user, isAuthenticated: true, isLoading: false });
    } catch (error) {
      console.error('Failed to load auth state from storage', error);
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_user');
      set({ user: null, isAuthenticated: false, isLoading: false });
    }
  }
}));
