"use client"

import * as React from "react"
import Link from "next/link"
import { ArrowRight, CheckCircle2, Home, Loader2, Route, Sparkles, TimerReset, CheckSquare, X } from "lucide-react"
import { toast } from "sonner"

import { PageHeader } from "@/components/layout/page-header"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Input } from "@/components/ui/input"
import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { useMySelections } from "@/hooks/use-selections"
import { useBatchUpdateStatusSteps } from "@/hooks/use-status-steps"
import { CandidateStatus, SelectionStatus } from "@/types"
import { cn } from "@/lib/utils"

const STEPS = ["Medical", "CoC", "Visa", "Ticket", "Arrival City"]
const STATUSES = ["pending", "in_progress", "completed", "failed"]

export default function TrackingHubPage() {
  const { isEthiopianAgent } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { data: selectionsData, isLoading, refetch } = useMySelections()
  const selections = React.useMemo(() => selectionsData?.selections || [], [selectionsData?.selections])
  const { mutateAsync: batchUpdate, isPending: isBatchUpdating } = useBatchUpdateStatusSteps()

  const trackingSelections = React.useMemo(
    () => (selections || []).filter((selection) => selection.status === SelectionStatus.APPROVED && selection.candidate),
    [selections]
  )

  const activeRecruitments = trackingSelections.filter((selection) => selection.candidate?.status === CandidateStatus.IN_PROGRESS)
  const completedRecruitments = trackingSelections.filter((selection) => selection.candidate?.status === CandidateStatus.COMPLETED)

  // Batch mode state
  const [batchMode, setBatchMode] = React.useState(false)
  const [selectedIds, setSelectedIds] = React.useState<Set<string>>(new Set())
  const [batchStep, setBatchStep] = React.useState("")
  const [batchStatus, setBatchStatus] = React.useState("")
  const [batchCoC, setBatchCoC] = React.useState("")
  const [batchCity, setBatchCity] = React.useState("")

  const toggleSelection = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const clearBatch = () => {
    setSelectedIds(new Set())
    setBatchMode(false)
    setBatchStep("")
    setBatchStatus("")
    setBatchCoC("")
    setBatchCity("")
  }

  const handleBatchApply = async () => {
    if (selectedIds.size === 0) {
      toast.error("Select at least one candidate")
      return
    }
    if (!batchStep) {
      toast.error("Select a step to update")
      return
    }
    if (!batchStatus) {
      toast.error("Select a status")
      return
    }

    const payload: {
      candidate_ids: string[]
      step_name: string
      status: string
      coc_status?: string
      arrival_city?: string
    } = {
      candidate_ids: Array.from(selectedIds),
      step_name: batchStep,
      status: batchStatus,
    }
    if (batchStep === "CoC" && batchCoC) payload.coc_status = batchCoC
    if (batchStep === "Arrival City" && batchCity) payload.arrival_city = batchCity

    try {
      await batchUpdate(payload)
      refetch()
      clearBatch()
    } catch {
      // handled by hook
    }
  }

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      <nav className="flex items-center text-sm font-medium text-muted-foreground">
        <Link href="/dashboard" className="flex items-center transition-colors hover:text-primary">
          <Home className="mr-1.5 h-4 w-4" />
          Dashboard
        </Link>
        <span className="mx-2 text-muted-foreground/50">/</span>
        <span className="font-semibold text-foreground">Process Tracking</span>
      </nav>

      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.18),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.18),_transparent_22%),linear-gradient(135deg,rgba(255,255,255,0.88),rgba(244,250,249,0.96))] shadow-glow dark:bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.24),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.22),_transparent_22%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,118,110,0.24))]">
        <CardContent className="flex flex-col gap-5 p-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="space-y-3">
              <PageHeader
                className="pb-0"
                heading="Process Tracking"
                text={
                  batchMode
                    ? "Select multiple candidates, choose a step and status, then update them all at once."
                    : isEthiopianAgent
                      ? "Open any approved recruitment and update the shared progress timeline."
                      : "Monitor every approved recruitment from one place."
                }
              />
              {activeWorkspace ? (
                <Badge variant="outline" className="w-fit rounded-full px-3 py-1 text-[11px] uppercase tracking-[0.2em]">
                  {activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}
                </Badge>
              ) : null}
            </div>
            <div className="flex flex-wrap gap-3 shrink-0">
              {isEthiopianAgent && !batchMode ? (
                <Button variant="outline" onClick={() => setBatchMode(true)}>
                  <CheckSquare className="mr-2 h-4 w-4" />
                  Batch Update
                </Button>
              ) : null}
              {isEthiopianAgent && batchMode ? (
                <Button variant="outline" onClick={clearBatch}>
                  <X className="mr-2 h-4 w-4" />
                  Cancel
                </Button>
              ) : null}
              <Button asChild>
                <Link href="/selections">
                  <Sparkles className="mr-2 h-4 w-4" />
                  My Selections
                </Link>
              </Button>
              <Button variant="outline" asChild>
                <Link href="/candidates">
                  <Route className="mr-2 h-4 w-4" />
                  Browse Candidates
                </Link>
              </Button>
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <SummaryCard label="Approved" value={trackingSelections.length} tone="from-sky-500/20 to-sky-400/5 text-sky-950 dark:text-sky-100" icon={<Sparkles className="h-4 w-4" />} />
            <SummaryCard label="In Progress" value={activeRecruitments.length} tone="from-amber-500/20 to-amber-400/5 text-amber-950 dark:text-amber-100" icon={<TimerReset className="h-4 w-4" />} />
            <SummaryCard label="Completed" value={completedRecruitments.length} tone="from-emerald-500/20 to-emerald-400/5 text-emerald-950 dark:text-emerald-100" icon={<CheckCircle2 className="h-4 w-4" />} />
          </div>
        </CardContent>
      </Card>

      <Card className="overflow-hidden shadow-sm">
        <CardContent className="p-6">
          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : trackingSelections.length > 0 ? (
            <div className="space-y-3">
              {trackingSelections.map((selection) => (
                <TrackingRow
                  key={selection.id}
                  selection={selection}
                  batchMode={batchMode}
                  selected={selectedIds.has(selection.candidate_id)}
                  onToggle={() => toggleSelection(selection.candidate_id)}
                />
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
              <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
                <CheckCircle2 className="h-10 w-10 text-muted-foreground" />
              </div>
              <div className="space-y-2">
                <h2 className="text-xl font-semibold">No approved recruitments yet</h2>
                <p className="max-w-md text-muted-foreground">
                  Once both parties approve a candidate, their tracking will appear here.
                </p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Batch action bar */}
      {batchMode && selectedIds.size > 0 ? (
        <div className="sticky bottom-4 z-50 rounded-2xl border border-primary/20 bg-background/95 shadow-2xl backdrop-blur">
          <div className="flex flex-wrap items-center gap-3 p-4">
            <span className="text-sm font-semibold shrink-0">
              {selectedIds.size} selected
            </span>

            <Select value={batchStep} onValueChange={setBatchStep}>
              <SelectTrigger className="h-9 w-[140px]">
                <SelectValue placeholder="Step" />
              </SelectTrigger>
              <SelectContent>
                {STEPS.map((s) => (
                  <SelectItem key={s} value={s}>{s}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={batchStatus} onValueChange={setBatchStatus}>
              <SelectTrigger className="h-9 w-[140px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                {STATUSES.map((s) => (
                  <SelectItem key={s} value={s}>{s.replace("_", " ")}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            {batchStep === "CoC" ? (
              <Select value={batchCoC} onValueChange={setBatchCoC}>
                <SelectTrigger className="h-9 w-[140px]">
                  <SelectValue placeholder="CoC status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="not_online">Not Online</SelectItem>
                  <SelectItem value="online">Online</SelectItem>
                </SelectContent>
              </Select>
            ) : null}

            {batchStep === "Arrival City" ? (
              <Input
                className="h-9 w-[160px]"
                placeholder="City name"
                value={batchCity}
                onChange={(e) => setBatchCity(e.target.value)}
              />
            ) : null}

            <Button
              onClick={handleBatchApply}
              disabled={isBatchUpdating || !batchStep || !batchStatus}
              className="ml-auto"
            >
              {isBatchUpdating ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : null}
              Update {selectedIds.size > 0 ? `${selectedIds.size}` : ""}
            </Button>
          </div>
        </div>
      ) : null}
    </div>
  )
}

function TrackingRow({
  selection,
  batchMode,
  selected,
  onToggle,
}: {
  selection: { id: string; candidate_id: string; candidate?: { full_name?: string; status?: string } }
  batchMode: boolean
  selected: boolean
  onToggle: () => void
}) {
  return (
    <div
      className={cn(
        "flex flex-col gap-3 rounded-[1.4rem] border border-border/70 bg-card/95 p-4 transition-all md:flex-row md:items-center md:justify-between",
        selected && "border-primary/40 bg-primary/5 ring-1 ring-primary/20",
      )}
    >
      <div className="flex items-center gap-3 min-w-0">
        {batchMode ? (
          <div
            className="flex h-6 w-6 shrink-0 cursor-pointer items-center justify-center rounded-md border border-border bg-background"
            onClick={onToggle}
          >
            {selected ? (
              <CheckCircle2 className="h-4 w-4 text-primary" />
            ) : null}
          </div>
        ) : null}
        <div className="min-w-0 space-y-1">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold truncate">
              {selection.candidate?.full_name || "Candidate"}
            </h2>
            <Badge className="bg-emerald-500 text-white hover:bg-emerald-500 h-5 text-[10px]">
              Approved
            </Badge>
            <Badge variant="outline" className="capitalize h-5 text-[10px]">
              {selection.candidate?.status?.replaceAll("_", " ") || "tracking"}
            </Badge>
          </div>
        </div>
      </div>

      <div className="flex flex-wrap gap-2 shrink-0">
        {!batchMode ? (
          <>
            <Button variant="outline" size="sm" className="h-8" asChild>
              <Link href={`/selections/${selection.id}`}>
                <Sparkles className="mr-1.5 h-3.5 w-3.5" />
                Selection
              </Link>
            </Button>
            <Button size="sm" className="h-8" asChild>
              <Link href={`/candidates/${selection.candidate_id}/tracking`}>
                <Route className="mr-1.5 h-3.5 w-3.5" />
                Tracking
                <ArrowRight className="ml-1.5 h-3.5 w-3.5" />
              </Link>
            </Button>
          </>
        ) : null}
      </div>
    </div>
  )
}

function SummaryCard({
  label,
  value,
  tone,
  icon,
}: {
  label: string
  value: number
  tone: string
  icon: React.ReactNode
}) {
  return (
    <Card className={`overflow-hidden border-white/20 bg-gradient-to-br ${tone} shadow-soft`}>
      <CardContent className="flex items-center justify-between p-4">
        <div>
          <p className="text-xs text-muted-foreground font-medium">{label}</p>
          <p className="mt-0.5 text-2xl font-bold">{value}</p>
        </div>
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-white/70 text-current shadow-sm dark:bg-white/10">
          {icon}
        </div>
      </CardContent>
    </Card>
  )
}
