"use client"

import * as React from "react"
import axios from "axios"
import { useRouter } from "next/navigation"
import { AlertCircle, CheckCircle2, Loader2, User } from "lucide-react"
import { toast } from "sonner"

import { Candidate } from "@/types"
import { useSelectCandidate } from "@/hooks/use-selections"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Separator } from "@/components/ui/separator"

interface SelectCandidateSuccessResponse {
  selection?: {
    id?: string
  }
}

interface ApiErrorResponse {
  error?: string
  message?: string
}

interface SelectCandidateDialogProps {
  candidate: Candidate
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SelectCandidateDialog({ candidate, open, onOpenChange }: SelectCandidateDialogProps) {
  const router = useRouter()
  const [agreed, setAgreed] = React.useState(false)
  const { mutate: selectCandidate, isPending } = useSelectCandidate(candidate.id)

  // Reset checkbox when dialog opens/closes
  React.useEffect(() => {
    if (!open) {
      setAgreed(false)
    }
  }, [open])

  const handleConfirm = () => {
    selectCandidate(undefined, {
        onSuccess: (data: SelectCandidateSuccessResponse) => {
        onOpenChange(false)
        toast.success("Candidate selected successfully!", {
          description: "Upload the employer contract and employer ID next so the agency can review and approve within 24 hours.",
        })
        // Redirect to selection detail page if we have the selection ID
        if (data?.selection?.id) {
          router.push(`/selections/${data.selection.id}`)
        } else {
          // Fallback to selections list
          router.push("/selections")
        }
      },
      onError: (error: unknown) => {
        const status = axios.isAxiosError<ApiErrorResponse>(error) ? error.response?.status : undefined
        const message = axios.isAxiosError<ApiErrorResponse>(error)
          ? error.response?.data?.message || error.response?.data?.error || error.message
          : error instanceof Error
            ? error.message
            : undefined

        if (status === 409) {
          toast.error("Candidate is no longer available", {
            description: "This candidate has been selected by another agency.",
          })
          onOpenChange(false)
          // Refresh the page to update candidate status
          router.refresh()
        } else {
          toast.error("Failed to select candidate", {
            description: message || "Please try again later.",
          })
        }
      },
    })
  }

  // Get photo URL from documents
  const photoUrl = candidate.documents?.find((doc) => doc.document_type === "photo")?.file_url

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-2xl">Confirm Candidate Selection</DialogTitle>
          <DialogDescription>
            Review the candidate details before confirming your selection.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Candidate Summary */}
          <div className="flex flex-col sm:flex-row gap-4 p-4 bg-muted/50 rounded-lg border">
            {/* Photo */}
            <div className="shrink-0">
              {photoUrl ? (
                <img
                  src={photoUrl}
                  alt={candidate.full_name}
                  className="h-24 w-24 rounded-lg object-cover border-2 border-border"
                />
              ) : (
                <div className="h-24 w-24 rounded-lg bg-muted flex items-center justify-center border-2 border-dashed">
                  <User className="h-10 w-10 text-muted-foreground" />
                </div>
              )}
            </div>

            {/* Info */}
            <div className="flex-1 space-y-2">
              <h3 className="text-xl font-bold">{candidate.full_name}</h3>
              <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                <span>{candidate.age ?? "N/A"} years old</span>
                <span>-</span>
                <span>{candidate.experience_years ?? 0} years experience</span>
              </div>

              {/* Languages */}
              {candidate.languages && candidate.languages.length > 0 && (
                <div className="flex flex-wrap gap-1.5 pt-1">
                  <span className="text-xs font-medium text-muted-foreground mr-1">Languages:</span>
                  {candidate.languages.map((lang, index) => (
                    <Badge key={index} variant="outline" className="text-xs">
                      {lang}
                    </Badge>
                  ))}
                </div>
              )}

              {/* Skills */}
              {candidate.skills && candidate.skills.length > 0 && (
                <div className="flex flex-wrap gap-1.5 pt-1">
                  <span className="text-xs font-medium text-muted-foreground mr-1">Skills:</span>
                  {candidate.skills.slice(0, 5).map((skill, index) => (
                    <Badge key={index} variant="secondary" className="text-xs">
                      {skill}
                    </Badge>
                  ))}
                  {candidate.skills.length > 5 && (
                    <Badge variant="secondary" className="text-xs">
                      +{candidate.skills.length - 5} more
                    </Badge>
                  )}
                </div>
              )}
            </div>
          </div>

          <Separator />

          {/* Important Notice */}
          <div className="space-y-3 p-4 bg-amber-50 dark:bg-amber-950/20 border-2 border-amber-200 dark:border-amber-800 rounded-lg">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
              <div className="space-y-2 flex-1">
                <h4 className="font-semibold text-amber-900 dark:text-amber-100">
                  Important Information
                </h4>
                <ul className="space-y-1.5 text-sm text-amber-800 dark:text-amber-200">
                  <li className="flex items-start gap-2">
                    <span className="text-amber-600 dark:text-amber-400 mt-0.5">-</span>
                    <span>This candidate will be locked for 24 hours exclusively for your agency</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-amber-600 dark:text-amber-400 mt-0.5">-</span>
                    <span>Both parties (you and the Ethiopian agent) must approve within the time window</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-amber-600 dark:text-amber-400 mt-0.5">-</span>
                    <span>If not approved by both parties, the candidate will be automatically released</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-amber-600 dark:text-amber-400 mt-0.5">-</span>
                    <span>You will be notified once the Ethiopian agent responds to your selection</span>
                  </li>
                </ul>
              </div>
            </div>
          </div>

          {/* Agreement Checkbox */}
          <div className="flex items-start gap-3 p-4 bg-muted/30 rounded-lg border">
            <Checkbox
              id="agree"
              checked={agreed}
              onCheckedChange={(checked) => setAgreed(checked === true)}
              disabled={isPending}
              className="mt-1"
            />
            <label
              htmlFor="agree"
              className="text-sm font-medium leading-relaxed cursor-pointer select-none"
            >
              I understand the selection terms and want to proceed with selecting{" "}
              <span className="font-semibold">{candidate.full_name}</span> for my agency.
            </label>
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={!agreed || isPending}
            className="bg-green-600 hover:bg-green-700 text-white min-w-[140px]"
          >
            {isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Selecting...
              </>
            ) : (
              <>
                <CheckCircle2 className="mr-2 h-4 w-4" />
                Confirm Selection
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
