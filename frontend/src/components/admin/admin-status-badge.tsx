import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { titleize } from "@/lib/admin-utils"

const toneByStatus: Record<string, string> = {
  pending: "border-amber-400/30 bg-amber-400/10 text-amber-200",
  pending_approval: "border-amber-400/30 bg-amber-400/10 text-amber-200",
  active: "border-emerald-400/30 bg-emerald-400/10 text-emerald-200",
  approved: "border-emerald-400/30 bg-emerald-400/10 text-emerald-200",
  available: "border-emerald-400/30 bg-emerald-400/10 text-emerald-200",
  locked: "border-orange-400/30 bg-orange-400/10 text-orange-200",
  suspended: "border-rose-400/30 bg-rose-400/10 text-rose-200",
  rejected: "border-rose-400/30 bg-rose-400/10 text-rose-200",
  expired: "border-slate-500/40 bg-slate-500/10 text-slate-200",
  under_review: "border-sky-400/30 bg-sky-400/10 text-sky-200",
  in_progress: "border-indigo-400/30 bg-indigo-400/10 text-indigo-200",
  completed: "border-violet-400/30 bg-violet-400/10 text-violet-200",
}

export function AdminStatusBadge({ status, className }: { status: string; className?: string }) {
  const normalized = status.trim().toLowerCase()
  const tone = toneByStatus[normalized] ?? "border-slate-500/40 bg-slate-500/10 text-slate-200"

  return (
    <Badge variant="outline" className={cn("rounded-full px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide", tone, className)}>
      {titleize(normalized)}
    </Badge>
  )
}
