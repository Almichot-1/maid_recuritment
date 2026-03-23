"use client"

import * as React from "react"
import { useParams, useRouter } from "next/navigation"
import { format, formatDistanceToNow } from "date-fns"
import { 
  Calendar, 
  CheckCircle2, 
  Download, 
  Eye, 
  FileText, 
  Loader2, 
  PencilLine,
  Trash2, 
  Upload,
  UserCheck,
  XCircle,
  AlertCircle,
  Building2,
  ChevronRight,
  Home,
  Unplug
} from "lucide-react"
import Link from "next/link"

import { useCandidate, useDeleteCandidate, usePublishCandidate, useUploadDocument } from "@/hooks/use-candidates"
import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidateShares, usePairingContext, useUnshareCandidateFromWorkspace } from "@/hooks/use-pairings"
import { useCandidateProgress, useUpdateStatusStep } from "@/hooks/use-status-steps"
import { CandidateShareDialog } from "@/components/candidates/candidate-share-dialog"
import { CandidateDetailSkeleton } from "@/components/candidates/candidate-detail-skeleton"
import { DocumentUpload } from "@/components/candidates/document-upload"
import { StatusTimeline } from "@/components/candidates/status-timeline"
import { SelectCandidateDialog } from "@/components/selections/select-candidate-dialog"
import { LockCountdown } from "@/components/selections/lock-countdown"
import { Badge, BadgeProps } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { CandidateStatus } from "@/types"
import { cn } from "@/lib/utils"

export default function CandidateDetailPage() {
  const params = useParams()
  const router = useRouter()
  const candidateId = params.id as string
  
  const { user, isEthiopianAgent, isForeignAgent } = useCurrentUser()
  const { activePairingId, activeWorkspace } = usePairingContext()
  const { data: candidate, isLoading, error } = useCandidate(candidateId)
  const showProgress = candidate?.status === CandidateStatus.IN_PROGRESS || candidate?.status === CandidateStatus.COMPLETED
  const isOwner = isEthiopianAgent && candidate?.created_by === user?.id
  const { data: candidateShares = [] } = useCandidateShares(candidateId, Boolean(isOwner))
  const { data: progressData } = useCandidateProgress(candidateId, Boolean(showProgress))
  
  const { mutate: deleteCandidate, isPending: isDeleting } = useDeleteCandidate(candidateId)
  const { mutate: publishCandidate, isPending: isPublishing } = usePublishCandidate(candidateId)
  const { mutate: uploadDocument, isPending: isUploadingDocument } = useUploadDocument(candidateId)
  const { mutate: updateStep, isPending: isUpdatingStep } = useUpdateStatusStep(candidateId)
  const { mutate: unshareFromWorkspace, isPending: isRemovingFromWorkspace } = useUnshareCandidateFromWorkspace()

  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false)
  const [selectDialogOpen, setSelectDialogOpen] = React.useState(false)
  const [publishDialogOpen, setPublishDialogOpen] = React.useState(false)
  const [imagePreview, setImagePreview] = React.useState<string | null>(null)
  const [shareDialogOpen, setShareDialogOpen] = React.useState(false)

  // Show loading skeleton
  if (isLoading) {
    return (
      <div className="space-y-6 animate-in fade-in duration-500">
        <CandidateDetailSkeleton />
      </div>
    )
  }

  // Show 404 if not found
  if (error || !candidate) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
        <div className="flex h-20 w-20 items-center justify-center rounded-full bg-destructive/10">
          <XCircle className="h-10 w-10 text-destructive" />
        </div>
        <h2 className="text-2xl font-bold">Candidate Not Found</h2>
        <p className="text-muted-foreground text-center max-w-md">
          The candidate you&apos;re looking for doesn&apos;t exist or you don&apos;t have permission to view it.
        </p>
        <Button onClick={() => router.push("/candidates")}>
          Back to Candidates
        </Button>
      </div>
    )
  }

  const activeShares = candidateShares.filter((share) => share.is_active)
  const isSharedInActiveWorkspace = Boolean(
    activePairingId && activeShares.some((share) => share.pairing_id === activePairingId)
  )
  const canDelete = isOwner && (candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE)
  const canEdit = isOwner && (candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE)
  const canSelect = isForeignAgent && candidate.status === CandidateStatus.AVAILABLE

  const getStatusBadge = (status: CandidateStatus) => {
    const variants: Record<CandidateStatus, { variant: BadgeProps["variant"]; label: string; className: string }> = {
      [CandidateStatus.DRAFT]: { 
        variant: "secondary", 
        label: "Draft", 
        className: "bg-gray-500 hover:bg-gray-600 text-white" 
      },
      [CandidateStatus.AVAILABLE]: { 
        variant: "default", 
        label: "Available", 
        className: "bg-green-500 hover:bg-green-600 text-white" 
      },
      [CandidateStatus.LOCKED]: {
        variant: "secondary",
        label: "Locked",
        className: "bg-amber-500 hover:bg-amber-600 text-white"
      },
      [CandidateStatus.UNDER_REVIEW]: {
        variant: "secondary",
        label: "Under Review",
        className: "bg-blue-500 hover:bg-blue-600 text-white"
      },
      [CandidateStatus.APPROVED]: {
        variant: "secondary",
        label: "Approved",
        className: "bg-emerald-600 hover:bg-emerald-700 text-white"
      },
      [CandidateStatus.IN_PROGRESS]: { 
        variant: "secondary", 
        label: "In Process", 
        className: "bg-blue-500 hover:bg-blue-600 text-white" 
      },
      [CandidateStatus.COMPLETED]: {
        variant: "secondary",
        label: "Completed",
        className: "bg-slate-700 hover:bg-slate-800 text-white"
      },
      [CandidateStatus.REJECTED]: { 
        variant: "destructive", 
        label: "Rejected", 
        className: "bg-red-500 hover:bg-red-600 text-white" 
      },
    }
    const config = variants[status] || variants[CandidateStatus.AVAILABLE]
    return <Badge className={cn("text-sm px-3 py-1", config.className)}>{config.label}</Badge>
  }

  const getProficiencyBadge = (language: string) => {
    // For now, showing basic badges. In future, store proficiency data
    return (
      <Badge variant="outline" className="bg-blue-50 dark:bg-blue-950/20 text-blue-700 dark:text-blue-300 border-blue-200">
        {language}
      </Badge>
    )
  }

  const getDocument = (type: string) => {
    return candidate.documents?.find(doc => doc.document_type === type)
  }

  const handleDelete = () => {
    deleteCandidate()
    setDeleteDialogOpen(false)
  }

  const handleRemoveFromWorkspace = () => {
    if (!activePairingId) {
      return
    }
    unshareFromWorkspace({ pairingId: activePairingId, candidateId })
  }

  const handlePublish = () => {
    publishCandidate()
    setPublishDialogOpen(false)
  }

  const handleUpdateStep = (stepName: string, status: string, notes?: string) => {
    updateStep({ step_name: stepName, status, notes })
  }

  const missingRequiredDocuments = ["passport", "photo"].filter((documentType) => !getDocument(documentType))
  const canGenerateCV = missingRequiredDocuments.length === 0
  const cvPageHref = `/candidates/${candidate.id}/cv`
  const trackingPageHref = `/candidates/${candidate.id}/tracking`

  const breadcrumbs = (
    <nav className="flex items-center text-sm font-medium text-muted-foreground mb-6">
      <Link href="/dashboard" className="transition-all hover:text-primary flex items-center">
        <Home className="mr-1.5 h-4 w-4" />
        Dashboard
      </Link>
      <ChevronRight className="h-4 w-4 mx-1 opacity-50" />
      <Link href="/candidates" className="transition-colors hover:text-primary">
        Candidates
      </Link>
      <ChevronRight className="h-4 w-4 mx-1 opacity-50" />
      <span className="text-foreground font-semibold truncate max-w-[200px]">
        {candidate.full_name}
      </span>
    </nav>
  )

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      {breadcrumbs}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Profile Header */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col sm:flex-row gap-6">
                {/* Photo */}
                <div className="shrink-0">
                  {getDocument("photo") ? (
                    <img
                      src={getDocument("photo")!.file_url}
                      alt={candidate.full_name}
                      className="h-32 w-32 rounded-lg object-cover border-2 border-border shadow-md cursor-pointer hover:opacity-90 transition-opacity"
                      onClick={() => setImagePreview(getDocument("photo")!.file_url)}
                    />
                  ) : (
                    <div className="h-32 w-32 rounded-lg bg-muted flex items-center justify-center border-2 border-dashed">
                      <UserCheck className="h-12 w-12 text-muted-foreground" />
                    </div>
                  )}
                </div>

                {/* Info */}
                <div className="flex-1 space-y-3">
                  <div>
                    <h1 className="text-3xl font-bold text-foreground">{candidate.full_name}</h1>
                    <div className="flex flex-wrap items-center gap-3 mt-2">
                      {getStatusBadge(candidate.status)}
                      <span className="text-sm text-muted-foreground">
                        {candidate.age ?? "N/A"} years old - {candidate.experience_years ?? 0} years experience
                      </span>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Calendar className="h-4 w-4" />
                    <span>Added on {format(new Date(candidate.created_at), "MMMM dd, yyyy")}</span>
                  </div>

                  {candidate.status === CandidateStatus.LOCKED && (
                    <div className="flex flex-col gap-2 p-3 bg-purple-50 dark:bg-purple-950/20 text-purple-800 dark:text-purple-300 rounded-lg border border-purple-200 dark:border-purple-800">
                      <div className="flex items-center gap-2">
                        <Building2 className="h-4 w-4" />
                        <span className="text-sm font-medium">
                          Reserved by a foreign agency
                        </span>
                      </div>
                      {candidate.lock_expires_at && (
                        <LockCountdown 
                          expiresAt={candidate.lock_expires_at}
                          className="text-xs"
                        />
                      )}
                    </div>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Personal Details */}
          <Card>
            <CardHeader>
              <CardTitle>Personal Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div className="space-y-1">
                  <p className="text-sm font-medium text-muted-foreground">Full Name</p>
                  <p className="text-base font-semibold">{candidate.full_name}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-sm font-medium text-muted-foreground">Age</p>
                  <p className="text-base font-semibold">{candidate.age} years</p>
                </div>
                <div className="space-y-1">
                  <p className="text-sm font-medium text-muted-foreground">Experience</p>
                  <p className="text-base font-semibold">{candidate.experience_years ?? 0} years</p>
                </div>
                {isOwner && candidate.created_by ? (
                  <div className="space-y-1">
                    <p className="text-sm font-medium text-muted-foreground">Created By</p>
                    <p className="text-base font-semibold break-all">{candidate.created_by}</p>
                  </div>
                ) : null}
                {isOwner ? (
                  <div className="space-y-1 sm:col-span-2">
                    <p className="text-sm font-medium text-muted-foreground">Partner visibility</p>
                    <p className="text-base font-semibold">
                      Changes you make here update the same candidate for every partner agency that can already see this profile.
                    </p>
                  </div>
                ) : null}
              </div>
            </CardContent>
          </Card>

          {/* Languages */}
          <Card>
            <CardHeader>
              <CardTitle>Languages</CardTitle>
            </CardHeader>
            <CardContent>
              {candidate.languages && candidate.languages.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {candidate.languages.map((lang, index) => (
                    <React.Fragment key={index}>
                      {getProficiencyBadge(lang)}
                    </React.Fragment>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No languages specified</p>
              )}
            </CardContent>
          </Card>

          {/* Skills */}
          <Card>
            <CardHeader>
              <CardTitle>Skills & Expertise</CardTitle>
            </CardHeader>
            <CardContent>
              {candidate.skills && candidate.skills.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {candidate.skills.map((skill, index) => (
                    <Badge key={index} variant="secondary" className="px-3 py-1">
                      {skill}
                    </Badge>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No skills specified</p>
              )}
            </CardContent>
          </Card>

          {/* Documents */}
          <Card>
            <CardHeader>
              <CardTitle>Documents</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Passport */}
              <div className="space-y-2">
                <p className="text-sm font-medium">Passport</p>
                {getDocument("passport") ? (
                  <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg border">
                    <FileText className="h-8 w-8 text-blue-500" />
                    <div className="flex-1">
                      <p className="text-sm font-medium">{getDocument("passport")!.file_name}</p>
                      <p className="text-xs text-muted-foreground">Uploaded</p>
                    </div>
                    <Button size="sm" variant="outline" asChild>
                      <a href={getDocument("passport")!.file_url} target="_blank" rel="noopener noreferrer">
                        <Eye className="h-4 w-4 mr-2" />
                        View
                      </a>
                    </Button>
                  </div>
                ) : (
                  isEthiopianAgent ? (
                    <DocumentUpload
                      documentType="passport"
                      title="Passport document"
                      description="Drop a PDF, JPG, or PNG passport file here."
                      accept={{
                        "application/pdf": [".pdf"],
                        "image/jpeg": [".jpg", ".jpeg"],
                        "image/png": [".png"],
                      }}
                      maxSize={10485760}
                      mode="instant"
                      disabled={isUploadingDocument}
                      onUpload={(file) => uploadDocument({ file, type: "passport" })}
                    />
                  ) : (
                    <div className="flex items-center justify-between p-3 bg-muted/30 rounded-lg border border-dashed">
                      <p className="text-sm text-muted-foreground">Not uploaded</p>
                    </div>
                  )
                )}
              </div>

              {/* Photo */}
              <div className="space-y-2">
                <p className="text-sm font-medium">Full Photo</p>
                {getDocument("photo") ? (
                  <div className="relative">
                    <img
                      src={getDocument("photo")!.file_url}
                      alt="Candidate photo"
                      className="w-full max-w-md h-64 object-cover rounded-lg border shadow-sm cursor-pointer hover:opacity-90 transition-opacity"
                      onClick={() => setImagePreview(getDocument("photo")!.file_url)}
                    />
                  </div>
                ) : (
                  isEthiopianAgent ? (
                    <DocumentUpload
                      documentType="photo"
                      title="Full body photo"
                      description="Drag and drop a clean JPG or PNG image."
                      accept={{
                        "image/jpeg": [".jpg", ".jpeg"],
                        "image/png": [".png"],
                      }}
                      maxSize={10485760}
                      mode="instant"
                      disabled={isUploadingDocument}
                      onUpload={(file) => uploadDocument({ file, type: "photo" })}
                    />
                  ) : (
                    <div className="flex items-center justify-between p-3 bg-muted/30 rounded-lg border border-dashed">
                      <p className="text-sm text-muted-foreground">Not uploaded</p>
                    </div>
                  )
                )}
              </div>

              {/* Video Interview */}
              <div className="space-y-2">
                <p className="text-sm font-medium">Video Interview</p>
                {getDocument("video") ? (
                  <video
                    src={getDocument("video")!.file_url}
                    controls
                    className="w-full max-w-2xl rounded-lg border shadow-sm"
                  />
                ) : (
                  isEthiopianAgent ? (
                    <DocumentUpload
                      documentType="video"
                      title="Video interview"
                      description="Drag and drop the MP4 interview file."
                      accept={{ "video/mp4": [".mp4"] }}
                      maxSize={52428800}
                      mode="instant"
                      disabled={isUploadingDocument}
                      onUpload={(file) => uploadDocument({ file, type: "video" })}
                    />
                  ) : (
                    <div className="flex items-center justify-between p-3 bg-muted/30 rounded-lg border border-dashed">
                      <p className="text-sm text-muted-foreground">Not uploaded</p>
                    </div>
                  )
                )}
              </div>
            </CardContent>
          </Card>

          {/* CV Section */}
          <Card>
            <CardHeader>
              <CardTitle>Curriculum Vitae</CardTitle>
            </CardHeader>
            <CardContent>
              {candidate.cv_pdf_url ? (
                <div className="space-y-3">
                  <div className="flex items-center gap-3 p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
                    <CheckCircle2 className="h-5 w-5 text-green-600" />
                    <span className="text-sm font-medium text-green-800 dark:text-green-300">CV Generated</span>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Open the dedicated CV page to preview the final PDF and download it cleanly.
                  </p>
                  <Button variant="outline" asChild>
                    <Link href={cvPageHref}>
                      <Download className="h-4 w-4 mr-2" />
                      Download CV
                    </Link>
                  </Button>
                </div>
              ) : (
                <div className="space-y-3">
                  <p className="text-sm text-muted-foreground">No CV generated yet</p>
                  {missingRequiredDocuments.length > 0 && (
                    <p className="text-sm text-amber-700 dark:text-amber-300">
                      Upload {missingRequiredDocuments.join(" and ")} before the CV can be downloaded.
                    </p>
                  )}
                  {isEthiopianAgent && canGenerateCV ? (
                    <Button asChild>
                      <Link href={cvPageHref}>
                        <Download className="h-4 w-4 mr-2" />
                        Download CV
                      </Link>
                    </Button>
                  ) : null}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Recruitment Progress */}
          {showProgress && progressData && progressData.steps.length > 0 && (
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>Recruitment Progress</CardTitle>
                  <span className="text-sm font-semibold text-muted-foreground">
                    {Math.round(progressData.progress_percentage)}% Complete
                  </span>
                </div>
                <div className="w-full h-2 bg-muted rounded-full overflow-hidden mt-2">
                  <div
                    className="h-full bg-gradient-to-r from-blue-500 to-green-500 transition-all duration-500"
                    style={{ width: `${progressData.progress_percentage}%` }}
                  />
                </div>
              </CardHeader>
              <CardContent>
                <StatusTimeline
                  steps={progressData.steps}
                  canUpdate={isOwner}
                  onUpdateStep={handleUpdateStep}
                  isUpdating={isUpdatingStep}
                />
              </CardContent>
            </Card>
          )}
        </div>

        {/* Right Column - Actions & Info */}
        <div className="space-y-6">
          {/* Actions Card */}
          <Card className="lg:sticky lg:top-6">
            <CardHeader>
              <CardTitle>Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {/* Ethiopian Agent Actions */}
              {isEthiopianAgent && (
                <>
                  {canEdit && (
                    <Button className="w-full" variant="outline" asChild>
                      <Link href={`/candidates/${candidate.id}/edit`}>
                        <PencilLine className="mr-2 h-4 w-4" />
                        Edit Candidate
                      </Link>
                    </Button>
                  )}

                  {(candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE) && (
                    <Button className="w-full" variant="secondary" onClick={() => setShareDialogOpen(true)}>
                      <Upload className="mr-2 h-4 w-4" />
                      Share With Partners
                    </Button>
                  )}

                  {isSharedInActiveWorkspace && (candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE) && (
                    <Button
                      className="w-full"
                      variant="outline"
                      onClick={handleRemoveFromWorkspace}
                      disabled={isRemovingFromWorkspace}
                    >
                      {isRemovingFromWorkspace && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      {!isRemovingFromWorkspace && <Unplug className="mr-2 h-4 w-4" />}
                      Remove from This Partner
                    </Button>
                  )}

                  {candidate.status === CandidateStatus.DRAFT && (
                    <Button 
                      className="w-full" 
                      variant="outline"
                      onClick={() => setPublishDialogOpen(true)}
                      disabled={isPublishing}
                    >
                      {isPublishing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      <Upload className="h-4 w-4 mr-2" />
                      Publish Candidate
                    </Button>
                  )}

                  {(candidate.cv_pdf_url || canGenerateCV) && (
                    <Button className="w-full" variant="outline" asChild>
                      <Link href={cvPageHref}>
                        <Download className="h-4 w-4 mr-2" />
                        Download CV
                      </Link>
                    </Button>
                  )}

                  {showProgress && (
                    <Button className="w-full" variant="outline" asChild>
                      <Link href={trackingPageHref}>
                        <Eye className="h-4 w-4 mr-2" />
                        Open Process Tracking
                      </Link>
                    </Button>
                  )}

                  {canDelete && (
                    <>
                      <Separator />
                      <Button 
                        className="w-full" 
                        variant="destructive"
                        onClick={() => setDeleteDialogOpen(true)}
                        disabled={isDeleting}
                      >
                        {isDeleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete From Candidate Library
                      </Button>
                    </>
                  )}

                  {isOwner && candidate.status === CandidateStatus.AVAILABLE && (
                    <p className="rounded-lg border border-border/70 bg-muted/40 p-3 text-xs text-muted-foreground">
                      This candidate is already published in your library. Editing it updates every partner that can see it, removing it from one partner only affects that partner, and deleting it from your library removes it from all partners at once.
                    </p>
                  )}

                  {isOwner && activeShares.length > 0 && (
                    <div className="rounded-lg border border-sky-200 bg-sky-50 p-3 text-xs text-sky-900 dark:border-sky-900/60 dark:bg-sky-950/30 dark:text-sky-100">
                      This candidate is visible to {activeShares.length} partner agenc{activeShares.length === 1 ? "y" : "ies"}. Removing it from {activeWorkspace?.partner_agency.company_name || activeWorkspace?.partner_agency.full_name || "the current partner"} will not affect the others, and any edits you make here will stay synced everywhere it is still shared.
                    </div>
                  )}

                  {isOwner && canDelete && activeShares.length > 0 && (
                    <p className="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-900 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-100">
                      Deleting from your candidate library will remove this candidate from all partner agencies and clear it from every place it is currently shared.
                    </p>
                  )}
                </>
              )}

              {/* Foreign Agent Actions */}
              {isForeignAgent && (
                <>
                  {candidate.cv_pdf_url ? (
                    <Button className="w-full" variant="outline" asChild>
                      <Link href={cvPageHref}>
                        <Download className="h-4 w-4 mr-2" />
                        Download CV
                      </Link>
                    </Button>
                  ) : null}

                  {showProgress ? (
                    <Button className="w-full" variant="outline" asChild>
                      <Link href={trackingPageHref}>
                        <Eye className="h-4 w-4 mr-2" />
                        Open Process Tracking
                      </Link>
                    </Button>
                  ) : null}

                  {canSelect ? (
                    <Button 
                      className="w-full bg-green-600 hover:bg-green-700 text-white" 
                      size="lg"
                      onClick={() => setSelectDialogOpen(true)}
                    >
                      <UserCheck className="h-4 w-4 mr-2" />
                      Select Candidate
                    </Button>
                  ) : candidate.status === CandidateStatus.LOCKED ? (
                    <div className="space-y-3">
                      {candidate.locked_by === user?.id ? (
                        <div className="p-4 bg-purple-50 dark:bg-purple-950/20 rounded-lg border border-purple-200 dark:border-purple-800">
                          <div className="flex items-center gap-2 mb-3">
                            <CheckCircle2 className="h-5 w-5 text-purple-600" />
                            <p className="text-sm font-medium text-purple-800 dark:text-purple-300">
                              Selected by You
                            </p>
                          </div>
                          {candidate.lock_expires_at && (
                            <LockCountdown 
                              expiresAt={candidate.lock_expires_at}
                              className="text-sm"
                            />
                          )}
                        </div>
                      ) : (
                        <div className="p-4 bg-purple-50 dark:bg-purple-950/20 rounded-lg border border-purple-200 dark:border-purple-800 text-center">
                          <AlertCircle className="h-8 w-8 text-purple-600 mx-auto mb-2" />
                          <p className="text-sm font-medium text-purple-800 dark:text-purple-300">
                            Currently Unavailable
                          </p>
                          <p className="text-xs text-purple-600 dark:text-purple-400 mt-1">
                            Selected by another agency
                          </p>
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="p-4 bg-muted rounded-lg text-center">
                      <p className="text-sm text-muted-foreground">
                        This candidate is not available for selection
                      </p>
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>

          {/* Additional Info Card */}
          <Card>
            <CardHeader>
              <CardTitle>Additional Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">Experience</p>
                <p className="text-base font-semibold">{candidate.experience_years} years</p>
              </div>
              <Separator />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">Created</p>
                <p className="text-sm">{format(new Date(candidate.created_at), "MMM dd, yyyy")}</p>
                <p className="text-xs text-muted-foreground">
                  {formatDistanceToNow(new Date(candidate.created_at), { addSuffix: true })}
                </p>
              </div>
              <Separator />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">Last Updated</p>
                <p className="text-sm">{format(new Date(candidate.updated_at), "MMM dd, yyyy")}</p>
                <p className="text-xs text-muted-foreground">
                  {formatDistanceToNow(new Date(candidate.updated_at), { addSuffix: true })}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
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
            <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
              {isDeleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Selection Dialog */}
      <SelectCandidateDialog
        candidate={candidate}
        open={selectDialogOpen}
        onOpenChange={setSelectDialogOpen}
      />

      {/* Publish Confirmation Dialog */}
      <Dialog open={publishDialogOpen} onOpenChange={setPublishDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Publish Candidate</DialogTitle>
            <DialogDescription>
              Publishing will add {candidate.full_name} to your candidate library so you can share the profile with one or more partner agencies.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPublishDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handlePublish} disabled={isPublishing}>
              {isPublishing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Publish
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Image Preview Dialog */}
      <Dialog open={!!imagePreview} onOpenChange={() => setImagePreview(null)}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>Photo Preview</DialogTitle>
          </DialogHeader>
          {imagePreview && (
            <img
              src={imagePreview}
              alt="Preview"
              className="w-full h-auto rounded-lg"
            />
          )}
        </DialogContent>
      </Dialog>

      <CandidateShareDialog
        candidate={candidate}
        open={shareDialogOpen}
        onOpenChange={setShareDialogOpen}
      />
    </div>
  )
}
