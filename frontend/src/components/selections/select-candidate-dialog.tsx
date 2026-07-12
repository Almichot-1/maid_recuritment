"use client"

import * as React from "react"
import axios from "axios"
import { useRouter } from "next/navigation"
import { CheckCircle2, Loader2, User } from "lucide-react"
import { toast } from "sonner"

import Image from "next/image"
import { Candidate } from "@/types"
import { useSelectCandidate } from "@/hooks/use-selections"
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
  const { mutate: selectCandidate, isPending } = useSelectCandidate(candidate.id)

  const handleConfirm = () => {
    selectCandidate(undefined, {
        onSuccess: (data: SelectCandidateSuccessResponse) => {
        onOpenChange(false)
        toast.success("Candidate selected successfully!", {
          description: "The Ethiopian agency will review your selection.",
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
                <div className="relative h-24 w-24">
                  <Image
                    src={photoUrl}
                    alt={candidate.full_name}
                    fill
                    unoptimized
                    className="rounded-lg object-cover border-2 border-border"
                  />
                </div>
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
                  {(candidate.languages as Array<{ language: string; proficiency?: string }>).map((lang, index: number) => (
                    <Badge key={index} variant="outline" className="text-xs">
                      {typeof lang === "string" ? lang : `${lang.language}${lang.proficiency ? ` - ${lang.proficiency}` : ""}`}
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
            disabled={isPending}
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
