"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { CheckCircle2, Clock, User, Eye, Loader2, Route } from "lucide-react"
import { format } from "date-fns"

import { Selection, SelectionStatus } from "@/types"
import { useCurrentUser } from "@/hooks/use-auth"
import { useApproveSelection, useRejectSelection } from "@/hooks/use-selections"
import { LockCountdown } from "./lock-countdown"
import { ApprovalDialog } from "./approval-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
interface SelectionCardProps {
  selection: Selection
}

export function SelectionCard({ selection }: SelectionCardProps) {
  const router = useRouter()
  const { isEthiopianAgent } = useCurrentUser()
  const [approveDialogOpen, setApproveDialogOpen] = React.useState(false)
  const [rejectDialogOpen, setRejectDialogOpen] = React.useState(false)

  const { mutate: approveSelection, isPending: isApproving } = useApproveSelection(selection.id, selection.candidate_id)
  const { mutate: rejectSelection, isPending: isRejecting } = useRejectSelection(selection.id, selection.candidate_id)

  const candidate = selection.candidate

  if (!candidate) {
    return null
  }

  const photoUrl = candidate.photo_url

  // Determine user's approval status
  const userHasApproved = isEthiopianAgent 
    ? selection.ethiopian_approved 
    : selection.foreign_approved

  const isPending = selection.status === SelectionStatus.PENDING
  const isApproved = selection.status === SelectionStatus.APPROVED
  const isRejected = selection.status === SelectionStatus.REJECTED
  const isExpired = selection.status === SelectionStatus.EXPIRED
  const hasRequiredEmployerDocuments = !!selection.employer_contract?.file_url && !!selection.employer_id?.file_url
  const approvalBlockedByEmployerPackage = isEthiopianAgent && isPending && !hasRequiredEmployerDocuments

  const getStatusBadge = () => {
    switch (selection.status) {
      case SelectionStatus.PENDING:
        return (
          <Badge className="bg-yellow-500 hover:bg-yellow-600 text-white">
            Pending
          </Badge>
        )
      case SelectionStatus.APPROVED:
        return (
          <Badge className="bg-green-500 hover:bg-green-600 text-white">
            Approved
          </Badge>
        )
      case SelectionStatus.REJECTED:
        return (
          <Badge className="bg-red-500 hover:bg-red-600 text-white">
            Rejected
          </Badge>
        )
      case SelectionStatus.EXPIRED:
        return (
          <Badge className="bg-gray-500 hover:bg-gray-600 text-white">
            Expired
          </Badge>
        )
      default:
        return null
    }
  }

  const handleApprove = () => {
    approveSelection(undefined, {
      onSuccess: () => {
        setApproveDialogOpen(false)
      },
    })
  }

  const handleReject = (reason?: string) => {
    rejectSelection({ reason: reason || "" }, {
      onSuccess: () => {
        setRejectDialogOpen(false)
      },
    })
  }

  const handleViewDetails = () => {
    router.push(`/selections/${selection.id}`)
  }

  return (
    <>
      <Card className="hover:shadow-md transition-shadow">
        <CardContent className="p-4">
          <div className="flex flex-col lg:flex-row gap-4">
            {/* Left: Candidate Info */}
            <div className="flex gap-3 flex-1 min-w-0">
              {/* Photo */}
              <div className="shrink-0">
                {photoUrl ? (
                  <img
                    src={photoUrl}
                    alt={candidate.full_name}
                    className="h-16 w-16 rounded-lg object-cover border-2 border-border cursor-pointer hover:opacity-90 transition-opacity"
                    onClick={handleViewDetails}
                  />
                ) : (
                  <div className="h-16 w-16 rounded-lg bg-muted flex items-center justify-center border-2 border-dashed">
                    <User className="h-8 w-8 text-muted-foreground" />
                  </div>
                )}
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <h3 
                  className="font-semibold text-lg leading-tight truncate cursor-pointer hover:text-primary transition-colors"
                  onClick={handleViewDetails}
                >
                  {candidate.full_name}
                </h3>
                <div className="flex items-center gap-2 text-sm text-muted-foreground mt-1">
                  <span>{candidate.age ?? "N/A"} years</span>
                  <span>-</span>
                  <span>{candidate.experience_years ?? 0} years exp.</span>
                </div>
              </div>
            </div>

            <Separator className="lg:hidden" />

            {/* Center: Status & Date */}
            <div className="flex flex-col gap-2 lg:min-w-[200px]">
              <div className="flex items-center gap-2">
                {getStatusBadge()}
              </div>
              <p className="text-xs text-muted-foreground">
                Selected {format(new Date(selection.created_at), "MMM dd, yyyy")}
              </p>
              {isPending && (
                <LockCountdown 
                  expiresAt={selection.expires_at}
                  className="text-xs"
                  showIcon={true}
                />
              )}
            </div>

            <Separator className="lg:hidden" />

            {/* Right: Approval Status & Actions */}
            <div className="flex flex-col gap-3 lg:min-w-[250px]">
              {/* Approval Indicators */}
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-medium text-muted-foreground">Ethiopian:</span>
                  {selection.ethiopian_approved ? (
                    <CheckCircle2 className="h-5 w-5 text-green-600" />
                  ) : (
                    <Clock className="h-5 w-5 text-gray-400" />
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-medium text-muted-foreground">Foreign:</span>
                  {selection.foreign_approved ? (
                    <CheckCircle2 className="h-5 w-5 text-green-600" />
                  ) : (
                    <Clock className="h-5 w-5 text-gray-400" />
                  )}
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-wrap gap-2">
                {isPending && !userHasApproved && (
                  <>
                    {approvalBlockedByEmployerPackage && (
                      <div className="w-full rounded-md border border-amber-300/50 bg-amber-50/80 px-3 py-2 text-center text-xs text-amber-900 dark:border-amber-900/40 dark:bg-amber-950/30 dark:text-amber-100">
                        Waiting for the foreign agency to upload the contract package.
                      </div>
                    )}
                    <Button
                      size="sm"
                      onClick={() => setApproveDialogOpen(true)}
                      className="bg-green-600 hover:bg-green-700 flex-1"
                      disabled={isApproving || isRejecting || approvalBlockedByEmployerPackage}
                    >
                      {isApproving && <Loader2 className="mr-2 h-3 w-3 animate-spin" />}
                      Approve
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setRejectDialogOpen(true)}
                      className="text-red-600 border-red-600 hover:bg-red-50 dark:hover:bg-red-950/20 flex-1"
                      disabled={isApproving || isRejecting}
                    >
                      {isRejecting && <Loader2 className="mr-2 h-3 w-3 animate-spin" />}
                      Reject
                    </Button>
                  </>
                )}

                {isPending && userHasApproved && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground p-2 bg-muted/50 rounded-md w-full justify-center">
                    <Clock className="h-4 w-4" />
                    <span>Waiting for other party...</span>
                  </div>
                )}

                {isApproved && (
                  <Button
                    size="sm"
                    onClick={() => router.push(`/candidates/${selection.candidate_id}/tracking`)}
                    className="w-full"
                  >
                    <Route className="mr-2 h-4 w-4" />
                    View Progress
                  </Button>
                )}

                {(isRejected || isExpired) && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={handleViewDetails}
                    className="w-full"
                  >
                    <Eye className="mr-2 h-4 w-4" />
                    View Details
                  </Button>
                )}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Approval Dialogs */}
      <ApprovalDialog
        open={approveDialogOpen}
        onOpenChange={setApproveDialogOpen}
        candidateName={candidate.full_name}
        type="approve"
        onConfirm={handleApprove}
        isLoading={isApproving}
      />

      <ApprovalDialog
        open={rejectDialogOpen}
        onOpenChange={setRejectDialogOpen}
        candidateName={candidate.full_name}
        type="reject"
        onConfirm={handleReject}
        isLoading={isRejecting}
      />
    </>
  )
}
