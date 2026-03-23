import { create } from 'zustand';
import { User } from '@/types';
import { usePairingStore } from '@/stores/pairing-store';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setAuth: (user: User, token: string) => void;
  updateUser: (updates: Partial<User>) => void;
  logout: () => void;
  loadFromStorage: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: true, // Start true until hydrated
  
  setAuth: (user: User, token: string) => {
    localStorage.setItem('auth_token', token);
    localStorage.setItem('auth_user', JSON.stringify(user));
    usePairingStore.getState().clear();
    set({ user, token, isAuthenticated: true, isLoading: false });
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
    set({ user: null, token: null, isAuthenticated: false, isLoading: false });
  },
  
  loadFromStorage: () => {
    try {
      if (typeof window !== 'undefined') {
        const token = localStorage.getItem('auth_token');
        const userStr = localStorage.getItem('auth_user');
        
        if (token && userStr) {
          const user = JSON.parse(userStr) as User;
          set({ user, token, isAuthenticated: true, isLoading: false });
        } else {
          set({ user: null, token: null, isAuthenticated: false, isLoading: false });
        }
      }
    } catch (error) {
      console.error('Failed to load auth state from storage', error);
      set({ user: null, token: null, isAuthenticated: false, isLoading: false });
    }
  }
}));
