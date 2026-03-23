import * as React from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { AxiosError } from "axios"
import { toast } from "sonner"

import api from "@/lib/api"
import adminApi from "@/lib/admin-api"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingStore } from "@/stores/pairing-store"
import { AdminPairing, CandidatePairShare, PairingContext } from "@/types"

export function usePairingContext() {
  const { user } = useCurrentUser()
  const context = usePairingStore((state) => state.context)
  const activePairingId = usePairingStore((state) => state.activePairingId)
  const isReady = usePairingStore((state) => state.isReady)
  const setContext = usePairingStore((state) => state.setContext)
  const setActivePairingId = usePairingStore((state) => state.setActivePairingId)

  const query = useQuery({
    queryKey: ["pairings", "me", user?.id],
    queryFn: async () => {
      const response = await api.get<{ context: PairingContext }>("/pairings/me")
      return response.data.context
    },
    enabled: Boolean(user),
    staleTime: 300_000,
    refetchOnWindowFocus: false,
  })

  React.useEffect(() => {
    if (query.data && user?.id) {
      setContext(query.data, user.id)
    }
  }, [query.data, setContext, user?.id])

  React.useEffect(() => {
    if (!user) {
      setContext(null)
    }
  }, [setContext, user])

  const activeWorkspace = React.useMemo(
    () => context?.workspaces.find((workspace) => workspace.id === activePairingId) || null,
    [activePairingId, context?.workspaces]
  )

  return {
    ...query,
    context,
    activePairingId,
    activeWorkspace,
    hasActivePairs: Boolean(context?.has_active_pairs),
    isReady,
    setActivePairingId: (pairingId: string | null) => setActivePairingId(pairingId, user?.id),
  }
}

export function useCandidateShares(candidateId?: string, enabled = true) {
  return useQuery({
    queryKey: ["candidate-shares", candidateId],
    queryFn: async () => {
      const response = await api.get<{ shares: CandidatePairShare[] }>(`/candidates/${candidateId}/shares`)
      return response.data.shares
    },
    enabled: Boolean(candidateId) && enabled,
    retry: (failureCount, error) => {
      const status = (error as AxiosError)?.response?.status
      if (status === 401 || status === 403 || status === 404) {
        return false
      }

      return failureCount < 2
    },
  })
}

export function useShareCandidateToWorkspace() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ pairingId, candidateId }: { pairingId: string; candidateId: string }) => {
      await api.post(`/pairings/${pairingId}/candidates/${candidateId}/share`)
    },
    onSuccess: async (_, variables) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["candidate-shares", variables.candidateId] }),
        queryClient.invalidateQueries({ queryKey: ["candidates"] }),
        queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] }),
      ])
      toast.success("Candidate shared to workspace")
    },
  })
}

export function useUnshareCandidateFromWorkspace() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ pairingId, candidateId }: { pairingId: string; candidateId: string }) => {
      await api.delete(`/pairings/${pairingId}/candidates/${candidateId}/share`)
    },
    onSuccess: async (_, variables) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["candidate-shares", variables.candidateId] }),
        queryClient.invalidateQueries({ queryKey: ["candidates"] }),
        queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] }),
      ])
      toast.success("Candidate removed from workspace")
    },
  })
}

export function useAdminPairings(filters?: { agency_id?: string; ethiopian_user_id?: string; foreign_user_id?: string; status?: string }) {
  return useQuery({
    queryKey: ["admin-pairings", filters ?? {}],
    queryFn: async () => {
      const response = await adminApi.get<{ pairings: AdminPairing[] }>("/admin/pairings", { params: filters })
      return response.data.pairings
    },
  })
}

export function useAgencyPairings(agencyId?: string) {
  return useQuery({
    queryKey: ["admin-agency-pairings", agencyId],
    queryFn: async () => {
      const response = await adminApi.get<{ pairings: AdminPairing[] }>(`/admin/agencies/${agencyId}/pairings`)
      return response.data.pairings
    },
    enabled: Boolean(agencyId),
  })
}

export function useCreateAdminPairing() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (payload: { ethiopian_user_id: string; foreign_user_id: string; notes?: string }) => {
      const response = await adminApi.post<{ pairing: AdminPairing }>("/admin/pairings", payload)
      return response.data.pairing
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-pairings"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-agency-pairings"] }),
      ])
    },
  })
}

export function useUpdateAdminPairing() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, status, notes }: { id: string; status: string; notes?: string }) => {
      const response = await adminApi.patch<{ pairing: AdminPairing }>(`/admin/pairings/${id}`, { status, notes })
      return response.data.pairing
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["admin-pairings"] }),
        queryClient.invalidateQueries({ queryKey: ["admin-agency-pairings"] }),
      ])
    },
  })
}
