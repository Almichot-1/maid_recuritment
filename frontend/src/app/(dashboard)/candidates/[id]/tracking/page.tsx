"use client"

import * as React from "react"
import Link from "next/link"
import { useParams } from "next/navigation"
import { AlertTriangle, ArrowLeft, Loader2, XCircle } from "lucide-react"

import { StatusTimeline } from "@/components/candidates/status-timeline"
import { PageHeader } from "@/components/layout/page-header"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidate, useUploadDocument } from "@/hooks/use-candidates"
import { useCandidateProgress, useUpdateStatusStep } from "@/hooks/use-status-steps"
import { CandidateStatus } from "@/types"

export default function CandidateTrackingPage() {
  const params = useParams()
  const candidateId = String(params.id || "")
  const { user, isEthiopianAgent } = useCurrentUser()
  const { data: candidate, isLoading: isCandidateLoading, error } = useCandidate(candidateId)
  const { data: progressData, isLoading: isProgressLoading } = useCandidateProgress(candidateId)
  const { mutate: updateStep, isPending: isUpdatingStep } = useUpdateStatusStep(candidateId)
  const { mutateAsync: uploadDocument, isPending: isUploadingDocument } = useUploadDocument(candidateId)
  const canUpdateProgress = isEthiopianAgent && candidate?.created_by === user?.id

  const handleUpdateStep = (stepName: string, status: string, notes?: string) => {
    if (!canUpdateProgress) {
      return
    }
    updateStep({ step_name: stepName, status, notes })
  }

  if (isCandidateLoading || isProgressLoading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
      </div>
    )
  }

  if (error || !candidate) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <div className="flex h-20 w-20 items-center justify-center rounded-full bg-destructive/10">
          <XCircle className="h-10 w-10 text-destructive" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-bold">Tracking unavailable</h1>
          <p className="max-w-md text-muted-foreground">
            This candidate could not be loaded or you no longer have access.
          </p>
        </div>
        <Button asChild>
          <Link href="/candidates">Back to candidates</Link>
        </Button>
      </div>
    )
  }

  const hasTracking = !!progressData && progressData.steps.length > 0
  const canTrack =
    hasTracking ||
    candidate.status === CandidateStatus.IN_PROGRESS ||
    candidate.status === CandidateStatus.COMPLETED
  const failedStep = progressData?.steps.find((step) => step.step_status === "failed")
  const activeStep = progressData?.steps.find((step) => step.step_status === "in_progress")
  const nextPending = progressData?.steps.find((step) => step.step_status === "pending")

  const currentFocus = failedStep || activeStep || nextPending

  return (
    <div className="space-y-6 pb-10">
      <PageHeader
        heading={`Tracking — ${candidate.full_name}`}
        text={
          canUpdateProgress
            ? "Update each milestone as the case moves forward."
            : "View-only timeline shared with your partner agency."
        }
        action={
          <Button variant="outline" asChild>
            <Link href={`/candidates/${candidate.id}`}>
              <ArrowLeft className="mr-2 h-4 w-4" />
              Candidate
            </Link>
          </Button>
        }
      />

      {!canTrack ? (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-sm text-muted-foreground">
              Tracking opens after both agencies approve the selection.
            </p>
          </CardContent>
        </Card>
      ) : hasTracking ? (
        <div className="space-y-4">
          <Card>
            <CardHeader className="space-y-3">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <CardTitle className="text-lg">Progress</CardTitle>
                <Badge variant="outline">{Math.round(progressData.progress_percentage)}%</Badge>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full bg-primary transition-all duration-500"
                  style={{ width: `${progressData.progress_percentage}%` }}
                />
              </div>
              {currentFocus ? (
                <p className="text-sm text-muted-foreground">
                  <span className="font-medium text-foreground">Now: </span>
                  {currentFocus.step_name}
                  {failedStep ? " — needs attention" : ""}
                </p>
              ) : null}
            </CardHeader>
            <CardContent className="space-y-4">
              {failedStep ? (
                <div className="flex items-start gap-3 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm">
                  <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
                  <div>
                    <p className="font-medium text-foreground">{failedStep.step_name}</p>
                    <p className="text-muted-foreground">
                      {failedStep.notes || "Marked as blocked. Add a note when you retry."}
                    </p>
                  </div>
                </div>
              ) : null}

              {!canUpdateProgress && isEthiopianAgent ? (
                <p className="text-sm text-muted-foreground">
                  Only the owning agency can update steps on this candidate.
                </p>
              ) : null}

              <StatusTimeline
                steps={progressData.steps}
                canUpdate={canUpdateProgress}
                onUpdateStep={handleUpdateStep}
                isUpdating={isUpdatingStep}
                onUploadMedicalDocument={
                  canUpdateProgress ? (file) => uploadDocument({ file, type: "medical" }) : undefined
                }
                isUploadingMedicalDocument={isUploadingDocument}
              />
            </CardContent>
          </Card>
        </div>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">Setting up tracking steps…</p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
