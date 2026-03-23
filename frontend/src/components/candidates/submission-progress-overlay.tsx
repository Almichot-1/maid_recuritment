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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4 backdrop-blur-sm">
      <Card className="w-full max-w-xl overflow-hidden border-0 shadow-2xl">
        <CardContent className="space-y-6 p-0">
          <div className="bg-[radial-gradient(circle_at_top_left,_rgba(59,130,246,0.16),_transparent_26%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.18),_transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.96))] px-6 py-5 text-white">
            <div className="flex items-start gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-white/10">
                <Loader2 className="h-6 w-6 animate-spin text-sky-200" />
              </div>
              <div className="space-y-1">
                <p className="text-xs uppercase tracking-[0.24em] text-sky-200">Processing</p>
                <h2 className="text-2xl font-semibold tracking-tight">{title}</h2>
                <p className="text-sm text-slate-200/90">{description}</p>
              </div>
            </div>
          </div>

          <div className="space-y-3 px-6 pb-6">
            {steps.map((step) => (
              <div
                key={step.label}
                className={cn(
                  "flex items-start gap-3 rounded-2xl border px-4 py-3 transition-all duration-300",
                  step.status === "complete" && "border-emerald-200 bg-emerald-50",
                  step.status === "active" && "border-sky-200 bg-sky-50 shadow-sm",
                  step.status === "pending" && "border-border/70 bg-card"
                )}
              >
                <div className="mt-0.5">
                  {step.status === "complete" ? (
                    <CheckCircle2 className="h-5 w-5 text-emerald-600" />
                  ) : step.status === "active" ? (
                    <Loader2 className="h-5 w-5 animate-spin text-sky-600" />
                  ) : (
                    <CircleDashed className="h-5 w-5 text-muted-foreground" />
                  )}
                </div>
                <div className="space-y-1">
                  <p
                    className={cn(
                      "text-sm font-semibold",
                      step.status === "pending" ? "text-foreground" : ""
                    )}
                  >
                    {step.label}
                  </p>
                  <p className="text-sm text-muted-foreground">{step.description}</p>
                </div>
              </div>
            ))}

            {footer ? (
              <div className="rounded-2xl border border-dashed border-border/70 bg-muted/20 px-4 py-3 text-sm text-muted-foreground">
                {footer}
              </div>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
