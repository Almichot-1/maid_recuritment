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
        "relative overflow-hidden rounded-[32px] border border-slate-200/80 bg-[radial-gradient(circle_at_top_right,_rgba(251,191,36,0.16),_transparent_24%),radial-gradient(circle_at_bottom_left,_rgba(14,165,233,0.12),_transparent_22%),linear-gradient(135deg,rgba(255,255,255,0.98),rgba(248,250,252,0.96),rgba(226,232,240,0.94))] p-6 shadow-[0_24px_55px_-34px_rgba(15,23,42,0.18)] backdrop-blur-xl dark:border-slate-800/90 dark:bg-[radial-gradient(circle_at_top_right,_rgba(251,191,36,0.16),_transparent_24%),radial-gradient(circle_at_bottom_left,_rgba(14,165,233,0.12),_transparent_22%),linear-gradient(135deg,rgba(15,23,42,0.96),rgba(17,24,39,0.94),rgba(2,6,23,0.96))] dark:shadow-[0_24px_55px_-30px_rgba(2,6,23,0.68)]",
        "before:absolute before:inset-x-0 before:top-0 before:h-px before:bg-gradient-to-r before:from-transparent before:via-amber-500 before:to-transparent dark:before:via-amber-300",
        className
      )}
      {...props}
    >
      <div className="relative flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-2.5">
          <div className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-700 dark:text-amber-300">Admin Mode</div>
          <div className="space-y-1">
            <h1 className="text-2xl font-semibold tracking-tight text-slate-950 dark:text-slate-50 sm:text-3xl">{title}</h1>
            {description ? <p className="max-w-3xl text-sm text-slate-600 dark:text-slate-300 sm:text-base">{description}</p> : null}
          </div>
        </div>
        {action ? <div className="flex flex-wrap items-center gap-3 sm:justify-end">{action}</div> : null}
      </div>
    </div>
  )
}
