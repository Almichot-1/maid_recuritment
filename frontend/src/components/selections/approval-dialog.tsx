"use client"

import * as React from "react"
import { AlertCircle, Loader2, XCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"

interface RejectSelectionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (reason?: string) => void
  isLoading?: boolean
}

/** Confirmation for rejecting a selection (optional reason). Approve uses direct action + toast. */
export function RejectSelectionDialog({
  open,
  onOpenChange,
  onConfirm,
  isLoading,
}: RejectSelectionDialogProps) {
  const [reason, setReason] = React.useState("")

  React.useEffect(() => {
    if (!open) {
      setReason("")
    }
  }, [open])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <XCircle className="h-5 w-5 text-red-600" />
            Reject selection?
          </DialogTitle>
          <DialogDescription>
            The candidate will be released and both parties will be notified. This cannot be undone.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2">
          <Label htmlFor="reason">Reason (optional)</Label>
          <Textarea
            id="reason"
            placeholder="Share a short reason with the other agency..."
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="min-h-[100px] resize-none"
            disabled={isLoading}
          />
        </div>

        <div className="flex items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-800 dark:bg-red-950/20">
          <AlertCircle className="mt-0.5 h-5 w-5 shrink-0 text-red-600" />
          <p className="text-sm text-red-800 dark:text-red-200">
            Rejection ends this selection for both agencies.
          </p>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={() => onConfirm(reason || undefined)} disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Reject
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
