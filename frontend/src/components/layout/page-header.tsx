import * as React from "react"
import { useI18n } from "@/lib/i18n"
import { cn } from "@/lib/utils"

interface PageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  heading: string
  text?: string
  action?: React.ReactNode
}

export function PageHeader({ heading, text, action, className, ...props }: PageHeaderProps) {
  const { t } = useI18n()

  return (
    <div className={cn("flex flex-col justify-between gap-4 border-b border-border pb-6 sm:flex-row sm:items-end", className)} {...props}>
      <div className="space-y-3">
        <div className="section-kicker">{t("pageHeader.kicker")}</div>
        <div className="space-y-2">
          <h1 className="font-display text-4xl leading-none text-foreground">{heading}</h1>
          {text ? <p className="max-w-3xl text-sm text-muted-foreground sm:text-base">{text}</p> : null}
        </div>
      </div>
      {action && <div className="flex items-center">{action}</div>}
    </div>
  )
}
