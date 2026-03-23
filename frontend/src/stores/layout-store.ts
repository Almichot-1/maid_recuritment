import { create } from 'zustand'

interface LayoutState {
  isSidebarCollapsed: boolean
  toggleSidebar: () => void
  setSidebarCollapsed: (value: boolean) => void
}

export const useLayoutStore = create<LayoutState>((set) => ({
  isSidebarCollapsed: false,
  toggleSidebar: () => set((state) => {
    const newVal = !state.isSidebarCollapsed
    if (typeof window !== 'undefined') {
      localStorage.setItem("sidebar-collapsed", String(newVal))
    }
    return { isSidebarCollapsed: newVal }
  }),
  setSidebarCollapsed: (value) => set({ isSidebarCollapsed: value }),
}))
