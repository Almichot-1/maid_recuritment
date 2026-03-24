"use client"

import * as React from "react"
import { AlertCircle, CheckCircle2, Loader2, XCircle } from "lucide-react"

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

interface ApprovalDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  candidateName: string
  type: "approve" | "reject"
  onConfirm: (reason?: string) => void
  isLoading?: boolean
}

export function ApprovalDialog({
  open,
  onOpenChange,
  candidateName,
  type,
  onConfirm,
  isLoading,
}: ApprovalDialogProps) {
  const [reason, setReason] = React.useState("")

  React.useEffect(() => {
    if (!open) {
      setReason("")
    }
  }, [open])

  const handleConfirm = () => {
    onConfirm(type === "reject" ? reason : undefined)
  }

  if (type === "approve") {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CheckCircle2 className="h-5 w-5 text-green-600" />
              Approve Selection
            </DialogTitle>
            <DialogDescription>
              Are you sure you want to approve the selection of{" "}
              <span className="font-semibold text-foreground">{candidateName}</span>?
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <div className="flex items-start gap-3 rounded-lg border border-green-200 bg-green-50 p-3 dark:border-green-800 dark:bg-green-950/20">
              <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-green-600" />
              <div className="text-sm text-green-800 dark:text-green-200">
                <p className="mb-1 font-medium">This will:</p>
                <ul className="space-y-1 text-xs">
                  <li>- Confirm your approval of this selection</li>
                  <li>- Notify the other party</li>
                  <li>- Move to recruitment process if both parties approve</li>
                </ul>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              onClick={handleConfirm}
              disabled={isLoading}
              className="bg-green-600 hover:bg-green-700"
            >
              {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Confirm Approval
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <XCircle className="h-5 w-5 text-red-600" />
            Reject Selection
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to reject the selection of{" "}
            <span className="font-semibold text-foreground">{candidateName}</span>?
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="flex items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-800 dark:bg-red-950/20">
            <AlertCircle className="mt-0.5 h-5 w-5 shrink-0 text-red-600" />
            <div className="text-sm text-red-800 dark:text-red-200">
              <p className="mb-1 font-medium">Warning:</p>
              <ul className="space-y-1 text-xs">
                <li>- This will reject the selection for both parties</li>
                <li>- The candidate will be released and become available again</li>
                <li>- This action cannot be undone</li>
              </ul>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="reason">Reason (Optional)</Label>
            <Textarea
              id="reason"
              placeholder="Provide a reason for rejection..."
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="min-h-[100px] resize-none"
              disabled={isLoading}
            />
            <p className="text-xs text-muted-foreground">
              This reason will be shared with the other party.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleConfirm}
            disabled={isLoading}
          >
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Confirm Rejection
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
