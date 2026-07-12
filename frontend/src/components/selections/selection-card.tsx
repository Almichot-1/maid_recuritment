"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { CheckCircle2, Clock, User, Eye, Loader2, Route } from "lucide-react"
import { format } from "date-fns"

import Image from "next/image"
import { Selection, SelectionStatus } from "@/types"
import { useCurrentUser } from "@/hooks/use-auth"
import { useApproveSelection, useRejectSelection } from "@/hooks/use-selections"
import { ApprovalDialog } from "./approval-dialog"
import { ProgressBadges } from "./progress-badges"
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

  // This should not happen as SelectionList filters out invalid candidates
  // But keep as safety check
  if (!candidate || !candidate.id || !candidate.full_name) {
    return null
  }

  const photoUrl = candidate.photo_url

  const isPending = selection.status === SelectionStatus.PENDING
  const isApproved = selection.status === SelectionStatus.APPROVED
  const isRejected = selection.status === SelectionStatus.REJECTED
  const isExpired = selection.status === SelectionStatus.EXPIRED

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
      case SelectionStatus.RELEASED:
        return (
          <Badge className="bg-orange-500 hover:bg-orange-600 text-white">
            Released
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
                  <div className="relative h-16 w-16 cursor-pointer hover:opacity-90 transition-opacity" onClick={handleViewDetails}>
                    <Image
                      src={photoUrl}
                      alt={candidate.full_name}
                      fill
                      unoptimized
                      className="rounded-lg object-cover border-2 border-border"
                    />
                  </div>
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
              {/* Show progress badges for approved selections */}
              {isApproved && selection.progress && (
                <div className="mt-2">
                  <ProgressBadges progress={selection.progress} />
                </div>
              )}
            </div>

            <Separator className="lg:hidden" />

            {/* Right: Approval Status & Actions */}
            <div className="flex flex-col gap-3 lg:min-w-[250px]">
              {/* Approval Status */}
              {isPending && (
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Clock className="h-4 w-4" />
                  <span>Waiting for Ethiopian agency approval</span>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex flex-wrap gap-2">
                {isPending && isEthiopianAgent && !selection.ethiopian_approved && (
                  <>
                    <Button
                      size="sm"
                      onClick={() => setApproveDialogOpen(true)}
                      className="bg-green-600 hover:bg-green-700 flex-1"
                      disabled={isApproving || isRejecting}
                    >
                      {isApproving && <Loader2 className="mr-2 h-3 w-3 animate-spin" />}
                      Approve
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setRejectDialogOpen(true)}
                      className="text-orange-600 border-orange-600 hover:bg-orange-50 dark:hover:bg-orange-950/20 flex-1"
                      disabled={isApproving || isRejecting}
                    >
                      Unlock
                    </Button>
                  </>
                )}

                {isPending && isEthiopianAgent && selection.ethiopian_approved && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground p-2 bg-muted/50 rounded-md w-full justify-center">
                    <CheckCircle2 className="h-4 w-4 text-green-600" />
                    <span>Approved</span>
                  </div>
                )}

                {isPending && !isEthiopianAgent && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground p-2 bg-muted/50 rounded-md w-full justify-center">
                    <Clock className="h-4 w-4" />
                    <span>Waiting for Ethiopian agency...</span>
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

                {(isRejected || isExpired || selection.status === 'released') && (
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
