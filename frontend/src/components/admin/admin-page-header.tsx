import * as React from "react"
import { cn } from "@/lib/utils"

interface AdminPageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string
  description?: string
  action?: React.ReactNode
}

export function AdminPageHeader({ title, description, action, className, ...props }: AdminPageHeaderProps) {
  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-[28px] border border-slate-800/80 bg-[radial-gradient(circle_at_top_left,_rgba(245,158,11,0.14),_transparent_28%),linear-gradient(145deg,rgba(15,23,42,0.98),rgba(15,23,42,0.9)_55%,rgba(2,6,23,0.98))] p-6 shadow-[0_24px_80px_-48px_rgba(15,23,42,0.85)] backdrop-blur",
        "before:absolute before:inset-x-0 before:top-0 before:h-px before:bg-gradient-to-r before:from-transparent before:via-amber-300/80 before:to-transparent",
        className
      )}
      {...props}
    >
      <div className="relative flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-2">
          <div className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-300">Admin Mode</div>
          <div className="space-y-1">
            <h1 className="text-2xl font-semibold tracking-tight text-white sm:text-3xl">{title}</h1>
            {description ? <p className="max-w-3xl text-sm text-slate-300 sm:text-base">{description}</p> : null}
          </div>
        </div>
        {action ? <div className="flex items-center gap-3">{action}</div> : null}
      </div>
    </div>
  )
}
