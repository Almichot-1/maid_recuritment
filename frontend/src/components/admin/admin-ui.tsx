"use client"

import * as React from "react"

import { Card, CardContent } from "@/components/ui/card"
import { cn } from "@/lib/utils"

export function AdminSurface({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <Card className={cn("admin-surface", className)} {...props} />
}

export function AdminToolbar({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("admin-toolbar", className)} {...props} />
}

export function AdminStatCard({
  label,
  value,
  detail,
  icon: Icon,
  className,
}: {
  label: string
  value: React.ReactNode
  detail?: string
  icon?: React.ElementType
  className?: string
}) {
  return (
    <AdminSurface className={cn("overflow-hidden", className)}>
      <CardContent className="flex items-start justify-between gap-4 p-6">
        <div className="space-y-2">
          <p className="text-sm font-medium text-slate-500 dark:text-slate-400">{label}</p>
          <p className="text-3xl font-semibold tracking-tight text-slate-950 dark:text-slate-50">{value}</p>
          {detail ? <p className="text-sm text-slate-500 dark:text-slate-400">{detail}</p> : null}
        </div>
        {Icon ? (
          <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-300">
            <Icon className="h-5 w-5" />
          </div>
        ) : null}
      </CardContent>
    </AdminSurface>
  )
}

export function AdminEmptyState({
  title,
  description,
  action,
  className,
}: {
  title: string
  description: string
  action?: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("admin-empty-state", className)}>
      <p className="text-base font-semibold text-slate-900 dark:text-slate-100">{title}</p>
      <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">{description}</p>
      {action ? <div className="mt-4 flex justify-center">{action}</div> : null}
    </div>
  )
}

export function AdminInfoTile({
  label,
  value,
  className,
}: {
  label: string
  value: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("admin-info-tile", className)}>
      <p className="admin-info-label">{label}</p>
      <div className="admin-info-value">{value}</div>
    </div>
  )
}

export function AdminMetricRow({
  label,
  value,
  className,
}: {
  label: string
  value: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("flex items-center justify-between rounded-2xl border border-slate-200/75 bg-slate-50/70 px-4 py-3 dark:border-slate-800 dark:bg-slate-900/75", className)}>
      <span className="text-sm text-slate-500 dark:text-slate-400">{label}</span>
      <span className="text-lg font-semibold text-slate-950 dark:text-slate-100">{value}</span>
    </div>
  )
}
