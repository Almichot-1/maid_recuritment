import { CandidateStatus, SelectionStatus } from "@/types"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

type Status = CandidateStatus | SelectionStatus | string
type Size = "sm" | "md" | "lg"

interface StatusBadgeProps {
  status: Status
  size?: Size
  className?: string
}

export function StatusBadge({ status, size = "md", className }: StatusBadgeProps) {
  const getStatusConfig = (status: Status) => {
    const statusLower = status.toLowerCase()

    const configs: Record<string, { label: string; className: string }> = {
      draft: {
        label: "Draft",
        className: "bg-gray-500 hover:bg-gray-600 text-white",
      },
      available: {
        label: "Available",
        className: "bg-green-500 hover:bg-green-600 text-white",
      },
      locked: {
        label: "Locked",
        className: "bg-amber-500 hover:bg-amber-600 text-white",
      },
      under_review: {
        label: "Under Review",
        className: "bg-blue-500 hover:bg-blue-600 text-white",
      },
      approved: {
        label: "Approved",
        className: "bg-emerald-500 hover:bg-emerald-600 text-white",
      },
      rejected: {
        label: "Rejected",
        className: "bg-red-500 hover:bg-red-600 text-white",
      },
      in_progress: {
        label: "In Process",
        className: "bg-purple-500 hover:bg-purple-600 text-white",
      },
      selected: {
        label: "Selected",
        className: "bg-purple-500 hover:bg-purple-600 text-white",
      },
      completed: {
        label: "Completed",
        className: "bg-green-500 hover:bg-green-600 text-white",
      },
      expired: {
        label: "Expired",
        className: "bg-gray-500 hover:bg-gray-600 text-white",
      },
      pending: {
        label: "Pending",
        className: "bg-yellow-500 hover:bg-yellow-600 text-white",
      },
    }

    return configs[statusLower] || { label: status, className: "bg-gray-500 text-white" }
  }

  const sizeClasses = {
    sm: "text-xs px-2 py-0.5",
    md: "text-sm px-3 py-1",
    lg: "text-base px-4 py-1.5",
  }

  const config = getStatusConfig(status)

  return (
    <Badge className={cn(config.className, sizeClasses[size], className)}>
      {config.label}
    </Badge>
  )
}
