"use client"

import * as React from "react"
import dynamic from "next/dynamic"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { Loader2, Lock, PencilLine, User, X, CheckCircle, Circle, Trash2, Download, Share2, Unlock } from "lucide-react"

import Image from "next/image"
import { Candidate, CandidateStatus } from "@/types"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePublishCandidate, useDeleteCandidate, useLockCandidate, useUnlockCandidate, downloadCandidateCVFile } from "@/hooks/use-candidates"
import { PublishButton } from "./publish-button"
import { usePairingContext } from "@/hooks/use-pairings"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { toast } from "sonner"

import { cn } from "@/lib/utils"
import { buildCandidateMessage, shareOnWhatsApp } from "@/lib/whatsapp"

const CandidateShareDialog = dynamic(
  () => import("@/components/candidates/candidate-share-dialog").then((module) => module.CandidateShareDialog)
)
const SelectCandidateDialog = dynamic(
  () => import("@/components/selections/select-candidate-dialog").then((module) => module.SelectCandidateDialog)
)

interface CandidateCardProps {
  candidate: Candidate
  selectable?: boolean
  selected?: boolean
  onSelectionChange?: (candidateId: string, selected: boolean) => void
}

export function CandidateCard({ candidate, selectable = false, selected = false, onSelectionChange }: CandidateCardProps) {
  const router = useRouter()
  const { user, isEthiopianAgent, isForeignAgent } = useCurrentUser()
  const { context, activeWorkspace } = usePairingContext()
  const [selectDialogOpen, setSelectDialogOpen] = React.useState(false)
  const [shareDialogOpen, setShareDialogOpen] = React.useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false)
  const [isPhotoExpanded, setIsPhotoExpanded] = React.useState(false)
  const [isDownloadingCV, setIsDownloadingCV] = React.useState(false)
  const { mutateAsync: publishCandidate, isPending: isPublishing } = usePublishCandidate(candidate.id)
  const { mutate: deleteCandidate, isPending: isDeleting } = useDeleteCandidate(candidate.id)
  const { mutate: lockCandidate, isPending: isLocking } = useLockCandidate()
  const { mutate: unlockCandidate, isPending: isUnlocking } = useUnlockCandidate()

  const handleCardClick = () => {
    if (selectable && onSelectionChange) {
      onSelectionChange(candidate.id, !selected)
    } else {
      router.push(`/candidates/${candidate.id}`)
    }
  }

  const handleSelectClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    setSelectDialogOpen(true)
  }

  const handleDownloadCV = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!candidate.cv_pdf_url) return
    try {
      setIsDownloadingCV(true)
      await downloadCandidateCVFile(candidate.id, candidate.full_name)
    } catch {
      toast.error("Failed to download CV")
    } finally {
      setIsDownloadingCV(false)
    }
  }

  const handleWhatsAppShare = (e: React.MouseEvent) => {
    e.stopPropagation()
    const msg = buildCandidateMessage(candidate)
    shareOnWhatsApp(msg)
  }

  const handleLock = (e: React.MouseEvent) => {
    e.stopPropagation()
    lockCandidate(candidate.id)
  }

  const handleUnlock = (e: React.MouseEvent) => {
    e.stopPropagation()
    unlockCandidate(candidate.id)
  }

  // Get photo URL from documents
  const photoUrl = candidate.documents?.find((doc) => doc.document_type === "photo")?.file_url

  // Derive country + salary from override or workspace defaults
  const override = candidate.pair_overrides?.find((o) => o.pairing_id === activeWorkspace?.id)
  const displayCountry = override?.country_applied || activeWorkspace?.default_country
  const displaySalary = override?.salary_offered || activeWorkspace?.default_salary

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
          <div 
            className="relative w-full aspect-square bg-muted overflow-hidden cursor-zoom-in"
            onClick={(e) => {
              if (photoUrl) {
                e.stopPropagation()
                setIsPhotoExpanded(true)
              }
            }}
          >
            {photoUrl ? (
              <Image
                src={photoUrl}
                alt={candidate.full_name}
                fill
                unoptimized
                className="object-cover object-top transition-transform duration-300 group-hover:scale-105"
                loading="lazy"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-primary/10 to-primary/5">
                <User className="h-20 w-20 text-muted-foreground/40" />
              </div>
            )}
            
            {/* Status Badge */}
            <div className="absolute top-3 left-3">
              <StatusBadge status={candidate.status} />
            </div>

            {/* Selection Checkbox */}
            {selectable && (
              <div
                className="absolute top-3 right-3 z-10"
                onClick={(e) => {
                  e.stopPropagation()
                  if (onSelectionChange) {
                    onSelectionChange(candidate.id, !selected)
                  }
                }}
              >
                <div className="bg-white dark:bg-background rounded-full p-1 shadow-md cursor-pointer hover:scale-110 transition-transform">
                  {selected ? (
                    <CheckCircle className="h-6 w-6 text-emerald-600 fill-emerald-100" />
                  ) : (
                    <Circle className="h-6 w-6 text-gray-400" />
                  )}
                </div>
              </div>
            )}

            {/* Lock Indicator */}
            {!selectable && (candidate.status === CandidateStatus.LOCKED || candidate.status === CandidateStatus.IN_PROGRESS) && (
              <div className="absolute top-3 right-3">
                <div className="flex items-center gap-1.5 bg-purple-500/90 text-white px-2 py-1 rounded-md text-xs font-semibold shadow-md">
                  <Lock className="h-3 w-3" />
                  <span>Locked</span>
                </div>
              </div>
            )}
          </div>

          {/* Content */}
          <div className="p-3 space-y-2">
            {/* Name */}
            <h3 className="font-semibold text-base leading-tight line-clamp-1">
              {candidate.full_name}
            </h3>

            {/* Age, Experience & Religion */}
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span>{candidate.age ?? "N/A"} years</span>
              <span>-</span>
              <span>{candidate.experience_years ?? 0} yrs exp.</span>
              {candidate.religion && (
                <>
                  <span>-</span>
                  <span>{candidate.religion}</span>
                </>
              )}
            </div>

            {/* Country & Salary */}
            {(displayCountry || displaySalary) && (
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                {displayCountry && <span>{displayCountry}</span>}
                {displayCountry && displaySalary && <span>|</span>}
                {displaySalary && <span>{displaySalary}</span>}
              </div>
            )}

            {(candidate.gender || candidate.experience_abroad) && (
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                {candidate.gender && <span>{candidate.gender}</span>}
                {candidate.gender && candidate.experience_abroad && candidate.experience_abroad.length > 0 && <span>-</span>}
                {candidate.experience_abroad && candidate.experience_abroad.length > 0 && (
                  <span>
                    {Array.isArray(candidate.experience_abroad)
                      ? candidate.experience_abroad.map((e) => e.country).join(", ")
                      : candidate.experience_abroad}
                  </span>
                )}
              </div>
            )}

            {/* Languages */}
            {candidate.languages.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {candidate.languages.map((lang, idx) => (
                  <Badge key={idx} variant="outline" className="text-[10px] px-1.5 py-0 h-4">
                    {typeof lang === "string" ? lang : `${lang.language}${lang.proficiency ? ` - ${lang.proficiency}` : ""}`}
                  </Badge>
                ))}
              </div>
            )}



            {/* Foreign Agent Actions */}
            {isForeignAgent && candidate.status === CandidateStatus.AVAILABLE && (
              <div className="mt-2 space-y-2">
                <div className="flex gap-2">
                  <Button
                    onClick={handleLock}
                    variant="outline"
                    size="sm"
                    className="flex-1 border-amber-300 text-amber-700 hover:bg-amber-50 dark:border-amber-700 dark:text-amber-400 dark:hover:bg-amber-950/30"
                    disabled={isLocking}
                  >
                    {isLocking ? <Loader2 className="h-4 w-4 animate-spin" /> : <Lock className="h-4 w-4" />}
                    <span className="ml-1">Hold</span>
                  </Button>
                  <Button
                    onClick={handleSelectClick}
                    size="sm"
                    className="flex-1 bg-green-600 hover:bg-green-700"
                  >
                    <span>Select</span>
                  </Button>
                </div>
                {candidate.cv_pdf_url && (
                  <div className="flex gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="flex-1 text-xs"
                      onClick={handleDownloadCV}
                      disabled={isDownloadingCV}
                    >
                      {isDownloadingCV ? <Loader2 className="h-3 w-3 animate-spin mr-1" /> : <Download className="h-3 w-3 mr-1" />}
                      CV
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="flex-1 text-xs"
                      onClick={handleWhatsAppShare}
                    >
                      <Share2 className="h-3 w-3 mr-1" />
                      WhatsApp
                    </Button>
                  </div>
                )}
              </div>
            )}

            {/* Locked by Current User - Status */}
            {isForeignAgent && isLockedByCurrentUser && (
              <div className="mt-2 space-y-2">
                <div className="p-2 bg-purple-50 dark:bg-purple-950/20 rounded-md border border-purple-200 dark:border-purple-800 text-center">
                  <p className="text-xs font-medium text-purple-900 dark:text-purple-100">
                    On hold — select or release
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button
                    onClick={handleSelectClick}
                    size="sm"
                    className="flex-1 bg-green-600 hover:bg-green-700"
                  >
                    Select
                  </Button>
                  <Button
                    onClick={handleUnlock}
                    variant="outline"
                    size="sm"
                    className="flex-1"
                    disabled={isUnlocking}
                  >
                    {isUnlocking ? <Loader2 className="h-4 w-4 animate-spin" /> : <Unlock className="h-4 w-4" />}
                    <span className="ml-1">Release</span>
                  </Button>
                </div>
                {candidate.cv_pdf_url && (
                  <div className="flex gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="flex-1 text-xs"
                      onClick={handleDownloadCV}
                      disabled={isDownloadingCV}
                    >
                      {isDownloadingCV ? <Loader2 className="h-3 w-3 animate-spin mr-1" /> : <Download className="h-3 w-3 mr-1" />}
                      CV
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="flex-1 text-xs"
                      onClick={handleWhatsAppShare}
                    >
                      <Share2 className="h-3 w-3 mr-1" />
                      WhatsApp
                    </Button>
                  </div>
                )}
              </div>
            )}

            {/* Locked by Another Agent - hidden by backend filter, but keep fallback */}
            {isForeignAgent && isLockedByOther && (
              <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground mt-2 p-2 bg-muted/50 rounded-md">
                <Lock className="h-4 w-4" />
                <span>Locked by another agency</span>
              </div>
            )}

            {/* Publish Button for Draft Candidates (Ethiopian Agent Owner) */}
            {isOwner && candidate.status === CandidateStatus.DRAFT && (
              <div onClick={(e) => e.stopPropagation()}>
                <PublishButton
                  workspaces={context?.workspaces || []}
                  isPublishing={isPublishing}
                  onPublish={async (pairingId) => {
                    await publishCandidate({ pairingId })
                  }}
                />
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
                {canEdit ? (
                  <Button
                    variant="destructive"
                    className="flex-1"
                    size="sm"
                    onClick={(event) => {
                      event.stopPropagation()
                      setDeleteDialogOpen(true)
                    }}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </Button>
                ) : null}
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Photo Modal */}
      {isPhotoExpanded && photoUrl && (
        <div 
          className="fixed inset-0 z-[100] flex items-center justify-center bg-black/90 backdrop-blur-sm cursor-zoom-out p-4"
          onClick={(e) => {
            e.stopPropagation()
            setIsPhotoExpanded(false)
          }}
        >
          <div className="relative max-w-4xl max-h-[90vh] w-full h-full flex items-center justify-center">
            <Button
              variant="ghost"
              size="icon"
              className="absolute top-0 right-0 text-white/70 hover:text-white hover:bg-white/20 rounded-full"
              onClick={(e) => {
                e.stopPropagation()
                setIsPhotoExpanded(false)
              }}
            >
              <X className="h-6 w-6" />
            </Button>
            <Image 
              src={photoUrl} 
              alt={candidate.full_name} 
              fill
              unoptimized
              className="object-contain rounded-lg shadow-2xl"
            />
          </div>
        </div>
      )}

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



      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent onClick={(e) => e.stopPropagation()}>
          <DialogHeader>
            <DialogTitle>Delete From Candidate Library</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete {candidate.full_name} from your candidate library? This will remove the candidate from every partner agency and cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                deleteCandidate()
                setDeleteDialogOpen(false)
              }}
              disabled={isDeleting}
            >
              {isDeleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
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
