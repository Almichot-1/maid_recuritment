import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { titleize } from "@/lib/admin-utils"

const toneByStatus: Record<string, string> = {
  pending: "border-[color:var(--color-warning)] text-[color:var(--color-warning)]",
  pending_approval: "border-[color:var(--color-warning)] text-[color:var(--color-warning)]",
  active: "border-[color:var(--color-success)] text-[color:var(--color-success)]",
  approved: "border-[color:var(--color-success)] text-[color:var(--color-success)]",
  available: "border-[color:var(--color-success)] text-[color:var(--color-success)]",
  locked: "border-primary text-primary",
  suspended: "border-[color:var(--color-danger)] text-[color:var(--color-danger)]",
  rejected: "border-[color:var(--color-danger)] text-[color:var(--color-danger)]",
  expired: "border-border text-muted-foreground",
  under_review: "border-[color:var(--color-info)] text-[color:var(--color-info)]",
  in_progress: "border-[color:var(--color-info)] text-[color:var(--color-info)]",
  completed: "border-foreground text-foreground",
}

export function AdminStatusBadge({ status, className }: { status: string; className?: string }) {
  const normalized = status.trim().toLowerCase()
  const tone = toneByStatus[normalized] ?? "border-border bg-muted text-muted-foreground"

  return (
    <Badge variant="outline" className={cn("px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.08em]", tone, className)}>
      {titleize(normalized)}
    </Badge>
  )
}
