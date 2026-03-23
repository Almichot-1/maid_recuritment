"use client"

import * as React from "react"
import { Building2, Loader2, Share2, Unplug } from "lucide-react"

import { useCandidateShares, usePairingContext, useShareCandidateToWorkspace, useUnshareCandidateFromWorkspace } from "@/hooks/use-pairings"
import { Candidate, CandidateStatus } from "@/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

interface CandidateShareDialogProps {
  candidate: Candidate
  open: boolean
  onOpenChange: (open: boolean) => void
}

function workspaceTitle(companyName?: string, fullName?: string) {
  return companyName?.trim() || fullName?.trim() || "Partner workspace"
}

export function CandidateShareDialog({ candidate, open, onOpenChange }: CandidateShareDialogProps) {
  const { context } = usePairingContext()
  const { data: shares = [], isLoading } = useCandidateShares(candidate.id, open)
  const shareMutation = useShareCandidateToWorkspace()
  const unshareMutation = useUnshareCandidateFromWorkspace()
  const [pendingPairingId, setPendingPairingId] = React.useState<string | null>(null)

  const shareable = candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE
  const sharedPairingIds = React.useMemo(
    () => new Set(shares.filter((share) => share.is_active).map((share) => share.pairing_id)),
    [shares]
  )

  const handleToggle = async (pairingId: string, isShared: boolean) => {
    setPendingPairingId(pairingId)
    try {
      if (isShared) {
        await unshareMutation.mutateAsync({ pairingId, candidateId: candidate.id })
      } else {
        await shareMutation.mutateAsync({ pairingId, candidateId: candidate.id })
      }
    } finally {
      setPendingPairingId(null)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Share With Partner Agencies</DialogTitle>
          <DialogDescription>
            Choose which partner agencies can see <span className="font-semibold text-foreground">{candidate.full_name}</span>. Sharing it here gives that partner access to this same profile before selection.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-2xl border border-border/70 bg-muted/40 p-4 text-sm text-muted-foreground">
            This profile lives in your agency library. If you edit it later, the updated details will automatically appear for every partner agency that can still see it.
          </div>

          {!shareable ? (
            <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
              This candidate is currently <span className="font-semibold">{candidate.status.replaceAll("_", " ")}</span>, so sharing is locked until the candidate returns to draft or available status.
            </div>
          ) : null}

          {isLoading ? (
            <div className="flex items-center justify-center py-10">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : context?.workspaces?.length ? (
            <div className="space-y-3">
              {context.workspaces.map((workspace) => {
                const isShared = sharedPairingIds.has(workspace.id)
                const isPending = pendingPairingId === workspace.id
                return (
                  <div
                    key={workspace.id}
                    className="flex flex-col gap-4 rounded-2xl border border-border/70 bg-card/80 p-4 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <div className="space-y-2">
                      <div className="flex flex-wrap items-center gap-2">
                        <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                          <Building2 className="h-4 w-4" />
                        </div>
                        <div>
                          <p className="font-semibold text-foreground">
                            {workspaceTitle(workspace.partner_agency.company_name, workspace.partner_agency.full_name)}
                          </p>
                          <p className="text-xs text-muted-foreground">{workspace.partner_agency.email}</p>
                        </div>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        <Badge variant={isShared ? "default" : "outline"}>
                          {isShared ? "Visible" : "Hidden"}
                        </Badge>
                        <Badge variant="outline">{workspace.status.replaceAll("_", " ")}</Badge>
                      </div>
                    </div>

                    <Button
                      variant={isShared ? "outline" : "default"}
                      disabled={!shareable || isPending}
                      onClick={() => handleToggle(workspace.id, isShared)}
                      className="sm:min-w-[170px]"
                    >
                      {isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : isShared ? <Unplug className="mr-2 h-4 w-4" /> : <Share2 className="mr-2 h-4 w-4" />}
                      {isShared ? "Remove from partner" : "Share with partner"}
                    </Button>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="rounded-2xl border border-dashed border-border p-6 text-sm text-muted-foreground">
              No partner agencies are active yet. Once an admin creates pairings for this agency, they will appear here.
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
