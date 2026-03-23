import * as React from "react"
import { cn } from "@/lib/utils"

interface PageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  heading: string
  text?: string
  action?: React.ReactNode
}

export function PageHeader({ heading, text, action, className, ...props }: PageHeaderProps) {
  return (
    <div className={cn("flex flex-col justify-between gap-4 pb-6 sm:flex-row sm:items-center", className)} {...props}>
      <div className="space-y-2">
        <span className="inline-flex w-fit items-center rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.24em] text-primary">
          Workspace
        </span>
        <div className="space-y-1.5">
          <h1 className="text-3xl font-bold tracking-tight text-foreground">{heading}</h1>
          {text && <p className="max-w-2xl text-muted-foreground">{text}</p>}
        </div>
      </div>
      {action && <div className="flex items-center">{action}</div>}
    </div>
  )
}
