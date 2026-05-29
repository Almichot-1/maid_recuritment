"use client"

import * as React from "react"
import dynamic from "next/dynamic"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { Lock, PencilLine, User } from "lucide-react"

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

  const photoUrl = candidate.documents?.find((doc) => doc.document_type === "photo")?.file_url
  const displaySkills = candidate.skills.slice(0, 3)
  const remainingSkills = candidate.skills.length - 3
  const isLockedByCurrentUser = candidate.status === CandidateStatus.LOCKED && candidate.locked_by === user?.id
  const isLockedByOther = candidate.status === CandidateStatus.LOCKED && candidate.locked_by !== user?.id
  const isOwner = isEthiopianAgent && candidate.created_by === user?.id
  const canEdit = isOwner && (candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE)

  const openCandidate = () => {
    router.push(`/candidates/${candidate.id}`)
  }

  const handleCardKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault()
      openCandidate()
    }
  }

  return (
    <>
      <Card
        role="button"
        tabIndex={0}
        className="cursor-pointer card-lift overflow-hidden"
        onClick={openCandidate}
        onKeyDown={handleCardKeyDown}
      >
        <CardContent className="p-0">
          <div className="relative aspect-square w-full overflow-hidden border-b border-border bg-muted/20">
            {photoUrl ? (
              <img src={photoUrl} alt={candidate.full_name} className="h-full w-full object-cover" loading="lazy" />
            ) : (
              <div className="flex h-full w-full items-center justify-center bg-background">
                <User className="h-20 w-20 text-muted-foreground/40" />
              </div>
            )}

            <div className="absolute right-3 top-3">
              <StatusBadge status={candidate.status} />
            </div>

            {(candidate.status === CandidateStatus.LOCKED || candidate.status === CandidateStatus.IN_PROGRESS) && (
              <div className="absolute left-3 top-3">
                <Badge variant="outline" className="border-primary text-primary">
                  <Lock className="h-3 w-3" />
                  Locked
                </Badge>
              </div>
            )}
          </div>

          <div className="space-y-4 p-4">
            <div className="space-y-2">
              <p className="font-display text-3xl leading-none text-foreground">{candidate.full_name}</p>
              <p className="text-sm text-muted-foreground">
                {candidate.age ?? "N/A"} years old · {candidate.experience_years ?? 0} years experience
              </p>
            </div>

            {candidate.languages.length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {candidate.languages.map((lang) => (
                  <Badge key={lang} variant="outline">
                    {lang}
                  </Badge>
                ))}
              </div>
            ) : null}

            {candidate.skills.length > 0 ? (
              <div className="space-y-2">
                <p className="route-stamp text-[11px] text-muted-foreground">Skills</p>
                <div className="flex flex-wrap gap-2">
                  {displaySkills.map((skill) => (
                    <Badge key={skill} variant="outline">
                      {skill}
                    </Badge>
                  ))}
                  {remainingSkills > 0 ? <Badge variant="outline">+{remainingSkills}</Badge> : null}
                </div>
              </div>
            ) : null}

            {isForeignAgent && candidate.status === CandidateStatus.AVAILABLE ? (
              <Button
                onClick={(event) => {
                  event.stopPropagation()
                  setSelectDialogOpen(true)
                }}
                className="w-full"
                size="sm"
              >
                Select candidate
              </Button>
            ) : null}

            {isForeignAgent && isLockedByCurrentUser && candidate.lock_expires_at ? (
              <div className="border border-border bg-muted/20 p-3">
                <p className="text-xs font-bold uppercase tracking-[0.06em] text-foreground">Selected by you</p>
                <div className="mt-2">
                  <LockCountdown expiresAt={candidate.lock_expires_at} className="text-xs" showIcon />
                </div>
              </div>
            ) : null}

            {isForeignAgent && isLockedByOther ? (
              <div className="border border-border bg-muted/20 p-3 text-sm text-muted-foreground">
                Locked by another agency
              </div>
            ) : null}

            {isOwner ? (
              <div className="flex gap-2">
                <Button
                  variant="outline"
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
                      <PencilLine className="h-4 w-4" />
                      Edit
                    </Link>
                  </Button>
                ) : null}
                <Button variant="ghost" className="flex-1" size="sm" asChild onClick={(event) => event.stopPropagation()}>
                  <Link href={`/candidates/${candidate.id}`}>Open</Link>
                </Button>
              </div>
            ) : null}
          </div>
        </CardContent>
      </Card>

      <SelectCandidateDialog candidate={candidate} open={selectDialogOpen} onOpenChange={setSelectDialogOpen} />
      <CandidateShareDialog candidate={candidate} open={shareDialogOpen} onOpenChange={setShareDialogOpen} />
    </>
  )
}

function StatusBadge({ status }: { status: CandidateStatus }) {
  const variants: Record<CandidateStatus, { label: string; className: string }> = {
    [CandidateStatus.DRAFT]: {
      label: "Draft",
      className: "border-border text-muted-foreground",
    },
    [CandidateStatus.AVAILABLE]: {
      label: "Available",
      className: "border-[color:var(--color-success)] text-[color:var(--color-success)]",
    },
    [CandidateStatus.LOCKED]: {
      label: "Locked",
      className: "border-[color:var(--color-warning)] text-[color:var(--color-warning)]",
    },
    [CandidateStatus.UNDER_REVIEW]: {
      label: "Under review",
      className: "border-[color:var(--color-info)] text-[color:var(--color-info)]",
    },
    [CandidateStatus.APPROVED]: {
      label: "Approved",
      className: "border-[color:var(--color-success)] text-[color:var(--color-success)]",
    },
    [CandidateStatus.IN_PROGRESS]: {
      label: "In progress",
      className: "border-primary text-primary",
    },
    [CandidateStatus.COMPLETED]: {
      label: "Completed",
      className: "border-foreground text-foreground",
    },
    [CandidateStatus.REJECTED]: {
      label: "Rejected",
      className: "border-[color:var(--color-danger)] text-[color:var(--color-danger)]",
    },
  }

  const variant = variants[status] || variants[CandidateStatus.AVAILABLE]

  return <Badge variant="outline" className={cn(variant.className)}>{variant.label}</Badge>
}
