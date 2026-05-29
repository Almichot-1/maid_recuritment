import * as React from "react"
import { cn } from "@/lib/utils"
import { useI18n } from "@/lib/i18n"

interface AdminPageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string
  description?: string
  action?: React.ReactNode
}

export function AdminPageHeader({ title, description, action, className, ...props }: AdminPageHeaderProps) {
  const { t } = useI18n()

  return (
    <div
      className={cn(
        "border border-border bg-card p-6",
        className
      )}
      {...props}
    >
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-2">
          <div className="section-kicker">{t("admin.mode")}</div>
          <div className="space-y-1">
            <h1 className="font-display text-4xl text-foreground">{title}</h1>
            {description ? <p className="max-w-3xl text-sm text-muted-foreground sm:text-base">{description}</p> : null}
          </div>
        </div>
        {action ? <div className="flex items-center gap-3">{action}</div> : null}
      </div>
    </div>
  )
}
