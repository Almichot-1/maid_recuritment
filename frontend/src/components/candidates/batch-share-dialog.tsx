"use client"

import * as React from "react"
import { Loader2, Share2, Unplug } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import api from "@/lib/api"
import { useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import type { WorkspaceSummary } from "@/types"

interface BatchShareDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  candidateIds: string[]
  workspaces: WorkspaceSummary[]
  onSuccess?: () => void
}

export function BatchShareDialog({
  open,
  onOpenChange,
  candidateIds,
  workspaces,
  onSuccess,
}: BatchShareDialogProps) {
  const queryClient = useQueryClient()
  const [selectedPartnerIds, setSelectedPartnerIds] = React.useState<string[]>([])
  const [isSharing, setIsSharing] = React.useState(false)
  const [isUnsharing, setIsUnsharing] = React.useState(false)

  // Reset selections when dialog opens
  React.useEffect(() => {
    if (open) {
      setSelectedPartnerIds(workspaces.map((ws) => ws.id))
    }
  }, [open, workspaces])

  const togglePartner = (id: string) => {
    setSelectedPartnerIds((prev) =>
      prev.includes(id) ? prev.filter((p) => p !== id) : [...prev, id]
    )
  }

  const handleShare = async () => {
    if (selectedPartnerIds.length === 0 || candidateIds.length === 0) return
    setIsSharing(true)
    let successCount = 0
    let errorCount = 0
    const partnerCount = selectedPartnerIds.length

    for (const pairingId of selectedPartnerIds) {
      for (const candidateId of candidateIds) {
        try {
          await api.post(`/pairings/${pairingId}/candidates/${candidateId}/share`)
          successCount++
        } catch {
          errorCount++
        }
      }
    }

    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["candidates"] }),
      queryClient.invalidateQueries({ queryKey: ["candidate-shares"] }),
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] }),
    ])

    if (errorCount > 0) {
      toast.warning(`Shared ${successCount} candidate(s) with ${partnerCount} partner(s), ${errorCount} operation(s) failed`)
    } else {
      toast.success(`Shared ${successCount} candidate(s) with ${partnerCount} partner(s)`)
    }

    setIsSharing(false)
    onOpenChange(false)
    onSuccess?.()
  }

  const handleUnshare = async () => {
    if (selectedPartnerIds.length === 0 || candidateIds.length === 0) return
    setIsUnsharing(true)
    let successCount = 0
    let errorCount = 0
    const partnerCount = selectedPartnerIds.length

    for (const pairingId of selectedPartnerIds) {
      for (const candidateId of candidateIds) {
        try {
          await api.delete(`/pairings/${pairingId}/candidates/${candidateId}/share`)
          successCount++
        } catch {
          errorCount++
        }
      }
    }

    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["candidates"] }),
      queryClient.invalidateQueries({ queryKey: ["candidate-shares"] }),
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] }),
    ])

    if (errorCount > 0) {
      toast.warning(`Unshared ${successCount} candidate(s) from ${partnerCount} partner(s), ${errorCount} operation(s) failed`)
    } else {
      toast.success(`Unshared ${successCount} candidate(s) from ${partnerCount} partner(s)`)
    }

    setIsUnsharing(false)
    onOpenChange(false)
    onSuccess?.()
  }

  const isPending = isSharing || isUnsharing

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Share with Partners</DialogTitle>
          <DialogDescription>
            Select partners to share or unshare {candidateIds.length} candidate(s) with.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {workspaces.length === 0 && (
            <p className="text-sm text-muted-foreground">No partners available.</p>
          )}
          {workspaces.map((ws) => (
            <div key={ws.id} className="flex items-center gap-3">
              <Checkbox
                id={ws.id}
                checked={selectedPartnerIds.includes(ws.id)}
                onCheckedChange={() => togglePartner(ws.id)}
              />
              <Label htmlFor={ws.id} className="flex items-center justify-between w-full cursor-pointer">
                <span>{ws.partner_agency?.company_name || ws.partner_agency?.full_name || ws.partner_agency?.email || "Partner"}</span>
              </Label>
            </div>
          ))}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
            Cancel
          </Button>
          <Button
            variant="secondary"
            onClick={handleUnshare}
            disabled={isPending || selectedPartnerIds.length === 0}
          >
            {isUnsharing ? (
              <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Unsharing...</>
            ) : (
              <><Unplug className="mr-2 h-4 w-4" /> Unshare</>
            )}
          </Button>
          <Button
            onClick={handleShare}
            disabled={isPending || selectedPartnerIds.length === 0}
          >
            {isSharing ? (
              <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Sharing...</>
            ) : (
              <><Share2 className="mr-2 h-4 w-4" /> Share</>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
