import { Badge } from "@/components/ui/badge"
import { titleize } from "@/lib/admin-utils"
import { cn } from "@/lib/utils"

const toneByStatus: Record<string, string> = {
  pending: "border-amber-200 bg-amber-100 text-amber-800 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-200",
  pending_approval: "border-amber-200 bg-amber-100 text-amber-800 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-200",
  active: "border-emerald-200 bg-emerald-100 text-emerald-800 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-200",
  approved: "border-emerald-200 bg-emerald-100 text-emerald-800 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-200",
  available: "border-emerald-200 bg-emerald-100 text-emerald-800 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-200",
  locked: "border-orange-200 bg-orange-100 text-orange-800 dark:border-orange-400/20 dark:bg-orange-400/10 dark:text-orange-200",
  suspended: "border-rose-200 bg-rose-100 text-rose-800 dark:border-rose-400/20 dark:bg-rose-400/10 dark:text-rose-200",
  rejected: "border-rose-200 bg-rose-100 text-rose-800 dark:border-rose-400/20 dark:bg-rose-400/10 dark:text-rose-200",
  expired: "border-slate-200 bg-slate-100 text-slate-700 dark:border-slate-600/40 dark:bg-slate-400/10 dark:text-slate-200",
  under_review: "border-sky-200 bg-sky-100 text-sky-800 dark:border-sky-400/20 dark:bg-sky-400/10 dark:text-sky-200",
  in_progress: "border-indigo-200 bg-indigo-100 text-indigo-800 dark:border-indigo-400/20 dark:bg-indigo-400/10 dark:text-indigo-200",
  completed: "border-violet-200 bg-violet-100 text-violet-800 dark:border-violet-400/20 dark:bg-violet-400/10 dark:text-violet-200",
}

export function AdminStatusBadge({ status, className }: { status: string; className?: string }) {
  const normalized = status.trim().toLowerCase()
  const tone = toneByStatus[normalized] ?? "border-slate-200 bg-slate-100 text-slate-700 dark:border-slate-600/40 dark:bg-slate-400/10 dark:text-slate-200"

  return (
    <Badge variant="outline" className={cn("rounded-full px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide", tone, className)}>
      {titleize(normalized)}
    </Badge>
  )
}
