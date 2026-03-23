import { create } from "zustand"

import { PairingContext } from "@/types"

interface PairingState {
  context: PairingContext | null
  activePairingId: string | null
  isReady: boolean
  setContext: (context: PairingContext | null, userId?: string) => void
  setActivePairingId: (pairingId: string | null, userId?: string) => void
  clear: () => void
}

function storageKeyForUser(userId?: string) {
  return userId ? `active_pairing_${userId}` : "active_pairing"
}

export const usePairingStore = create<PairingState>((set) => ({
  context: null,
  activePairingId: null,
  isReady: false,

  setContext: (context, userId) => {
    if (typeof window === "undefined") {
      set({
        context,
        activePairingId: context?.active_pairing_id || context?.workspaces?.[0]?.id || null,
        isReady: true,
      })
      return
    }

    if (!context) {
      localStorage.removeItem(storageKeyForUser(userId))
      set({ context: null, activePairingId: null, isReady: true })
      return
    }

    const storedPairingId = localStorage.getItem(storageKeyForUser(userId))
    const validPairingIds = new Set((context.workspaces || []).map((workspace) => workspace.id))
    const nextPairingId =
      (storedPairingId && validPairingIds.has(storedPairingId) ? storedPairingId : null) ||
      (context.active_pairing_id && validPairingIds.has(context.active_pairing_id) ? context.active_pairing_id : null) ||
      context.workspaces?.[0]?.id ||
      null

    if (nextPairingId) {
      localStorage.setItem(storageKeyForUser(userId), nextPairingId)
    } else {
      localStorage.removeItem(storageKeyForUser(userId))
    }

    set({ context, activePairingId: nextPairingId, isReady: true })
  },

  setActivePairingId: (pairingId, userId) => {
    if (typeof window !== "undefined") {
      if (pairingId) {
        localStorage.setItem(storageKeyForUser(userId), pairingId)
      } else {
        localStorage.removeItem(storageKeyForUser(userId))
      }
    }

    set({ activePairingId: pairingId })
  },

  clear: () => {
    set({ context: null, activePairingId: null, isReady: false })
  },
}))
