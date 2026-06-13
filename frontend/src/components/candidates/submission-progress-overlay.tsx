"use client"

import { CheckCircle2, CircleDashed, Loader2 } from "lucide-react"

import { Card, CardContent } from "@/components/ui/card"
import { cn } from "@/lib/utils"

type StepStatus = "pending" | "active" | "complete"

export interface SubmissionProgressStep {
  label: string
  description: string
  status: StepStatus
}

interface SubmissionProgressOverlayProps {
  open: boolean
  title: string
  description: string
  steps: SubmissionProgressStep[]
  footer?: string
}

export function SubmissionProgressOverlay({
  open,
  title,
  description,
  steps,
  footer,
}: SubmissionProgressOverlayProps) {
  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-md px-4 animate-in fade-in duration-200">
      <Card className="w-full max-w-2xl overflow-hidden border shadow-2xl animate-in zoom-in-95 duration-300">
        <CardContent className="space-y-6 p-0">
          <div className="relative overflow-hidden bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 px-8 py-8">
            <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_rgba(59,130,246,0.15),transparent_50%),radial-gradient(ellipse_at_bottom_left,_rgba(16,185,129,0.15),transparent_50%)]" />
            <div className="relative flex items-start gap-5">
              <div className="flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500/20 to-emerald-500/20 ring-1 ring-white/10">
                <Loader2 className="h-8 w-8 animate-spin text-blue-400" />
              </div>
              <div className="space-y-2 pt-1">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold uppercase tracking-wider text-blue-300">Processing</span>
                </div>
                <h2 className="text-3xl font-bold tracking-tight text-white">{title}</h2>
                <p className="text-base leading-relaxed text-slate-300">{description}</p>
              </div>
            </div>
          </div>

          <div className="space-y-4 px-8 pb-8">
            {steps.map((step) => (
              <div
                key={step.label}
                className={cn(
                  "group flex items-start gap-4 rounded-xl border-2 px-5 py-4 transition-all duration-300",
                  step.status === "complete" && "border-emerald-500/50 bg-emerald-50 dark:border-emerald-500/30 dark:bg-emerald-950/20",
                  step.status === "active" && "border-blue-500/50 bg-blue-50 shadow-lg shadow-blue-500/10 dark:border-blue-500/30 dark:bg-blue-950/20",
                  step.status === "pending" && "border-slate-200 bg-slate-50 dark:border-slate-800 dark:bg-slate-900/50"
                )}
              >
                <div className="mt-0.5 shrink-0">
                  {step.status === "complete" ? (
                    <CheckCircle2 className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
                  ) : step.status === "active" ? (
                    <Loader2 className="h-6 w-6 animate-spin text-blue-600 dark:text-blue-400" />
                  ) : (
                    <CircleDashed className="h-6 w-6 text-slate-400 dark:text-slate-600" />
                  )}
                </div>
                <div className="min-w-0 flex-1 space-y-1.5">
                  <p
                    className={cn(
                      "text-base font-semibold leading-snug",
                      step.status === "complete" && "text-emerald-900 dark:text-emerald-100",
                      step.status === "active" && "text-blue-900 dark:text-blue-100",
                      step.status === "pending" && "text-slate-700 dark:text-slate-300"
                    )}
                  >
                    {step.label}
                  </p>
                  <p 
                    className={cn(
                      "text-sm leading-relaxed",
                      step.status === "complete" && "text-emerald-700 dark:text-emerald-300",
                      step.status === "active" && "text-blue-700 dark:text-blue-300",
                      step.status === "pending" && "text-slate-600 dark:text-slate-400"
                    )}
                  >
                    {step.description}
                  </p>
                </div>
              </div>
            ))}

            {footer ? (
              <div className="mt-6 rounded-xl border-2 border-dashed border-slate-300 bg-slate-100 px-5 py-4 dark:border-slate-700 dark:bg-slate-900/50">
                <p className="text-sm leading-relaxed text-slate-700 dark:text-slate-300">{footer}</p>
              </div>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
