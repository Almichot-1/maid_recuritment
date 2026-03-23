"use client"

import * as React from "react"
import Link from "next/link"
import { useParams } from "next/navigation"
import {
  AlertTriangle,
  ArrowLeft,
  ChevronRight,
  Clock3,
  Gauge,
  Home,
  Loader2,
  ShieldCheck,
  Sparkles,
  XCircle,
} from "lucide-react"

import { StatusTimeline } from "@/components/candidates/status-timeline"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidate } from "@/hooks/use-candidates"
import { useCandidateProgress, useUpdateStatusStep } from "@/hooks/use-status-steps"
import { CandidateStatus } from "@/types"

export default function CandidateTrackingPage() {
  const params = useParams()
  const candidateId = String(params.id || "")
  const { user, isEthiopianAgent } = useCurrentUser()
  const { data: candidate, isLoading: isCandidateLoading, error } = useCandidate(candidateId)
  const { data: progressData, isLoading: isProgressLoading } = useCandidateProgress(candidateId)
  const { mutate: updateStep, isPending: isUpdatingStep } = useUpdateStatusStep(candidateId)
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
          <h1 className="text-2xl font-bold">Process tracking is unavailable</h1>
          <p className="max-w-md text-muted-foreground">
            The candidate could not be loaded, or you no longer have access to this record.
          </p>
        </div>
        <Button asChild>
          <Link href="/candidates">Back to candidates</Link>
        </Button>
      </div>
    )
  }

  const hasTracking = !!progressData && progressData.steps.length > 0
  const canTrack = hasTracking || candidate.status === CandidateStatus.IN_PROGRESS || candidate.status === CandidateStatus.COMPLETED
  const completedSteps = progressData?.steps.filter((step) => step.step_status === "completed").length ?? 0
  const activeStep = progressData?.steps.find((step) => step.step_status === "in_progress")
  const failedStep = progressData?.steps.find((step) => step.step_status === "failed")

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      <nav className="flex items-center text-sm font-medium text-muted-foreground">
        <Link href="/dashboard" className="flex items-center transition-colors hover:text-primary">
          <Home className="mr-1.5 h-4 w-4" />
          Dashboard
        </Link>
        <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
        <Link href="/candidates" className="transition-colors hover:text-primary">
          Candidates
        </Link>
        <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
        <Link href={`/candidates/${candidate.id}`} className="transition-colors hover:text-primary">
          {candidate.full_name}
        </Link>
        <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
        <span className="font-semibold text-foreground">Process Tracking</span>
      </nav>

      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(59,130,246,0.16),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.18),_transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.96))] text-white shadow-xl">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_300px]">
          <div className="space-y-4">
            <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-sky-200 hover:bg-white/15">
              Process tracking
            </Badge>
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold tracking-tight">
                Shared recruitment tracking for {candidate.full_name}
              </h1>
              <p className="max-w-2xl text-sm text-slate-200/90">
                Once both agencies approve a selection, the recruitment process lives here so both sides can follow medical, CoC, LMIS, ticket, and arrival progress clearly.
              </p>
            </div>
            <div className="flex flex-wrap gap-3">
              <Button className="bg-white text-slate-950 hover:bg-slate-100" asChild>
                <Link href={`/candidates/${candidate.id}`}>
                  <ArrowLeft className="mr-2 h-4 w-4" />
                  Back to candidate
                </Link>
              </Button>
              <Button variant="outline" className="border-white/20 bg-white/10 text-white hover:bg-white/15 hover:text-white" asChild>
                <Link href="/selections">
                  <Sparkles className="mr-2 h-4 w-4" />
                  Open selections
                </Link>
              </Button>
            </div>
          </div>

          <div className="rounded-3xl border border-white/10 bg-white/10 p-5 backdrop-blur">
            <p className="text-xs uppercase tracking-[0.24em] text-sky-200">Tracking summary</p>
            <div className="mt-4 space-y-4">
              <MetricCard
                icon={<ShieldCheck className="h-5 w-5 text-emerald-300" />}
                label="Candidate status"
                value={candidate.status.replaceAll("_", " ")}
              />
              <MetricCard
                icon={<Gauge className="h-5 w-5 text-sky-300" />}
                label={failedStep ? "Blocked at" : "Progress"}
                value={failedStep ? failedStep.step_name : hasTracking ? `${completedSteps}/${progressData.steps.length} steps` : "Not started"}
              />
              <MetricCard
                icon={<Clock3 className="h-5 w-5 text-amber-300" />}
                label="Editable by"
                value={
                  canUpdateProgress
                    ? "You can update steps"
                    : isEthiopianAgent
                      ? "Owner agency only"
                      : "View only"
                }
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {!canTrack ? (
        <Card className="shadow-sm">
          <CardContent className="py-16">
            <div className="mx-auto max-w-lg space-y-3 text-center">
              <h2 className="text-2xl font-semibold tracking-tight">Tracking starts after both approvals</h2>
              <p className="text-muted-foreground">
                This candidate has not entered the shared process-tracking phase yet. Once both the employer and the agency approve the selection, the recruitment steps will appear here.
              </p>
            </div>
          </CardContent>
        </Card>
      ) : hasTracking ? (
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_320px]">
            <Card className="animated-border shadow-sm">
            <CardHeader>
              <div className="flex items-center justify-between gap-4">
                <CardTitle>Step-by-step progress</CardTitle>
                <span className="text-sm font-semibold text-muted-foreground">
                  {Math.round(progressData.progress_percentage)}% complete
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full bg-gradient-to-r from-blue-500 to-green-500 transition-all duration-500"
                  style={{ width: `${progressData.progress_percentage}%` }}
                />
              </div>
            </CardHeader>
            <CardContent>
              {failedStep ? (
                <div className="mb-5 rounded-[1.4rem] border border-rose-300/40 bg-rose-50/85 px-4 py-4 text-sm text-rose-900 dark:border-rose-900/50 dark:bg-rose-950/30 dark:text-rose-100">
                  <div className="flex items-start gap-3">
                    <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
                    <div className="space-y-1">
                      <p className="font-semibold">Issue reported in {failedStep.step_name}</p>
                      <p>{failedStep.notes || "The Ethiopian agency marked this milestone as failed and will update it here once it is resolved."}</p>
                    </div>
                  </div>
                </div>
              ) : null}

              {!canUpdateProgress && isEthiopianAgent ? (
                <div className="mb-5 rounded-[1.4rem] border border-amber-300/40 bg-amber-50/80 px-4 py-3 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-100">
                  This candidate belongs to another Ethiopian agency account, so you can view the process but you cannot update it.
                </div>
              ) : null}
              <StatusTimeline
                steps={progressData.steps}
                canUpdate={canUpdateProgress}
                onUpdateStep={handleUpdateStep}
                isUpdating={isUpdatingStep}
              />
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle>Live Process Snapshot</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-muted-foreground">
              <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Current focus</p>
                <p className="mt-2 text-base font-semibold text-foreground">
                  {failedStep
                    ? `${failedStep.step_name} needs attention`
                    : activeStep
                      ? activeStep.step_name
                      : candidate.status === CandidateStatus.COMPLETED
                        ? "Recruitment completed"
                        : "Waiting for the next milestone"}
                </p>
                <p className="mt-2">
                  {failedStep
                    ? failedStep.notes || "The current milestone has been marked as failed. The written reason will stay visible here for both agencies until the Ethiopian agency resumes it."
                    : activeStep
                      ? "This is the milestone that is currently active in the shared process."
                      : "As soon as a step starts, it will show here for both agencies."}
                </p>
              </div>
              <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                The Ethiopian agency owns the updates for each recruitment step, while the employer sees the same timeline in real time.
              </div>
              <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                Every completed step stays visible so both sides keep one shared source of truth from medical through arrival.
              </div>
            </CardContent>
          </Card>
        </div>
      ) : (
        <Card className="shadow-sm">
          <CardContent className="py-16">
            <div className="mx-auto max-w-lg space-y-3 text-center">
              <Loader2 className="mx-auto h-10 w-10 animate-spin text-primary" />
              <h2 className="text-2xl font-semibold tracking-tight">Preparing the tracking workspace</h2>
              <p className="text-muted-foreground">
                The selection is approved, and the shared tracking steps are still being initialized.
              </p>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

function MetricCard({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode
  label: string
  value: string
}) {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
      <div className="flex items-center gap-2 text-xs uppercase tracking-[0.2em] text-slate-300">
        {icon}
        <span>{label}</span>
      </div>
      <p className="mt-2 text-lg font-semibold text-white capitalize">{value}</p>
    </div>
  )
}
