"use client"

import * as React from "react"
import { CheckCircle2, Circle, Loader2, Lock, PlayCircle } from "lucide-react"
import { format, isValid } from "date-fns"

import { StatusStep } from "@/types"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

interface StatusTimelineProps {
  steps: StatusStep[]
  canUpdate?: boolean
  onUpdateStep?: (stepName: string, status: string, notes?: string) => void
  isUpdating?: boolean
}

type StepTone = {
  badgeClassName: string
  iconClassName: string
  glowClassName: string
  cardClassName: string
  label: string
}

const STEP_TONES: Record<StatusStep["step_status"], StepTone> = {
  pending: {
    badgeClassName: "border-slate-300 bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200",
    iconClassName: "border-slate-300 bg-white text-slate-500 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-300",
    glowClassName: "bg-slate-300/40 dark:bg-slate-700/30",
    cardClassName: "border-border/70 bg-card/95",
    label: "Pending",
  },
  in_progress: {
    badgeClassName: "border-sky-200 bg-sky-100 text-sky-800 dark:border-sky-900/70 dark:bg-sky-950/50 dark:text-sky-200",
    iconClassName: "border-sky-300 bg-sky-50 text-sky-700 shadow-[0_0_0_6px_rgba(14,165,233,0.12)] dark:border-sky-700 dark:bg-sky-950/60 dark:text-sky-200",
    glowClassName: "bg-sky-400/30 dark:bg-sky-500/20",
    cardClassName: "border-sky-200/70 bg-sky-50/65 dark:border-sky-900/40 dark:bg-sky-950/25",
    label: "In Progress",
  },
  completed: {
    badgeClassName: "border-emerald-200 bg-emerald-100 text-emerald-800 dark:border-emerald-900/60 dark:bg-emerald-950/40 dark:text-emerald-200",
    iconClassName: "border-emerald-300 bg-emerald-50 text-emerald-700 shadow-[0_0_0_6px_rgba(16,185,129,0.12)] dark:border-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-200",
    glowClassName: "bg-emerald-400/30 dark:bg-emerald-500/20",
    cardClassName: "border-emerald-200/70 bg-emerald-50/65 dark:border-emerald-900/40 dark:bg-emerald-950/25",
    label: "Completed",
  },
}

export function StatusTimeline({ steps, canUpdate, onUpdateStep, isUpdating }: StatusTimelineProps) {
  const [editingStep, setEditingStep] = React.useState<string | null>(null)
  const [notes, setNotes] = React.useState("")

  const handleUpdateStatus = (stepName: string, newStatus: string) => {
    if (!onUpdateStep) {
      return
    }

    onUpdateStep(stepName, newStatus, notes || undefined)
    setEditingStep(null)
    setNotes("")
  }

  return (
    <div className="space-y-4">
      {steps.map((step, index) => {
        const tone = STEP_TONES[step.step_status]
        const isEditing = editingStep === step.step_name
        const previousStepsIncomplete = steps.slice(0, index).some((previousStep) => previousStep.step_status !== "completed")
        const canStartStep = !previousStepsIncomplete && step.step_status === "pending"

        return (
          <div
            key={step.id}
            className={cn(
              "group relative overflow-hidden rounded-[1.75rem] border p-5 shadow-soft transition-all duration-300 hover:-translate-y-1 hover:shadow-glow",
              tone.cardClassName
            )}
          >
            {index < steps.length - 1 ? (
              <div className="pointer-events-none absolute left-10 top-[5.25rem] h-[calc(100%-2.5rem)] w-px bg-gradient-to-b from-primary/40 via-border to-transparent" />
            ) : null}

            <div className="flex gap-4">
              <div className="relative z-10 pt-1">
                <span className={cn("absolute inset-0 rounded-full blur-xl", tone.glowClassName)} />
                <div
                  className={cn(
                    "relative flex h-12 w-12 items-center justify-center rounded-2xl border bg-background/90 transition-transform duration-300 group-hover:scale-105",
                    tone.iconClassName
                  )}
                >
                  {renderStatusIcon(step.step_status)}
                </div>
              </div>

              <div className="min-w-0 flex-1 space-y-4">
                <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="text-xs font-semibold uppercase tracking-[0.24em] text-muted-foreground">
                        Step {index + 1}
                      </p>
                      <Badge variant="outline" className={cn("rounded-full px-3 py-1 font-semibold", tone.badgeClassName)}>
                        {tone.label}
                      </Badge>
                    </div>
                    <div>
                      <h4 className="text-lg font-semibold text-foreground">{step.step_name}</h4>
                      <p className="mt-1 text-sm text-muted-foreground">
                        {step.step_status === "completed"
                          ? "This milestone has been completed and recorded in the shared process."
                          : step.step_status === "in_progress"
                            ? "This milestone is currently active and waiting for the next update."
                            : "This milestone has not started yet."}
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center gap-2">
                    {canUpdate && step.step_status === "pending" ? (
                      <Button
                        size="sm"
                        variant={canStartStep ? "default" : "outline"}
                        disabled={isUpdating || !canStartStep}
                        onClick={() => setEditingStep(step.step_name)}
                      >
                        {canStartStep ? (
                          <>
                            <PlayCircle className="h-3.5 w-3.5" />
                            Start Step
                          </>
                        ) : (
                          <>
                            <Lock className="h-3.5 w-3.5" />
                            Finish Previous Step First
                          </>
                        )}
                      </Button>
                    ) : null}

                    {canUpdate && step.step_status === "in_progress" ? (
                      <Button
                        size="sm"
                        onClick={() => handleUpdateStatus(step.step_name, "completed")}
                        disabled={isUpdating}
                        className="bg-emerald-600 text-white hover:bg-emerald-700"
                      >
                        {isUpdating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <CheckCircle2 className="h-3.5 w-3.5" />}
                        Complete Step
                      </Button>
                    ) : null}
                  </div>
                </div>

                <div className="grid gap-3 text-sm text-muted-foreground md:grid-cols-2 xl:grid-cols-3">
                  <InfoPill
                    label="Updated"
                    value={formatStepDate(step.updated_at, "Awaiting first update")}
                  />
                  <InfoPill
                    label="Updated by"
                    value={step.updated_by?.name || "System"}
                  />
                  <InfoPill
                    label="Completed"
                    value={step.completed_at ? formatStepDate(step.completed_at, "Not yet") : "Not yet"}
                  />
                </div>

                {step.notes ? (
                  <div className="rounded-2xl border border-border/70 bg-background/80 px-4 py-3 text-sm text-muted-foreground">
                    <span className="mb-1 block text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">
                      Notes
                    </span>
                    {step.notes}
                  </div>
                ) : null}

                {canUpdate && isEditing ? (
                  <div className="rounded-[1.4rem] border border-primary/20 bg-background/85 p-4 shadow-soft">
                    <div className="mb-3 flex items-center gap-2">
                      <Badge className="rounded-full bg-primary/12 px-3 py-1 text-primary hover:bg-primary/12">
                        Update {step.step_name}
                      </Badge>
                    </div>
                    <Textarea
                      placeholder="Add an update note so both agencies understand what changed..."
                      value={notes}
                      onChange={(event) => setNotes(event.target.value)}
                      className="min-h-[96px] resize-none"
                    />
                    <div className="mt-3 flex flex-wrap gap-2">
                      <Button
                        size="sm"
                        onClick={() => handleUpdateStatus(step.step_name, "in_progress")}
                        disabled={isUpdating}
                      >
                        {isUpdating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <PlayCircle className="h-3.5 w-3.5" />}
                        Mark In Progress
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          setEditingStep(null)
                          setNotes("")
                        }}
                        disabled={isUpdating}
                      >
                        Cancel
                      </Button>
                    </div>
                  </div>
                ) : null}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}

function renderStatusIcon(status: StatusStep["step_status"]) {
  switch (status) {
    case "completed":
      return <CheckCircle2 className="h-5 w-5" />
    case "in_progress":
      return <Loader2 className="h-5 w-5 animate-spin" />
    default:
      return <Circle className="h-5 w-5" />
  }
}

function InfoPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-border/60 bg-background/75 px-4 py-3">
      <span className="block text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">
        {label}
      </span>
      <p className="mt-1 font-medium text-foreground">{value}</p>
    </div>
  )
}

function formatStepDate(value: string | undefined, fallback: string) {
  if (!value) {
    return fallback
  }

  const parsed = new Date(value)
  if (!isValid(parsed)) {
    return fallback
  }

  return format(parsed, "MMM dd, yyyy 'at' h:mm a")
}
