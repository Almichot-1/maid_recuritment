"use client"

import * as React from "react"
import dynamic from "next/dynamic"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { User, Lock, PencilLine } from "lucide-react"

import { Candidate, CandidateStatus } from "@/types"
import { useCurrentUser } from "@/hooks/use-auth"
import { LockCountdown } from "@/components/selections/lock-countdown"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { cn } from "@/lib/utils"

const CandidateShareDialog = dynamic(
  () => import("@/components/candidates/candidate-share-dialog").then((module) => module.CandidateShareDialog)
)
const SelectCandidateDialog = dynamic(
  () => import("@/components/selections/select-candidate-dialog").then((module) => module.SelectCandidateDialog)
)

interface CandidateCardProps {
  candidate: Candidate
}

export function CandidateCard({ candidate }: CandidateCardProps) {
  const router = useRouter()
  const { user, isEthiopianAgent, isForeignAgent } = useCurrentUser()
  const [selectDialogOpen, setSelectDialogOpen] = React.useState(false)
  const [shareDialogOpen, setShareDialogOpen] = React.useState(false)

  const handleCardClick = () => {
    router.push(`/candidates/${candidate.id}`)
  }

  const handleSelectClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    setSelectDialogOpen(true)
  }

  // Get photo URL from documents
  const photoUrl = candidate.documents?.find((doc) => doc.document_type === "photo")?.file_url

  // Display first 3 skills
  const displaySkills = candidate.skills.slice(0, 3)
  const remainingSkills = candidate.skills.length - 3

  // Check if this candidate is locked by current user
  const isLockedByCurrentUser = candidate.status === CandidateStatus.LOCKED &&
    candidate.locked_by === user?.id

  // Check if locked by someone else
  const isLockedByOther = candidate.status === CandidateStatus.LOCKED &&
    candidate.locked_by !== user?.id

  const isOwner = isEthiopianAgent && candidate.created_by === user?.id
  const canEdit = isOwner && (candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE)

  return (
    <>
      <Card
        className="group cursor-pointer transition-all duration-200 hover:shadow-lg hover:scale-[1.02] overflow-hidden"
        onClick={handleCardClick}
      >
        <CardContent className="p-0">
          {/* Photo */}
          <div className="relative w-full aspect-square bg-muted overflow-hidden">
            {photoUrl ? (
              <img
                src={photoUrl}
                alt={candidate.full_name}
                className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                loading="lazy"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-primary/10 to-primary/5">
                <User className="h-20 w-20 text-muted-foreground/40" />
              </div>
            )}
            
            {/* Status Badge */}
            <div className="absolute top-3 right-3">
              <StatusBadge status={candidate.status} />
            </div>

            {/* Lock Indicator */}
            {(candidate.status === CandidateStatus.LOCKED || candidate.status === CandidateStatus.IN_PROGRESS) && (
              <div className="absolute top-3 left-3">
                <div className="flex items-center gap-1.5 bg-purple-500/90 text-white px-2 py-1 rounded-md text-xs font-semibold shadow-md">
                  <Lock className="h-3 w-3" />
                  <span>Locked</span>
                </div>
              </div>
            )}
          </div>

          {/* Content */}
          <div className="p-4 space-y-3">
            {/* Name */}
            <h3 className="font-semibold text-lg leading-tight line-clamp-1">
              {candidate.full_name}
            </h3>

            {/* Age & Experience */}
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
              <span>{candidate.age ?? "N/A"} years old</span>
              <span>-</span>
              <span>{candidate.experience_years ?? 0} years exp.</span>
            </div>

            {/* Languages */}
            {candidate.languages.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {candidate.languages.map((lang) => (
                  <Badge key={lang} variant="outline" className="text-xs">
                    {lang}
                  </Badge>
                ))}
              </div>
            )}

            {/* Skills */}
            {candidate.skills.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {displaySkills.map((skill) => (
                  <Badge key={skill} variant="secondary" className="text-xs">
                    {skill}
                  </Badge>
                ))}
                {remainingSkills > 0 && (
                  <Badge variant="secondary" className="text-xs">
                    +{remainingSkills} more
                  </Badge>
                )}
              </div>
            )}

            {/* Action Button for Foreign Agent */}
            {isForeignAgent && candidate.status === CandidateStatus.AVAILABLE && (
              <Button
                onClick={handleSelectClick}
                className="w-full mt-2 bg-green-600 hover:bg-green-700"
                size="sm"
              >
                Select Candidate
              </Button>
            )}

            {/* Locked by Current User - Show Countdown */}
            {isForeignAgent && isLockedByCurrentUser && candidate.lock_expires_at && (
              <div className="mt-2 p-3 bg-purple-50 dark:bg-purple-950/20 rounded-lg border border-purple-200 dark:border-purple-800">
                <p className="text-xs font-medium text-purple-900 dark:text-purple-100 mb-2">
                  Selected by you
                </p>
                <LockCountdown 
                  expiresAt={candidate.lock_expires_at}
                  className="text-xs"
                  showIcon={true}
                />
              </div>
            )}

            {/* Locked by Another Agent */}
            {isForeignAgent && isLockedByOther && (
              <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground mt-2 p-2 bg-muted/50 rounded-md">
                <Lock className="h-4 w-4" />
                <span>Locked by another agency</span>
              </div>
            )}

            {isOwner && (
              <div className="mt-2 flex gap-2">
                <Button
                  variant="secondary"
                  className="flex-1"
                  size="sm"
                  onClick={(event) => {
                    event.stopPropagation()
                    setShareDialogOpen(true)
                  }}
                >
                  Share
                </Button>
                {canEdit ? (
                  <Button variant="outline" className="flex-1" size="sm" asChild onClick={(event) => event.stopPropagation()}>
                    <Link href={`/candidates/${candidate.id}/edit`}>
                      <PencilLine className="mr-2 h-4 w-4" />
                      Edit
                    </Link>
                  </Button>
                ) : null}
                <Button variant="outline" className="flex-1" size="sm" asChild onClick={(event) => event.stopPropagation()}>
                  <Link href={`/candidates/${candidate.id}`}>Open</Link>
                </Button>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Selection Dialog */}
      <SelectCandidateDialog
        candidate={candidate}
        open={selectDialogOpen}
        onOpenChange={setSelectDialogOpen}
      />

      <CandidateShareDialog
        candidate={candidate}
        open={shareDialogOpen}
        onOpenChange={setShareDialogOpen}
      />
    </>
  )
}

function StatusBadge({ status }: { status: CandidateStatus }) {
  const variants: Record<CandidateStatus, { label: string; className: string }> = {
    [CandidateStatus.DRAFT]: {
      label: "Draft",
      className: "bg-gray-500 text-white border-gray-600",
    },
    [CandidateStatus.AVAILABLE]: {
      label: "Available",
      className: "bg-green-500 text-white border-green-600",
    },
    [CandidateStatus.LOCKED]: {
      label: "Locked",
      className: "bg-amber-500 text-white border-amber-600",
    },
    [CandidateStatus.UNDER_REVIEW]: {
      label: "Under Review",
      className: "bg-blue-500 text-white border-blue-600",
    },
    [CandidateStatus.APPROVED]: {
      label: "Approved",
      className: "bg-emerald-500 text-white border-emerald-600",
    },
    [CandidateStatus.IN_PROGRESS]: {
      label: "In Process",
      className: "bg-purple-500 text-white border-purple-600",
    },
    [CandidateStatus.COMPLETED]: {
      label: "Completed",
      className: "bg-slate-700 text-white border-slate-800",
    },
    [CandidateStatus.REJECTED]: {
      label: "Rejected",
      className: "bg-red-500 text-white border-red-600",
    },
  }

  const variant = variants[status] || variants[CandidateStatus.AVAILABLE]

  return (
    <Badge className={cn("shadow-md font-semibold", variant.className)}>
      {variant.label}
    </Badge>
  )
}
