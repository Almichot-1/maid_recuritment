"use client"

import * as React from "react"
import { AxiosError } from "axios"
import Link from "next/link"
import { useParams, useRouter } from "next/navigation"
import { format } from "date-fns"
import {
  AlertTriangle,
  BadgeCheck,
  ChevronRight,
  Clock,
  Eye,
  FileText,
  Home,
  Loader2,
  MessageSquare,
  Sparkles,
  User,
  XCircle,
} from "lucide-react"
import { toast } from "sonner"

import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidate, useDeleteCandidateDocument, useUploadDocument } from "@/hooks/use-candidates"
import {
  useApproveSelection,
  useRejectSelection,
  useSelection,
  useSelectionApprovals,
  useUploadSelectionDocument,
} from "@/hooks/use-selections"
import { useSelectionProgress, useUpdateProgress } from "@/hooks/use-selection-progress"
import { mapProgressToSteps, mapOldStatusToNew, stepToPayload } from "@/components/process-tracking/process-tracking-page"
import { resolveCandidateChatThread } from "@/hooks/use-chat"
import { SelectionStatus } from "@/types"
import { ApprovalDialog } from "@/components/selections/approval-dialog"
import { StatusTimeline } from "@/components/candidates/status-timeline"
import { DocumentUpload } from "@/components/candidates/document-upload"
import Image from "next/image"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

export default function SelectionDetailPage() {
  const params = useParams()
  const router = useRouter()
  const selectionId = String(params.id || "")
  const { user, isEthiopianAgent } = useCurrentUser()
  const { data: selection, isLoading: isSelectionLoading } = useSelection(selectionId)
  const candidateId = selection?.candidate_id
  const { data: trackingCandidate } = useCandidate(candidateId)
  const { data: selectionProgress, isLoading: isProgressLoading } = useSelectionProgress(selectionId)
  const { data: approvalStatus, isLoading: isApprovalsLoading } = useSelectionApprovals(selectionId)
  const { mutate: approveSelection, isPending: isApproving } = useApproveSelection(selectionId, candidateId)
  const { mutate: rejectSelection, isPending: isRejecting } = useRejectSelection(selectionId, candidateId)
  const { mutateAsync: uploadSelectionDocument, isPending: isUploadingSelectionDocument } = useUploadSelectionDocument(selectionId)
  const updateProgress = useUpdateProgress(selectionId || "")
  const isUpdatingStep = updateProgress.isPending
  const { mutateAsync: uploadCandidateDocument, isPending: isUploadingMedicalDocument } = useUploadDocument(candidateId || "")
  const { mutateAsync: removeCandidateDocument, isPending: isRemovingMedicalDocument } = useDeleteCandidateDocument(candidateId || "")

  const [approveDialogOpen, setApproveDialogOpen] = React.useState(false)
  const [rejectDialogOpen, setRejectDialogOpen] = React.useState(false)
  const [isOpeningChat, setIsOpeningChat] = React.useState(false)
  const [replacingDocumentType, setReplacingDocumentType] = React.useState<"contract" | "employer_id" | null>(null)
  const [activeUploadType, setActiveUploadType] = React.useState<"contract" | "employer_id" | null>(null)
  const [uploadProgress, setUploadProgress] = React.useState<Record<"contract" | "employer_id", number>>({
    contract: 0,
    employer_id: 0,
  })

  if (isSelectionLoading || isApprovalsLoading || (candidateId && isProgressLoading)) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!selection || !selection.candidate) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 text-center">
        <XCircle className="h-10 w-10 text-destructive" />
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold">Selection not found</h1>
          <p className="text-muted-foreground">This selection is unavailable or you do not have access to it.</p>
        </div>
        <Button onClick={() => router.push("/selections")}>Back to Selections</Button>
      </div>
    )
  }

  const candidate = selection.candidate
  const userHasApproved = isEthiopianAgent ? selection.ethiopian_approved : selection.foreign_approved
  const isPending = selection.status === SelectionStatus.PENDING
  const isApproved = selection.status === SelectionStatus.APPROVED
  const isRejected = selection.status === SelectionStatus.REJECTED
  const isExpired = selection.status === SelectionStatus.EXPIRED
  const steps = selectionProgress ? mapProgressToSteps(selectionProgress) : []
  const showTrackingTimeline = steps.length > 0
  const trackingPageHref = `/candidates/${selection.candidate_id}/tracking`
  const canUpdateProgress = isEthiopianAgent && candidate.created_by === user?.id
  const hasEmployerContract = !!selection.employer_contract?.file_url
  const failedStep = steps.find((step) => step.step_status === "failed")
  const completedSteps = steps.filter((step) => step.step_status === "completed").length
  const progressPercentage = steps.length > 0 ? Math.round((completedSteps / steps.length) * 100) : 0
  const medicalDocument = trackingCandidate?.documents?.find((document) => document.document_type === "medical")

  const handleUpdateStep = async (stepName: string, status: string, notes?: string, cocStatus?: string, arrivalCity?: string) => {
    if (!selectionId || !canUpdateProgress) {
      return
    }
    const mappedStatus = mapOldStatusToNew(stepName, status)
    const payload: Record<string, string> = {
      ...stepToPayload(stepName, mappedStatus),
      ...(cocStatus ? { coc_type: cocStatus } : {}),
      ...(arrivalCity ? { arrival_city: arrivalCity } : {}),
    }
    await updateProgress.mutateAsync(payload)
  }

  const handleRemoveMedicalDocument = async () => {
    if (!medicalDocument?.id || !canUpdateProgress) {
      return
    }
    await removeCandidateDocument({ documentId: medicalDocument.id })
  }

  const handleUploadSelectionDocument = async (type: "contract" | "employer_id", file: File) => {
    setActiveUploadType(type)
    setUploadProgress((current) => ({ ...current, [type]: 0 }))

    try {
      await uploadSelectionDocument({
        type,
        file,
        onProgress: (progress) => {
          setUploadProgress((current) => ({ ...current, [type]: progress }))
        },
      })
      setReplacingDocumentType((current) => (current === type ? null : current))
    } finally {
      setActiveUploadType((current) => (current === type ? null : current))
      setUploadProgress((current) => ({ ...current, [type]: 0 }))
    }
  }

  const handleApprove = () => {
    approveSelection(undefined, {
      onSuccess: () => setApproveDialogOpen(false),
    })
  }

  const handleReject = (reason?: string) => {
    rejectSelection(
      { reason: reason || "" },
      {
        onSuccess: () => setRejectDialogOpen(false),
      }
    )
  }

  const handleOpenCandidateChat = async () => {
    if (!selection.candidate_id) {
      return
    }

    setIsOpeningChat(true)
    try {
      await resolveCandidateChatThread(selection.candidate_id)
      router.push(`/partners/chat?candidate_id=${selection.candidate_id}`)
    } catch (error) {
      const status = (error as AxiosError<{ error?: string; message?: string }>).response?.status
      if (status === 403 || status === 404) {
        toast.error("Candidate chat is not accessible in this workspace.")
      } else if (status === 401) {
        toast.error("Your session has expired. Please sign in again.")
      } else {
        toast.error("Could not open candidate chat right now.")
      }
    } finally {
      setIsOpeningChat(false)
    }
  }

  return (
    <div className="space-y-6 pb-10">
      <nav className="flex items-center text-sm font-medium text-muted-foreground">
        <Link href="/dashboard" className="transition-colors hover:text-primary flex items-center">
          <Home className="mr-1.5 h-4 w-4" />
          Dashboard
        </Link>
        <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
        <Link href="/selections" className="transition-colors hover:text-primary">
          Selections
        </Link>
        <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
        <span className="text-foreground font-semibold">{candidate.full_name}</span>
      </nav>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col gap-6 sm:flex-row">
                <div className="shrink-0">
                  {candidate.photo_url ? (
                    <div className="relative h-32 w-32">
                      <Image
                        src={candidate.photo_url}
                        alt={candidate.full_name}
                        fill
                        unoptimized
                        className="rounded-lg border object-cover shadow-sm"
                      />
                    </div>
                  ) : (
                    <div className="flex h-32 w-32 items-center justify-center rounded-lg border border-dashed bg-muted">
                      <User className="h-10 w-10 text-muted-foreground" />
                    </div>
                  )}
                </div>

                <div className="flex-1 space-y-3">
                  <div>
                    <h1 className="text-3xl font-bold">{candidate.full_name}</h1>
                    <div className="mt-2 flex flex-wrap items-center gap-3">
                      <SelectionStatusBadge status={selection.status} />
                      <span className="text-sm text-muted-foreground">
                        {candidate.age ?? "N/A"} years old - {candidate.experience_years ?? 0} years experience
                      </span>
                    </div>
                  </div>

                  <div className="space-y-1 text-sm text-muted-foreground">
                    <p>Selected on {format(new Date(selection.created_at), "MMMM dd, yyyy")}</p>
                    <p>Candidate status: {candidate.status.replaceAll("_", " ")}</p>
                  </div>

                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Selection Status</CardTitle>
              <CardDescription>The current state of this candidate selection.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-3">
                <SelectionStatusBadge status={selection.status} />
                {isPending && (
                  <span className="text-sm text-muted-foreground">
                    Waiting for Ethiopian agency to respond
                  </span>
                )}
              </div>

              <div className="text-sm text-muted-foreground space-y-1">
                <p>Selected by foreign agency</p>
                <p>on {format(new Date(selection.created_at), "MMMM dd, yyyy")}</p>
              </div>

              {approvalStatus && approvalStatus.approvals.length > 0 && (
                <>
                  <Separator />
                  <div className="space-y-3">
                    <h3 className="text-sm font-semibold">Decision Log</h3>
                    {approvalStatus.approvals.map((approval) => (
                      <div key={`${approval.user_id}-${approval.decided_at}`} className="rounded-lg border p-3">
                        <div className="flex items-center justify-between gap-3">
                          <div>
                            <p className="font-medium">{approval.user_name || approval.role}</p>
                            <p className="text-xs text-muted-foreground">{approval.role.replaceAll("_", " ")}</p>
                          </div>
                          <Badge variant={approval.decision === "approved" ? "default" : "destructive"}>
                            {approval.decision}
                          </Badge>
                        </div>
                        <p className="mt-2 text-xs text-muted-foreground">
                          {format(new Date(approval.decided_at), "MMM dd, yyyy h:mm a")}
                        </p>
                      </div>
                    ))}
                  </div>
                </>
              )}
            </CardContent>
          </Card>

          <Card className="overflow-hidden">
            <CardHeader>
              <CardTitle>Employer Contract Package</CardTitle>
              <CardDescription>
                The foreign employer can optionally upload contract documents here.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <div className="grid gap-4 md:grid-cols-2">
                <SupportingDocumentCard
                  icon={<FileText className="h-4 w-4" />}
                  label="Contract file"
                  description="Offer letter, signed contract, or requested working terms."
                  document={selection.employer_contract}
                  canReplace={!isEthiopianAgent && isPending}
                  onReplace={() => setReplacingDocumentType("contract")}
                  uploading={activeUploadType === "contract"}
                  progress={uploadProgress.contract}
                />
              </div>


              {!isEthiopianAgent && isPending ? (
                <div className="grid gap-4 md:grid-cols-2">
                  {!hasEmployerContract || replacingDocumentType === "contract" ? (
                    <DocumentUpload
                      documentType="contract"
                      title={hasEmployerContract ? "Replace contract" : "Upload contract"}
                      description={
                        activeUploadType === "contract" && uploadProgress.contract > 0
                          ? `Uploading... ${uploadProgress.contract}%`
                          : "Drop a PDF, JPG, or PNG contract file."
                      }
                      accept={{
                        "application/pdf": [".pdf"],
                        "image/jpeg": [".jpg", ".jpeg"],
                        "image/png": [".png"],
                      }}
                      maxSize={10485760}
                      mode="instant"
                      disabled={isUploadingSelectionDocument && activeUploadType !== "contract"}
                      onRemove={() => setReplacingDocumentType((current) => (current === "contract" ? null : current))}
                      onUpload={(file) => handleUploadSelectionDocument("contract", file)}
                    />
                  ) : null}
                </div>
              ) : null}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Recruitment Tracking</CardTitle>
              <CardDescription>
                Once the Ethiopian agency approves, the recruitment tracking steps become visible here.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {showTrackingTimeline ? (
                <>
                  {failedStep ? (
                    <div className="rounded-[1.4rem] border border-rose-300/40 bg-rose-50/80 px-4 py-4 text-sm text-rose-900 dark:border-rose-900/50 dark:bg-rose-950/30 dark:text-rose-100">
                      <div className="flex items-start gap-3">
                        <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
                        <div className="space-y-1">
                          <p className="font-semibold">Issue reported in {failedStep.step_name}</p>
                          <p>
                            {failedStep.notes || "The Ethiopian agency marked this milestone as failed and will post the recovery update here."}
                          </p>
                        </div>
                      </div>
                    </div>
                  ) : null}

                  <div className="space-y-2">
                  <div className="flex items-center justify-between gap-3 text-sm">
                        <span className="text-muted-foreground">Overall candidate status</span>
                        <span className="font-semibold capitalize">
                          {trackingCandidate?.status?.replaceAll("_", " ") ?? "pending"}
                        </span>
                      </div>
                      <div className="flex items-center justify-between gap-3 text-sm">
                        <span className="text-muted-foreground">Progress</span>
                        <span className="font-semibold">{progressPercentage}%</span>
                      </div>
                      <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                        <div
                          className="h-full bg-gradient-to-r from-blue-500 to-green-500 transition-all duration-500"
                          style={{ width: `${progressPercentage}%` }}
                        />
                      </div>
                  </div>

                  <Separator />

                  {!canUpdateProgress && isEthiopianAgent ? (
                    <div className="rounded-[1.4rem] border border-amber-300/40 bg-amber-50/80 px-4 py-3 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-100">
                      You can monitor this process here, but only the Ethiopian agency that created this candidate can update the milestones.
                    </div>
                  ) : null}

                  <div className="flex justify-end">
                    <Button variant="outline" asChild>
                      <Link href={trackingPageHref}>
                        <Eye className="mr-2 h-4 w-4" />
                        Open {candidate.full_name} Tracking
                      </Link>
                    </Button>
                  </div>

                  <StatusTimeline
                    steps={steps}
                    canUpdate={canUpdateProgress}
                    onUpdateStep={handleUpdateStep}
                    isUpdating={isUpdatingStep}
                    onUploadMedicalDocument={canUpdateProgress ? (file) => uploadCandidateDocument({ file, type: "medical" }) : undefined}
                    isUploadingMedicalDocument={isUploadingMedicalDocument}
                    onRemoveMedicalDocument={canUpdateProgress ? handleRemoveMedicalDocument : undefined}
                    isRemovingMedicalDocument={isRemovingMedicalDocument}
                  />
                </>
              ) : (
                <div className="rounded-lg border bg-muted/40 p-4 text-sm text-muted-foreground">
                  {isPending
                    ? `${candidate.full_name}'s tracking has not started yet. Once both agencies approve, the Ethiopian agency can update medical, CoC, LMIS, ticket, and arrival steps here.`
                    : `${candidate.full_name}'s selection is approved, but the tracking steps are still being prepared.`}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {isPending && isEthiopianAgent && !selection.ethiopian_approved && (
                <>
                  <Button
                    className="w-full bg-green-600 hover:bg-green-700"
                    onClick={() => setApproveDialogOpen(true)}
                    disabled={isApproving || isRejecting}
                  >
                    {isApproving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    Approve Selection
                  </Button>
                  <Button
                    variant="outline"
                    className="w-full text-orange-600 border-orange-600 hover:bg-orange-50 dark:hover:bg-orange-950/20"
                    onClick={() => setRejectDialogOpen(true)}
                    disabled={isApproving || isRejecting}
                  >
                    {isRejecting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    Unlock Candidate
                  </Button>
                </>
              )}

              {isPending && isEthiopianAgent && selection.ethiopian_approved && (
                <div className="rounded-lg border bg-muted/40 p-3 text-sm text-muted-foreground">
                  You have approved this selection. The recruitment process can begin.
                </div>
              )}

              {isPending && !isEthiopianAgent && (
                <div className="rounded-lg border bg-muted/40 p-3 text-sm text-muted-foreground">
                  Waiting for the Ethiopian agency to respond to your selection.
                </div>
              )}

              {isPending && userHasApproved && (
                <div className="rounded-lg border bg-muted/40 p-3 text-sm text-muted-foreground">
                  You have already approved this selection. Waiting for the other party.
                </div>
              )}

              {isApproved && (
                <div className="rounded-lg border border-green-200 bg-green-50 p-3 text-sm text-green-800 dark:border-green-900/50 dark:bg-green-950/20 dark:text-green-200">
                  Selection approved. The candidate can continue through the recruitment steps.
                </div>
              )}

              {isRejected && (
                <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800 dark:border-red-900/50 dark:bg-red-950/20 dark:text-red-200">
                  This selection has been rejected and will not move forward.
                </div>
              )}

              {isExpired && (
                <div className="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/20 dark:text-amber-200">
                  This selection expired before both approvals were completed.
                </div>
              )}

              <Separator />

              <Button variant="outline" className="w-full" asChild>
                <Link href={`/candidates/${selection.candidate_id}`}>
                  <Eye className="mr-2 h-4 w-4" />
                  Open Candidate
                </Link>
              </Button>

              <Button
                variant="outline"
                className="w-full"
                onClick={() => void handleOpenCandidateChat()}
                disabled={isOpeningChat}
              >
                {isOpeningChat ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <MessageSquare className="mr-2 h-4 w-4" />}
                Open chat about this candidate
              </Button>

              {showTrackingTimeline ? (
                <Button variant="outline" className="w-full" asChild>
                  <Link href={trackingPageHref}>
                    <Sparkles className="mr-2 h-4 w-4" />
                    Open {candidate.full_name} Tracking
                  </Link>
                </Button>
              ) : (isApproved && steps.length === 0) ? (
                <Button variant="outline" className="w-full" asChild>
                  <Link href={trackingPageHref}>
                    <Sparkles className="mr-2 h-4 w-4" />
                    Open {candidate.full_name} Tracking
                  </Link>
                </Button>
              ) : null}
            </CardContent>
          </Card>

          {failedStep ? (
            <Card>
              <CardHeader>
                <CardTitle>Current Issue</CardTitle>
                <CardDescription>The shared recruitment process is waiting on this issue to be resolved.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                <div className="rounded-2xl border border-rose-300/50 bg-rose-50/80 p-4 text-rose-900 dark:border-rose-900/40 dark:bg-rose-950/25 dark:text-rose-100">
                  <p className="font-semibold">{failedStep.step_name}</p>
                  <p className="mt-2">{failedStep.notes || "The Ethiopian agency has not added a written reason yet."}</p>
                </div>
              </CardContent>
            </Card>
          ) : null}
        </div>
      </div>

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
    </div>
  )
}

function SelectionStatusBadge({ status }: { status: SelectionStatus }) {
  switch (status) {
    case SelectionStatus.PENDING:
      return <Badge className="bg-yellow-500 hover:bg-yellow-600 text-white">Pending</Badge>
    case SelectionStatus.APPROVED:
      return <Badge className="bg-green-500 hover:bg-green-600 text-white">Approved</Badge>
    case SelectionStatus.REJECTED:
      return <Badge className="bg-red-500 hover:bg-red-600 text-white">Rejected</Badge>
    case SelectionStatus.EXPIRED:
      return <Badge className="bg-slate-500 hover:bg-slate-600 text-white">Expired</Badge>
    case SelectionStatus.RELEASED:
      return <Badge className="bg-orange-500 hover:bg-orange-600 text-white">Released</Badge>
    default:
      return <Badge variant="secondary">{status}</Badge>
  }
}

function SupportingDocumentCard({
  icon,
  label,
  description,
  document,
  canReplace = false,
  onReplace,
  uploading = false,
  progress = 0,
}: {
  icon: React.ReactNode
  label: string
  description: string
  document?: { file_url: string; file_name: string; uploaded_at?: string }
  canReplace?: boolean
  onReplace?: () => void
  uploading?: boolean
  progress?: number
}) {
  if (!document) {
    return (
      <div className="rounded-[1.4rem] border border-dashed border-border/70 bg-muted/20 p-4">
        <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
          {icon}
          {label}
        </div>
        <p className="mt-2 text-sm text-muted-foreground">{description}</p>
        <p className="mt-4 inline-flex items-center gap-2 rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-800 dark:bg-amber-950/40 dark:text-amber-200">
          <Clock className="h-3.5 w-3.5" />
          Not uploaded yet
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-[1.4rem] border border-emerald-200/70 bg-emerald-50/70 p-4 dark:border-emerald-900/40 dark:bg-emerald-950/20">
      <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
        {icon}
        {label}
      </div>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
        <div className="mt-4 space-y-3 rounded-2xl border border-emerald-200/60 bg-background/90 p-3 dark:border-emerald-900/30">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="truncate font-medium text-foreground">{document.file_name}</p>
            <p className="text-xs text-muted-foreground">
              {document.uploaded_at ? `Uploaded ${format(new Date(document.uploaded_at), "MMM dd, yyyy h:mm a")}` : "Uploaded"}
            </p>
          </div>
          <Badge className="bg-emerald-600 text-white hover:bg-emerald-600">
            <BadgeCheck className="mr-1 h-3.5 w-3.5" />
            Ready
          </Badge>
        </div>
        {uploading ? (
          <p className="text-xs font-medium text-primary">Uploading replacement... {progress}%</p>
        ) : null}
        <Button variant="outline" size="sm" asChild>
          <a href={document.file_url} target="_blank" rel="noreferrer">
            <Eye className="mr-2 h-4 w-4" />
            View File
          </a>
        </Button>
        {canReplace && onReplace ? (
          <Button variant="ghost" size="sm" onClick={onReplace}>
            <FileText className="mr-2 h-4 w-4" />
            Replace File
          </Button>
        ) : null}
      </div>
    </div>
  )
}
