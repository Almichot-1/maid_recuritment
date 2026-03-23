import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { titleize } from "@/lib/admin-utils"

const toneByStatus: Record<string, string> = {
  pending: "border-amber-200 bg-amber-100 text-amber-800",
  pending_approval: "border-amber-200 bg-amber-100 text-amber-800",
  active: "border-emerald-200 bg-emerald-100 text-emerald-800",
  approved: "border-emerald-200 bg-emerald-100 text-emerald-800",
  available: "border-emerald-200 bg-emerald-100 text-emerald-800",
  locked: "border-orange-200 bg-orange-100 text-orange-800",
  suspended: "border-rose-200 bg-rose-100 text-rose-800",
  rejected: "border-rose-200 bg-rose-100 text-rose-800",
  expired: "border-slate-200 bg-slate-100 text-slate-700",
  under_review: "border-sky-200 bg-sky-100 text-sky-800",
  in_progress: "border-indigo-200 bg-indigo-100 text-indigo-800",
  completed: "border-violet-200 bg-violet-100 text-violet-800",
}

export function AdminStatusBadge({ status, className }: { status: string; className?: string }) {
  const normalized = status.trim().toLowerCase()
  const tone = toneByStatus[normalized] ?? "border-slate-200 bg-slate-100 text-slate-700"

  return (
    <Badge variant="outline" className={cn("rounded-full px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide", tone, className)}>
      {titleize(normalized)}
    </Badge>
  )
}
